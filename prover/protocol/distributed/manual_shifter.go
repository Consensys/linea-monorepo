package distributed

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// AssignManuallyShiftedColumn implements [wizard.ProverAction] and
// is responsible for assigning the manually shifted columns with the right values at the right round.
type AssignManualShifts struct {
	ManualShifts []*dedicated.ManuallyShifted
}

// Run implements [wizard.ProverAction] interface
func (a *AssignManualShifts) Run(runtime *wizard.ProverRuntime) {
	for i := range a.ManualShifts {
		a.ManualShifts[i].GetColAssignment(runtime)
	}
}

/*
compileManualShifter creates a small compiler job iterating on every queries that are not ignored.
- for lookup, permutation, projection: whenever we meet a shifted column, we mark the query as ignored
- we create a replacement column for every problematic columns involved in the query, i.e., a shifted column
- we create a new equivalent query, replacing the shifted columns by the manually shifted ones
- we schedule the assignment of the manually shifted columns for the right round
*/
func compileManualShifter(comp *wizard.CompiledIOP) {
	logrus.Info("compiling manual shifter")
	defer logrus.Info("finished compiling manual shifter")

	for roundId := 0; roundId < comp.NumRounds(); roundId++ {
		for _, qName := range comp.QueriesNoParams.AllKeysAt(roundId) {
			if comp.QueriesNoParams.IsIgnored(qName) {
				continue
			}
			q := comp.QueriesNoParams.Data(qName)

			switch q := q.(type) {
			case query.Inclusion:
				// @arijit: avoid calling len(q.GetShiftedRelatedColumns()) twice
				if len(q.GetShiftedRelatedColumns()) > 0 {
					replayQueryWithManualShiftInclusion(comp, qName, len(q.GetShiftedRelatedColumns()))
					continue
				}
				continue
			case query.Permutation:
				if len(q.GetShiftedRelatedColumns()) > 0 {
					replayQueryWithManualShiftPermutation(comp, qName, len(q.GetShiftedRelatedColumns()))
					continue
				}
				continue
			case query.Projection:
				if len(q.GetShiftedRelatedColumns()) > 0 {
					replayQueryWithManualShiftProjection(comp, qName, len(q.GetShiftedRelatedColumns()))
					continue
				}
				continue
			default:
				continue
			}
		}
	}
}

// replayQueryWithManualShiftInclusion creates a new inclusion query with the same semantics as the original one, but replacing the shifted columns by manually shifted ones. It also marks the original query as ignored and schedule the assignment of the manually shifted columns at the right round.
func replayQueryWithManualShiftInclusion(comp *wizard.CompiledIOP, qName ifaces.QueryID, numShift int) {
	// sanity check: make sure it is an inclusion query
	q, ok := comp.QueriesNoParams.Data(qName).(query.Inclusion)
	if !ok {
		panic("replayQueryWithManualShiftInclusion called on a non-inclusion query")
	}

	// allocate the shifted cols
	shiftedCols := make([]*dedicated.ManuallyShifted, 0, numShift)
	newQ := query.Inclusion{
		ID:             ifaces.QueryIDf("%v_WITH_MANUALLY_SHIFTED_COL", q.ID),
		Included:       make([]ifaces.Column, len(q.Included)),
		Including:      make([][]ifaces.Column, len(q.Including)),
		IncludedFilter: q.IncludedFilter,
		IncludingFilter: func() []ifaces.Column {
			if q.IncludingFilter == nil {
				return nil
			}
			res := make([]ifaces.Column, len(q.IncludingFilter))
			return res
		}(),
	}

	// replace the shifted columns with the manually shifted ones
	if q.IncludedFilter != nil && q.IncludedFilter.IsComposite() {
		switch col := q.IncludedFilter.(type) {
		case column.Shifted:
			offset := col.Offset
			shiftedCol := dedicated.ManuallyShift(comp, q.IncludedFilter, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", q.IncludedFilter.GetColID()))
			shiftedCols = append(shiftedCols, shiftedCol)
			newQ.IncludedFilter = shiftedCol.Natural
		default:
			newQ.IncludedFilter = q.IncludedFilter
		}
	} else {
		newQ.IncludedFilter = q.IncludedFilter
	}

	for i, included := range q.Included {
		if included.IsComposite() {
			switch col := included.(type) {
			case column.Shifted:
				offset := col.Offset
				shiftedCol := dedicated.ManuallyShift(comp, included, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", included.GetColID()))
				shiftedCols = append(shiftedCols, shiftedCol)
				newQ.Included[i] = shiftedCol.Natural
			default:
				newQ.Included[i] = included
			}
		} else {
			newQ.Included[i] = included
		}
	}

	for frag := range q.Including {

		if q.IncludingFilter != nil && q.IncludingFilter[frag] != nil {
			if q.IncludingFilter[frag].IsComposite() {
				switch col := q.IncludingFilter[frag].(type) {
				case column.Shifted:
					offset := col.Offset
					shiftedCol := dedicated.ManuallyShift(comp, q.IncludingFilter[frag], offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", q.IncludingFilter[frag].GetColID()))
					shiftedCols = append(shiftedCols, shiftedCol)
					newQ.IncludingFilter[frag] = shiftedCol.Natural
				default:
					newQ.IncludingFilter[frag] = q.IncludingFilter[frag]
				}
			} else {
				newQ.IncludingFilter[frag] = q.IncludingFilter[frag]
			}
		}

		newQ.Including[frag] = make([]ifaces.Column, len(q.Including[frag]))
		for i, including := range q.Including[frag] {
			if including.IsComposite() {
				switch col := including.(type) {
				case column.Shifted:
					offset := col.Offset
					shiftedCol := dedicated.ManuallyShift(comp, including, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", including.GetColID()))
					shiftedCols = append(shiftedCols, shiftedCol)
					newQ.Including[frag][i] = shiftedCol.Natural
				default:
					newQ.Including[frag][i] = including
				}
			} else {
				newQ.Including[frag][i] = including
			}
		}
	}
	// insert the new query
	comp.QueriesNoParams.AddToRound(q.Included[0].Round(),
		newQ.ID, newQ)
	// register the prover action to assign the manually shifted columns at the right round
	comp.RegisterProverAction(q.Included[0].Round(),
		&AssignManualShifts{ManualShifts: shiftedCols})
	// ignore the current query
	comp.QueriesNoParams.MarkAsIgnored(qName)
}

// replayQueryWithManualShiftPermutation creates a new permutation query with the same semantics as the original one, but replacing the shifted columns by manually shifted ones. It also marks the original query as ignored and schedule the assignment of the manually shifted columns at the right round.
func replayQueryWithManualShiftPermutation(comp *wizard.CompiledIOP, qName ifaces.QueryID, numShift int) {
	// sanity check: make sure it is a permutation query
	q, ok := comp.QueriesNoParams.Data(qName).(query.Permutation)
	if !ok {
		panic("replayQueryWithManualShiftPermutation called on a non-permutation query")
	}

	// allocate the shifted cols
	shiftedCols := make([]*dedicated.ManuallyShifted, 0, numShift)
	newQ := query.Permutation{
		ID: ifaces.QueryIDf("%v_WITH_MANUALLY_SHIFTED_COL", q.ID),
		A:  make([][]ifaces.Column, len(q.A)),
		B:  make([][]ifaces.Column, len(q.B)),
	}

	// replace the shifted columns with the manually shifted ones
	for frag := range q.A {
		newQ.A[frag] = make([]ifaces.Column, len(q.A[frag]))
		for i, col := range q.A[frag] {
			if col.IsComposite() {
				switch col := col.(type) {
				case column.Shifted:
					offset := col.Offset
					shiftedCol := dedicated.ManuallyShift(comp, col, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", col.GetColID()))
					shiftedCols = append(shiftedCols, shiftedCol)
					newQ.A[frag][i] = shiftedCol.Natural
				default:
					newQ.A[frag][i] = col
				}
			} else {
				newQ.A[frag][i] = col
			}
		}
	}

	for frag := range q.B {
		newQ.B[frag] = make([]ifaces.Column, len(q.B[frag]))
		for i, col := range q.B[frag] {
			if col.IsComposite() {
				switch col := col.(type) {
				case column.Shifted:
					offset := col.Offset
					shiftedCol := dedicated.ManuallyShift(comp, col, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", col.GetColID()))
					shiftedCols = append(shiftedCols, shiftedCol)
					newQ.B[frag][i] = shiftedCol.Natural
				default:
					newQ.B[frag][i] = col
				}
			} else {
				newQ.B[frag][i] = col
			}
		}
	}
	// insert the new query
	comp.QueriesNoParams.AddToRound(q.A[0][0].Round(),
		newQ.ID, newQ)
	// register the prover action to assign the manually shifted columns at the right round
	comp.RegisterProverAction(q.A[0][0].Round(),
		&AssignManualShifts{ManualShifts: shiftedCols})
	// ignore the current query
	comp.QueriesNoParams.MarkAsIgnored(q.ID)
}

// replayQueryWithManualShiftProjection creates a new projection query with the same semantics as the original one, but replacing the shifted columns by manually shifted ones. It also marks the original query as ignored and schedule the assignment of the manually shifted columns at the right round.
func replayQueryWithManualShiftProjection(comp *wizard.CompiledIOP, qName ifaces.QueryID, numShift int) {
	// sanity check: make sure it is a projection query
	q, ok := comp.QueriesNoParams.Data(qName).(query.Projection)
	if !ok {
		panic("replayQueryWithManualShiftProjection called on a non-projection query")
	}

	// allocate the shifted cols
	shiftedCols := make([]*dedicated.ManuallyShifted, 0, numShift)
	newQ := query.Projection{
		Round: q.Round,
		ID:    ifaces.QueryIDf("%v_WITH_MANUALLY_SHIFTED_COL", q.ID),
		Inp:   query.ProjectionMultiAryInput{},
	}
	// replace the shifted columns with the manually shifted ones
	if q.Inp.FiltersA != nil {
		newQ.Inp.FiltersA = make([]ifaces.Column, len(q.Inp.FiltersA))
		for i, f := range q.Inp.FiltersA {
			if f.IsComposite() {
				switch col := f.(type) {
				case column.Shifted:
					offset := col.Offset
					shiftedCol := dedicated.ManuallyShift(comp, f, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", f.GetColID()))
					shiftedCols = append(shiftedCols, shiftedCol)
					newQ.Inp.FiltersA[i] = shiftedCol.Natural
				default:
					newQ.Inp.FiltersA[i] = f
				}
			} else {
				newQ.Inp.FiltersA[i] = f
			}
		}
	}

	if q.Inp.FiltersB != nil {
		newQ.Inp.FiltersB = make([]ifaces.Column, len(q.Inp.FiltersB))
		for i, f := range q.Inp.FiltersB {
			if f.IsComposite() {
				switch col := f.(type) {
				case column.Shifted:
					offset := col.Offset
					shiftedCol := dedicated.ManuallyShift(comp, f, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", f.GetColID()))
					shiftedCols = append(shiftedCols, shiftedCol)
					newQ.Inp.FiltersB[i] = shiftedCol.Natural
				default:
					newQ.Inp.FiltersB[i] = f
				}
			} else {
				newQ.Inp.FiltersB[i] = f
			}
		}
	}

	newQ.Inp.ColumnsA = make([][]ifaces.Column, len(q.Inp.ColumnsA))
	for i := range q.Inp.ColumnsA {
		newQ.Inp.ColumnsA[i] = make([]ifaces.Column, len(q.Inp.ColumnsA[i]))
		for j, col := range q.Inp.ColumnsA[i] {
			if col.IsComposite() {
				switch col := col.(type) {
				case column.Shifted:
					offset := col.Offset
					shiftedCol := dedicated.ManuallyShift(comp, col, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", col.GetColID()))
					shiftedCols = append(shiftedCols, shiftedCol)
					newQ.Inp.ColumnsA[i][j] = shiftedCol.Natural
				default:
					newQ.Inp.ColumnsA[i][j] = col
				}
			} else {
				newQ.Inp.ColumnsA[i][j] = col
			}
		}
	}

	newQ.Inp.ColumnsB = make([][]ifaces.Column, len(q.Inp.ColumnsB))
	for i := range q.Inp.ColumnsB {
		newQ.Inp.ColumnsB[i] = make([]ifaces.Column, len(q.Inp.ColumnsB[i]))
		for j, col := range q.Inp.ColumnsB[i] {
			if col.IsComposite() {
				switch col := col.(type) {
				case column.Shifted:
					offset := col.Offset
					shiftedCol := dedicated.ManuallyShift(comp, col, offset, fmt.Sprintf("MANUALLY_SHIFTED_%v", col.GetColID()))
					shiftedCols = append(shiftedCols, shiftedCol)
					newQ.Inp.ColumnsB[i][j] = shiftedCol.Natural
				default:
					newQ.Inp.ColumnsB[i][j] = col
				}
			} else {
				newQ.Inp.ColumnsB[i][j] = col
			}
		}
	}
	// insert the new query
	comp.QueriesNoParams.AddToRound(q.Round,
		newQ.ID, newQ)
	// register the prover action to assign the manually shifted columns at the right round
	comp.RegisterProverAction(q.Round,
		&AssignManualShifts{ManualShifts: shiftedCols})
	// ignore the current query
	comp.QueriesNoParams.MarkAsIgnored(qName)
}

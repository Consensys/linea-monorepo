package execution_data_collector

import (
	"fmt"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// PadderPacker is used to format the data so that we can compute a Poseidon-hash
// of the limbs in the execution data collector.
type PadderPacker struct {
	// the input limbs to be packed and padded
	InputLimbs [common.NbLimbU128]ifaces.Column
	// The number of bytes in the limbs.
	InputNoBytes ifaces.Column
	// the is active part of the input
	InputIsActive ifaces.Column
	// Step 1 columns
	// OneColumn is a column that contains the limbs of the input consecutively, so that we can apply the periodic filters on it and get the sum of bytes in each segment of 8 limbs.
	OneColumn         ifaces.Column
	OneColumnFilter   ifaces.Column
	OneColumnBytes    ifaces.Column
	OneColumnBytesSum ifaces.Column
	// Helper filters that do not depend on the data. These are periodic filters with period 8.
	PeriodicFilter [8]ifaces.Column
	// for the 0 index period filter, we need a trimmed version of it which only lights up on OneColumnFilter active rows
	TrimmedPeriodicFilter ifaces.Column
	// Step 2 columns. The inactive gaps of the OneColumn are removed here.
	OneColumnWithoutGaps ifaces.Column
	FilterWithoutGaps    ifaces.Column
	// Helper columns that allow to pad with zeroes up to a multiple of 8.
	CounterColumn       ifaces.Column
	CounterColumnFilled ifaces.Column
	CounterColumnAdded  ifaces.Column
	// selectors for the prover to compute the values in the SelectorCounterColumnFilled column
	SelectorCounterColumnFilled        ifaces.Column
	ComputeSelectorCounterColumnFilled wizard.ProverAction
	// Step 3 columns. The final output columns after padding and packing
	// OuterColumns are the output columns that will be used to compute the Poseidon hash.
	OuterColumns [8]ifaces.Column
	// the isActive part of the output
	OuterIsActive ifaces.Column
}

// NewPoseidonPadderPacker returns a new GenericPadderPacker with initialized columns that are not constrained.
func NewPadderPacker(comp *wizard.CompiledIOP, inputLimbs [common.NbLimbU128]ifaces.Column, inputNoBytes, inputIsActive ifaces.Column, name string) PadderPacker {
	var (
		res     PadderPacker
		newSize int
	)
	res.InputLimbs = inputLimbs
	res.InputNoBytes = inputNoBytes
	res.InputIsActive = inputIsActive

	oldSize := res.InputLimbs[0].Size()
	newSize = oldSize * common.NbLimbU128
	res.OneColumn = util.CreateCol(name, "ONE_COLUMN", newSize, comp)
	res.OneColumnFilter = util.CreateCol(name, "ONE_COLUMN_FILTER", newSize, comp)
	res.OneColumnBytes = util.CreateCol(name, "ONE_COLUMN_BYTES", newSize, comp)
	res.OneColumnBytesSum = util.CreateCol(name, "ONE_COLUMN_BYTE_SUM", newSize, comp)
	res.OneColumnWithoutGaps = util.CreateCol(name, "ONE_COLUMN_WITHOUT_GAPS", newSize, comp)
	res.FilterWithoutGaps = util.CreateCol(name, "FILTER_WITHOUT_GAPS", newSize, comp)
	res.CounterColumn = util.CreateCol(name, "COUNTER_COLUMN", newSize, comp)
	res.CounterColumnFilled = util.CreateCol(name, "COUNTER_COLUMN_FILLED", newSize, comp)
	res.CounterColumnAdded = util.CreateCol(name, "COUNTER_COLUMN_ADDED", newSize, comp)

	for i := range res.PeriodicFilter {
		res.PeriodicFilter[i] = util.CreateCol(name, fmt.Sprintf("PERIODIC_FILTER_%d", i), newSize, comp)
	}
	res.TrimmedPeriodicFilter = util.CreateCol(name, fmt.Sprintf("TRIMMED_PERIODIC_FILTER"), newSize, comp)

	for i := range res.OuterColumns {
		res.OuterColumns[i] = util.CreateCol(name, fmt.Sprintf("INTER_COLUMN_%d", i), oldSize, comp)
	}
	res.OuterIsActive = util.CreateCol(name, "OUTPUT_IS_ACTIVE", oldSize, comp)
	return res
}

// DefineHasher specifies the constraints of the GenericPadderPacker with respect to the ExtractedData fetched from the arithmetization
func DefinePadderPacker(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	DefinePeriodicFilters(comp, ppp, name)
	// Step 1
	DefineOneColumnNBytesConstraints(comp, ppp, name)
	DefineTrimmedPeriodicFilter(comp, ppp, name)
	DefineOneColumnFilter(comp, ppp, name)
	DefineStepOneProjectionQueries(comp, ppp, name)
	// Step 2
	DefineFilterWithoutGaps(comp, ppp, name)
	DefineStepTwoProjectionQueries(comp, ppp, name)
	// Step 3
	DefinePadderPackerSelectorConstraints(comp, ppp, name)
	DefineCounterPadding(comp, ppp, name)
	DefineCounterColumnAdded(comp, ppp, name)
	DefineOuterActiveFilter(comp, ppp, name)
	DefineStepThreeProjectionQueries(comp, ppp, name)
}

func DefinePeriodicFilters(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	for j := range ppp.PeriodicFilter {
		// constraints for the filter at index j
		util.MustBeBinary(comp, ppp.PeriodicFilter[j])

		comp.InsertGlobal(0,
			ifaces.QueryIDf("%s_PERIODIC_PATTERN_SAME_VALUE_AFTER_EACH_SEGMENT_%d", name, j),
			sym.Mul(
				ppp.PeriodicFilter[j],
				sym.Sub(
					ppp.PeriodicFilter[j],
					column.Shift(ppp.PeriodicFilter[j], -8),
				),
			),
		)

		comp.InsertLocal(0,
			ifaces.QueryIDf("%s_PERIODIC_PATTERN_INIT_ONE_%d", name, j),
			sym.Sub(
				column.Shift(ppp.PeriodicFilter[j], j),
				1),
		)
		for index := 0; index < 8; index++ {
			if index != j {
				// for all these positions, we must have zeroes
				comp.InsertLocal(0,
					ifaces.QueryIDf("%s_PERIODIC_PATTERN_INIT_ZERO_%d_%d", name, j, index),
					ifaces.ColumnAsVariable(column.Shift(ppp.PeriodicFilter[j], index)),
				)
			}
		}
	}
}

func DefineCounterPadding(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	// Constrain CounterColumn
	for i := 0; i < 8; i++ {
		comp.InsertGlobal(0,
			ifaces.QueryIDf("%s_COUNTER_VALUE_%d", name, i),
			sym.Mul(
				ppp.FilterWithoutGaps,                                   // on the active part of the column without gaps
				ppp.PeriodicFilter[i],                                   // at position i in each segment of 8
				sym.Sub(ppp.CounterColumn, field.NewElement(uint64(i))), // CounterColumn must be equal to i
			),
		)
	}

	// Constraint: After last active row, the counter filling continues until reaching 0 (multiple of 8)
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_COUNTER_PADDING_BORDER", name),
		sym.Mul(
			sym.Sub(field.NewElement(1), column.Shift(ppp.FilterWithoutGaps, 1)),
			ppp.FilterWithoutGaps, // at the border of FilterWithoutGaps, the active part of the column without gaps
			ppp.CounterColumn,     // must be > 0, if the CounterColumn is 0, it means we are in the lucky case where we are already at a multiple of 8 and we do not need to fill in more
			sym.Sub(
				// CounterColumnFilled increases by 1 compared to CounterColumn
				// at the border
				column.Shift(ppp.CounterColumnFilled, 1),
				ppp.CounterColumn,
				1,
			),
		),
	)

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_COUNTER_PADDING_FILLING_CORRECTNESS", name),
		sym.Mul(
			ppp.CounterColumnFilled,           // CounterColumnFilled > 0
			sym.Sub(1, ppp.PeriodicFilter[7]), // we are not at the end of the size 8 block
			sym.Sub(
				// CounterColumnFilled increases by 1 compared to CounterColumnFilled
				column.Shift(ppp.CounterColumnFilled, 1),
				ppp.CounterColumnFilled,
				1,
			),
		),
	)

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_COUNTER_PADDING_FILLING_ZEROIZATION", name),
		sym.Mul(
			ppp.PeriodicFilter[0],                  // we are at the end of the size 8 block
			ifaces.Column(ppp.CounterColumnFilled), // force CounterColumnFilled to be zero
		),
	)

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_COUNTER_PADDING_FILLING_NEVER_REACTIVATES", name),
		sym.Mul(
			ppp.PeriodicFilter[0],                  // we are at the end of the size 8 block
			ifaces.Column(ppp.CounterColumnFilled), // force CounterColumnFilled to be zero
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_COUNTER_PADDING_FILLING_NEVER_INCREASES_FROM_O_TO_1_AGAIN", name),
		sym.Mul(
			sym.Sub(1, ppp.FilterWithoutGaps),        // on the inactive part of the column without gaps
			ppp.SelectorCounterColumnFilled,          // 1 when CounterColumnFilled is 0
			column.Shift(ppp.CounterColumnFilled, 1), // require that when CounterColumnFilled is 0, the next one is also 0
		),
	)
}

func DefinePadderPackerSelectorConstraints(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	// We first compute the prover actions
	ppp.SelectorCounterColumnFilled, ppp.ComputeSelectorCounterColumnFilled = dedicated.IsZero(
		comp,
		ifaces.ColumnAsVariable(ppp.CounterColumnFilled),
	).GetColumnAndProverAction()
}

func DefineOneColumnNBytesConstraints(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_ONE_COLUMN_NBYTES_SUM", name),
		sym.Mul(
			ppp.PeriodicFilter[0], // when the Periodic filter is at the beginning of a segment of 8
			sym.Sub(
				ppp.OneColumnBytesSum, // OneColumnBytesSum must be equal to the sum of OneColumnBytes over the current and next 7 rows
				ppp.OneColumnBytes,
				column.Shift(ppp.OneColumnBytes, 1),
				column.Shift(ppp.OneColumnBytes, 2),
				column.Shift(ppp.OneColumnBytes, 3),
				column.Shift(ppp.OneColumnBytes, 4),
				column.Shift(ppp.OneColumnBytes, 5),
				column.Shift(ppp.OneColumnBytes, 6),
				column.Shift(ppp.OneColumnBytes, 7),
			),
		),
	)

	for j := 0; j < common.NbLimbU128-1; j++ {
		// for all the periodic filters except the last one
		// the last one being excluded ensures that the blocks get reset
		// the number of bytes loaded must have the pattern 2, 2, 2, 0, 0... in a segment of 8
		// once it reaches 0, it must stay 0 until the end of the segment of 8
		// j stops before common.NbLimbU128-1, because we allow for a reset at the start of the next segment
		comp.InsertGlobal(0,
			ifaces.QueryIDf("%s_NBYTES_ZERO_PROPAGATION_%d", name, j),
			sym.Mul(
				ppp.PeriodicFilter[j],
				sym.Mul( // (2-OneColumnBytes[i])*OneColumnBytes(i+1)=0
					sym.Sub(
						field.NewElement(2),
						ppp.OneColumnBytes,
					),
					column.Shift(ppp.OneColumnBytes, 1),
				),
			),
		)
	}
}

func DefineTrimmedPeriodicFilter(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_TRIMMED_FIRST_PERIODIC_FILTER", name),
		sym.Sub(
			ppp.TrimmedPeriodicFilter, // when the Periodic filter is at the beginning of a segment of 8
			sym.Mul(
				ppp.OneColumnFilter,
				ppp.PeriodicFilter[0],
			),
		),
	)
}

func DefineOneColumnFilter(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	// the OneColumnFilter is also implicitly constrained by the step 1 projection queries
	util.MustBeBinary(comp, ppp.OneColumnFilter)
	IsActivePattern(comp, ppp.OneColumnFilter)
}

func DefineStepOneProjectionQueries(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	// copy the number of bytes in a row of the execution data collector into the OneColumnBytesSum column at the start of each segment of 8 rows
	comp.InsertProjection(
		ifaces.QueryIDf("%s_PADDER_PACKER_ONE_COLUMN_PROJECTION_BYTES_SUM", name),
		query.ProjectionInput{
			ColumnA: []ifaces.Column{ppp.InputNoBytes},
			ColumnB: []ifaces.Column{ppp.OneColumnBytesSum},
			FilterA: ppp.InputIsActive,
			FilterB: ppp.TrimmedPeriodicFilter})

	// projection query to copy the limbs into the OneColumn column
	// this also helps to ensure that the OneColumnFilter is correctly set
	comp.InsertProjection(
		ifaces.QueryIDf("%s_PROJECTION_STEP_1", name),
		query.ProjectionMultiAryInput{
			ColumnsA: [][]ifaces.Column{
				[]ifaces.Column{ppp.InputLimbs[0]},
				[]ifaces.Column{ppp.InputLimbs[1]},
				[]ifaces.Column{ppp.InputLimbs[2]},
				[]ifaces.Column{ppp.InputLimbs[3]},
				[]ifaces.Column{ppp.InputLimbs[4]},
				[]ifaces.Column{ppp.InputLimbs[5]},
				[]ifaces.Column{ppp.InputLimbs[6]},
				[]ifaces.Column{ppp.InputLimbs[7]},
			},
			ColumnsB: [][]ifaces.Column{
				[]ifaces.Column{ppp.OneColumn},
			},
			FiltersA: []ifaces.Column{
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
			},
			FiltersB: []ifaces.Column{
				ppp.OneColumnFilter,
			},
		},
	)
}

func DefineStepTwoProjectionQueries(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	comp.InsertProjection(
		ifaces.QueryIDf("%s_PADDER_PACKER_STEP_TWO_ONE_COLUMN_NO_GAPS_PROJECTION", name),
		query.ProjectionInput{
			ColumnA: []ifaces.Column{ppp.OneColumn},
			ColumnB: []ifaces.Column{ppp.OneColumnWithoutGaps},
			FilterA: ppp.OneColumnBytes, // we use OneColumnBytes, since this will represent which limbs are 0 and non-zero
			FilterB: ppp.FilterWithoutGaps})
}

func DefineFilterWithoutGaps(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	util.MustBeBinary(comp, ppp.FilterWithoutGaps)
	IsActivePattern(comp, ppp.FilterWithoutGaps)
}

func DefineStepThreeProjectionQueries(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	comp.InsertProjection(
		ifaces.QueryIDf("%s_PADDER_PACKER_STEP_3_FROM_ONE_COLUMN_TO_8_COLUMNS", name),
		query.ProjectionMultiAryInput{
			ColumnsA: [][]ifaces.Column{
				[]ifaces.Column{ppp.OneColumnWithoutGaps},
			},
			ColumnsB: [][]ifaces.Column{
				[]ifaces.Column{ppp.OuterColumns[0]},
				[]ifaces.Column{ppp.OuterColumns[1]},
				[]ifaces.Column{ppp.OuterColumns[2]},
				[]ifaces.Column{ppp.OuterColumns[3]},
				[]ifaces.Column{ppp.OuterColumns[4]},
				[]ifaces.Column{ppp.OuterColumns[5]},
				[]ifaces.Column{ppp.OuterColumns[6]},
				[]ifaces.Column{ppp.OuterColumns[7]},
			},
			FiltersA: []ifaces.Column{
				ppp.CounterColumnAdded,
			},
			FiltersB: []ifaces.Column{
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
			},
		},
	)
}

func DefineCounterColumnAdded(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_COUNTER_COLUMN_ADDED", name),
		sym.Sub(
			ppp.CounterColumnAdded, // when the Periodic filter is at the beginning of a segment of 8
			ppp.FilterWithoutGaps,
			ppp.CounterColumnFilled,
		),
	)
}

func DefineOuterActiveFilter(comp *wizard.CompiledIOP, ppp *PadderPacker, name string) {
	IsActivePattern(comp, ppp.OuterIsActive)
}

func IsActivePattern(comp *wizard.CompiledIOP, col ifaces.Column) {
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_IS_ACTIVE_CONSTRAINT_NO_0_TO_1", col.GetColID()),
		sym.Sub(
			col,
			sym.Mul(
				column.Shift(col, -1),
				col,
			),
		),
	)
}

// AssignHasher assigns the data in the GenericPadderPacker using the ExtractedData fetched from the arithmetization
func AssignPadderPacker(run *wizard.ProverRuntime, ppp PadderPacker) {
	oneColumn := make([]field.Element, ppp.OneColumn.Size())
	oneColumnWithoutGaps := make([]field.Element, ppp.OneColumnWithoutGaps.Size())
	filterWithoutGaps := make([]field.Element, ppp.FilterWithoutGaps.Size())
	oneColumnFilter := make([]field.Element, ppp.OneColumn.Size())
	periodicFilter := make([][]field.Element, len(ppp.PeriodicFilter))
	oneColumnBytes := make([]field.Element, ppp.OneColumn.Size())
	oneColumnBytesSum := make([]field.Element, ppp.OneColumn.Size())
	counterColumn := make([]field.Element, ppp.CounterColumn.Size())
	counterColumnFilled := make([]field.Element, ppp.CounterColumnFilled.Size())
	for j := range ppp.PeriodicFilter {
		periodicFilter[j] = make([]field.Element, ppp.PeriodicFilter[j].Size())
	}
	outer := make([][]field.Element, len(ppp.OuterColumns))
	for j := range ppp.OuterColumns {
		outer[j] = make([]field.Element, ppp.OuterColumns[j].Size())
	}
	outerIsActive := make([]field.Element, ppp.OuterIsActive.Size())

	counterRow := 0
	for i := 0; i < ppp.InputLimbs[0].Size(); i++ {
		for j := 0; j < common.NbLimbU128; j++ {
			periodicFilter[j][counterRow].SetOne()
			counterRow++
		}
	}

	counterRow = 0

	for i := 0; i < ppp.InputLimbs[0].Size(); i++ {
		isActive := ppp.InputIsActive.GetColAssignmentAt(run, i)
		nBytesLimb := ppp.InputNoBytes.GetColAssignmentAt(run, i)
		nBytesLimbInt := int(nBytesLimb.Uint64())
		remainingNBytes := nBytesLimbInt
		if isActive.IsOne() {
			for j := 0; j < common.NbLimbU128; j++ {
				if j == 0 {
					oneColumnBytesSum[counterRow].SetUint64(uint64(nBytesLimbInt))
				}
				// periodicFilter[j][counterRow].SetOne()
				limbValue := ppp.InputLimbs[j].GetColAssignmentAt(run, i)
				limbBytes := limbValue.Bytes()
				oneColumnFilter[counterRow].SetOne()
				if remainingNBytes > 0 {
					oneColumnBytes[counterRow].SetUint64(2)
					oneColumn[counterRow].SetBytes(limbBytes[2:4])
					// oneColumnFilter[counterRow].SetOne()
					remainingNBytes -= 2
				}
				counterRow++
			}
		}
	}

	counterRow = 0
	for i := 0; i < len(oneColumn); i++ {
		if !oneColumnBytes[i].IsZero() {
			oneColumnWithoutGaps[counterRow].Set(&oneColumn[i])
			filterWithoutGaps[counterRow].SetOne()
			counterRow++
		}
	}

	lastRow := 0
	for i := 0; i < len(oneColumnWithoutGaps); i++ {
		if filterWithoutGaps[i].IsOne() {
			outer[i%common.NbLimbU128][i/common.NbLimbU128].Set(&oneColumnWithoutGaps[i])
			// outerIsActive[i%common.NbLimbU128][i/common.NbLimbU128].SetOne()
			outerIsActive[i/common.NbLimbU128].SetOne()
			counterColumn[i].SetUint64(uint64(i % common.NbLimbU128))
			// counterColumnFilled[i].SetUint64(uint64(i % common.NbLimbU128))
			lastRow = i
		}
	}
	if lastRow%common.NbLimbU128 != 0 {
		for i := lastRow + 1; i%common.NbLimbU128 != 0; i++ {
			counterColumnFilled[i].SetUint64(uint64(i % common.NbLimbU128))
		}
	}

	run.AssignColumn(ppp.OneColumn.GetColID(), sv.NewRegular(oneColumn))
	run.AssignColumn(ppp.OneColumnFilter.GetColID(), sv.NewRegular(oneColumnFilter))
	run.AssignColumn(ppp.OneColumnWithoutGaps.GetColID(), sv.NewRegular(oneColumnWithoutGaps))
	run.AssignColumn(ppp.FilterWithoutGaps.GetColID(), sv.NewRegular(filterWithoutGaps))
	run.AssignColumn(ppp.OneColumnBytes.GetColID(), sv.NewRegular(oneColumnBytes))
	run.AssignColumn(ppp.OneColumnBytesSum.GetColID(), sv.NewRegular(oneColumnBytesSum))
	run.AssignColumn(ppp.CounterColumn.GetColID(), sv.NewRegular(counterColumn))
	run.AssignColumn(ppp.CounterColumnFilled.GetColID(), sv.NewRegular(counterColumnFilled))
	run.AssignColumn(ppp.CounterColumnAdded.GetColID(), sv.Add(sv.NewRegular(filterWithoutGaps), sv.NewRegular(counterColumnFilled)))
	for j := range ppp.PeriodicFilter {
		run.AssignColumn(ppp.PeriodicFilter[j].GetColID(), sv.NewRegular(periodicFilter[j]))
	}
	run.AssignColumn(ppp.TrimmedPeriodicFilter.GetColID(), sv.Mul(sv.NewRegular(periodicFilter[0]), sv.NewRegular(oneColumnFilter)))

	for j := range ppp.OuterColumns {
		run.AssignColumn(ppp.OuterColumns[j].GetColID(), sv.NewRegular(outer[j]))
	}
	run.AssignColumn(ppp.OuterIsActive.GetColID(), sv.NewRegular(outerIsActive))

	for i := 0; i < ppp.OuterColumns[0].Size(); i++ {
		for j := range ppp.OuterColumns {
			fetchedValue := ppp.OuterColumns[j].GetColAssignmentAt(run, i)
			bytes := fetchedValue.Bytes()
			fmt.Println(utils.HexEncodeToString(bytes[:]))
		}
	}

	ppp.ComputeSelectorCounterColumnFilled.Run(run)
}

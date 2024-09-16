package stitcher

import (
	"strings"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/profiling"
)

type stitchingContext struct {
	// The compiled IOP
	comp *wizard.CompiledIOP
	// All columns under the minSize are ignored.
	// No stitching goes beyond MaxSize.
	MinSize, MaxSize int

	// It collects the information about subColumns and their stitchings.
	// The index of Stitchings is over the rounds.
	Stitchings []struct {
		// associate a group of the sub columns to their stitching
		ByStitching map[ifaces.ColID][]ifaces.Column
		// for a sub column, it indicates its stitching column and its position in the stitching.
		BySubCol map[ifaces.ColID]struct {
			NameStitching ifaces.ColID
			PosInNew      int
		}
	}
}

// It stores the information regarding a single stitching.
// used to update the content of [stitchingContex.Stitchings]
type stitching struct {
	// the result of the stitching
	stitchingRes ifaces.Column
	// sub columns that are stitched together
	subCol []ifaces.Column
	// the round in which the sub columns are committed.
	round int
	// status of the sub columns
	// the only valid status for the eligible sub columns are;
	// committed, Precomputed, VerifierDefined
	status column.Status
}

// Stitcher applies the stitching over the eligible sub columns and adjusts the constraints accordingly.
func Stitcher(minSize, maxSize int) func(comp *wizard.CompiledIOP) {

	return func(comp *wizard.CompiledIOP) {
		// it creates stitchings from the eligible columns and commits to the them.
		ctx := newStitcher(comp, minSize, maxSize)

		//  adjust the constraints accordingly over the stitchings of the sub columns.
		ctx.constraints()

		// it assigns the stitching columns and delete the assignment of the sub columns.
		comp.SubProvers.AppendToInner(comp.NumRounds()-1, func(run *wizard.ProverRuntime) {
			for round := range comp.NumRounds() {
				for subCol := range ctx.Stitchings[round].BySubCol {
					run.Columns.TryDel(subCol)
				}
			}
		})
	}
}

// it commits to the stitchings of the eligible sub columns.
func newStitcher(comp *wizard.CompiledIOP, minSize, maxSize int) stitchingContext {
	numRounds := comp.NumRounds()
	res := stitchingContext{
		comp:    comp,
		MinSize: minSize,
		MaxSize: maxSize,
		// initialize the stitichings
		Stitchings: make([]struct {
			ByStitching map[ifaces.ColID][]ifaces.Column
			BySubCol    map[ifaces.ColID]struct {
				NameStitching ifaces.ColID
				PosInNew      int
			}
		}, numRounds),
	}
	// it scans the compiler trace for the eligible columns, creates stitchings from the sub columns and commits to the them.
	res.ScanStitchCommit()
	return res
}

// ScanStitchCommit scans compiler trace and classifies the sub columns eligible to the stitching.
// It then stitches the sub columns, commits to them and update stitchingContext.
// It also forces the compiler to set the status of the sub columns to 'ignored'.
// since the sub columns are technically replaced with their stitching.
func (ctx *stitchingContext) ScanStitchCommit() {
	for round := 0; round < ctx.comp.NumRounds(); round++ {

		// scan the compiler trace to find the eligible columns for stitching
		columnsBySize := scanAndClassifyEligibleColumns(*ctx, round)

		for size, cols := range columnsBySize {

			var (
				precomputedCols = make([]ifaces.Column, 0, len(cols))
				committedCols   = make([]ifaces.Column, 0, len(cols))
			)

			// collect the the columns with valid status; Precomputed, committed
			// verifierDefined is valid but is not present in the compiler trace we handle it directly during the constraints.
			for _, col := range cols {
				status := ctx.comp.Columns.Status(col.GetColID())
				switch status {
				case column.Precomputed:
					precomputedCols = append(precomputedCols, col)
				case column.Committed:
					committedCols = append(committedCols, col)

				default:
					utils.Panic("found the column %v with the invalid status %v for stitching", col.GetColID(), status.String())
				}

				// Mark it as ignored, so that it is no longer considered as
				// queryable (since we are replacing it with its stitching).
				ctx.comp.Columns.MarkAsIgnored(col.GetColID())
			}

			if len(precomputedCols) != 0 {
				// classify the columns to the groups, each of size ctx.MaxSize
				preComputedGroups := groupCols(precomputedCols, ctx.MaxSize/size)

				for _, group := range preComputedGroups {
					// prepare a group for stitching
					stitching := stitching{
						subCol: group,
						round:  round,
						status: column.Precomputed,
					}
					// stitch the group
					ctx.stitchGroup(stitching)
				}
			}

			if len(committedCols) != 0 {
				committedGroups := groupCols(committedCols, ctx.MaxSize/size)

				for _, group := range committedGroups {
					stitching := stitching{
						subCol: group,
						round:  round,
						status: column.Committed,
					}
					ctx.stitchGroup(stitching)

				}
			}

		}

		ctx.comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
			stopTimer := profiling.LogTimer("stitching compiler")
			defer stopTimer()
			for id, subColumns := range ctx.Stitchings[round].ByStitching {
				// Trick, in order to compute the assignment of stitching column, we
				// extract the witness of the interleaving of the grouped
				// columns.
				witnesses := make([]smartvectors.SmartVector, len(subColumns))
				for i := range witnesses {
					witnesses[i] = subColumns[i].GetColAssignment(run)
				}
				assignement := smartvectors.
					AllocateRegular(len(subColumns) * witnesses[0].Len()).(*smartvectors.Regular)
				for i := range subColumns {
					for j := 0; j < witnesses[0].Len(); j++ {
						(*assignement)[i+j*len(subColumns)] = witnesses[i].Get(j)
					}
				}
				run.AssignColumn(id, assignement)
			}
		})
	}
}

// It scan the compiler trace for a given round and classifies the columns eligible to the stitching, by their size.
func scanAndClassifyEligibleColumns(ctx stitchingContext, round int) map[int][]ifaces.Column {
	columnsBySize := map[int][]ifaces.Column{}

	for _, colName := range ctx.comp.Columns.AllKeysAt(round) {

		status := ctx.comp.Columns.Status(colName)
		col := ctx.comp.Columns.GetHandle(colName)

		// 1. we expect no constraints over a mix of eligible columns and proof, thus ignore Proof columns
		// 2. we expect no verifingKey column to fall withing the stitching interval (ctx.MinSize, ctx.MaxSize)
		// 3. we expect no query over the ignored columns.
		if status == column.Ignored || status == column.Proof || status == column.VerifyingKey {
			continue
		}

		// If the sizes are either too small or too large, stitcher does not manipulate the column.
		if ctx.MinSize > col.Size() || col.Size() >= ctx.MaxSize {
			continue
		}

		// Initialization clause of `sizes`
		if _, ok := columnsBySize[col.Size()]; !ok {
			columnsBySize[col.Size()] = []ifaces.Column{}
		}

		columnsBySize[col.Size()] = append(columnsBySize[col.Size()], col)
	}
	return columnsBySize
}

// group the cols with the same size
func groupCols(cols []ifaces.Column, numToStitch int) (groups [][]ifaces.Column) {

	numGroups := utils.DivCeil(len(cols), numToStitch)
	groups = make([][]ifaces.Column, numGroups)

	size := cols[0].Size()

	for i, col := range cols {
		if col.Size() != size {
			utils.Panic(
				"column %v of size %v has been grouped with %v of size %v",
				col.GetColID(), col.Size(), cols[0].GetColID(), cols[0].Size(),
			)
		}
		groups[i/numToStitch] = append(groups[i/numToStitch], col)
	}

	lastGroup := &groups[len(groups)-1]
	zeroCol := verifiercol.NewConstantCol(field.Zero(), size)

	for i := len(*lastGroup); i < numToStitch; i++ {
		*lastGroup = append(*lastGroup, zeroCol)
	}

	return groups
}

func groupedName(group []ifaces.Column) ifaces.ColID {
	fmtted := make([]string, len(group))
	for i := range fmtted {
		fmtted[i] = group[i].String()
	}
	return ifaces.ColIDf("STITCHER_%v", strings.Join(fmtted, "_"))
}

// for a group of sub columns it creates their stitching.
func (ctx *stitchingContext) stitchGroup(s stitching) {
	var (
		group        = s.subCol
		stitchingCol ifaces.Column
		status       = s.status
	)
	// Declare the new columns
	switch status {
	case column.Precomputed:
		values := make([][]field.Element, len(group))
		for j := range values {
			values[j] = smartvectors.IntoRegVec(ctx.comp.Precomputed.MustGet(group[j].GetColID()))
		}
		assignement := vector.Interleave(values...)
		stitchingCol = ctx.comp.InsertPrecomputed(
			groupedName(group),
			smartvectors.NewRegular(assignement),
		)
	case column.Committed:
		stitchingCol = ctx.comp.InsertCommit(
			s.round,
			groupedName(s.subCol),
			ctx.MaxSize,
		)

	default:
		panic("The status is not valid for the stitching")

	}

	s.stitchingRes = stitchingCol
	ctx.insertNew(s)
}

// it inserts the new stitching to [stitchingContex]
func (ctx *stitchingContext) insertNew(s stitching) {
	// Initialize the bySubCol if necessary
	if ctx.Stitchings[s.round].BySubCol == nil {
		ctx.Stitchings[s.round].BySubCol = map[ifaces.ColID]struct {
			NameStitching ifaces.ColID
			PosInNew      int
		}{}
	}

	// Populate the bySubCol
	for posInNew, c := range s.subCol {
		ctx.Stitchings[s.round].BySubCol[c.GetColID()] = struct {
			NameStitching ifaces.ColID
			PosInNew      int
		}{
			NameStitching: s.stitchingRes.GetColID(),
			PosInNew:      posInNew,
		}
	}

	//Initialize byStitching f necessary
	if ctx.Stitchings[s.round].ByStitching == nil {
		ctx.Stitchings[s.round].ByStitching = make(map[ifaces.ColID][]ifaces.Column)
	}
	//populate byStitching
	ctx.Stitchings[s.round].ByStitching[s.stitchingRes.GetColID()] = s.subCol
}

// it checks if the column belongs to a stitching.
func isColEligible(ctx stitchingContext, col ifaces.Column) bool {
	natural := column.RootParents(col)[0]
	_, found := ctx.Stitchings[col.Round()].BySubCol[natural.GetColID()]
	return found
}

// It checks if the expression is over a set of the columns eligible to the stitching.
// Namely, it contains columns of proper size with status Precomputed, Committed, or Verifiercol.
// It panics if the expression includes a mixture of eligible columns and columns with status Proof/VerifiyingKey/Ignored.
//
// If all the columns are verifierCol the expression is not eligible to the compilation.
// This is an expected behavior, since the verifier checks such expression by itself.
func isExprEligible(ctx stitchingContext, board symbolic.ExpressionBoard) bool {
	metadata := board.ListVariableMetadata()
	hasAtLeastOneEligible := false
	allAreEligible := true
	allAreVeriferCol := true
	for i := range metadata {
		switch m := metadata[i].(type) {
		// reminder: [verifiercol.VerifierCol] , [column.Natural] and [column.Shifted]
		// all implement [ifaces.Column]
		case ifaces.Column: // it is a Committed, Precomputed or verifierCol
			natural := column.RootParents(m)[0]
			switch natural.(type) {
			case column.Natural: // then it is not a verifiercol
				allAreVeriferCol = false
				b := isColEligible(ctx, m)

				hasAtLeastOneEligible = hasAtLeastOneEligible || b
				allAreEligible = allAreEligible && b
				if m.Size() == 0 {
					panic("found no columns in the expression")
				}
			}

		}

	}

	if hasAtLeastOneEligible && !allAreEligible {
		// 1. we expect no expression including Proof columns
		// 2. we expect no expression over ignored columns
		// 3. we expect no VerifiyingKey withing the stitching range.
		panic("the expression is not valid, it is mixed with invalid columns of status Proof/Ingnored/verifierKey")
	}
	if allAreVeriferCol {
		// 4. we expect no expression involving only and only the verifierCols.
		// We expect that this case wont happen.
		// Otherwise should be handled in the [github.com/consensys/zkevm-monorepo/prover/protocol/query] package.
		// Namely, Local/Global queries should be checked directly by the verifer.
		panic("all the columns in the expression are verifierCols, unsupported by the compiler")
	}

	return hasAtLeastOneEligible
}

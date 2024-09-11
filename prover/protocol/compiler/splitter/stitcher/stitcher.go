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
	isPreComputed bool
	//TBD: supporting verifierDefined via expansion
	// for the moment it is handled similar to committed columns (which is not very efficient).
}

// Stitcher applies the stitching over the eligible sub columns and adjusts the constraints accordingly.
func Stitcher(minSize, maxSize int) func(comp *wizard.CompiledIOP) {

	return func(comp *wizard.CompiledIOP) {
		// it creates stitchings from the eligible columns and commits to the them.
		ctx := newStitcher(comp, minSize, maxSize)

		// ignore the constraints over the subColumns.
		// ctx.IgnoreConstraintsOverSubColumns()

		//  adjust the constraints accordingly over the stitchings of the sub columns.
		ctx.constraints()

		// it assign the stitching columns and delete the assignment of the sub columns.
		comp.SubProvers.AppendToInner(comp.NumRounds()-1, func(run *wizard.ProverRuntime) {
			for round := range comp.NumRounds() {
				for subCol, _ := range ctx.Stitchings[round].BySubCol {
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
				normalCols      = make([]ifaces.Column, 0, len(cols))
			)

			// collect the precomputedColumns and normal columns.
			for _, col := range cols {
				if ctx.comp.Columns.Status(col.GetColID()) == column.Precomputed {
					precomputedCols = append(precomputedCols, col)
				} else {
					normalCols = append(normalCols, col)
				}
			}

			if len(precomputedCols) != 0 {
				// classify the columns to the groups, each of size ctx.MaxSize
				preComputedGroups := groupCols(precomputedCols, ctx.MaxSize/size)

				for _, group := range preComputedGroups {
					// prepare a group for stitching
					stitching := stitching{
						subCol:        group,
						round:         round,
						isPreComputed: true,
					}
					// stitch the group
					ctx.stitchGroup(stitching)
				}
			}

			if len(normalCols) != 0 {
				normalGroups := groupCols(normalCols, ctx.MaxSize/size)

				for _, group := range normalGroups {
					stitching := stitching{
						subCol:        group,
						round:         round,
						isPreComputed: false,
					}
					ctx.stitchGroup(stitching)

				}
			}

		}

		ctx.comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
			stopTimer := profiling.LogTimer("stitching compiler")
			defer stopTimer()
			for id, subColumns := range ctx.Stitchings[round].ByStitching {
				// Trick, in order to compute the assignment of newName we
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

		// Mark it as ignored, so that it is no longer considered as
		// queryable (since we are replacing it with its stitching).
		ctx.comp.Columns.MarkAsIgnored(colName)

		// Initialization clause of `sizes`
		if _, ok := columnsBySize[col.Size()]; !ok {
			columnsBySize[col.Size()] = []ifaces.Column{}
		}

		columnsBySize[col.Size()] = append(columnsBySize[col.Size()], col)
	}
	return columnsBySize
}

// group the cols with the same size
func groupCols(cols []ifaces.Column, numToStick int) (groups [][]ifaces.Column) {

	numGroups := utils.DivCeil(len(cols), numToStick)
	groups = make([][]ifaces.Column, numGroups)

	size := cols[0].Size()

	for i, col := range cols {
		if col.Size() != size {
			utils.Panic(
				"column %v of size %v has been grouped with %v of size %v",
				col.GetColID(), col.Size(), cols[0].GetColID(), cols[0].Size(),
			)
		}
		groups[i/numToStick] = append(groups[i/numToStick], col)
	}

	lastGroup := &groups[len(groups)-1]
	zeroCol := verifiercol.NewConstantCol(field.Zero(), size)

	for i := len(*lastGroup); i < numToStick; i++ {
		*lastGroup = append(*lastGroup, zeroCol)
	}

	return groups
}

func groupedName(group []ifaces.Column) ifaces.ColID {
	fmtted := make([]string, len(group))
	for i := range fmtted {
		fmtted[i] = group[i].String()
	}
	return ifaces.ColIDf("STICKER_%v", strings.Join(fmtted, "_"))
}

// for a group of sub columns it creates their stitching.
func (ctx *stitchingContext) stitchGroup(s stitching) {
	var (
		group        = s.subCol
		stitchingCol ifaces.Column
	)
	// Declare the new columns
	if s.isPreComputed {
		values := make([][]field.Element, len(group))
		for j := range values {
			values[j] = smartvectors.IntoRegVec(ctx.comp.Precomputed.MustGet(group[j].GetColID()))
		}
		assignement := vector.Interleave(values...)
		stitchingCol = ctx.comp.InsertPrecomputed(
			groupedName(group),
			smartvectors.NewRegular(assignement),
		)
	} else {
		stitchingCol = ctx.comp.InsertCommit(
			s.round,
			groupedName(s.subCol),
			ctx.MaxSize,
		)

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
// It panics if the expression includes a mixture of eligible columns and Proof/VerifiyingKey/Ignored status.
func isExprEligible(ctx stitchingContext, board symbolic.ExpressionBoard) bool {
	metadata := board.ListVariableMetadata()
	hasAtLeastOneEligible := false
	allAreEligible := true
	for i := range metadata {
		switch m := metadata[i].(type) {
		case ifaces.Column:
			b := isColEligible(ctx, m)

			hasAtLeastOneEligible = hasAtLeastOneEligible || b
			allAreEligible = allAreEligible && b

			if m.Size() == 0 {
				panic("found no columns in the expression")
			}
		}

	}

	if hasAtLeastOneEligible && !allAreEligible {
		// 1. we expect no expression including Proof columns
		// 2. we expect no expression over ignored columns
		// 3. we expect no VerifiyingKey withing the stitching range.
		panic("the expression is not valid, it incudes a mix of eligible columns and invalid columns i.e., Ignored,Proof, VerifingKey")
	}

	return hasAtLeastOneEligible
}

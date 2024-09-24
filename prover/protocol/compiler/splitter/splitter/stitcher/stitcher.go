package stitcher

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	alliance "github.com/consensys/linea-monorepo/prover/protocol/compiler/splitter/splitter"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
)

type stitchingContext struct {
	// The compiled IOP
	comp *wizard.CompiledIOP
	// All columns under the minSize are ignored.
	// No stitching goes beyond MaxSize.
	MinSize, MaxSize int

	// It collects the information about subColumns and their stitchings.
	// The index of Stitchings is over the rounds.
	Stitchings []alliance.SummerizedAlliances
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
		Stitchings: make([]alliance.SummerizedAlliances, numRounds),
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
					// note that status of verifercol/ veriferDefined is not available via compiler trace.
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
					stitching := alliance.Alliance{
						SubCols: group,
						Round:   round,
						Status:  column.Precomputed,
					}
					// stitch the group
					ctx.stitchGroup(stitching)
				}
			}

			if len(committedCols) != 0 {
				committedGroups := groupCols(committedCols, ctx.MaxSize/size)

				for _, group := range committedGroups {
					stitching := alliance.Alliance{
						SubCols: group,
						Round:   round,
						Status:  column.Committed,
					}
					ctx.stitchGroup(stitching)

				}
			}

		}
		// @Azam Precomputed ones are double assigned by this?
		ctx.comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
			stopTimer := profiling.LogTimer("stitching compiler")
			defer stopTimer()
			for idBigCol, subColumns := range ctx.Stitchings[round].ByBigCol {

				// Sanity-check
				sizeBigCol := ctx.comp.Columns.GetHandle(idBigCol).Size()
				if sizeBigCol != ctx.MaxSize {
					utils.Panic("Unexpected size %v != %v", sizeBigCol, ctx.MaxSize)
				}

				// If the column is precomputed, it is already assigned
				if ctx.comp.Precomputed.Exists(idBigCol) {
					continue
				}

				// get the assignment of the subColumns and interleave them
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
				run.AssignColumn(idBigCol, assignement)
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

		// 1. we expect no constraint over a mix of eligible columns and proof, thus ignore Proof columns
		// 2. we expect no verifingKey column to fall withing the stitching interval (ctx.MinSize, ctx.MaxSize)
		// 3. we expect no query over the ignored columns.
		if status == column.Ignored || status == column.Proof || status == column.VerifyingKey {
			continue
		}

		// If the column is too big, the stitcher does not manipulate the column.
		if col.Size() >= ctx.MaxSize {
			continue
		}

		//  If the column is very small, make it public.
		if col.Size() < ctx.MinSize {
			if status.IsPublic() {
				// Nothing to do : the column is already public and we will ask the
				// verifier to perform the operation itself.
				continue
			}
			ctx.makeColumnPublic(col, status)
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
func (ctx *stitchingContext) stitchGroup(s alliance.Alliance) {
	var (
		group        = s.SubCols
		stitchingCol ifaces.Column
		status       = s.Status
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
			s.Round,
			groupedName(s.SubCols),
			ctx.MaxSize,
		)

	default:
		panic("The status is not valid for the stitching")

	}

	s.BigCol = stitchingCol
	(alliance.MultiSummary)(ctx.Stitchings).InsertNew(s)
}

// it checks if the column belongs to a stitching.
func isColEligible(stitchings alliance.MultiSummary, col ifaces.Column) bool {
	natural := column.RootParents(col)[0]
	_, found := stitchings[col.Round()].BySubCol[natural.GetColID()]
	return found
}

// It makes the given colum public.
// If the colum is Precomputed it becomes the VerifierKey, otherwise it becomes Proof.
func (ctx stitchingContext) makeColumnPublic(col ifaces.Column, status column.Status) {
	switch status {
	case column.Precomputed:
		// send it to the verifier directly as part of the verifying key
		ctx.comp.Columns.SetStatus(col.GetColID(), column.VerifyingKey)
	case column.Committed:
		// send it to the verifier directly as part of the proof
		ctx.comp.Columns.SetStatus(col.GetColID(), column.Proof)
	default:
		utils.Panic("Unknown status : %v", status.String())
	}
}

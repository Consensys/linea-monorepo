package splitter

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	alliance "github.com/consensys/linea-monorepo/prover/protocol/compiler/stitch_split"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
)

// Splitter
func Splitter(size int) func(*wizard.CompiledIOP) {
	return func(comp *wizard.CompiledIOP) {
		// it creates the splitting columns (from the eligible columns), and commits to them
		ctx := newSplitter(comp, size)
		// adjust the constraints accordingly over the stitchings of the sub columns.
		ctx.constraints()

		// it assigns the stitching columns and delete the assignment of the sub columns.
		comp.RegisterProverAction(comp.NumRounds()-1, &splitterAssignProverAction{ctx: ctx})
	}
}

// splitterAssignProverAction handles the assignment and deletion of columns in the splitter.
// It implements the [wizard.ProverAction] interface.
type splitterAssignProverAction struct {
	ctx splitterContext
}

// Run executes the column deletion for the splitter.
func (a *splitterAssignProverAction) Run(run *wizard.ProverRuntime) {
	for round := range a.ctx.Splittings {
		for bigCol := range a.ctx.Splittings[round].ByBigCol {
			run.Columns.TryDel(bigCol)
		}
	}
}

type splitterContext struct {
	// the compiled IOP
	comp *wizard.CompiledIOP
	// the size for splitting the big columns
	size int
	// It collects the information about the splitting and subColumns.
	// The index of Splittings is over the rounds.
	Splittings []alliance.SummerizedAlliances
}

func newSplitter(comp *wizard.CompiledIOP, size int) splitterContext {
	numRound := comp.NumRounds()
	ctx := splitterContext{
		comp:       comp,
		size:       size,
		Splittings: make([]alliance.SummerizedAlliances, numRound),
	}

	ctx.ScanSplitCommit()
	return ctx
}

func (ctx *splitterContext) ScanSplitCommit() {
	comp := ctx.comp
	for round := 0; round < comp.NumRounds(); round++ {
		for _, col := range comp.Columns.AllHandlesAtRound(round) {

			status := comp.Columns.Status(col.GetColID())

			// 1. we expect no constraint over a mix of eligible columns and proof, thus ignore Proof columns
			// 2. we expect no verifingKey column to fall withing the stitching interval (ctx.MinSize, ctx.MaxSize)
			// 3. we expect no query over the ignored columns.
			if status == column.Ignored || status == column.Proof || status == column.VerifyingKey {
				continue
			}
			if col.Size() < ctx.size {
				utils.Panic("stitcher is not working correctly, the small columns should have been handled by the stitcher")
			}

			if col.Size()%ctx.size != 0 {
				utils.Panic("the column size %v does not divide the splitting size %v", col.Size(), ctx.size)
			}
			if col.Size() == ctx.size {
				continue
			}

			// Create the subslices and give them the same status as their parents
			numSubSlices := col.Size() / ctx.size
			subSlices := make([]ifaces.Column, numSubSlices)

			// Note that the status of Verifiercol is not available via compiler.
			// Thus, it is not handles here, we rather handle it during the constraints.
			switch status {
			// if the veriferKey is big, it should be declared as precomputed
			case column.Precomputed, column.VerifyingKey:
				// Then, on top of defining the new split column. We need to assign it
				// directly.
				precomp := comp.Precomputed.MustGet(col.GetColID())
				for i := 0; i < len(subSlices); i++ {
					// InsertPrecomputed() automatically assigns the
					// status "Precomputed".
					subSlices[i] = comp.InsertPrecomputed(
						nameHandleSlice(col, i, col.Size()/ctx.size),
						precomp.SubVector(i*ctx.size, (i+1)*ctx.size),
					)
					// For the verifyingKey, declare the status manually.
					if status != column.Precomputed {
						comp.Columns.SetStatus(subSlices[i].GetColID(), status)
					}
				}

			case column.Committed:
				for i := 0; i < len(subSlices); i++ {
					subSlices[i] = comp.InsertCommit(round,
						nameHandleSlice(col, i, col.Size()/ctx.size),
						ctx.size,
					)
				}
			default:
				panic("Invalid Status")
			}

			splitting := alliance.Alliance{
				BigCol:  col,
				SubCols: subSlices,
				Round:   round,
				Status:  status,
			}

			(alliance.MultiSummary)(ctx.Splittings).InsertNew(splitting)

			// Mark the handle as ignored
			comp.Columns.MarkAsIgnored(col.GetColID())
		}
		ctx.comp.RegisterProverAction(round, &scanSplitCommitAssignProverAction{ctx: ctx, round: round})
	}
}

// scanSplitCommitAssignProverAction handles the prover action for splitting columns.
// It implements the [wizard.ProverAction] interface.
type scanSplitCommitAssignProverAction struct {
	ctx   *splitterContext
	round int
}

// Run executes the prover action for the specified round.
func (a *scanSplitCommitAssignProverAction) Run(run *wizard.ProverRuntime) {
	a.ctx.Prove(a.round)(run)
}

func nameHandleSlice(h ifaces.Column, num, numSlots int) ifaces.ColID {
	return ifaces.ColIDf("%v_SUBSLICE_%v_OVER_%v", h.GetColID(), num, numSlots)
}

func (ctx splitterContext) Prove(round int) wizard.MainProverStep {

	return func(run *wizard.ProverRuntime) {
		stopTimer := profiling.LogTimer("splitter compiler")
		defer stopTimer()

		for idBigCol, subCols := range ctx.Splittings[round].ByBigCol {

			// Sanity-check
			bigCol := ctx.comp.Columns.GetHandle(idBigCol)
			if len(subCols)*ctx.size != bigCol.Size() {
				utils.Panic("Unexpected sizes %v  * %v != %v", len(subCols), ctx.size, bigCol.Size())
			}

			// If the column is precomputed, it was already assigned
			if ctx.comp.Precomputed.Exists(idBigCol) {
				continue
			}

			// assign the subColumns
			witness := bigCol.GetColAssignment(run)
			for i := 0; i < len(subCols); i++ {
				run.AssignColumn(subCols[i].GetColID(), witness.SubVector(i*ctx.size, (i+1)*ctx.size))
			}
		}
	}
}

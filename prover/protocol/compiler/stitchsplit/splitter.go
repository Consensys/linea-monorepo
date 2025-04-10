package stitchsplit

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

type splitProverAction struct {
	splittings []SummerizedAlliances
}

func (a *splitProverAction) Run(run *wizard.ProverRuntime) {
	for round := range a.splittings {
		for bigCol := range a.splittings[round].ByBigCol {
			run.Columns.TryDel(bigCol)
		}
	}
}

// Splitter
func Splitter(size int) func(*wizard.CompiledIOP) {
	return func(comp *wizard.CompiledIOP) {
		ctx := newSplitter(comp, size)
		ctx.constraints()
		comp.RegisterProverAction(comp.NumRounds()-1, &splitProverAction{
			splittings: ctx.Splittings,
		})
	}
}

type splitterContext struct {
	// the compiled IOP
	comp *wizard.CompiledIOP
	// the size for splitting the big columns
	size int
	// It collects the information about the splitting and subColumns.
	// The index of Splittings is over the rounds.
	Splittings []SummerizedAlliances
}

func newSplitter(comp *wizard.CompiledIOP, size int) splitterContext {
	numRound := comp.NumRounds()
	ctx := splitterContext{
		comp:       comp,
		size:       size,
		Splittings: make([]SummerizedAlliances, numRound),
	}

	ctx.ScanSplitCommit()
	return ctx
}

type proveRoundProverAction struct {
	ctx   *splitterContext
	round int
}

func (a *proveRoundProverAction) Run(run *wizard.ProverRuntime) {
	stopTimer := profiling.LogTimer("splitter compiler")
	defer stopTimer()

	for idBigCol, subCols := range a.ctx.Splittings[a.round].ByBigCol {
		bigCol := a.ctx.comp.Columns.GetHandle(idBigCol)
		if len(subCols)*a.ctx.size != bigCol.Size() {
			utils.Panic("Unexpected sizes %v * %v != %v", len(subCols), a.ctx.size, bigCol.Size())
		}

		if a.ctx.comp.Precomputed.Exists(idBigCol) {
			continue
		}

		witness := bigCol.GetColAssignment(run)
		for i := 0; i < len(subCols); i++ {
			run.AssignColumn(subCols[i].GetColID(), witness.SubVector(i*a.ctx.size, (i+1)*a.ctx.size))
		}
	}
}

func (ctx *splitterContext) ScanSplitCommit() {
	comp := ctx.comp
	for round := 0; round < comp.NumRounds(); round++ {
		for _, col := range comp.Columns.AllHandlesAtRound(round) {
			status := comp.Columns.Status(col.GetColID())
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
			numSubSlices := col.Size() / ctx.size
			subSlices := make([]ifaces.Column, numSubSlices)
			switch status {
			case column.Precomputed, column.VerifyingKey:
				precomp := comp.Precomputed.MustGet(col.GetColID())
				for i := 0; i < len(subSlices); i++ {
					subSlices[i] = comp.InsertPrecomputed(
						nameHandleSlice(col, i, numSubSlices),
						precomp.SubVector(i*ctx.size, (i+1)*ctx.size),
					)
					if status != column.Precomputed {
						comp.Columns.SetStatus(subSlices[i].GetColID(), status)
					}
				}
			case column.Committed:
				for i := 0; i < len(subSlices); i++ {
					subSliceName := nameHandleSlice(col, i, numSubSlices)
					if !comp.Columns.Exists(subSliceName) {
						subSlices[i] = comp.InsertCommit(round, subSliceName, ctx.size)
					} else {
						subSlices[i] = comp.Columns.GetHandle(subSliceName)
						logrus.Infof("Reusing existing subsliced column: %v", subSliceName)
					}
				}
			default:
				panic("Invalid Status")
			}
			splitting := Alliance{
				BigCol:  col,
				SubCols: subSlices,
				Round:   round,
				Status:  status,
			}
			(MultiSummary)(ctx.Splittings).InsertNew(splitting)
			comp.Columns.MarkAsIgnored(col.GetColID())
		}
		if len(ctx.Splittings[round].ByBigCol) == 0 {
			continue
		}
		ctx.comp.RegisterProverAction(round, &proveRoundProverAction{
			ctx:   ctx,
			round: round,
		})
	}
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

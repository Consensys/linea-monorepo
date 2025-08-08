package stitchsplit

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

type SplitProverAction struct {
	Splittings []SummerizedAlliances
}

func (a *SplitProverAction) Run(run *wizard.ProverRuntime) {
	for round := range a.Splittings {
		// This is an iteration over a map so the order is non-deterministic
		// but this does not matter here.
		for bigCol := range a.Splittings[round].ByBigCol {
			run.Columns.TryDel(bigCol)
		}
	}
}

func Splitter(size int) func(*wizard.CompiledIOP) {
	return func(comp *wizard.CompiledIOP) {
		ctx := newSplitter(comp, size)
		ctx.constraints()
		comp.RegisterProverAction(comp.NumRounds()-1, &SplitProverAction{
			Splittings: ctx.Splittings,
		})
	}
}

type SplitterContext struct {
	Comp       *wizard.CompiledIOP
	Size       int
	Splittings []SummerizedAlliances
}

func newSplitter(comp *wizard.CompiledIOP, size int) SplitterContext {
	numRound := comp.NumRounds()
	ctx := SplitterContext{
		Comp:       comp,
		Size:       size,
		Splittings: make([]SummerizedAlliances, numRound),
	}
	ctx.ScanSplitCommit()
	return ctx
}

type ProveRoundProverAction struct {
	Ctx   *SplitterContext
	Round int
}

func (a *ProveRoundProverAction) Run(run *wizard.ProverRuntime) {
	stopTimer := profiling.LogTimer("splitter compiler")
	defer stopTimer()

	// This sorting is necessary to ensure that we iterate in deterministic
	// order over the [ByBigCol] map.
	idBigCols := utils.SortedKeysOf(a.Ctx.Splittings[a.Round].ByBigCol, func(a, b ifaces.ColID) bool {
		return a < b
	})

	for _, idBigCol := range idBigCols {

		subCols := a.Ctx.Splittings[a.Round].ByBigCol[idBigCol]
		bigCol := a.Ctx.Comp.Columns.GetHandle(idBigCol)

		if len(subCols)*a.Ctx.Size != bigCol.Size() {
			utils.Panic("Unexpected sizes %v * %v != %v", len(subCols), a.Ctx.Size, bigCol.Size())
		}
		if a.Ctx.Comp.Precomputed.Exists(idBigCol) {
			continue
		}
		witness := bigCol.GetColAssignment(run)
		for i := 0; i < len(subCols); i++ {
			run.AssignColumn(subCols[i].GetColID(), witness.SubVector(i*a.Ctx.Size, (i+1)*a.Ctx.Size))
		}
	}
}

func (ctx *SplitterContext) ScanSplitCommit() {
	comp := ctx.Comp
	for round := 0; round < comp.NumRounds(); round++ {
		for _, col := range comp.Columns.AllHandlesAtRound(round) {
			status := comp.Columns.Status(col.GetColID())
			if status == column.Ignored || status == column.Proof || status == column.VerifyingKey {
				continue
			}
			if col.Size() < ctx.Size {
				utils.Panic("stitcher is not working correctly, the small columns should have been handled by the stitcher")
			}
			if col.Size()%ctx.Size != 0 {
				utils.Panic("the column size %v does not divide the splitting size %v", col.Size(), ctx.Size)
			}
			if col.Size() == ctx.Size {
				continue
			}
			numSubSlices := col.Size() / ctx.Size
			subSlices := make([]ifaces.Column, numSubSlices)
			switch status {
			case column.Precomputed, column.VerifyingKey:
				precomp := comp.Precomputed.MustGet(col.GetColID())
				for i := 0; i < len(subSlices); i++ {
					subSlices[i] = comp.InsertPrecomputed(
						nameHandleSlice(col, i, numSubSlices),
						precomp.SubVector(i*ctx.Size, (i+1)*ctx.Size),
					)
					if status != column.Precomputed {
						comp.Columns.SetStatus(subSlices[i].GetColID(), status)
					}
				}
			case column.Committed:
				for i := 0; i < len(subSlices); i++ {
					subSliceName := nameHandleSlice(col, i, numSubSlices)
					if !comp.Columns.Exists(subSliceName) {
						subSlices[i] = comp.InsertCommit(round, subSliceName, ctx.Size)
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
		ctx.Comp.RegisterProverAction(round, &ProveRoundProverAction{
			Ctx:   ctx,
			Round: round,
		})
	}
}

func nameHandleSlice(h ifaces.Column, num, numSlots int) ifaces.ColID {
	return ifaces.ColIDf("%v_SUBSLICE_%v_OVER_%v", h.GetColID(), num, numSlots)
}

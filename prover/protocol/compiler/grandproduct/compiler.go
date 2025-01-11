package grandproduct

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// CompileGrandProductDist compiles [query.GrandProduct] queries and
func CompileGrandProductDist(comp *wizard.CompiledIOP) {
	var (
		allProverActions = make([]permutation.ProverTaskAtRound, comp.NumRounds()+1)
		// zCatalog stores a mapping (round, size) into ZCtx and helps finding
		// which Z context should be used to handle a part of a given permutation
		// query.
		zCatalog = map[[2]int]*permutation.ZCtx{}
	)

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {

		// Filter out non grand product queries
		grandproduct, ok := comp.QueriesParams.Data(qName).(query.GrandProduct)
		if !ok {
			continue
		}

		// This ensures that the grand product query is not used again in the
		// compilation process. We know that the query was not already ignored at the beginning because we are iterating over the unignored keys.
		comp.QueriesParams.MarkAsIgnored(qName)
		round := comp.QueriesParams.Round(qName)

		dispatchGrandProduct(zCatalog, round, grandproduct)
	}

	for entry, zC := range zCatalog {
		zC.Compile(comp)
		round := entry[0]
		allProverActions[round] = append(allProverActions[round], zC)
	}

	for round := range allProverActions {
		if len(allProverActions[round]) > 0 {
			comp.RegisterProverAction(round, allProverActions[round])
			comp.RegisterVerifierAction(round, permutation.VerifierCtx(allProverActions[round]))
		}
	}

}

// dispatchGrandProduct applies the grand product argument compilation over
// a specific [query.GrandProduct]
func dispatchGrandProduct(
	zCatalog map[[2]int]*permutation.ZCtx,
	round int,
	q query.GrandProduct,
) {
	for size, gpInputs := range q.Inputs {
		var (
			catalogEntry = [2]int{round + 1, size}
		)
		if _, ok := zCatalog[catalogEntry]; !ok {
			zCatalog[catalogEntry] = &permutation.ZCtx{
				Size:  size,
				Round: round + 1,
			}
		}
		ctx := zCatalog[catalogEntry]
		ctx.NumeratorFactors = append(ctx.NumeratorFactors, gpInputs.Numerators...)
		ctx.DenominatorFactors = append(ctx.DenominatorFactors, gpInputs.Denominators...)

	}

}

package permutation

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func IntoGrandProduct(comp *wizard.CompiledIOP) {
	numRounds := comp.NumRounds()

	/*
		Handles the lookups and permutations checks
	*/
	for i := 0; i < numRounds; i++ {
		queries := comp.QueriesNoParams.AllKeysAt(i)
		for _, qName := range queries {
			// Skip if it was already compiled
			if comp.QueriesNoParams.IsIgnored(qName) {
				continue
			}

			switch q_ := comp.QueriesNoParams.Data(qName).(type) {
			case query.Permutation:
				reducePermutationIntoGrandProduct(comp, q_, i)
			}
		}
	}
}
// The below function does the following:
// 1. Register beta and alpha (for the random linear combination in case A and B are multi-columns)
// 2. Tell the prover that they are not needed to be sampled as they are to be fetched from the randomness beacon
func reducePermutationIntoGrandProduct(comp *wizard.CompiledIOP, q query.Permutation, round int) {
	var (
		isMultiColumn = len(q.A[0]) > 1
		alpha         coin.Info
		// beta has to be different for different for different queries for the soundness of z-packing
		beta          = comp.InsertCoin(round+1, permutation.DeriveName[coin.Name](q, "BETA"), coin.Field)
	)

	if isMultiColumn {
		alpha = comp.InsertCoin(round+1, permutation.DeriveName[coin.Name](q, "ALPHA"), coin.Field)
	}

	// Reduce a permutation query into a GrandProduct query
	comp.InsertGrandProduct(round, q.Name(), alpha, beta)


}

package permutation

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// CompileIntoGdProduct scans comp, looking for [query.Permutation] queries
// and compile them into global [query.GrandProduct]. One for every-round.
func CompileIntoGdProduct(comp *wizard.CompiledIOP) {

	var (
		// zCatalog stores a mapping (round, size) into ZCtx and helps finding
		// which Z context should be used to handle a part of a given permutation
		// query.
		zCatalog = map[[2]int]*ZCtx{}
		numPub   = []*symbolic.Expression{}
		denPub   = []*symbolic.Expression{}
	)

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {

		// Filter out non permutation queries
		permutation, ok := comp.QueriesNoParams.Data(qName).(query.Permutation)
		if !ok {
			continue
		}

		// This ensures that the lookup query is not used again in the
		// compilation process. We know that the query was already ignored at
		// the beginning because we are iterating over the unignored keys.
		comp.QueriesNoParams.MarkAsIgnored(qName)
		round := comp.QueriesNoParams.Round(qName)

		_numPub, _denPub := dispatchPermutation(comp, zCatalog, round, permutation)
		numPub = append(numPub, _numPub...)
		denPub = append(denPub, _denPub...)
	}

	zCatalogKeys := utils.SortedKeysOf(zCatalog, func(a, b [2]int) bool {
		if a[0] != b[0] {
			return a[0] < b[0]
		}
		return a[1] < b[1]
	})

	queries := make([]*query.GrandProduct, comp.NumRounds()+1)

	for _, entry := range zCatalogKeys {

		var (
			zctx  = zCatalog[entry]
			round = entry[0]
			size  = entry[1]
		)

		if queries[round] == nil {
			queries[round] = &query.GrandProduct{
				Round:  round,
				ID:     ifaces.QueryIDf("GD_PRODUCT_%v_%v", comp.SelfRecursionCount, round),
				Inputs: make(map[int]*query.GrandProductInput),
			}
		}

		gdProductQuery := queries[round]

		if gdProductQuery.Inputs[size] == nil {
			gdProductQuery.Inputs[size] = &query.GrandProductInput{
				Size: size,
			}
		}

		gdProductInput := gdProductQuery.Inputs[size]

		gdProductInput.Numerators = append(gdProductInput.Numerators, zctx.NumeratorFactors...)
		gdProductInput.Denominators = append(gdProductInput.Denominators, zctx.DenominatorFactors...)
	}

	for _, query := range queries {

		if query == nil {
			continue
		}

		comp.InsertGrandProduct(query.Round, query.ID, query.Inputs)
		comp.RegisterProverAction(query.Round, &AssignPermutationGrandProduct{Query: query, IsPartial: len(numPub)+len(denPub) > 0})
		comp.RegisterVerifierAction(query.Round, &CheckGrandProductIsOne{Query: query, ExplicitNum: numPub, ExplicitDen: denPub})
	}
}

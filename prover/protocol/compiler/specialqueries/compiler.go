package specialqueries

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

/*
Reduce the fixed permutations into
*/
func CompileFixedPermutations(comp *wizard.CompiledIOP) {

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
			case query.FixedPermutation:
				reduceFixedPermutation(comp, q_, i)
			}
		}
	}
}

/*
Derive a name for a a coin created during the compilation process
*/
func deriveName[R ~string](context string, q ifaces.QueryID, name string) R {
	var res string
	if q == "" {
		res = fmt.Sprintf("%v_%v", context, name)
	} else {
		res = fmt.Sprintf("%v_%v_%v", q, context, name)
	}
	return R(res)
}

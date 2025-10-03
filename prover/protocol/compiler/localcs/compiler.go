package localcs

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

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

/*
Compiles the local constraints
*/
func Compile[T zk.Element](comp *wizard.CompiledIOP[T]) {

	compileOpeningsToConstraints(comp)

	numRounds := comp.NumRounds()

	/*
		First compile all local constraints
	*/
	for i := 0; i < numRounds; i++ {
		queries := comp.QueriesNoParams.AllKeysAt(i)

		for _, qName := range queries {

			// Skip if it was already compiled
			if comp.QueriesNoParams.IsIgnored(qName) {
				continue
			}

			if q_, ok := comp.QueriesNoParams.Data(qName).(query.LocalConstraint[T]); ok {
				ReduceLocalConstraint(comp, q_, i)
			}
		}
	}
}

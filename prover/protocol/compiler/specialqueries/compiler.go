package specialqueries

import (
	"fmt"
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"
	"github.com/sirupsen/logrus"
)

/*
Reduces the special queries of the wizard
*/
func CompilePermutations(comp *wizard.CompiledIOP) {

	numRounds := comp.NumRounds()
	meta := &metaCtx{provers: collection.NewVecVec[wizard.ProverStep](numRounds)}

	/*
		Handles the permutation checks
	*/
	for i := 0; i < numRounds; i++ {
		queries := comp.QueriesNoParams.AllKeysAt(i)
		for _, qName := range queries {
			// Skip if it was already compiled
			if comp.QueriesNoParams.IsIgnored(qName) {
				if _, ok := comp.QueriesNoParams.Data(qName).(query.Permutation); ok {
					logrus.Debugf("Permutation query %v is ignored", qName)
				}
				continue
			}

			switch q_ := comp.QueriesNoParams.Data(qName).(type) {
			case query.Permutation:
				reducePermutation(comp, meta, q_, i)
			}
		}
	}

	/*
		Registers the provers context
	*/
	for i := 0; i < meta.provers.Len(); i++ {
		comp.SubProvers.AppendToInner(i, meta.Prover(i))
	}
}

/*
Compilation context of all the inclusion queries. It is used to wrap all the
prover steps in parallel.
*/
type metaCtx struct {
	provers collection.VecVec[wizard.ProverStep]
}

func (m *metaCtx) Prover(round int) wizard.ProverStep {
	return func(run *wizard.ProverRuntime) {
		steps := m.provers.MustGet(round)
		stopTimer := profiling.LogTimer("meta-prover : %v steps to run for round %v", len(steps), round)
		parallel.ExecuteChunky(len(steps), func(start, stop int) {
			for i := start; i < stop; i++ {
				steps[i](run)
			}
		})
		stopTimer()
	}
}

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

/*
Derive a name for a a coin created during the compilation process
*/
func deriveNameWithCnt[R ~string](comp *wizard.CompiledIOP, extra ...any) R {
	extraFmt := make([]string, 0, len(extra)+1)
	extraFmt = append(extraFmt, fmt.Sprintf("%v", comp.SelfRecursionCount))
	for _, e := range extra {
		extraFmt = append(extraFmt, fmt.Sprintf("%v", e))
	}
	return R(strings.Join(extraFmt, "_"))
}

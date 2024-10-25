package innerproduct

import (
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Compile applies the inner-product compilation pass over `comp` it marks all
// the inner-product queries as `Ignored` and adds protocol items to justify
// these compiled inner-products.
//
// The inner-product queries are processed in groups relating to column of the
// same size.
func Compile(comp *wizard.CompiledIOP) {

	var (
		// round stores the latest definition round for all the unignored
		// compilation queries. If `round` is left as -1 at the end of this
		// query capture phase, we conclude that no-inner product queries were
		// found.
		round = -1
		// queryMap organizes all the encountered queries by column size. These
		// form groups that are independently compiled.
		queryMap = map[int][]query.InnerProduct{}
		// sizes lists all the column sizes that have been encountered by
		// the compiler in chronological order. This is necessary this let us
		// iterate over theses in deterministic order. If we only had `queryMap`
		// such a thing would be impossible. We stress that deterministic order
		// iteration is essential as it ensures that compiler yield exactly the
		// same protocol if the same step of compilation are applied in the same
		// order.
		sizes = []int{}
		// contextsForSize list all the sub-compilation context
		// in the same order as `sizes`.
		// proverTaskCollaps indicates when we have more than one pair of inner-product with the same size
		// and thus collapsing all pairs to a single column is required.
		proverTaskNoCollaps, proverTaskCollpas proverTask
	)

	for _, qName := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(qName).(query.InnerProduct)
		if !ok {
			// not an inner-product query, ignore
			continue
		}

		comp.QueriesParams.MarkAsIgnored(qName)
		round = utils.Max(round, comp.QueriesParams.Round(qName))
		size := q.A.Size()

		if _, ok := queryMap[size]; !ok {
			sizes = append(sizes, size)
		}

		queryMap[size] = append(queryMap[size], q)
	}

	if round < 0 {
		// We return because we found out that there were no queries to compile
		return
	}

	for _, size := range sizes {
		ctx := compileForSize(comp, round, queryMap[size])
		switch ctx.round {
		case round:
			proverTaskNoCollaps = append(proverTaskNoCollaps, ctx)
		case round + 1:
			proverTaskCollpas = append(proverTaskCollpas, ctx)
		default:
			utils.Panic("round before compilation was  %v and after compilation %v", round, ctx.round)
		}

	}
	// run the prover of the relevant round
	if len(proverTaskNoCollaps) >= 1 {
		comp.RegisterProverAction(round, proverTaskNoCollaps)
	}

	if len(proverTaskCollpas) >= 1 {
		comp.RegisterProverAction(round+1, proverTaskCollpas)
	}

}

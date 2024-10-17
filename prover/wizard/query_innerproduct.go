package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// Represent a batch of inner-product <a, b0>, <a, b1>, <a, b2> ...
type QueryInnerProduct struct {
	A        Column
	B        Column
	metadata *metadata
	*subQuery
}

// Constructor for inner-product.
// The list of polynomial Bs must be deduplicated.
func (api *API) NewQueryInnerProduct(a Column, b Column) *QueryInnerProduct {

	if a.Size() != b.Size() {
		utils.Panic(
			"a and b do not have the same size (%v->%v) (%v->%v)",
			a.String(), a.Size(), b.String(), b.Size(),
		)
	}

	var (
		round = max(a.Round(), b.Round())
		res   = &QueryInnerProduct{
			A:        a,
			B:        b,
			metadata: api.newMetadata(),
			subQuery: &subQuery{
				round: round,
			},
		}
	)

	api.queries.addToRound(round, res)
	return res
}

func (r QueryInnerProduct) computeResult(run Runtime) QueryResult {
	var (
		a = r.A.GetAssignment(run)
		b = r.B.GetAssignment(run)
	)
	return &QueryResFE{R: smartvectors.InnerProduct(a, b)}
}

func (r QueryInnerProduct) computeResultGnark(api frontend.API, run RuntimeGnark) QueryResultGnark {
	var (
		a   = r.A.GetAssignmentGnark(api, run)
		b   = r.B.GetAssignmentGnark(api, run)
		res = frontend.Variable(0)
	)

	for i := range a {
		res = api.Add(res, api.Mul(a[i], b[i]))
	}
	return &QueryResFEGnark{R: res}
}

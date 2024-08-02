package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/fft/fastpoly"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// Multiple polynomials, one point
type QueryUnivariateEval struct {
	Pol      Column
	X        Accessor
	metadata *metadata
	*subQuery
}

// NewQueryUnivariateEval Constructor for univariate evaluation queries
// The list of polynomial must be deduplicated.
func (api *API) NewQueryUnivariateEval(pol Column, x Accessor) *QueryUnivariateEval {
	var (
		round = x.Round()
		res   = &QueryUnivariateEval{
			Pol:      pol,
			X:        x,
			metadata: api.newMetadata(),
			subQuery: &subQuery{
				round: round,
			},
		}
	)

	api.queries.addToRound(round, res)
	return res
}

// Test that the polynomial evaluation holds
func (r QueryUnivariateEval) ComputeResult(run Runtime) QueryResult {

	var (
		pol = r.Pol.GetAssignment(run)
		x   = r.X.GetVal(run)
		y   = smartvectors.Interpolate(pol, x)
	)

	return &QueryResFE{
		R: y,
	}
}

// Test that the polynomial evaluation holds
func (r QueryUnivariateEval) ComputeResultGnark(api frontend.API, run GnarkRuntime) QueryResultGnark {

	var (
		pol = r.Pol.GetAssignmentGnark(api, run)
		x   = r.X.GetValGnark(api, run)
		y   = fastpoly.InterpolateGnark(api, pol, x)
	)

	return &QueryResFEGnark{
		R: y,
	}
}

// BatchComputeUnivariateEval is a batching utility for QueryUnivariateEval
// which allows computing the result of many queries at the same point at the
// same time. This is more efficient than evaluating each query separately.
//
// The call will also register the result of the query. All qs should be using
// the same accessor otherwise the call will panic.
func BatchComputeUnivariateEval(run *RuntimeProver, qs []QueryUnivariateEval) []QueryResult {

	xString := qs[0].X.String()
	for i := range qs {
		if qs[i].X.String() != xString {
			utils.Panic("expects all the queries to have the same X point")
		}
	}

	var (
		pols = make([]smartvectors.SmartVector, len(qs))
		x    = qs[0].X.GetVal(run)
		res  = make([]QueryResult, len(qs))
	)

	for i := range qs {
		pols[i] = qs[i].Pol.GetAssignment(run)
	}

	ys := smartvectors.BatchInterpolate(pols, x)

	for i := range qs {
		res[i] = &QueryResFE{
			R: ys[i],
		}
		run.queryRes.InsertNew(qs[i].id(), res[i])
	}

	return res
}

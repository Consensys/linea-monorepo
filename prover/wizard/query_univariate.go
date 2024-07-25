package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/fft/fastpoly"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// Multiple polynomials, one point
type QueryUnivariateEval struct {
	Pols     []Column
	X        Accessor
	metadata *metadata
	*subQuery
}

// NewQueryUnivariateEval Constructor for univariate evaluation queries
// The list of polynomial must be deduplicated.
func (api *API) NewQueryUnivariateEval(pols []Column, x Accessor) *QueryUnivariateEval {
	var (
		round  = x.Round()
		_, err = utils.AllReturnEqual(Column.Size, pols)
		res    = &QueryUnivariateEval{
			Pols:     pols,
			X:        x,
			metadata: api.newMetadata(),
			subQuery: &subQuery{
				round: round,
			},
		}
	)

	if err != nil {
		utils.Panic("all columns should have the same size: %v", err)
	}

	api.queries.addToRound(round, res)
	return res
}

// Test that the polynomial evaluation holds
func (r QueryUnivariateEval) ComputeResult(run Runtime) QueryResult {

	var (
		pols = make([]smartvectors.SmartVector, len(r.Pols))
		x    = r.X.GetVal(run)
	)

	for i := range r.Pols {
		pols[i] = r.Pols[i].GetAssignment(run)
	}

	return &QueryResFESlice{
		R: smartvectors.BatchInterpolate(pols, x),
	}
}

// Test that the polynomial evaluation holds
func (r QueryUnivariateEval) ComputeResultGnark(api frontend.API, run GnarkRuntime) QueryResultGnark {

	var (
		ys = make([]frontend.Variable, len(r.Pols))
		x  = r.X.GetValGnark(api, run)
	)

	for i := range r.Pols {
		pol := r.Pols[i].GetAssignmentGnark(api, run)
		ys[i] = fastpoly.InterpolateGnark(api, pol, x)
	}

	return &QueryResFESliceGnark{
		R: ys,
	}
}

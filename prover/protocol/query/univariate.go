package query

import (
	"errors"
	"fmt"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectorsext"
	"github.com/consensys/linea-monorepo/prover/maths/fft/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/fft/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// Multiple polynomials, one point
type UnivariateEval struct {
	Pols    []ifaces.Column
	QueryID ifaces.QueryID
}

// Parameters for an univariate evaluation
type UnivariateEvalParams struct {
	X  interface{}
	Ys []interface{}
}

/*
Constructor for univariate evaluation queries
The list of polynomial must be deduplicated.
*/
func NewUnivariateEval(id ifaces.QueryID, pols ...ifaces.Column) UnivariateEval {
	// Panics if there is a duplicate
	polsSet := collection.NewSet[ifaces.ColID]()

	if len(pols) == 0 {
		utils.Panic("Univariate eval declared with zero polynomials")
	}

	for _, pol := range pols {

		if len(pol.GetColID()) == 0 {
			utils.Panic("Assigned a polynomial ifaces.QueryID with an empty length")
		}

		if polsSet.Insert(pol.GetColID()) {
			utils.Panic("(query %v) Got a duplicate entry %v in %v\n", id, pol, pols)
		}
	}

	return UnivariateEval{QueryID: id, Pols: pols}
}

// Name implements the [ifaces.Query] interface
func (r UnivariateEval) Name() ifaces.QueryID {
	return r.QueryID
}

// Constructor for non-fixed point univariate evaluation query parameters
func NewUnivariateEvalParams(x interface{}, ys ...interface{}) UnivariateEvalParams {
	return UnivariateEvalParams{
		X:  x,
		Ys: ys,
	}
}

// Update the fiat-shamir state with the alleged evaluations. We assume that
// the verifer always computes the values of X upfront on his own. Therefore
// there is no need to include them in the FS.
func (p UnivariateEvalParams) UpdateFS(state *fiatshamir.State) {
	state.UpdateMixed(p.Ys...)
}

// Test that the polynomial evaluation holds
func (r UnivariateEval) Check(run ifaces.Runtime) error {
	params := run.GetParams(r.QueryID).(UnivariateEvalParams)

	errMsg := "univariate query check failed\n"
	anyErr := false

	// check whether X is a base field element
	baseX, isBaseX := params.X.(field.Element)

	for k, pol := range r.Pols {
		wit := pol.GetColAssignment(run)

		// check whether we are dealing with base field elements
		if _, isBaseErr := wit.GetBase(0); isBaseErr == nil && isBaseX {
			// the smartvector is composed of base field elements and X is also a base field element
			actualY := smartvectors.Interpolate(wit, baseX)
			expectedY := params.Ys[k].(field.Element)

			if actualY != expectedY {
				anyErr = true
				errMsg += fmt.Sprintf("expected P(x) = %s but got %s for %v\n", expectedY.String(), actualY.String(), pol.GetColID())
			}
		} else {
			// we are dealing with extension field elements
			var processedX fext.Element
			if isBaseX {
				processedX = fext.NewFromBase(baseX)
			} else {
				processedX = params.X.(fext.Element)
			}
			// now, check that the Y matches the expected one
			actualY := smartvectorsext.Interpolate(wit, processedX)
			expectedY := params.Ys[k].(fext.Element)

			if actualY != expectedY {
				anyErr = true
				errMsg += fmt.Sprintf("expected P(x) = %s but got %s for %v\n", expectedY.String(), actualY.String(), pol.GetColID())
			}

		}

	}

	if anyErr {
		return errors.New(errMsg)
	}

	return nil
}

// Test that the polynomial evaluation holds
func (r UnivariateEval) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	params := run.GetParams(r.QueryID).(GnarkUnivariateEvalParams)

	if params.IsBase {
		for k, pol := range r.Pols {
			wit := pol.GetColAssignmentGnark(run)
			actualY := fastpoly.InterpolateGnark(api, wit, params.BaseX)
			api.AssertIsEqual(actualY, params.BaseYs[k])
		}
	} else {
		outerApi := gnarkfext.API{Inner: api}
		for k, pol := range r.Pols {
			wit := pol.GetColAssignmentGnarkExt(run)
			actualY := fastpolyext.InterpolateGnark(outerApi, wit, params.ExtX)
			outerApi.AssertIsEqual(actualY, params.ExtYs[k])
		}
	}
}

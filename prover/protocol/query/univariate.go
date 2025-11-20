package query

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark/frontend"
	fiatshamir "github.com/consensys/linea-monorepo/prover/crypto/fiatshamir_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/google/uuid"
)

// Multiple polynomials, one point
type UnivariateEval struct {
	Pols    []ifaces.Column
	QueryID ifaces.QueryID
	uuid    uuid.UUID `serde:"omit"`
}

// Parameters for an univariate evaluation
type UnivariateEvalParams struct {
	X      field.Element
	Ys     []field.Element
	ExtX   fext.Element
	ExtYs  []fext.Element
	IsBase bool
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

	return UnivariateEval{QueryID: id, Pols: pols, uuid: uuid.New()}
}

// Name implements the [ifaces.Query] interface
func (r UnivariateEval) Name() ifaces.QueryID {
	return r.QueryID
}

// Constructor for non-fixed point univariate evaluation query parameters
func NewUnivariateEvalParams(x field.Element, ys ...field.Element) UnivariateEvalParams {
	return UnivariateEvalParams{
		X:      x,
		Ys:     ys,
		IsBase: true,
	}
}
func NewUnivariateEvalParamsExt(x fext.Element, ys ...fext.Element) UnivariateEvalParams {
	return UnivariateEvalParams{
		ExtX:   x,
		ExtYs:  ys,
		IsBase: false,
	}
}

// Update the fiat-shamir state with the alleged evaluations. We assume that
// the verifer always computes the values of X upfront on his own. Therefore
// there is no need to include them in the FS.
func (p UnivariateEvalParams) UpdateFS(state *poseidon2_koalabear.MDHasher) {
	fiatshamir.Update(state, p.Ys...)
}

func (p UnivariateEvalParams) UpdateFSExt(state *poseidon2_koalabear.MDHasher) {
	fiatshamir.UpdateExt(state, p.ExtYs...)
}

// Test that the polynomial evaluation holds
func (r UnivariateEval) Check(run ifaces.Runtime) error {
	params := run.GetParams(r.QueryID).(UnivariateEvalParams)
	errMsg := "univariate query check failed\n"
	anyErr := false

	for k, pol := range r.Pols {
		wit := pol.GetColAssignment(run)
		actualY := smartvectors.EvaluateFextPolyLagrange(wit, params.ExtX)
		if actualY != params.ExtYs[k] {
			anyErr = true
			errMsg += fmt.Sprintf("expected P(x) = %s but got %s for %v\n", params.ExtYs[k].String(), actualY.String(), pol.GetColID())
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

	ext4, _ := gnarkfext.NewExt4(api)
	for k, pol := range r.Pols {
		wit := pol.GetColAssignmentGnark(run)
		actualY := fastpoly.EvaluateLagrangeGnarkMixed(api, wit, params.ExtX)
		ext4.AssertIsEqual(&actualY, &params.ExtYs[k])
	}
}

func (r UnivariateEval) UUID() uuid.UUID {
	return r.uuid
}

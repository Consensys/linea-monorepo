package query

import (
	"fmt"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Queries the opening of a handle at zero
type LocalOpening struct {
	Pol ifaces.Column
	ID  ifaces.QueryID
}

// Contains the result of a local opening
type LocalOpeningParams struct {
	BaseY  field.Element
	ExtY   fext.Element
	IsBase bool
}

// Updates a Fiat-Shamir state
func (lop LocalOpeningParams) UpdateFS(fs *fiatshamir.State) {
	if lop.IsBase {
		fs.Update(lop.BaseY)
	} else {
		// Change this for the actual extension!
		fs.Update(lop.ExtY.A0)
		fs.Update(lop.ExtY.A1)
	}
}

// Constructs a new local opening query
func NewLocalOpening(id ifaces.QueryID, pol ifaces.Column) LocalOpening {

	if len(pol.GetColID()) == 0 {
		utils.Panic("Assigned a polynomial name with an empty length")
	}

	return LocalOpening{ID: id, Pol: pol}
}

// Name implements the [ifaces.Query] interface
func (r LocalOpening) Name() ifaces.QueryID {
	return r.ID
}

// Constructor for non-fixed point univariate evaluation query parameters
func NewLocalOpeningParams(y field.Element) LocalOpeningParams {
	return LocalOpeningParams{
		BaseY:  y,
		ExtY:   fext.Element{A0: y, A1: field.Zero()},
		IsBase: true,
	}
}

func NewLocalOpeningParamsExt(z fext.Element) LocalOpeningParams {
	return LocalOpeningParams{
		BaseY:  field.Zero(),
		ExtY:   z,
		IsBase: false,
	}
}

// Test that the polynomial evaluation holds
func (r LocalOpening) Check(run ifaces.Runtime) error {
	params := run.GetParams(r.ID).(LocalOpeningParams)
	if params.IsBase {
		actualY := r.Pol.GetColAssignmentAt(run, 0)
		if actualY != params.BaseY {
			return fmt.Errorf("expected P(x) = %s but got %s for %v", params.BaseY.String(), actualY.String(), r.Pol.GetColID())
		}
	} else {
		actualY := r.Pol.GetColAssignmentAtExt(run, 0)
		if actualY != params.ExtY {
			return fmt.Errorf("expected P(x) = %s but got %s for %v", params.ExtY.String(), actualY.String(), r.Pol.GetColID())
		}
	}

	return nil
}

// Test that the polynomial evaluation holds
func (r LocalOpening) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	params := run.GetParams(r.ID).(GnarkLocalOpeningParams)
	if params.isBase {
		actualY := r.Pol.GetColAssignmentGnarkAt(run, 0)
		api.AssertIsEqual(params.BaseY, actualY)
	} else {
		apiExt := gnarkfext.API{api}
		actualY := r.Pol.GetColAssignmentGnarkAtExt(run, 0)
		apiExt.AssertIsEqual(params.ExtY, actualY)
	}

}

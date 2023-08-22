package query

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/fiatshamir"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

// Queries the opening of a handle at zero
type LocalOpening struct {
	Pol ifaces.Column
	ID  ifaces.QueryID
}

// Contains the result of a local opening
type LocalOpeningParams struct {
	Y field.Element
}

// Updates a Fiat-Shamir state
func (lop LocalOpeningParams) UpdateFS(fs *fiatshamir.State) {
	fs.Update(lop.Y)
}

// Constructs a new local opening query
func NewLocalOpening(id ifaces.QueryID, pol ifaces.Column) LocalOpening {

	// For simplicity, we enforce the `pol` to be either Natural or Shifted(Natural)
	// Allegedly, this does not block any-case
	switch h := pol.(type) {
	case column.Natural:
		// allowed
	case column.Shifted:
		if _, ok := h.Parent.(column.Natural); !ok {
			utils.Panic("Unsupported handle should only be a shifted or a natural %v", pol)
		}
		// allowed
	default:
		utils.Panic("Unsupported handle should only be a shifted %v", pol)
	}

	if len(pol.GetColID()) == 0 {
		utils.Panic("Assigned a polynomial name with an empty length")
	}

	return LocalOpening{ID: id, Pol: pol}
}

// Constructor for non-fixed point univariate evaluation query parameters
func NewLocalOpeningParams(y field.Element) LocalOpeningParams {
	return LocalOpeningParams{Y: y}
}

// Test that the polynomial evaluation holds
func (r LocalOpening) Check(run ifaces.Runtime) error {
	params := run.GetParams(r.ID).(LocalOpeningParams)
	actualY := r.Pol.GetColAssignmentAt(run, 0)

	if actualY != params.Y {
		return fmt.Errorf("expected P(x) = %s but got %s for %v", params.Y.String(), actualY.String(), r.Pol.GetColID())
	}

	return nil
}

// Test that the polynomial evaluation holds
func (r LocalOpening) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	params := run.GetParams(r.ID).(GnarkLocalOpeningParams)
	actualY := r.Pol.GetColAssignmentGnarkAt(run, 0)
	api.AssertIsEqual(params.Y, actualY)
}

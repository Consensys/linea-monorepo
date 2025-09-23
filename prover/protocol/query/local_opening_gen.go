package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/google/uuid"
)

// Queries the opening of a handle at zero
type LocalOpeningGen[T zk.Element] struct {
	Pol  ifaces.ColumnGen[T]
	ID   ifaces.QueryID
	uuid uuid.UUID `serde:"omit"`
}

// Contains the result of a local opening
// type LocalOpeningParams struct {
// 	BaseY  field.Element
// 	ExtY   fext.Element
// 	IsBase bool
// }

// // Updates a Fiat-Shamir state
// func (lop LocalOpeningParams) UpdateFS(fs hash.StateStorer) {
// 	if lop.IsBase {
// 		fiatshamir.Update(fs, lop.BaseY)
// 	} else {
// 		// Change this for the actual extension!
// 		fiatshamir.UpdateExt(fs, lop.ExtY)
// 	}
// }

// Constructs a new local opening query
func NewLocalOpeningGen[T zk.Element](id ifaces.QueryID, pol ifaces.ColumnGen[T]) LocalOpeningGen[T] {

	if len(pol.GetColID()) == 0 {
		utils.Panic("Assigned a polynomial name with an empty length")
	}

	return LocalOpeningGen[T]{ID: id, Pol: pol, uuid: uuid.New()}
}

// Name implements the [ifaces.Query] interface
func (r LocalOpeningGen[T]) Name() ifaces.QueryID {
	return r.ID
}

// IsBase returns if the column is a base-field column
func (r LocalOpeningGen[T]) IsBase() bool {
	return r.Pol.IsBase()
}

// Constructor for [LocalOpeningParams] when y is a base field element.
// func NewLocalOpeningParams(y field.Element) LocalOpeningParams {
// 	return LocalOpeningParams{
// 		BaseY:  y,
// 		ExtY:   fext.Lift(y),
// 		IsBase: true,
// 	}
// }

// Constructor for [LocalOpeningParams] when y is a base field element.
// func NewLocalOpeningParamsExt(z fext.Element) LocalOpeningParams {
// 	return LocalOpeningParams{
// 		ExtY:   z,
// 		IsBase: false,
// 	}
// }

// func (lop LocalOpeningParams) ToGenericGroupElement() fext.GenericFieldElem {
// 	if lop.IsBase {
// 		return fext.NewESHashFromBase(lop.BaseY)
// 	} else {
// 		return fext.NewESHashFromExt(lop.ExtY)
// 	}
// }

// Test that the polynomial evaluation holds
func (r LocalOpeningGen[T]) Check(run ifaces.Runtime) error {
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
func (r LocalOpeningGen[T]) CheckGnark(api frontend.API, run ifaces.GnarkRuntimeGen[T]) {

	e4Ext, err := gnarkfext.NewExt4[T](api)
	if err != nil {
		panic(err)
	}

	params := run.GetParams(r.ID).(GnarkLocalOpeningParamsGen[T])
	if params.IsBase {
		actualY := r.Pol.GetColAssignmentGnarkAt(run, 0)
		api.AssertIsEqual(params.BaseY, actualY)
	} else {
		actualY := r.Pol.GetColAssignmentGnarkAtExt(run, 0)

		// params.ExtY.AssertIsEqual(api, actualY)
		e4Ext.AssertIsEqual(&actualY, &params.ExtY)
	}
}

func (r LocalOpeningGen[T]) UUID() uuid.UUID {
	return r.uuid
}

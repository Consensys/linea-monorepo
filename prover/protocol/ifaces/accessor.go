package ifaces

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/gnark/frontend"
)

// Function that can be used to retrieve the result of a
// functional wizard. It also satisfy the the `symbolic.Variable`
// interface. It can be used to generalize coins or any single-valued
// value known by the verifier (coin, Params, message)
type Accessor struct {
	Name                string
	GetVal              func(run Runtime) field.Element
	GetFrontendVariable func(api frontend.API, c GnarkRuntime) frontend.Variable
	Round               int
}

// Construct a new accessor
func NewAccessor(
	name string,
	getVal func(run Runtime) field.Element,
	getFrontendVariable func(api frontend.API, c GnarkRuntime) frontend.Variable,
	round int,
) *Accessor {
	return &Accessor{
		Name:                name,
		GetVal:              getVal,
		GetFrontendVariable: getFrontendVariable,
		Round:               round,
	}
}

// We importantly pass a ptr because we want DeepEqual to work.
// If we pass it to copies of the accessor, it will not work
// because the accessor contains function pointers. But if we pass
// a pointer, DeepEqual will detect that the two instances are
// dereferencing to the same place.
func (acc *Accessor) String() string {
	return fmt.Sprintf("ACCESSOR_%v", acc.Name)
}

// We importantly pass a ptr because we want DeepEqual to work.
// If we pass it to copies of the accessor, it will not work
// because the accessor contains function pointers. But if we pass
// a pointer, DeepEqual will detect that the two instances are
// dereferencing to the same place.
func (acc *Accessor) AsVariable() *symbolic.Expression {
	return symbolic.NewVariable(acc)
}

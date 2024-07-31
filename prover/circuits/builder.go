package circuits

import "github.com/consensys/gnark/constraint"

type Builder interface {
	Compile() (constraint.ConstraintSystem, error)
}

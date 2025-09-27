package symbolic

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
)

// Constant is an implementation of [Operator] which represents a constant value
type Constant[T zk.Element] struct {
	Val fext.GenericFieldElem
}

// Degree implements the [Operator] interface
func (Constant[T]) Degree([]int) int {
	panic("we never call it for a constant")
}

// Evaluates implements the [Operator] interface
func (c Constant[T]) Evaluate([]sv.SmartVector) sv.SmartVector {
	panic("we never call it for a constant")
}

func (c Constant[T]) EvaluateExt([]sv.SmartVector) sv.SmartVector {
	panic("we never call EvaluateExt for a constant")
}

func (c Constant[T]) EvaluateMixed([]sv.SmartVector) sv.SmartVector {
	panic("we never call EvaluateMixed for a constant")
}

// GnarkEval implements the [Operator] interface.
func (c Constant[T]) GnarkEval(api frontend.API, inputs []T) T {
	panic("we never call it for a constant")
}

// GnarkEvalExt implements the [Operator] interface.
func (c Constant[T]) GnarkEvalExt(api frontend.API, inputs []gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T] {
	panic("we never call it for a constant")
}

// NewConstant creates a new [Constant]. The function admits any input types
// that is either: field.Element, int, uint or decimal string.
func NewConstant[T zk.Element](val interface{}) *Expression[T] {
	var x fext.Element
	if _, err := fext.SetInterface(&x, val); err != nil {
		panic(err)
	}

	newHash := fext.NewMinimalESHashFromExt(&x)
	//Create the expression
	res := &Expression[T]{
		Operator: Constant[T]{Val: *newHash},
		Children: []*Expression[T]{},
		ESHash:   *new(fext.GenericFieldElem).Set(newHash),
		IsBase:   fext.IsBase(&x),
	}
	return res
}

/*
Validate implements the [Operator] interface.
*/
func (c Constant[T]) Validate(expr *Expression[T]) error {
	if !reflect.DeepEqual(c, expr.Operator) {
		panic("expr.operator != c")
	}

	if len(expr.Children) != 0 {
		return fmt.Errorf("mismatch in the size of the children and coefficients")
	}

	return nil
}

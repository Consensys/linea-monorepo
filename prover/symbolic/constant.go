package symbolic

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
)

// Constant is an implementation of [Operator] which represents a constant value
type Constant struct {
	Val fext.Element
}

// Degree implements the [Operator] interface
func (Constant) Degree([]int) int {
	panic("we never call it for a constant")
}

// Evaluates implements the [Operator] interface
func (c Constant) Evaluate([]sv.SmartVector, ...mempool.MemPool) sv.SmartVector {
	panic("we never call it for a constant")
}

func (c Constant) EvaluateExt([]sv.SmartVector, ...mempool.MemPool) sv.SmartVector {
	panic("we never call EvaluateExt for a constant")
}

// GnarkEval implements the [Operator] interface.
func (c Constant) GnarkEval(api frontend.API, inputs []frontend.Variable) frontend.Variable {
	panic("we never call it for a constant")
}

// NewConstant creates a new [Constant]. The function admits any input types
// that is either: field.Element, int, uint or decimal string.
func NewConstant(val interface{}) *Expression {
	var x fext.Element
	if _, err := x.SetInterface(val); err != nil {
		panic(err)
	}

	// isBase is true if the value is a field.Element, otherwise it is false
	_, isBase := val.(field.Element)

	res := &Expression{
		Operator: Constant{Val: x},
		Children: []*Expression{},
		ESHash:   x,
		isBase:   isBase,
	}

	// Pass the string and not the field.Element itself
	res.ESHash.Set(&x)
	return res
}

/*
Validate implements the [Operator] interface.
*/
func (c Constant) Validate(expr *Expression) error {
	if !reflect.DeepEqual(c, expr.Operator) {
		panic("expr.operator != c")
	}

	if len(expr.Children) != 0 {
		return fmt.Errorf("mismatch in the size of the children and coefficients")
	}

	return nil
}

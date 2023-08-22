package symbolic

import (
	"fmt"
	"math/big"
	"reflect"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark/frontend"
)

/*
Constant operator
*/
type Constant struct {
	Val field.Element
}

func (Constant) Degree([]int) int {
	panic("we never call it for a constant")
}

func (c Constant) Evaluate([]sv.SmartVector) sv.SmartVector {
	panic("we never call it for a constant")
}

func (c Constant) GnarkEval(api frontend.API, inputs []frontend.Variable) frontend.Variable {
	panic("we never call it for a constant")
}

// Creates a new constant
func NewConstant(val interface{}) *Expression {
	var x field.Element
	if _, err := x.SetInterface(val); err != nil {
		panic(err)
	}

	res := &Expression{
		Operator: Constant{Val: x},
		Children: []*Expression{},
		ESHash:   x,
	}

	// Pass the string and not the field.Element itself
	var sig big.Int
	res.ESHash.SetBigInt(x.BigInt(&sig))
	return res
}

/*
Validates that the constant is well-formed
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

package symbolic

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils"
)

/*
	Dedicated type of expression for evaluating a polynomial
	whose coefficients are sub-expressions. It is use for batching
	multiple evaluation constraints into a single check using the
	Schwartz-Zippel lemma
*/

type PolyEval[T zk.Element] struct{}

func NewPolyEval[T zk.Element](x *Expression[T], coeffs []*Expression[T]) *Expression[T] {

	/*
		No coeff is unexpected. If this panics on a legit use-case,
		feel free to return the constant zero instead.
	*/
	if len(coeffs) == 0 {
		utils.Panic("Polynomial with no coeffs")
	}

	/*
		Only one coeff. In this case no need to use `x`. We just return
		the original expression
	*/
	if len(coeffs) == 1 {
		return coeffs[0]
	}

	eshashes := []fext.GenericFieldElem{}
	for i := range coeffs {
		eshashes = append(eshashes, coeffs[i].ESHash)
	}

	esh := polyext.EvalUnivariateMixed(eshashes, x.ESHash)
	children := append([]*Expression[T]{x}, coeffs...)

	return &Expression[T]{
		Operator: PolyEval[T]{},
		Children: children,
		ESHash:   esh,
	}
}

/*
Returns the degree of the operation given, as input, the degree of the children
*/
func (PolyEval[T]) Degree(inputDegrees []int) int {
	return utils.Max(inputDegrees...)
}

/*
Evaluates a polynomial evaluation
*/
func (PolyEval[T]) Evaluate(inputs []sv.SmartVector) sv.SmartVector {
	// We assume that the first element is always a scalar
	// Get the constant value. We use Get(0) to get the value, but any integer would
	// also work provided it is also in range. 0 ensures that.
	x := inputs[0].(*sv.Constant).Get(0)
	return sv.LinearCombination(inputs[1:], x)
}

/*
Validates that the LC is well-formed
*/
func (PolyEval[T]) Validate(expr *Expression[T]) error {
	if len(expr.Children) < 2 {
		return fmt.Errorf("poly eval of degree 0")
	}
	return nil
}

/*
Evaluate the expression in a gnark circuit
Does not support vector evaluation
*/
func (PolyEval[T]) GnarkEval(api frontend.API, inputs []T) T {
	/*
		We use the Horner method
	*/
	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	x := inputs[0]
	res := inputs[len(inputs)-1]

	for i := len(inputs) - 2; i >= 1; i-- {
		res = *apiGen.Mul(&res, &x)
		c := inputs[i]
		res = *apiGen.Add(&res, &c)
	}

	return res
}

/*
EvaluateExt the expression in a gnark circuit
Does not support vector evaluation
*/
func (PolyEval[T]) GnarkEvalExt(api frontend.API, inputs []gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T] {
	/*
		We use the Horner method
	*/
	x := inputs[0]
	res := inputs[len(inputs)-1]

	e4Api, err := gnarkfext.NewExt4[T](api)
	if err != nil {

	}

	// outerApi := gnarkfext.NewExtApi(api)

	for i := len(inputs) - 2; i >= 1; i-- {
		res = *e4Api.Mul(&res, &x)
		c := inputs[i]
		res = *e4Api.Add(&res, &c)
	}

	return res
}

func (PolyEval[T]) EvaluateExt(inputs []sv.SmartVector) sv.SmartVector {
	// We assume that the first element is always a scalar
	// Get the constant value. We use Get(0) to get the value, but any integer would
	// also work provided it is also in range. 0 ensures that.
	x := inputs[0].(*sv.ConstantExt).GetExt(0) // to ensure we panic if the input is not a constant
	return sv.LinearCombinationExt(inputs[1:], x)
}

func (PolyEval[T]) EvaluateMixed(inputs []sv.SmartVector) sv.SmartVector {
	if sv.AreAllBase(inputs) {
		return PolyEval[T]{}.Evaluate(inputs)
	} else {
		return PolyEval[T]{}.EvaluateExt(inputs)
	}
}

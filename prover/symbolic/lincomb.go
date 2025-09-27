package symbolic

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors_mixed"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// LinComb is an [Operator] symbolizing linear combinations of expressions with
// constant coefficients. It is used under the hood to generically represents
// additions or expression of variables.
//
// For expression building it is advised to use the constructors instead of this
// directly. The exposition of this Operator is meant to allow implementing
// symbolic expression optimizations.
type LinComb[T zk.Element] struct {
	// The Coeffs are typically small integers (1, -1)
	Coeffs []int
}

// NewLinComb creates an expression representing the linear combination of a
// list of expressions and coefficients.
// the constructor operates a sequence of optimization routines: flattening,
// regroupment and removal of zeroes.
//
// If provided an empty list of items, the function returns the zero constant
// and if the number of items does not match the number of coeffs, the function
// panics.
//
// Note: the function is not guaranteed to return a LinComb object, since the
// optimization routine may detect that this simplifies into a single-term
// expression or a constant expression.
func NewLinComb[T zk.Element](items []*Expression[T], coeffs []int) *Expression[T] {

	if len(items) != len(coeffs) {
		panic("unmatching lengths")
	}

	coeffs, items = expandTerms[T](&LinComb[T]{}, coeffs, items)
	coeffs, items, constCoeffs, constVal := regroupTerms(coeffs, items)

	// This regroups all the constants into a global constant with a coefficient
	// of 1.
	var t fext.GenericFieldElem
	c := fext.GenericFieldZero()
	for i := range constCoeffs {
		t.SetInt64(int64(constCoeffs[i]))
		t.Mul(&constVal[i])
		c.Add(&t)
	}

	if !c.IsZero() {
		coeffs = append(coeffs, 1)
		items = append(items, NewConstant[T](c))
	}

	coeffs, items = removeZeroCoeffs(coeffs, items)

	if len(items) == 0 {
		return NewConstant[T](0)
	}

	// The LinCombExt is just a single-term: more efficient to unwrap it
	if len(items) == 1 && coeffs[0] == 1 {
		return items[0]
	}

	e := &Expression[T]{
		Operator: LinComb[T]{Coeffs: coeffs},
		Children: items,
		IsBase:   computeIsBaseFromChildren(items),
	}

	// Now we need to assign the ESH
	eshashes := make([]sv.SmartVector, len(e.Children))
	for i := range e.Children {
		eshashes[i] = sv.ToConstantSmartvector(&e.Children[i].ESHash, 1)
	}

	if len(items) > 0 {
		// The cast back to sv.Constant is not functionally important but is an easy
		// sanity check.
		evalResult := e.Operator.EvaluateMixed(eshashes)
		tmp := sv.GetGenericElemOfSmartvector(evalResult, 0)
		e.ESHash.Set(&tmp)
	}

	return e
}

// Degree implements the [Operator] interface and returns the maximum degree of
// the underlying expression.
func (LinComb[T]) Degree(inputDegrees []int) int {
	return utils.Max(inputDegrees...)
}

// Evaluate implements the [Operator] interface.
func (lc LinComb[T]) Evaluate(inputs []sv.SmartVector) sv.SmartVector {
	return sv.LinComb(lc.Coeffs, inputs)
}

func (lc LinComb[T]) EvaluateExt(inputs []sv.SmartVector) sv.SmartVector {
	return sv.LinCombExt(lc.Coeffs, inputs)
}

func (lc LinComb[T]) EvaluateMixed(inputs []sv.SmartVector) sv.SmartVector {
	return smartvectors_mixed.LinCombMixed(lc.Coeffs, inputs)
}

// Validate implements the [Operator] interface
func (lc LinComb[T]) Validate(expr *Expression[T]) error {
	if !reflect.DeepEqual(lc, expr.Operator) {
		panic("expr.operator != lc")
	}

	if len(lc.Coeffs) != len(expr.Children) {
		return fmt.Errorf("mismatch in the size of the children and coefficients")
	}

	return nil
}

// GnarkEval implements the [GnarkEval] interface
func (lc LinComb[T]) GnarkEval(api frontend.API, inputs []T) T {

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	// res := frontend.Variable(0)
	res := apiGen.ValueOf(0)

	// There should be as many inputs as there are coeffs
	if len(inputs) != len(lc.Coeffs) {
		utils.Panic("%v inputs but %v coeffs", len(inputs), len(lc.Coeffs))
	}

	/*
		Accumulate the scalars
	*/
	for i, input := range inputs {
		coeff := apiGen.ValueOf(lc.Coeffs[i])
		res = apiGen.Add(res, apiGen.Mul(coeff, &input))
	}

	return *res
}

// GnarkEval implements the [GnarkEvalExt] interface
func (lc LinComb[T]) GnarkEvalExt(api frontend.API, inputs []gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T] {

	e4Api, err := gnarkfext.NewExt4[T](api)
	if err != nil {
		panic(err)
	}

	res := e4Api.Zero()

	if len(inputs) != len(lc.Coeffs) {
		utils.Panic("%v inputs but %v coeffs", len(inputs), len(lc.Coeffs))
	}

	var tmp gnarkfext.E4Gen[T]
	for i, input := range inputs {
		coeff := e4Api.NewFromBase(lc.Coeffs[i])
		tmp = *e4Api.Mul(coeff, &input)
		res = e4Api.Add(res, &tmp)
	}

	return *res
}

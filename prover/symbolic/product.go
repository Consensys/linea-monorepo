package symbolic

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors_mixed"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var (
	ErrTermIsProduct    error = errors.New("term is product")
	ErrTermIsNotProduct error = errors.New("term is not product")
)

// Product is an implementation of the [Operator] interface and represents a
// product of terms with exponents.
//
// For expression building it is advised to use the constructors instead of this
// directly. The exposition of this Operator is meant to allow implementing
// symbolic expression optimizations.
//
// Note that the library does not support exponents with negative values and
// such expressions would be treated as invalid and yield panic errors if the
// package sees it.
type Product[T zk.Element] struct {
	// Exponents for each term in the multiplication
	Exponents []int
}

// NewProduct returns an expression representing a product of items applying
// exponents to them. The newly constructed expression is subjected to basic
// optimizations routines: detection of zero factors, expansion by
// associativity, regroupment of terms and removal or terms with a coefficient
// of zero.
//
// Thus, the returned expression is not guaranteed to be of type [Product]. To
// actually multiply or exponentiate [Expression] objects, the user is advised
// to use [Mul] [Square] or [Pow] instead.
//
// If provided an empty list of items/exponents the function returns 1 as a
// default value and if the lengths of the two parameters do not match, the
// function panics.
func NewProduct[T zk.Element](items []*Expression[T], exponents []int) *Expression[T] {

	if len(items) != len(exponents) {
		panic("unmatching lengths")
	}

	for i := range exponents {
		if exponents[i] < 0 {
			panic("negative exponents are not allowed")
		}
	}

	for i := range items {
		if items[i].ESHash.IsZero() && exponents[i] != 0 {
			return NewConstant[T](0)
		}
	}

	exponents, items = expandTerms[T](&Product[T]{}, exponents, items)
	exponents, items, constExponents, constVal := regroupTerms(exponents, items)

	// This regroups all the constants into a global constant with a coefficient
	// of 1.
	var t fext.GenericFieldElem
	c := fext.GenericFieldOne()
	for i := range constExponents {
		t.Exp(&constVal[i], big.NewInt(int64(constExponents[i])))
		c.Mul(&t)
	}

	if !c.IsOne() {
		exponents = append(exponents, 1)
		items = append(items, NewConstant[T](c))
	}

	exponents, items = removeZeroCoeffs(exponents, items)

	if len(items) == 0 {
		return NewConstant[T](1)
	}

	if len(items) == 1 && exponents[0] == 1 {
		return items[0]
	}

	e := &Expression[T]{
		Operator: Product[T]{Exponents: exponents},
		Children: items,
		ESHash:   fext.GenericFieldOne(),
		IsBase:   computeIsBaseFromChildren(items),
	}

	for i := range e.Children {
		tmp := fext.GenericFieldOne()
		switch {
		case exponents[i] == 1:
			e.ESHash.Mul(&e.Children[i].ESHash)
		case exponents[i] == 2:
			tmp.Square(&e.Children[i].ESHash)
			e.ESHash.Mul(&tmp)
		case exponents[i] == 3:
			tmp.Square(&e.Children[i].ESHash)
			tmp.Mul(&e.Children[i].ESHash)
			e.ESHash.Mul(&tmp)
		case exponents[i] == 4:
			tmp.Square(&e.Children[i].ESHash)
			tmp.Square(&tmp)
			e.ESHash.Mul(&tmp)
		default:
			exponent := big.NewInt(int64(exponents[i]))
			tmp.Exp(&e.Children[i].ESHash, exponent)
			e.ESHash.Mul(&tmp)
		}
	}

	return e
}

// Degree implements the [Operator] interface and returns the sum of the degree
// of all the operands weighted by the exponents.
func (prod Product[T]) Degree(inputDegrees []int) int {
	res := 0
	// Just the sum of all the degrees
	for i, exp := range prod.Exponents {
		res += exp * inputDegrees[i]
	}
	return res
}

// Evaluate implements the [Operator] interface.
func (prod Product[T]) Evaluate(inputs []sv.SmartVector) sv.SmartVector {
	return sv.Product(prod.Exponents, inputs)
}

// Validate implements the [Operator] interface.
func (prod Product[T]) Validate(expr *Expression[T]) error {
	if !reflect.DeepEqual(prod, expr.Operator) {
		panic("expr.operator != prod")
	}

	if len(prod.Exponents) != len(expr.Children) {
		return fmt.Errorf("mismatch in the size of the children and coefficients")
	}

	for _, e := range prod.Exponents {
		if e < 0 {
			panic("found a negative exponent")
		}
	}

	return nil
}

// GnarkEval implements the [Operator] interface.
func (prod Product[T]) GnarkEval(api frontend.API, inputs []T) T {

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	res := apiGen.ValueOf(1)

	// There should be as many inputs as there are coeffs
	if len(inputs) != len(prod.Exponents) {
		utils.Panic("%v inputs but %v coeffs", len(inputs), len(prod.Exponents))
	}

	/*
		Accumulate the scalars
	*/
	var bExp big.Int
	for i, input := range inputs {
		bExp.SetUint64(uint64(prod.Exponents[i]))
		term := field.Exp(apiGen, input, &bExp)
		res = apiGen.Mul(res, term)
	}

	return *res
}

// GnarkEval implements the [Operator] interface.
func (prod Product[T]) GnarkEvalExt(api frontend.API, inputs []gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T] {

	e4Api, err := gnarkfext.NewExt4[T](api)
	if err != nil {
		panic(err)
	}

	res := e4Api.One()

	// There should be as many inputs as there are coeffs
	if len(inputs) != len(prod.Exponents) {
		utils.Panic("%v inputs but %v coeffs", len(inputs), len(prod.Exponents))
	}

	// outerApi := gnarkfext.NewExtApi(api)
	/*
		Accumulate the scalars
	*/

	var bExp big.Int
	for i, input := range inputs {
		bExp.SetUint64(uint64(prod.Exponents[i]))
		term := e4Api.Exp(&input, &bExp)
		res = e4Api.Mul(res, term)
	}

	return *res
}

func (prod Product[T]) EvaluateExt(inputs []sv.SmartVector) sv.SmartVector {
	return sv.ProductExt(prod.Exponents, inputs)
}

func (prod Product[T]) EvaluateMixed(inputs []sv.SmartVector) sv.SmartVector {
	return smartvectors_mixed.ProductMixed(prod.Exponents, inputs)
}

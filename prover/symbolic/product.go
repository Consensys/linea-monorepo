package symbolic

import (
	"fmt"
	"math/big"
	"reflect"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
)

/*
Operator symbolizing "Multiplication"
*/
type Product struct {
	// Exponents for each term in the multiplication
	Exponents []int
}

/*
Returns the degree of the operation given, as input, the degree of its children
*/
func (m Product) Degree(inputDegrees []int) int {
	res := 0
	// Just the sum of all the degrees
	for i, exp := range m.Exponents {
		res += exp * inputDegrees[i]
	}
	return res
}

/*
Evaluates a k-ary multiplication. It can take vectors or scalars as inputs
*/
func (m Product) Evaluate(inputs []sv.SmartVector) sv.SmartVector {
	return sv.Product(m.Exponents, inputs)
}

/*
Creates a new linear combination. Performs no simplification
*/
func NewProduct(items []*Expression, exponents []int) *Expression {

	if len(items) != len(exponents) {
		panic("unmatching lengths")
	}

	if len(items) == 0 {
		panic("no item are forbidden")
	}

	e := &Expression{
		Operator: Product{Exponents: exponents},
		Children: items,
	}

	// Now we need to assign the ESH
	eshashes := make([]sv.SmartVector, len(e.Children))
	for i := range e.Children {
		eshashes[i] = sv.NewConstant(e.Children[i].ESHash, 1)
	}

	if len(items) > 0 {
		// The cast back to sv.Constant is not functionally important but is an easy
		// sanity check.
		e.ESHash = e.Operator.Evaluate(eshashes).(*sv.Constant).Get(0)
	}

	return e
}

/*
Add to non-LC terms
*/
func MulTwoNonProd(a, b *Expression) *Expression {

	if a.IsMul() || b.IsMul() {
		panic("Add two terms with an LC term")
	}

	if a.ESHash == b.ESHash {
		return NewSingleTermProduct(a, 2)
	}

	// Otherwise, addition
	res := &Expression{
		Operator: Product{Exponents: []int{1, 1}},
		Children: []*Expression{a, b},
	}

	res.ESHash.Mul(&a.ESHash, &b.ESHash)
	return res
}

/*
Returns a sum of linear combinations as another linear combination
*/
func MulProducts(exprs ...*Expression) *Expression {

	newExponents := []int{}
	newChildren := []*Expression{}
	ESHtoExponentsPos := map[field.Element]int{}

	for _, expr := range exprs {
		// This will panic if the expression is not a Mul but this is expected
		// behaviour.
		curExponents := expr.Operator.(Product).Exponents

		for i := range expr.Children {
			currESHash := expr.Children[i].ESHash
			// Check if the ESH was already discoved
			if pos, ok := ESHtoExponentsPos[currESHash]; ok {
				// Found, add the coefficient
				newExponents[pos] += curExponents[i]
			} else {
				// Not found, add an entry
				ESHtoExponentsPos[currESHash] = len(newExponents)
				newExponents = append(newExponents, curExponents[i])
				newChildren = append(newChildren, expr.Children[i])
			}
		}
	}

	/*
		We do not support division so there is no chances that one of the
		exponents cancels out. Thus no need to prune the product from the
		zero-exponents term.
	*/

	newESH := field.One()
	for i := range exprs {
		newESH.Mul(&newESH, &exprs[i].ESHash)
	}

	/*
		Construct the merged linear combination
	*/
	return &Expression{
		Children: newChildren,
		Operator: Product{Exponents: newExponents},
		ESHash:   newESH,
	}
}

/*
Append a new term into a Product
*/
func MulNewTermProduct(prod *Expression, newExponent int, newTerm *Expression) *Expression {

	// Sanity-check, the newTerm should not be Product
	if _, ok := newTerm.Operator.(Product); ok {
		panic("newTerm is Product")
	}

	newExponents := append([]int{}, prod.Operator.(Product).Exponents...)
	children := prod.Children

	found := false
	for i, child := range prod.Children {
		if child.ESHash == newTerm.ESHash {
			found = true
			newExponents[i] += newExponent
		}
	}

	if !found {
		newExponents = append(newExponents, newExponent)
		children = append(children, newTerm)
	}

	var newESH field.Element
	newESH.Exp(newTerm.ESHash, big.NewInt(int64(newExponent)))
	newESH.Mul(&newESH, &prod.ESHash)

	return &Expression{
		ESHash:   newESH,
		Children: children,
		Operator: Product{Exponents: newExponents},
	}

}

/*
Creates a linerar combination from a single term
*/
func NewSingleTermProduct(term *Expression, exponent int) *Expression {
	// Sanity-check, the term should not be Product
	if _, ok := term.Operator.(Product); ok {
		panic("term is Product")
	}

	if exponent == 0 {
		return NewConstant(1)
	}

	// Circuit-breaker : coeff == 1 means it is pointless to wrap it in a LC
	if exponent == 1 {
		return term
	}

	newESH := term.ESHash
	newESH.Exp(newESH, big.NewInt(int64(exponent)))

	return &Expression{
		ESHash:   newESH,
		Children: []*Expression{term},
		Operator: Product{Exponents: []int{exponent}},
	}
}

/*
Validates that the product is well-formed
*/
func (prod Product) Validate(expr *Expression) error {
	if !reflect.DeepEqual(prod, expr.Operator) {
		panic("expr.operator != prod")
	}

	if len(prod.Exponents) != len(expr.Children) {
		return fmt.Errorf("mismatch in the size of the children and coefficients")
	}

	return nil
}

/*
Evaluate the expression in a gnark circuit
Does not support vector evaluation
*/
func (prod Product) GnarkEval(api frontend.API, inputs []frontend.Variable) frontend.Variable {

	res := frontend.Variable(1)

	// There should be as many inputs as there are coeffs
	if len(inputs) != len(prod.Exponents) {
		utils.Panic("%v inputs but %v coeffs", len(inputs), len(prod.Exponents))
	}

	/*
		Accumulate the scalars
	*/
	for i, input := range inputs {
		term := gnarkutil.Exp(api, input, prod.Exponents[i])
		res = api.Mul(res, term)
	}

	return res
}

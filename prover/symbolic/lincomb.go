package symbolic

import (
	"fmt"
	"reflect"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

/*
Operator symbolizing "Additions"
*/
type LinComb struct {
	// The Coeffs are typically small integers (1, -1)
	Coeffs []int
}

/*
Returns the degree of the operation given, as input, the degree of the children
*/
func (LinComb) Degree(inputDegrees []int) int {
	return utils.Max(inputDegrees...)
}

/*
Evaluates a linear. It can take vectors or scalars as inputs
*/
func (l LinComb) Evaluate(inputs []sv.SmartVector) sv.SmartVector {
	return sv.LinComb(l.Coeffs, inputs)
}

/*
Creates a new linear combination. Performs no simplification
*/
func NewLinComb(items []*Expression, coeffs []int) *Expression {

	if len(items) != len(coeffs) {
		panic("unmatching lengths")
	}

	if len(items) == 0 {
		panic("no item are forbidden")
	}

	e := &Expression{
		Operator: LinComb{Coeffs: coeffs},
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
func AddTwoNonLC(a, b *Expression) *Expression {

	if a.IsLC() || b.IsLC() {
		panic("Add two terms with an LC term")
	}

	if a.ESHash == b.ESHash {
		return NewSingleTermLC(a, 2)
	}

	// Otherwise, addition
	res := &Expression{
		Operator: LinComb{Coeffs: []int{1, 1}},
		Children: []*Expression{a, b},
	}

	res.ESHash.Add(&a.ESHash, &b.ESHash)
	return res
}

/*
Returns a LinComb representing the negation of l
*/
func NegateLC(e *Expression) *Expression {
	// Will panic if the input is not a LinearCombination
	// but this is expected.
	resCoeffs := append([]int{}, e.Operator.(LinComb).Coeffs...)
	for i := range resCoeffs {
		resCoeffs[i] = -resCoeffs[i]
	}

	res := Expression{
		Children: e.Children,
		Operator: LinComb{Coeffs: resCoeffs},
		ESHash:   e.ESHash,
	}

	res.ESHash.Neg(&res.ESHash)
	return &res
}

/*
Returns a sum of linear combinations as another linear combination
*/
func AddLCs(exprs ...*Expression) *Expression {

	newCoeffs := []int{}
	newChildren := []*Expression{}
	ESHtoCoeffsPos := map[field.Element]int{}

	for _, expr := range exprs {
		curCoeffs := expr.Operator.(LinComb).Coeffs
		for i := range expr.Children {
			currESHash := expr.Children[i].ESHash
			// Check if the ESH was already discoved
			if pos, ok := ESHtoCoeffsPos[currESHash]; ok {
				// Found, add the coefficient
				newCoeffs[pos] += curCoeffs[i]
			} else {
				// Not found, add an entry
				ESHtoCoeffsPos[currESHash] = len(newCoeffs)
				newCoeffs = append(newCoeffs, curCoeffs[i])
				newChildren = append(newChildren, expr.Children[i])
			}
		}
	}

	/*
		Removes the cancelled coefficients from the expression
	*/
	trimCursor := 0
	for range newCoeffs {
		if newCoeffs[trimCursor] == 0 {
			// Edge case : we can't trim the last entry by overwriting it with the next one
			if trimCursor+1 >= len(newChildren) {
				// this would be removed at the end
				continue
			}
			// This duplicates the last element but we remove it at the end of the loop
			copy(newChildren[trimCursor:], newChildren[trimCursor+1:])
			copy(newCoeffs[trimCursor:], newCoeffs[trimCursor+1:])
			continue
		}
		// Examine the next entry
		trimCursor++
	}
	// At this stage, the last entry has been duplicated many times
	// We remove it by trimming the end of the coeffs and the children
	newCoeffs = newCoeffs[:trimCursor]
	newChildren = newChildren[:trimCursor]

	if len(newCoeffs) == 1 {
		return NewSingleTermLC(newChildren[0], newCoeffs[0])
	}

	var newESH field.Element
	for i := range exprs {
		newESH.Add(&newESH, &exprs[i].ESHash)
	}

	/*
		Construct the merged linear combination
	*/
	return &Expression{
		Children: newChildren,
		Operator: LinComb{Coeffs: newCoeffs},
		ESHash:   newESH,
	}
}

/*
Append a new term into a LC
*/
func AddNewTermLC(lc *Expression, newCoeff int, newTerm *Expression) *Expression {

	// Sanity-check, the newTerm should not be LC
	if _, ok := newTerm.Operator.(LinComb); ok {
		panic("newTerm is LC")
	}

	// Sanity-check the arity must match

	newCoeffs := append([]int{}, lc.Operator.(LinComb).Coeffs...)
	children := lc.Children

	found := false
	for i, child := range lc.Children {
		if child.ESHash == newTerm.ESHash {
			found = true
			newCoeffs[i] += newCoeff
		}
	}

	if !found {
		newCoeffs = append(newCoeffs, newCoeff)
		children = append(children, newTerm)
	}

	var newESH field.Element
	newESH.SetInterface(newCoeff)
	newESH.Mul(&newESH, &newTerm.ESHash)
	newESH.Add(&newESH, &lc.ESHash)

	return &Expression{
		ESHash:   newESH,
		Children: children,
		Operator: LinComb{Coeffs: newCoeffs},
	}

}

/*
Creates a linerar combination from a single term
*/
func NewSingleTermLC(term *Expression, coeff int) *Expression {
	// Sanity-check, the term should not be LC
	if _, ok := term.Operator.(LinComb); ok {
		panic("term is LC")
	}

	// Circuit-breaker : coeff == 1 means it is pointless to wrap it in a LC
	if coeff == 1 {
		return term
	}

	var newESH field.Element
	newESH.SetInterface(coeff)
	newESH.Mul(&newESH, &term.ESHash)

	return &Expression{
		ESHash:   newESH,
		Children: []*Expression{term},
		Operator: LinComb{Coeffs: []int{coeff}},
	}
}

/*
Validates that the LC is well-formed
*/
func (lc LinComb) Validate(expr *Expression) error {
	if !reflect.DeepEqual(lc, expr.Operator) {
		panic("expr.operator != lc")
	}

	if len(lc.Coeffs) != len(expr.Children) {
		return fmt.Errorf("mismatch in the size of the children and coefficients")
	}

	return nil
}

/*
Evaluate the expression in a gnark circuit
Does not support vector evaluation
*/
func (lc LinComb) GnarkEval(api frontend.API, inputs []frontend.Variable) frontend.Variable {

	res := frontend.Variable(0)

	// There should be as many inputs as there are coeffs
	if len(inputs) != len(lc.Coeffs) {
		utils.Panic("%v inputs but %v coeffs", len(inputs), len(lc.Coeffs))
	}

	/*
		Accumulate the scalars
	*/
	for i, input := range inputs {
		coeff := frontend.Variable(lc.Coeffs[i])
		res = api.Add(res, api.Mul(coeff, input))
	}

	return res
}

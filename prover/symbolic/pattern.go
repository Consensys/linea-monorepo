package symbolic

import "github.com/consensys/accelerated-crypto-monorepo/maths/field"

/*
Returns true if an expression is a mul at the top-level
*/
func (f *Expression) IsMul() bool {
	_, ok := f.Operator.(Product)
	return ok
}

/*
Returns true if the expression is a product with a single-factor
(possibly with multiplicity).

Returns the exponents if true
*/
func (f *Expression) IsSingleFactor() (bool, int) {
	prod, ok := f.Operator.(Product)
	if ok && len(prod.Exponents) == 1 {
		return true, prod.Exponents[0]
	}
	return false, 0
}

/*
Returns true if the expression is a single-term LC with a -1 as a coeff
*/
func (f *Expression) IsSingleTerm() (bool, int) {
	lc, ok := f.Operator.(LinComb)
	if ok && len(lc.Coeffs) == 1 {
		return true, lc.Coeffs[0]
	}
	return false, 0
}

/*
Returns true if it is a constant and return its value if so
*/
func (f *Expression) IsConstant() (bool, field.Element) {
	if cons, ok := f.Operator.(Constant); ok {
		return ok, cons.Val
	}
	return false, field.Zero()
}

/*
Returns true if the expression is a LC
*/
func (f *Expression) IsLC() bool {
	_, ok := f.Operator.(LinComb)
	return ok
}

/*
Returns true if the expression is a variable
*/
func (f *Expression) IsVariable() bool {
	_, ok := f.Operator.(Variable)
	return ok
}

package symbolic

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

// Get the expression obtained by adding two expressions together
func (expr *Expression) Add(other *Expression) *Expression {

	/*
		This happens if the zkEVM returns a poisoned variable and
		this variable gets used in an expression in either a global
		or a local constraint. In this case, we reject when the global
		constraint is instantiated (e.g., after the expression if built).
		As a consequence, we need to build "poisoned expressions" (i.e.,
		a nil pointer). This will be caught later on.
	*/
	if expr == nil || other == nil {
		return nil
	}

	// Edge-case : adding by zero
	if expr.ESHash.IsZero() {
		return other
	}

	// Edge-case : adding by zero
	if other.ESHash.IsZero() {
		return expr
	}

	// Edge case : the two operands are constants
	{
		okExpr, valExpr := expr.IsConstant()
		okOther, valOther := other.IsConstant()
		if okExpr && okOther {
			var res field.Element
			res.Add(&valExpr, &valOther)
			return NewConstant(res)
		}
	}

	// Edge case : one of the two operand is a LC
	_, ok := expr.Operator.(LinComb)
	_, ok_ := other.Operator.(LinComb)

	switch {
	case ok && ok_:
		return AddLCs(expr, other)
	case ok && !ok_:
		return AddNewTermLC(expr, 1, other)
	case !ok && ok_:
		return AddNewTermLC(other, 1, expr)
	}

	return AddTwoNonLC(expr, other)
}

// Get the expression obtained by multiplying two expressions together
func (expr *Expression) Mul(other *Expression) *Expression {
	/*
		This happens if the zkEVM returns a poisoned variable and
		this variable gets used in an expression in either a global
		or a local constraint. In this case, we reject when the global
		constraint is instantiated (e.g., after the expression if built).
		As a consequence, we need to build "poisoned expressions" (i.e.,
		a nil pointer). This will be caught later on.
	*/
	if expr == nil || other == nil {
		return nil
	}

	// Edge case : one of the two operand are zero
	// Then just return the constant zero
	if expr.ESHash.IsZero() || other.ESHash.IsZero() {
		return NewConstant(0)
	}

	// Edge case : the two operands are constants
	// Just return a constant containing the result
	{
		okExpr, valExpr := expr.IsConstant()
		okOther, valOther := other.IsConstant()
		if okExpr && okOther {
			var res field.Element
			res.Mul(&valExpr, &valOther)
			return NewConstant(res)
		}
	}

	// Edge case, one of the two operand is one
	// Return the other
	if expr.ESHash.IsOne() {
		return other
	}

	if other.ESHash.IsOne() {
		return expr
	}

	var esh field.Element
	esh.Mul(&expr.ESHash, &other.ESHash)

	/*
		If one of the two operand is a product, we accumulate it
	*/
	switch {
	case expr.IsMul() && other.IsMul():
		// Just append the multiplications
		return MulProducts(expr, other)
	case expr.IsMul() != other.IsMul():

		// Ensures that expr is Mul and not other. At this point
		// we should not use expr.IsMul() or other.IsMul() anymore.
		if !expr.IsMul() {
			expr, other = other, expr
		}

		// If other has a single term, then we should move the coeff factor out
		if ok, coeff := other.IsSingleTerm(); ok {
			innerTerm := expr.Mul(other.Children[0])
			return NewSingleTermLC(innerTerm, coeff)
		}

		return MulNewTermProduct(expr, 1, other)
	}

	ok, coeff := expr.IsSingleTerm()
	ok_, coeff_ := other.IsSingleTerm()

	switch {
	case ok && ok_:
		// Wraps the operator in a LC, this simplifies the expression
		innerTerm := expr.Children[0].Mul(other.Children[0])
		return NewSingleTermLC(innerTerm, coeff*coeff_)
	case ok != ok_:
		if !ok {
			expr, other = other, expr
			coeff = coeff_ // coeff = coeff_ would just assign a dummy value
		}

		// Now expr has a single term and other does not
		innerTerm := expr.Children[0].Mul(other)
		return NewSingleTermLC(innerTerm, coeff)
	}

	return MulTwoNonProd(expr, other)
}

// Returns a negation of expr
func (expr *Expression) Neg() *Expression {

	/*
		This happens if the zkEVM returns a poisoned variable and
		this variable gets used in an expression in either a global
		or a local constraint. In this case, we reject when the global
		constraint is instantiated (e.g., after the expression if built).
		As a consequence, we need to build "poisoned expressions" (i.e.,
		a nil pointer). This will be caught later on.
	*/
	if expr == nil {
		return nil
	}

	// Edge case : the two operands are constants
	{
		ok, cons := expr.IsConstant()
		if ok {
			var res field.Element
			res.Neg(&cons)
			return NewConstant(res)
		}
	}

	// If it is a LC, just revert all the coefficients
	if expr.IsLC() {
		return NegateLC(expr)
	}

	// Otherwise, wrap it in a singleTerm LC
	return NewSingleTermLC(expr, -1)
}

/*
Substract two expressions
*/
func (f *Expression) Sub(other *Expression) *Expression {
	return f.Add(other.Neg())
}

/*
Square an expression
*/
func (f *Expression) Square() *Expression {
	return f.Mul(f)
}

/*
Square an expression
*/
func (f *Expression) Pow(n int) *Expression {

	f.AssertValid()

	if n < 0 {
		panic("negative exponent")
	}

	if n == 0 {
		return NewConstant(1)
	}

	if n == 1 {
		return f
	}

	if n%2 == 0 {
		x2 := f.Square()
		res := x2.Pow(n / 2)
		return res
	}

	if n%2 == 1 {
		x2 := f.Square()
		res := x2.Pow((n - 1) / 2).Mul(f)
		return res
	}

	panic("unreachable")
}

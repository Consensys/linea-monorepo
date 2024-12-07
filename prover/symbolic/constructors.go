package symbolic

// Get the expression obtained by adding two expressions together. The function
// some elementary simplification routines such as merging the LinComb nodes or
// collapsing the constants.
//
// Deprecated: Use [Add] instead. In the future, we will make this function
// completely internal.
func (expr *Expression) Add(other *Expression) *Expression {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if expr == nil || other == nil {
		return nil
	}

	return NewLinComb([]*Expression{expr, other}, []int{1, 1})

}

// Get the expression obtained by multiplying two expressions together
//
// Deprecated: Use [Mul] instead as this function will be made private.
func (expr *Expression) Mul(other *Expression) *Expression {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if expr == nil || other == nil {
		return nil
	}

	return NewProduct([]*Expression{expr, other}, []int{1, 1})
}

// Returns a negation of expr
//
// Deprecated: Use [Neg] instead of this function as it will be made private.
func (expr *Expression) Neg() *Expression {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if expr == nil {
		return nil
	}

	// Otherwise, wrap it in a singleTerm LC
	return NewLinComb([]*Expression{expr}, []int{-1})
}

/*
Subtract two expressions

Deprecated: Use [Sub] instead of this function as it will be made private.
*/
func (f *Expression) Sub(other *Expression) *Expression {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if f == nil || other == nil {
		return nil
	}

	return NewLinComb([]*Expression{f, other}, []int{1, -1})
}

/*
Square an expression

Deprecated: Use [Square] instead of this function as it will be made private.
*/
func (f *Expression) Square() *Expression {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if f == nil {
		return nil
	}

	return NewProduct([]*Expression{f}, []int{2})
}

/*
Square an expression

// Deprecated: Use [Pow] instead of this function as it will be made private.
*/
func (f *Expression) Pow(n int) *Expression {

	if f == nil {
		return nil
	}

	f.AssertValid()

	return NewProduct([]*Expression{f}, []int{n})
}

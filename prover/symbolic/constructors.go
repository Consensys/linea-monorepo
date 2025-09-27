package symbolic

// Get the expression obtained by adding two expressions together. The function
// some elementary simplification routines such as merging the LinComb nodes or
// collapsing the constants.
//
// Deprecated: Use [Add] instead. In the future, we will make this function
// completely internal.
func (expr *Expression[T]) Add(other *Expression[T]) *Expression[T] {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if expr == nil || other == nil {
		return nil
	}

	return NewLinComb([]*Expression[T]{expr, other}, []int{1, 1})

}

// Get the expression obtained by multiplying two expressions together
//
// Deprecated: Use [Mul] instead as this function will be made private.
func (expr *Expression[T]) Mul(other *Expression[T]) *Expression[T] {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if expr == nil || other == nil {
		return nil
	}

	return NewProduct([]*Expression[T]{expr, other}, []int{1, 1})
}

// Returns a negation of expr
//
// Deprecated: Use [Neg] instead of this function as it will be made private.
func (expr *Expression[T]) Neg() *Expression[T] {

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
	return NewLinComb([]*Expression[T]{expr}, []int{-1})
}

/*
Substract two expressions

Deprecated: Use [Sub] instead of this function as it will be made private.
*/
func (f *Expression[T]) Sub(other *Expression[T]) *Expression[T] {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if f == nil || other == nil {
		return nil
	}

	return NewLinComb([]*Expression[T]{f, other}, []int{1, -1})
}

/*
Square an expression

Deprecated: Use [Square] instead of this function as it will be made private.
*/
func (f *Expression[T]) Square() *Expression[T] {

	// This happens if the zkEVM returns a poisoned variable and
	// this variable gets used in an expression in either a global
	// or a local constraint. In this case, we reject when the global
	// constraint is instantiated (e.g., after the expression if built).
	// As a consequence, we need to build "poisoned expressions" (i.e.,
	// a nil pointer). This will be caught later on.
	if f == nil {
		return nil
	}

	return NewProduct([]*Expression[T]{f}, []int{2})
}

/*
Square an expression

// Deprecated: Use [Pow] instead of this function as it will be made private.
*/
func (f *Expression[T]) Pow(n int) *Expression[T] {

	if f == nil {
		return nil
	}

	f.AssertValid()

	return NewProduct([]*Expression[T]{f}, []int{n})
}

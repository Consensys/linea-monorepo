package symbolic

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

/*
IsMul returns true if an expression is a mul at the top-level
*/
func (f *Expression[T]) IsMul() bool {
	_, ok := f.Operator.(Product[T])
	return ok
}

/*
IsConstant returns true if it is a constant and return its value if so
*/
func (f *Expression[T]) IsConstant() (bool, fext.GenericFieldElem) {
	if cons, ok := f.Operator.(Constant[T]); ok {
		return ok, cons.Val
	}
	return false, fext.GenericFieldZero()
}

/*
IsLC returns true if the expression is a LC
*/
func (f *Expression[T]) IsLC() bool {
	_, ok := f.Operator.(LinComb[T])
	return ok
}

/*
IsVariable returns true if the expression is a variable
*/
func (f *Expression[T]) IsVariable() bool {
	_, ok := f.Operator.(Variable[T])
	return ok
}

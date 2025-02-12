package symbolic

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

/*
IsMul returns true if an expression is a mul at the top-level
*/
func (f *Expression) IsMul() bool {
	_, ok := f.Operator.(Product)
	return ok
}

/*
IsConstant returns true if it is a constant and return its value if so
*/
func (f *Expression) IsConstant() (bool, fext.Element) {
	if cons, ok := f.Operator.(Constant); ok {
		return ok, cons.Val
	}
	return false, fext.Zero()
}

/*
IsLC returns true if the expression is a LC
*/
func (f *Expression) IsLC() bool {
	_, ok := f.Operator.(LinComb)
	return ok
}

/*
IsVariable returns true if the expression is a variable
*/
func (f *Expression) IsVariable() bool {
	_, ok := f.Operator.(Variable)
	return ok
}

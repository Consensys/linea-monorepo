package wizard

import (
	"github.com/consensys/gnark/frontend"
)

type NativeFieldOps struct{ api frontend.API }

func (n NativeFieldOps) Mul(a, b *frontend.Variable) *frontend.Variable {
	r := n.api.Mul(*a, *b)
	return &r
}

func (n NativeFieldOps) Add(a, b *frontend.Variable) *frontend.Variable {
	r := n.api.Add(*a, *b)
	return &r
}

func (n NativeFieldOps) Neg(a *frontend.Variable) *frontend.Variable {
	r := n.api.Neg(*a)
	return &r
}

func (n NativeFieldOps) Sub(a, b *frontend.Variable) *frontend.Variable {
	r := n.api.Sub(*a, *b)
	return &r
}

func (n NativeFieldOps) Inverse(a *frontend.Variable) *frontend.Variable {
	r := n.api.Inverse(*a)
	return &r
}

func (n NativeFieldOps) ToBinary(a *frontend.Variable, m ...int) []frontend.Variable {
	r := n.api.ToBinary(*a, m...)
	return r
}

func (n NativeFieldOps) FromBinary(a ...frontend.Variable) *frontend.Variable {
	r := n.api.FromBinary(a...)
	return &r
}

func (n NativeFieldOps) Xor(a, b frontend.Variable) frontend.Variable {
	r := n.api.Xor(a, b)
	return r
}

func (n NativeFieldOps) Or(a, b frontend.Variable) frontend.Variable {
	r := n.api.Or(a, b)
	return r
}

func (n NativeFieldOps) And(a, b frontend.Variable) frontend.Variable {
	r := n.api.And(a, b)
	return r
}

func (n NativeFieldOps) Select(a frontend.Variable, i1, i2 *frontend.Variable) *frontend.Variable {
	r := n.api.Select(a, *i1, *i2)
	return &r
}

func (n NativeFieldOps) Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 *frontend.Variable) *frontend.Variable {
	r := n.api.Lookup2(b0, b1, *i0, *i1, *i2, *i3)
	return &r
}

func (n NativeFieldOps) IsZero(a *frontend.Variable) frontend.Variable {
	r := n.api.IsZero(*a)
	return r
}

func (n NativeFieldOps) Cmp(a, b *frontend.Variable) frontend.Variable {
	r := n.api.Cmp(*a, *b)
	return r
}

func (n NativeFieldOps) AssertIsEqual(a, b *frontend.Variable) {
	n.api.AssertIsEqual(*a, *b)
}

func (n NativeFieldOps) AssertIsDifferent(a, b *frontend.Variable) {
	n.api.AssertIsEqual(*a, *b)
}

func (n NativeFieldOps) AssertIsBoolean(a frontend.Variable) {
	n.api.AssertIsBoolean(a)
}

func (n NativeFieldOps) AssertIsCrumb(a *frontend.Variable) {
	n.api.AssertIsCrumb(*a)
}

func (n NativeFieldOps) AssertIsLessOrEqual(v *frontend.Variable, bound *frontend.Variable) {
	n.api.AssertIsLessOrEqual(*v, *bound)
}

func (n NativeFieldOps) Println(a ...frontend.Variable) {
	n.api.Println(a...)
}

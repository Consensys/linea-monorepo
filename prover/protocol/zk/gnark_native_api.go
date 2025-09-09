package zk

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
)

type NativeElement struct {
	V frontend.Variable
}

func packNE(v frontend.Variable) *NativeElement {
	return &NativeElement{V: v}
}

func valueOfNE(v any) NativeElement {
	return NativeElement{V: v}
}

type NativeAPI struct {
	api frontend.API
}

func newNativeAPI(api frontend.API) (*NativeAPI, error) {
	return &NativeAPI{api: api}, nil
}

var _ FieldOps[NativeElement] = &NativeAPI{}

func (n *NativeAPI) Mul(a, b *NativeElement) *NativeElement {
	return packNE(n.api.Mul(a.V, b.V))
}

func (n *NativeAPI) Add(a, b *NativeElement) *NativeElement {
	return packNE(n.api.Add(a.V, b.V))
}

func (n *NativeAPI) Neg(a *NativeElement) *NativeElement {
	return packNE(n.api.Neg(a.V))
}

func (n *NativeAPI) Sub(a, b *NativeElement) *NativeElement {
	return packNE(n.api.Sub(a.V, b.V))
}

func (n *NativeAPI) Inverse(a *NativeElement) *NativeElement {
	return packNE(n.api.Inverse(a.V))
}

func (n *NativeAPI) Div(a, b *NativeElement) *NativeElement {
	return packNE(n.api.Div(a.V, b.V))
}

func (n *NativeAPI) ToBinary(a *NativeElement, m ...int) []frontend.Variable {
	r := n.api.ToBinary(a.V, m...)
	return r
}

func (n *NativeAPI) FromBinary(a ...frontend.Variable) *NativeElement {
	return packNE(n.api.FromBinary(a...))
}

func (n *NativeAPI) And(a, b frontend.Variable) frontend.Variable {
	return n.api.And(a, b)
}

func (n *NativeAPI) Xor(a, b frontend.Variable) frontend.Variable {
	return n.api.Xor(a, b)
}

func (n *NativeAPI) Or(a, b frontend.Variable) frontend.Variable {
	return n.api.Or(a, b)
}

func (n *NativeAPI) Select(a frontend.Variable, i1, i2 *NativeElement) *NativeElement {
	return packNE(n.api.Select(a, i1.V, i2.V))
}

func (n *NativeAPI) Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 *NativeElement) *NativeElement {
	return packNE(n.api.Lookup2(b0, b1, i0.V, i1.V, i2.V, i3.V))
}

func (n *NativeAPI) IsZero(a *NativeElement) frontend.Variable {
	return n.api.IsZero(a.V)
}

func (n *NativeAPI) AssertIsEqual(a, b *NativeElement) {
	n.api.AssertIsEqual(a.V, b.V)
}

func (n *NativeAPI) AssertIsDifferent(a, b *NativeElement) {
	n.api.AssertIsDifferent(a.V, b.V)
}

func (n *NativeAPI) AssertIsLessOrEqual(v *NativeElement, bound *NativeElement) {
	n.api.AssertIsLessOrEqual(v.V, bound.V)
}

func (n *NativeAPI) FromUint(v uint64) *NativeElement {
	return packNE(v)
}

func (n *NativeAPI) FromKoalabear(v koalabear.Element) *NativeElement {
	return packNE(v)
}

func (n *NativeAPI) NewHint(f solver.Hint, nbOutputs int, inputs ...*NativeElement) ([]*NativeElement, error) {
	_inputs := make([]frontend.Variable, len(inputs))
	for i, r := range inputs {
		_inputs[i] = r.V
	}
	_r, err := n.api.NewHint(f, nbOutputs, _inputs...)
	if err != nil {
		return nil, err
	}
	res := make([]*NativeElement, nbOutputs)
	for i, r := range _r {
		res[i] = packNE(r)
	}
	return res, nil
}

func (n *NativeAPI) Println(a ...*NativeElement) {
	args := make([]frontend.Variable, len(a))
	for i, v := range a {
		args[i] = v.V
	}
	n.api.Println(args...)
}

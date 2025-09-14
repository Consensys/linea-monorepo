package zk

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
)

type EmulatedElement = emulated.Element[emulated.KoalaBear]

// EmulatedFieldOps struct implementing FieldOps[T FType].
// It is a wrapper around emulatedApi for koalabear.
type EmulatedFieldOps struct {
	api frontend.API
	ef  *emulated.Field[emulated.KoalaBear]
}

func newEmulatedAPI(api frontend.API) (*EmulatedFieldOps, error) {
	f, err := emulated.NewField[emulated.KoalaBear](api)
	if err != nil {
		return nil, err
	}
	return &EmulatedFieldOps{api: api, ef: f}, nil
}

var _ FieldOps[EmulatedElement] = &EmulatedFieldOps{}

func (e *EmulatedFieldOps) Mul(a, b *EmulatedElement) *EmulatedElement {
	return e.ef.Mul(a, b)
}

func (e *EmulatedFieldOps) MulConst(a *EmulatedElement, b *big.Int) *EmulatedElement {
	return e.ef.MulConst(a, b)
}

func (e *EmulatedFieldOps) Add(a, b *EmulatedElement) *EmulatedElement {
	return e.ef.Add(a, b)
}

func (e *EmulatedFieldOps) Neg(a *EmulatedElement) *EmulatedElement {
	return e.ef.Neg(a)
}

func (e *EmulatedFieldOps) Sub(a, b *EmulatedElement) *EmulatedElement {
	return e.ef.Sub(a, b)
}

func (e *EmulatedFieldOps) Inverse(a *EmulatedElement) *EmulatedElement {
	return e.ef.Inverse(a)
}

func (e *EmulatedFieldOps) Div(a, b *EmulatedElement) *EmulatedElement {
	return e.ef.Div(a, b)
}

func (e *EmulatedFieldOps) ToBinary(a *EmulatedElement, m ...int) []frontend.Variable {
	return e.ef.ToBits(a)
}

func (e *EmulatedFieldOps) FromBinary(a ...frontend.Variable) *EmulatedElement {
	return e.ef.FromBits(a...)
}

func (n *EmulatedFieldOps) And(a, b frontend.Variable) frontend.Variable {
	return n.api.And(a, b)
}

func (n *EmulatedFieldOps) Xor(a, b frontend.Variable) frontend.Variable {
	return n.api.Xor(a, b)
}

func (n *EmulatedFieldOps) Or(a, b frontend.Variable) frontend.Variable {
	return n.api.Or(a, b)
}

func (e *EmulatedFieldOps) Select(a frontend.Variable, i1, i2 *EmulatedElement) *EmulatedElement {
	return e.ef.Select(a, i1, i2)
}

func (e *EmulatedFieldOps) Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 *EmulatedElement) *EmulatedElement {
	return e.ef.Lookup2(b0, b1, i0, i1, i2, i3)
}

func (e *EmulatedFieldOps) IsZero(a *EmulatedElement) frontend.Variable {
	return e.ef.IsZero(a)
}

func (e *EmulatedFieldOps) AssertIsEqual(a, b *EmulatedElement) {
	e.ef.AssertIsEqual(a, b)
}

func (e *EmulatedFieldOps) AssertIsDifferent(a, b *EmulatedElement) {
	e.ef.AssertIsDifferent(a, b)
}

func (e *EmulatedFieldOps) AssertIsLessOrEqual(a, b *EmulatedElement) {
	e.ef.AssertIsLessOrEqual(a, b)
}

func (e *EmulatedFieldOps) FromUint(v uint64) *EmulatedElement {
	a := emulated.ValueOf[emulated.KoalaBear](v)
	return &a
}

func (e *EmulatedFieldOps) FromKoalabear(v koalabear.Element) *EmulatedElement {
	a := emulated.ValueOf[emulated.KoalaBear](v)
	return &a
}

func (e *EmulatedFieldOps) NewHint(f solver.Hint, nbOutputs int, inputs ...*EmulatedElement) ([]*EmulatedElement, error) {
	res, err := e.ef.NewHint(f, nbOutputs, inputs...)
	return res, err
}

func (e *EmulatedFieldOps) Println(a ...*EmulatedElement) {
	for i := 0; i < len(a); i++ {
		e.api.Println(a[i].Limbs...)
	}
}

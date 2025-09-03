package zk

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
)

// EmulatedFieldOps struct implementing FieldOps[T FType].
// It is a wrapper around emulatedApi for koalabear.
type EmulatedFieldOps struct {
	api frontend.API
	ef  *emulated.Field[emulated.KoalaBear]
}

func (e EmulatedFieldOps) Mul(a, b *emulated.Element[emulated.KoalaBear]) *emulated.Element[emulated.KoalaBear] {
	return e.ef.Mul(a, b)
}

func (e EmulatedFieldOps) Add(a, b *emulated.Element[emulated.KoalaBear]) *emulated.Element[emulated.KoalaBear] {
	return e.ef.Add(a, b)
}

func (e EmulatedFieldOps) Neg(a *emulated.Element[emulated.KoalaBear]) *emulated.Element[emulated.KoalaBear] {
	return e.ef.Neg(a)
}

func (e EmulatedFieldOps) Sub(a, b *emulated.Element[emulated.KoalaBear]) *emulated.Element[emulated.KoalaBear] {
	return e.ef.Sub(a, b)
}

func (e EmulatedFieldOps) Inverse(a *emulated.Element[emulated.KoalaBear]) *emulated.Element[emulated.KoalaBear] {
	return e.ef.Inverse(a)
}

func (e EmulatedFieldOps) ToBinary(a *emulated.Element[emulated.KoalaBear], m ...int) []frontend.Variable {
	return e.ef.ToBits(a)
}

func (e EmulatedFieldOps) FromBinary(a ...frontend.Variable) *emulated.Element[emulated.KoalaBear] {
	return e.ef.FromBits(a...)
}

func (e EmulatedFieldOps) Select(a frontend.Variable, i1, i2 *emulated.Element[emulated.KoalaBear]) *emulated.Element[emulated.KoalaBear] {
	return e.ef.Select(a, i1, i2)
}

func (e EmulatedFieldOps) Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 *emulated.Element[emulated.KoalaBear]) *emulated.Element[emulated.KoalaBear] {
	return e.ef.Lookup2(b0, b1, i0, i1, i2, i3)
}

func (e EmulatedFieldOps) IsZero(a *emulated.Element[emulated.KoalaBear]) frontend.Variable {
	return e.ef.IsZero(a)
}

func (e EmulatedFieldOps) AssertIsEqual(a, b *emulated.Element[emulated.KoalaBear]) {
	e.ef.AssertIsEqual(a, b)
}

func (e EmulatedFieldOps) AssertIsDifferent(a, b *emulated.Element[emulated.KoalaBear]) {
	e.ef.AssertIsDifferent(a, b)
}

func (e EmulatedFieldOps) AssertIsLessOrEqual(a, b *emulated.Element[emulated.KoalaBear]) {
	e.ef.AssertIsLessOrEqual(a, b)
}

func (e EmulatedFieldOps) SetFromUint(a *emulated.Element[emulated.KoalaBear], v uint64) {
	*a = emulated.ValueOf[emulated.KoalaBear](v)
}

func (e EmulatedFieldOps) Println(a ...emulated.Element[emulated.KoalaBear]) {
	for i := 0; i < len(a); i++ {
		e.api.Println(a[i].Limbs...)
	}
}

func (e EmulatedFieldOps) Api() frontend.API {
	return e.api
}

func getFieldOpEmulated(api frontend.API) (FieldOps[emulated.Element[emulated.KoalaBear]], error) {
	ff, err := emulated.NewField[emulated.KoalaBear](api)
	if err != nil {
		return EmulatedFieldOps{}, err
	}
	return EmulatedFieldOps{api: api, ef: ff}, nil
}

package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
)

// FieldType stores the circuit type: emulated or native.
type FieldType int

const (
	// Native koalabear circuit defined over koalabear
	Native FieldType = iota

	// Emulated koalabear circuit over bls12-377
	Emulated

	// Other cases
	Undefined
)

// FType supported field types in gnark circuit
type FType interface {
	emulated.Element[emulated.KoalaBear] | frontend.Variable
}

// FieldType returns the type of the circuit, emulated (koalabear over bls12-377) or native
// (koalabear over koalabear).
func getType[T FType]() FieldType {
	var a T
	if _, ok := any(a).(emulated.Element[emulated.KoalaBear]); ok {
		return Emulated
	}
	return Native
}

// Circuit 0: mixed
type TestMixedCircuitMixed[T FType] struct {
	A, B T
	R    T `gnark:",public"`
}

type FieldOps[T FType] interface {
	Mul(a, b *T) *T
	Add(a, b *T) *T
	Neg(a *T) *T
	Sub(a, b *T) *T
	Inverse(a *T) *T

	ToBinary(a *T, n ...int) []frontend.Variable
	FromBinary(b ...frontend.Variable) *T
	Xor(a, b frontend.Variable) frontend.Variable
	Or(a, b *T) *T
	And(a, b *T) *T

	Select(b frontend.Variable, i1, i2 *T) *T
	Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 *T) *T
	IsZero(i1 *T) frontend.Variable

	Cmp(i1, i2 *T) frontend.Variable

	AssertIsEqual(a, b *T)
	AssertIsDifferent(i1, i2 *T)
	AssertIsBoolean(i1 frontend.Variable)
	AssertIsCrumb(i1 *T)
	AssertIsLessOrEqual(v *T, bound *T)

	Println(a ...T)
}

type EmulatedFieldOps struct {
	api frontend.API
	ef  *emulated.Field[emulated.KoalaBear]
}

func (e EmulatedFieldOps) Mul(a, b *emulated.Element[emulated.KoalaBear]) *emulated.Element[emulated.KoalaBear] {
	return e.ef.Mul(a, b)
}

func (e EmulatedFieldOps) AssertIsEqual(a, b *emulated.Element[emulated.KoalaBear]) {
	e.ef.AssertIsEqual(a, b)
}

func (e EmulatedFieldOps) Println(a ...emulated.Element[emulated.KoalaBear]) {
	for i := 0; i < len(a); i++ {
		e.api.Println(a[i].Limbs...)
	}
}

func getFieldOpNative(api frontend.API) FieldOps[frontend.Variable] {
	return NativeFieldOps{api: api}
}

func getFieldOpEmulated(api frontend.API) (FieldOps[emulated.Element[emulated.KoalaBear]], error) {
	ff, err := emulated.NewField[emulated.KoalaBear](api)
	if err != nil {
		return EmulatedFieldOps{}, err
	}
	return EmulatedFieldOps{api: api, ef: ff}, nil
}

func (c *TestMixedCircuitMixed[T]) Define(api frontend.API) error {

	var wApi FieldOps[T]
	t := getType[T]()
	if t == Emulated {
		tmpApi, err := getFieldOpEmulated(api)
		if err != nil {
			return err
		}
		wApi = any(tmpApi).(FieldOps[T])
	} else {
		tmpApi := getFieldOpNative(api)
		wApi = any(tmpApi).(FieldOps[T])
	}

	wApi.Println(c.R)
	a := wApi.Mul(&c.A, &c.B)
	wApi.AssertIsEqual(a, &c.R)

	return nil
}

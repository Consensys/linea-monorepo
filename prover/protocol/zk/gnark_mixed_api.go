package zk

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
)

// fieldType stores the circuit type: emulated or native.
type fieldType int

const (
	// Native koalabear circuit defined over koalabear
	Native fieldType = iota

	// Emulated koalabear circuit over bls12-377
	Emulated

	// Other cases
	Undefined
)

// FType (for Field Type) supported field types in gnark circuit
type FType interface {
	emulated.Element[emulated.KoalaBear] | frontend.Variable
}

// getType returns the type of the circuit, emulated (koalabear over bls12-377) or native
// (koalabear over koalabear).
func getType[T FType]() fieldType {
	var a T
	if _, ok := any(a).(emulated.Element[emulated.KoalaBear]); ok {
		return Emulated
	}
	return Native
}

type FieldOps[T FType] interface {
	Mul(a, b *T) *T
	Add(a, b *T) *T
	Neg(a *T) *T
	Sub(a, b *T) *T
	Inverse(a *T) *T
	Div(a, b *T) *T

	ToBinary(a *T, n ...int) []frontend.Variable
	FromBinary(b ...frontend.Variable) *T

	And(a, b frontend.Variable) frontend.Variable

	Select(b frontend.Variable, i1, i2 *T) *T
	Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 *T) *T
	IsZero(i1 *T) frontend.Variable

	AssertIsEqual(a, b *T)
	AssertIsDifferent(i1, i2 *T)

	AssertIsLessOrEqual(v *T, bound *T)

	FromUint(v uint64) T

	FromKoalabear(v koalabear.Element) *T

	NewHint(f solver.Hint, nbOutputs int, inputs ...*T) ([]*T, error)

	Println(a ...T)

	NativeApi() frontend.API
}

// Circuit 0: mixed
// type TestMixedCircuitMixed[T FType] struct {
// 	A, B T
// 	R    T `gnark:",public"`
// }

// func (c *TestMixedCircuitMixed[T]) Define(api frontend.API) error {

// 	var wApi FieldOps[T]
// 	t := getType[T]()
// 	if t == Emulated {
// 		tmpApi, err := getFieldOpEmulated(api)
// 		if err != nil {
// 			return err
// 		}
// 		wApi = any(tmpApi).(FieldOps[T])
// 	} else {
// 		tmpApi := getFieldOpNative(api)
// 		wApi = any(tmpApi).(FieldOps[T])
// 	}

// 	wApi.Println(c.R)
// 	a := wApi.Mul(&c.A, &c.B)
// 	wApi.AssertIsEqual(a, &c.R)

// 	return nil
// }

package zk

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
)

// Element (for Field Type) supported field types in gnark circuit
type Element interface {
	EmulatedElement | NativeElement
}

type APIGen[T Element] interface {
	Mul(a, b *T) *T
	MulConst(a *T, b *big.Int) *T
	Add(a, b *T) *T
	Neg(a *T) *T
	Sub(a, b *T) *T
	Inverse(a *T) *T
	Div(a, b *T) *T

	ToBinary(a *T, n ...int) []frontend.Variable
	FromBinary(b ...frontend.Variable) *T

	And(a, b frontend.Variable) frontend.Variable
	Xor(a, b frontend.Variable) frontend.Variable
	Or(a, b frontend.Variable) frontend.Variable

	Select(b frontend.Variable, i1, i2 *T) *T
	Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 *T) *T
	IsZero(i1 *T) frontend.Variable

	AssertIsEqual(a, b *T)
	AssertIsDifferent(i1, i2 *T)

	AssertIsLessOrEqual(v *T, bound *T)

	FromUint(v uint64) *T

	FromKoalabear(v koalabear.Element) *T

	NewHint(f solver.Hint, nbOutputs int, inputs ...*T) ([]*T, error)

	Println(a ...*T)

	// GnarkAPI() frontend.API
}

func NewApi[T Element](api frontend.API) (APIGen[T], error) {
	var t T
	var ret APIGen[T]
	var ok bool
	switch any(t).(type) {
	case EmulatedElement:
		retV, err := newEmulatedAPI(api)
		if err != nil {
			return nil, err
		}
		ret, ok = any(retV).(APIGen[T])
		if !ok {
			panic("could not cast emulated API to requested type")
		}
	case NativeElement:
		retV, err := newNativeAPI(api)
		if err != nil {
			return nil, err
		}
		ret, ok = any(retV).(APIGen[T])
		if !ok {
			panic("could not cast native API to requested type")
		}
	default:
		return nil, fmt.Errorf("unsupported requested API type")
	}
	return ret, nil
}

func ValueOf[T Element](input any) T {
	var ret T
	switch v := any(&ret).(type) {
	case *EmulatedElement:
		*v = emulated.ValueOf[emulated.KoalaBear](input)
	case *NativeElement:
		*v = NativeElement{V: input}
	default:
		panic("unsupported type")
	}
	return ret
}

// Circuit 0: mixed
// type TestMixedCircuitMixed[T FType] struct {
// 	A, B T
// 	R    T `gnark:",public"`
// }

// func (c *TestMixedCircuitMixed[T]) Define(api frontend.API) error {

// 	var wApi APIGen[T]
// 	t := getType[T]()
// 	if t == Emulated {
// 		tmpApi, err := getFieldOpEmulated(api)
// 		if err != nil {
// 			return err
// 		}
// 		wApi = any(tmpApi).(APIGen[T])
// 	} else {
// 		tmpApi := getFieldOpNative(api)
// 		wApi = any(tmpApi).(APIGen[T])
// 	}

// 	wApi.Println(c.R)
// 	a := wApi.Mul(&c.A, &c.B)
// 	wApi.AssertIsEqual(a, &c.R)

// 	return nil
// }

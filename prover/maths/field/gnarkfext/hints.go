package gnarkfext

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

func init() {
	// solver.RegisterHint(inverseE4Hint[zk.NativeElement], inverseE4Hint[zk.EmulatedElement])
	solver.RegisterHint(
		// zk.MixedHint[zk.EmulatedElement](_divE4),
		// zk.MixedHint[zk.NativeElement](_divE4),
		// divE4Emulated,
		// _divE4,
		inverseE4Emulated,
		inverseE4Native)
}

func inverseE2Hint(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
	var a, c extensions.E2

	a.A0.SetBigInt(inputs[0])
	a.A1.SetBigInt(inputs[1])

	c.Inverse(&a)

	c.A0.BigInt(res[0])
	c.A1.BigInt(res[1])

	return nil
}

func divE4Emulated(_ *big.Int, nativeInputs []*big.Int, nativeOutputs []*big.Int) error {
	return emulated.UnwrapHint(nativeInputs, nativeOutputs, _divE4)
}

func _divE4(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
	var a, b, c fext.Element

	a.B0.A0.SetBigInt(inputs[0])
	a.B0.A1.SetBigInt(inputs[1])
	a.B1.A0.SetBigInt(inputs[2])
	a.B1.A1.SetBigInt(inputs[3])

	b.B0.A0.SetBigInt(inputs[4])
	b.B0.A1.SetBigInt(inputs[5])
	b.B1.A0.SetBigInt(inputs[6])
	b.B1.A1.SetBigInt(inputs[7])

	c.Div(&a, &b)

	c.B0.A0.BigInt(res[0])
	c.B0.A1.BigInt(res[1])
	c.B1.A0.BigInt(res[2])
	c.B1.A1.BigInt(res[3])

	return nil
}

func InverseHint[T zk.Element]() solver.Hint {
	var t T
	switch any(t).(type) {
	case zk.EmulatedElement:
		return inverseE4Emulated
	case zk.NativeElement:
		return inverseE4Native
	default:
		panic("unsupported requested API type")
	}
}

func inverseE4Emulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, inverseE4Native)
}

func inverseE4Native(_ *big.Int, inputs []*big.Int, output []*big.Int) error {

	var a, c fext.Element

	a.B0.A0.SetBigInt(inputs[0])
	a.B0.A1.SetBigInt(inputs[1])
	a.B1.A0.SetBigInt(inputs[2])
	a.B1.A1.SetBigInt(inputs[3])

	c.Inverse(&a)

	c.B0.A0.BigInt(output[0])
	c.B0.A1.BigInt(output[1])
	c.B1.A0.BigInt(output[2])
	c.B1.A1.BigInt(output[3])

	return nil

}

// func inverseE4Hint[T zk.Element](_ *big.Int, nativeInputs []*big.Int, nativeOutputs []*big.Int) error {
// 	var t T
// 	switch any(t).(type) {
// 	case zk.EmulatedElement:
// 		return emulated.UnwrapHint(nativeInputs, nativeOutputs, _inverseE4)
// 	case zk.NativeElement:
// 		return _inverseE4(nil, nativeInputs, nativeOutputs)
// 	default:
// 		return fmt.Errorf("unsupported requested API type")
// 	}
// }

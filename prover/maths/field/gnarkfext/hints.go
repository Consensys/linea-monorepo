package gnarkfext

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

func init() {
	solver.RegisterHint(
		inverseE2Hint,
		inverseE4HintNative, inverseE4HintEmulated,
		divE4HintNative, divE4HintEmulated,
		mulE4HintNative, mulE4HintEmulated)
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

func inverseE4HintNative(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
	var a, c fext.Element

	a.B0.A0.SetBigInt(inputs[0])
	a.B0.A1.SetBigInt(inputs[1])
	a.B1.A0.SetBigInt(inputs[2])
	a.B1.A1.SetBigInt(inputs[3])

	c.Inverse(&a)

	c.B0.A0.BigInt(res[0])
	c.B0.A1.BigInt(res[1])
	c.B1.A0.BigInt(res[2])
	c.B1.A1.BigInt(res[3])

	return nil
}

func inverseE4HintEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, inverseE4HintNative)
}

func inverseE4Hint(t zk.VType) solver.Hint {
	if t == zk.Native {
		return inverseE4HintNative
	} else {
		return inverseE4HintEmulated
	}
}

func divE4HintNative(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
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

func divE4HintEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, divE4HintNative)
}

func divE4Hint(t zk.VType) solver.Hint {
	if t == zk.Native {
		return divE4HintNative
	} else {
		return divE4HintEmulated
	}
}

// mulE4HintNative computes E4 multiplication in the hint
func mulE4HintNative(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
	var a, b, c fext.Element

	a.B0.A0.SetBigInt(inputs[0])
	a.B0.A1.SetBigInt(inputs[1])
	a.B1.A0.SetBigInt(inputs[2])
	a.B1.A1.SetBigInt(inputs[3])

	b.B0.A0.SetBigInt(inputs[4])
	b.B0.A1.SetBigInt(inputs[5])
	b.B1.A0.SetBigInt(inputs[6])
	b.B1.A1.SetBigInt(inputs[7])

	c.Mul(&a, &b)

	c.B0.A0.BigInt(res[0])
	c.B0.A1.BigInt(res[1])
	c.B1.A0.BigInt(res[2])
	c.B1.A1.BigInt(res[3])

	return nil
}

func mulE4HintEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, mulE4HintNative)
}

func mulE4Hint(t zk.VType) solver.Hint {
	if t == zk.Native {
		return mulE4HintNative
	} else {
		return mulE4HintEmulated
	}
}

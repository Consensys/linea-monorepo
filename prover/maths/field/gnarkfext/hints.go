package gnarkfext

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

func init() {
	solver.RegisterHint(inverseE2Hint, inverseE4Hint, divE4Hint)
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

func inverseE4Hint(_ *big.Int, nativeInputs []*big.Int, nativeOutputs []*big.Int) error {

	return emulated.UnwrapHint(nativeInputs, nativeOutputs,
		func(mod *big.Int, inputs, output []*big.Int) error {

			var a, c fext.Element

			fmt.Printf("bigInt = %s\n", inputs[0].String())
			fmt.Printf("bigInt = %s\n", inputs[1].String())
			fmt.Printf("bigInt = %s\n", inputs[2].String())
			fmt.Printf("bigInt = %s\n", inputs[3].String())

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
		})

}

func divE4Hint(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
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

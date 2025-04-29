package bls

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/math/emulated"
)

const (
	// BLS base field is 381 bits, and we use 4 limbs of 128 bits to represent
	// it. However, the highest limb is always zero, but the arithmetization
	// keeps it for nice alignment. We pass it to the circuit but check
	// explicitly that its 0.
	nbFpLimbs = 4 // (x_3, x_2, x_1, x_0) MSB order

	nbG1Limbs = 2 * nbFpLimbs  // (Ax, Ay)
	nbG2Limbs = 4 * nbFpLimbs  // (BxIm, BxRe, ByIm, ByRe)
	nbGtLimbs = 12 * nbFpLimbs // representation according to gnark - we don't use Gt in arithmetization, only in glue for accumulation

)

type g1ElementWizard struct {
	P [nbG1Limbs]frontend.Variable
}

func (c *g1ElementWizard) ToG1Element(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField]) sw_bls12381.G1Affine {
	panic("todo")
}

type g2ElementWizard struct {
	Q [nbG2Limbs]frontend.Variable
}

func (c *g2ElementWizard) ToG2Element(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField]) sw_bls12381.G2Affine {
	panic("todo")
}

type gtElementWizard struct {
	T [nbGtLimbs]frontend.Variable
}

func (c *gtElementWizard) ToGTElement(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField]) sw_bls12381.GTEl {
	panic("todo")
}

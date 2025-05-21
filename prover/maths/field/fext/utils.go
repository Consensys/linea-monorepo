package fext

import (
	"math/bits"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// One returns 1
func One() Element {
	var res Element
	res.B0.A0.SetOne()
	return res
}

// SetZero sets an E4 elmt to zero
func Zero() Element {
	var res Element
	return res
}

func Uint64(z *Element) (uint64, uint64, uint64, uint64) {
	return uint64(z.B0.A0.Bits()[0]), uint64(z.B0.A1.Bits()[0]), uint64(z.B1.A0.Bits()[0]), uint64(z.B1.A1.Bits()[0])
}

// SetInt64 sets z to v and returns z
// ./common/smartvectorsext/fuzzing.go:195:		coeffField.SetInt64(int64(tcase.coeffs[i]))
// ./common/smartvectorsext/arithmetic_op.go:95:		c.SetInt64(int64(coeff))
// ./common/smartvectorsext/arithmetic_op.go:119:		c.SetInt64(int64(coeff))
// ./common/smartvectorsext/arithmetic_op.go:149:		c.SetInt64(int64(coeff))
// ./common/smartvectorsext/arithmetic_op.go:168:		c.SetInt64(int64(coeff))
// ./common/smartvectors/fft_test.go:254:				xCoeff.SetInt64(2)
// ./common/smartvectors/fuzzing.go:193:		coeffField.SetInt64(int64(tcase.coeffs[i]))
// ./common/smartvectors/arithmetic_op.go:96:		c.SetInt64(int64(coeff))
// ./common/smartvectors/arithmetic_op.go:120:		c.SetInt64(int64(coeff))
// ./common/smartvectors/arithmetic_op.go:150:		c.SetInt64(int64(coeff))
// ./common/smartvectors/arithmetic_op.go:169:		c.SetInt64(int64(coeff))
// ./common/vector/vector_wizard.go:102:		res[i].SetInt64(int64(x))
func SetInt64(z *Element, v int64) *Element {
	z.B0.A0.SetInt64(v)
	z.B0.A1.SetZero()
	z.B1.SetZero()
	return z // z.toMont()
}

// SetInt64Pair sets z to the int64 pair corresponding to (v1, v2) and returns z
// ./maths/common/smartvectorsext/polynomial_test.go:147:				SetInt64Pair(
// ./maths/common/smartvectorsext/arithmetic_test.go:252:	sum := new(fext.Element).SetInt64Pair(int64(1+5*fext.RootPowers[1]), 10)
// ./maths/common/vectorext/vector_wizard.go:126:		res[i].SetInt64Pair(int64(xs[2*i]), int64(xs[2*i+1]))
// ./maths/common/polyext/poly_test.go:32:	require.Equal(t, y, *new(fext.Element).SetInt64Pair(int64(first), int64(second)))
func SetInt64Tuple(z *Element, v1, v2, v3, v4 int64) *Element {
	z.B0.A0.SetInt64(v1)
	z.B0.A1.SetInt64(v2)
	z.B1.A0.SetInt64(v3)
	z.B1.A1.SetInt64(v4)
	return z // z.toMont()
}

// FromBase sets z = v
func FromBase(z *Element, v *field.Element) {
	z.B0.A0.Set(v)
	z.B0.A1.SetZero()
	z.B1.A0.SetZero()
	z.B1.A1.SetZero()
}

// PseudoRand generates a field using a pseudo-random number generator
func RandomElement() Element {
	var res Element
	res.SetRandom()
	return res
}

func ExpToInt(z *Element, x Element, k int) *Element {
	if k == 0 {
		return z.SetOne()
	}

	if k < 0 {
		x.Inverse(&x)
		k = -k
	}

	z.Set(&x)

	for i := bits.Len(uint(k)) - 2; i >= 0; i-- {
		z.Square(z)
		if (k>>i)&1 == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}

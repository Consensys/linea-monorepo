package fext

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// SetUint64 sets z to v and returns z
func (z *Element) SetUint64(v uint64) *Element {
	//  sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
	temp := fr.Element{v}
	temp.Mul(&temp, &rSquare)
	res := Element{field.Zero(), temp}
	return &res // z.toMont()
}

// SetInt64 sets z to v and returns z
func (z *Element) SetInt64(v int64) *Element {

	// absolute value of v
	m := v >> 63
	z.SetUint64(uint64((v ^ m) - m))

	if m != 0 {
		// v is negative
		z.Neg(z)
	}

	return z
}

func (z *Element) Uint64() (uint64, uint64) {
	return z.A0.Bits()[0], z.A1.Bits()[0]
}

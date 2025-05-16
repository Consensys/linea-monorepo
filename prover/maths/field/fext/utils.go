package fext

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// SetUint64 sets z to v and returns z
func (z *Element) SetUint64(v uint64) *Element {
	//  sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
	z.B0.A0.SetUint64(v)
	z.B0.A1.SetZero()
	return z // z.toMont()
}

// SetInt64 sets z to v and returns z
func (z *Element) SetInt64(v int64) *Element {
	z.B0.A0.SetInt64(v)
	z.B0.A1.SetZero()
	return z // z.toMont()
}

func (z *Element) SetFromVector(inp [4]int) *Element {
	z.B0.A0.SetInt64(int64(inp[0]))
	z.B0.A1.SetInt64(int64(inp[1]))
	return z // z.toMont()
}

// SetInt64Pair sets z to the int64 pair corresponding to (v1, v2) and returns z
func (z *Element) SetInt64Pair(v1, v2 int64) *Element {
	z.B0.A0.SetInt64(v1)
	z.B0.A1.SetInt64(v2)
	return z // z.toMont()
}

func (z *Element) Uint64() (uint64, uint64) {
	return z.A0.Bits()[0], z.A1.Bits()[0]
}

func Butterfly(a, b *Element) {
	field.Butterfly(&a.B0.A0, &b.B0.A0)
	field.Butterfly(&a.B0.A1, &b.B0.A1)
}

func (z *Element) MulByBase(first *Element, second *Element) *Element {
	z.B0.A0.Mul(&first.A0, second)
	z.B0.A1.Mul(&first.A1, second)
	return z
}

func (z *Element) DivByBase(first *Element, second *Element) *Element {
	z.B0.A0.Div(&first.A0, second)
	z.B0.A1.Div(&first.A1, second)
	return z
}

package fext

// SetUint64 sets z to v and returns z
func SetUint64(z *Element, v uint64) *Element {
	//  sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
	z.B0.A0.SetUint64(v)
	z.B0.A1.SetZero()
	z.B1.SetZero()
	return z // z.toMont()
}

// SetInt64 sets z to v and returns z
func SetInt64(z *Element, v int64) *Element {
	z.B0.A0.SetInt64(v)
	z.B0.A1.SetZero()
	z.B1.SetZero()
	return z // z.toMont()
}

// TODO Only used in maths/common/vectorext/vector_wizard.go:112, remove this function
// func (z *Element) SetFromVector(inp [4]int) *Element {
// 	z.B0.A0.SetInt64(int64(inp[0]))
// 	z.B0.A1.SetInt64(int64(inp[1]))
// 	return z // z.toMont()
// }

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

// func (z *Element) Uint64() (uint64, uint64) {
// 	return z.A0.Bits()[0], z.A1.Bits()[0]
// }

// func Butterfly(a, b *Element) {
// 	field.Butterfly(&a.B0.A0, &b.B0.A0)
// 	field.Butterfly(&a.B0.A1, &b.B0.A1)
// }

// func (z *Element) MulByElement(first *Element, second *Element) *Element {
// 	z.B0.A0.Mul(&first.A0, second)
// 	z.B0.A1.Mul(&first.A1, second)
// 	return z
// }

// TODO Only used in FFT, remove this function
// func (z *Element) DivByBase(first *Element, second *Element) *Element {
// 	z.B0.A0.Div(&first.A0, second)
// 	z.B0.A1.Div(&first.A1, second)
// 	return z
// }

package fext

// SetUint64 sets z to v and returns z
func (z *Element) SetUint64(v uint64) *Element {
	//  sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
	z.A0.SetUint64(v)
	z.A1.SetZero()
	return z // z.toMont()
}

// SetInt64 sets z to v and returns z
func (z *Element) SetInt64(v int64) *Element {

	z.A0.SetInt64(v)
	z.A1.SetZero()
	return z // z.toMont()
}

func (z *Element) Uint64() (uint64, uint64) {
	return z.A0.Bits()[0], z.A1.Bits()[0]
}

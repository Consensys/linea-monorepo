package fft

import (
	field "github.com/consensys/gnark-crypto/field/koalabear"
	gnarkfft "github.com/consensys/gnark-crypto/field/koalabear/fft"
)

// BitReverse applies the bit-reversal permutation to v.
// len(v) must be a power of 2
func BitReverse(v []field.Element) {
	gnarkfft.BitReverse(v)
}

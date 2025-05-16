package fft

import (
	gnarkfft "github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// BitReverseExt applies the bit-reversal permutation to v.
// len(v) must be a power of 2
func BitReverseExt(v []fext.Element) {
	gnarkfft.BitReverse(v)
}

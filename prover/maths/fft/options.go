package fft

import (
	gnarkfft "github.com/consensys/gnark-crypto/field/koalabear/fft"
)

// Option defines option for altering the behavior of FFT methods.
// See the descriptions of functions returning instances of this type for
// particular options.
type Option = gnarkfft.Option

// OnCoset if provided, FFT(a) returns the evaluation of a on a coset.
func OnCoset() Option {
	return gnarkfft.OnCoset()
}

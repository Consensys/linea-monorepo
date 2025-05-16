package fft

import (
	gnarkfft "github.com/consensys/gnark-crypto/field/koalabear/fft"
)

// Option defines option for altering the behavior of FFT methods.
// See the descriptions of functions returning instances of this type for
// particular options.
type Option = gnarkfft.Option

type fftConfig struct {
	coset   bool
	nbTasks int
}

// OnCoset if provided, FFT(a) returns the evaluation of a on a coset.
func OnCoset() Option {
	return gnarkfft.OnCoset()
}

// WithNbTasks sets the max number of task (go routine) to spawn. Must be between 1 and 512.
func WithNbTasks(nbTasks int) Option {
	return gnarkfft.WithNbTasks(nbTasks)
}

package fft

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"

	gnarkfft "github.com/consensys/gnark-crypto/field/koalabear/fft"
)

// Decimation is used in the FFT call to select decimation in time or in frequency
type Decimation = gnarkfft.Decimation

const (
	DIT Decimation = gnarkfft.DIT
	DIF Decimation = gnarkfft.DIF
)

// parallelize threshold for a single butterfly op, if the fft stage is not parallelized already
const butterflyThreshold = 16

// FFT computes (recursively) the discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// if coset if set, the FFT(a) returns the evaluation of a on a coset.
func (domain *Domain) FFT(a []field.Element, decimation Decimation, opts ...Option) {

	domain.GnarkDomain.FFT(a, decimation, opts...)
}

// FFTInverse computes (recursively) the inverse discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// coset sets the shift of the fft (0 = no shift, standard fft)
// len(a) must be a power of 2, and w must be a len(a)th root of unity in field F.
func (domain *Domain) FFTInverse(a []field.Element, decimation Decimation, opts ...Option) {
	domain.GnarkDomain.FFTInverse(a, decimation, opts...)
}

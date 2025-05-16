package fft

import (
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

func (domain *Domain) FFTExt(a []fext.E4, decimation Decimation, opts ...Option) {
	domain.gnarkDomain.FFTExt(a, decimation, opts...)
}

// FFTInverseExt computes (recursively) the inverse discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// coset sets the shift of the fft (0 = no shift, standard fft)
// len(a) must be a power of 2, and w must be a len(a)th root of unity in field F.
func (domain *Domain) FFTInverseExt(a []fext.E4, decimation Decimation, opts ...Option) {
	domain.gnarkDomain.FFTInverseExt(a, decimation, opts...)
}

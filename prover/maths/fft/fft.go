package fft

import (
	"math/bits"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

// Decimation is used in the FFT call to select decimation in time or in frequency
type Decimation uint8

const (
	DIT Decimation = iota
	DIF
)

// FFT computes (recursively) the discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// if coset if set, the FFT(a) returns the evaluation of a on a coset.
func (domain *Domain) FFT(a []field.Element, decimation Decimation, coset ...bool) {

	_coset := false
	if len(coset) > 0 {
		_coset = coset[0]
	}

	// if coset != 0, scale by coset table
	if _coset {
		scale := func(cosetTable []field.Element) {
			for i := 0; i < len(a); i++ {
				a[i].Mul(&a[i], &cosetTable[i])
			}
		}
		if decimation == DIT {
			scale(domain.CosetTableReversed)

		} else {
			scale(domain.CosetTable)
		}
	}

	switch decimation {
	case DIF:
		difFFT(a, domain.Twiddles, 0)
	case DIT:
		ditFFT(a, domain.Twiddles, 0)
	default:
		panic("not implemented")
	}
}

// FFTInverse computes (recursively) the inverse discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// coset sets the shift of the fft (0 = no shift, standard fft)
// len(a) must be a power of 2, and w must be a len(a)th root of unity in field F.
func (domain *Domain) FFTInverse(a []field.Element, decimation Decimation, coset ...bool) {

	_coset := false
	if len(coset) > 0 {
		_coset = coset[0]
	}

	switch decimation {
	case DIF:
		difFFT(a, domain.TwiddlesInv, 0)
	case DIT:
		ditFFT(a, domain.TwiddlesInv, 0)
	default:
		panic("not implemented")
	}

	// scale by CardinalityInv
	if !_coset {
		for i := 0; i < len(a); i++ {
			a[i].Mul(&a[i], &domain.CardinalityInv)
		}
		return
	}

	scale := func(cosetTable []field.Element) {
		for i := 0; i < len(a); i++ {
			a[i].Mul(&a[i], &cosetTable[i]).
				Mul(&a[i], &domain.CardinalityInv)
		}
	}
	if decimation == DIT {
		scale(domain.CosetTableInv)
		return
	}

	// decimation == DIF
	scale(domain.CosetTableInvReversed)

}

func difFFT(a []field.Element, twiddles [][]field.Element, stage int) {

	n := len(a)
	if n == 1 {
		return
	} else if n == 8 {
		kerDIF8(a, twiddles, stage)
		return
	}
	m := n >> 1

	// i == 0
	field.Butterfly(&a[0], &a[m])
	for i := 1; i < m; i++ {
		field.Butterfly(&a[i], &a[i+m])
		a[i+m].Mul(&a[i+m], &twiddles[stage][i])
	}

	if m == 1 {
		return
	}

	nextStage := stage + 1
	difFFT(a[0:m], twiddles, nextStage)
	difFFT(a[m:n], twiddles, nextStage)
}

func ditFFT(a []field.Element, twiddles [][]field.Element, stage int) {

	n := len(a)
	if n == 1 {
		return
	} else if n == 8 {
		kerDIT8(a, twiddles, stage)
		return
	}
	m := n >> 1

	nextStage := stage + 1

	ditFFT(a[0:m], twiddles, nextStage)
	ditFFT(a[m:n], twiddles, nextStage)

	field.Butterfly(&a[0], &a[m])
	for k := 1; k < m; k++ {
		a[k+m].Mul(&a[k+m], &twiddles[stage][k])
		field.Butterfly(&a[k], &a[k+m])
	}
}

// BitReverse applies the bit-reversal permutation to a.
// len(a) must be a power of 2 (as in every single function in this file)
func BitReverse(a []field.Element) {
	n := uint64(len(a))
	nn := uint64(64 - bits.TrailingZeros64(n))

	for i := uint64(0); i < n; i++ {
		irev := bits.Reverse64(i) >> nn
		if irev > i {
			a[i], a[irev] = a[irev], a[i]
		}
	}
}

// kerDIT8 is a kernel that process a FFT of size 8
func kerDIT8(a []field.Element, twiddles [][]field.Element, stage int) {

	field.Butterfly(&a[0], &a[1])
	field.Butterfly(&a[2], &a[3])
	field.Butterfly(&a[4], &a[5])
	field.Butterfly(&a[6], &a[7])
	field.Butterfly(&a[0], &a[2])
	a[3].Mul(&a[3], &twiddles[stage+1][1])
	field.Butterfly(&a[1], &a[3])
	field.Butterfly(&a[4], &a[6])
	a[7].Mul(&a[7], &twiddles[stage+1][1])
	field.Butterfly(&a[5], &a[7])
	field.Butterfly(&a[0], &a[4])
	a[5].Mul(&a[5], &twiddles[stage+0][1])
	field.Butterfly(&a[1], &a[5])
	a[6].Mul(&a[6], &twiddles[stage+0][2])
	field.Butterfly(&a[2], &a[6])
	a[7].Mul(&a[7], &twiddles[stage+0][3])
	field.Butterfly(&a[3], &a[7])
}

// kerDIF8 is a kernel that process a FFT of size 8
func kerDIF8(a []field.Element, twiddles [][]field.Element, stage int) {

	field.Butterfly(&a[0], &a[4])
	field.Butterfly(&a[1], &a[5])
	field.Butterfly(&a[2], &a[6])
	field.Butterfly(&a[3], &a[7])
	a[5].Mul(&a[5], &twiddles[stage+0][1])
	a[6].Mul(&a[6], &twiddles[stage+0][2])
	a[7].Mul(&a[7], &twiddles[stage+0][3])
	field.Butterfly(&a[0], &a[2])
	field.Butterfly(&a[1], &a[3])
	field.Butterfly(&a[4], &a[6])
	field.Butterfly(&a[5], &a[7])
	a[3].Mul(&a[3], &twiddles[stage+1][1])
	a[7].Mul(&a[7], &twiddles[stage+1][1])
	field.Butterfly(&a[0], &a[1])
	field.Butterfly(&a[2], &a[3])
	field.Butterfly(&a[4], &a[5])
	field.Butterfly(&a[6], &a[7])
}

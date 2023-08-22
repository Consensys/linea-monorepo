package fastpoly

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/matrix"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

// Multiply twi polynomial modulo X^n - 1
// a and b must be in coefficient form, the result is in coefficient
// form
// a and b are destroyed during the operation
func MultModXMinus1(domain *fft.Domain, res, a, b []field.Element) {
	// All the item must be of the right size
	if len(a) != len(b) || len(a) != len(res) || len(a) != int(domain.Cardinality) {
		panic(
			fmt.Sprintf("All items should have the right size %v %v %v %v",
				domain.Cardinality, len(res), len(a), len(b)),
		)
	}

	domain.FFT(a, fft.DIF)
	domain.FFT(b, fft.DIF)
	vector.MulElementWise(res, a, b)
	domain.FFTInverse(res, fft.DIT)

}

// Multiply two polynomials modulo X^n - 1
// `a` must be in coefficient form
// `precomp` must be in evaluation form over the domain (DIT)
// `res` can be either `a` or any pre-allocated array
// res must pre-allocated in all cases
// a is destroyed during the operation
func MultModXnMinus1Precomputed(domain *fft.Domain, res, a, precomp []field.Element) {

	// All the item must be of the right size
	if len(a) != len(precomp) || len(a) != len(res) || len(a) != int(domain.Cardinality) {
		panic(
			fmt.Sprintf("All items should have the right size %v %v %v %v",
				domain.Cardinality, len(res), len(a), len(precomp)),
		)
	}

	domain.FFT(a, fft.DIF)
	vector.MulElementWise(res, a, precomp)
	domain.FFTInverse(res, fft.DIT)
}

// Batched version of `MultModXnMinus1Precomputed`
// `a` is a matrix such that
//   - The row i, contains the vector of i-th coefficients of each poly
//
// `precomp` must be in evaluation form over the domain (DIT)
// `res` is a preallocated matrix such that
//   - It has the same dimensions as `a`
//   - It must be preallocated
//
// `a` is destroyed during the operation
func BatchMultModXnMinus1Precomputed(domain *fft.Domain, res, a [][]field.Element, precomp []field.Element) {

	// All the item must be of the right size
	if len(a) != len(precomp) || len(a) != len(res) || len(a) != int(domain.Cardinality) {
		panic(
			fmt.Sprintf("All items should have the right size %v %v %v %v",
				domain.Cardinality, len(res), len(a), len(precomp)),
		)
	}

	domain.BatchFFT(a, fft.DIF)
	matrix.ScalarMulSubslices(res, a, precomp)
	domain.BatchFFTInverse(res, fft.DIT)
}

package fastpoly

import (
	"fmt"

	gnarkfft "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Multiply twi polynomial modulo X^n - 1
// a and b must be in coefficient form, the result is in coefficient
// form
// a and b are destroyed during the operation
func MultModXMinus1(domain *gnarkfft.Domain, res, a, b []field.Element) {
	// All the item must be of the right size
	if len(a) != len(b) || len(a) != len(res) || uint64(len(a)) != domain.Cardinality {
		panic(
			fmt.Sprintf("All items should have the right size %v %v %v %v",
				domain.Cardinality, len(res), len(a), len(b)),
		)
	}

	domain.FFT(a, gnarkfft.DIF)
	domain.FFT(b, gnarkfft.DIF)
	vector.MulElementWise(res, a, b)
	domain.FFTInverse(res, gnarkfft.DIT)

}

// Multiply two polynomials modulo X^n - 1
// `a` must be in coefficient form
// `precomp` must be in evaluation form over the domain (DIT)
// `res` can be either `a` or any pre-allocated array
// res must pre-allocated in all cases
// a is destroyed during the operation
func MultModXnMinus1Precomputed(domain *gnarkfft.Domain, res, a, precomp []field.Element) {

	// All the item must be of the right size
	if len(a) != len(precomp) || len(a) != len(res) || uint64(len(a)) != domain.Cardinality {
		panic(
			fmt.Sprintf("All items should have the right size %v %v %v %v",
				domain.Cardinality, len(res), len(a), len(precomp)),
		)
	}

	domain.FFT(a, gnarkfft.DIF)
	vector.MulElementWise(res, a, precomp)
	domain.FFTInverse(res, gnarkfft.DIT)
}

package fastpoly

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

/*
Given the evaluations of a polynomial on a domain (whose
size must be a power of two, or panic), return an evaluation
at a chosen x.

As an input the user can specify that the inputs are given
on a coset.

Interpolate(poly []E1, x E4)
*/
func Interpolate(poly []field.Element, x fext.Element, oncoset ...bool) fext.Element {
	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	n := len(poly)

	domain := fft.NewDomain(uint64(n))
	denominator := make([]fext.Element, n)

	one := fext.One()

	if len(oncoset) > 0 && oncoset[0] {
		x.MulByElement(&x, &domain.FrMultiplicativeGenInv)
	}

	/*
		First, we compute the denominator,

		D_x = \frac{X}{x} - g for x \in H
			where H is the subgroup of the roots of unity (not the coset)
			and g a field element such that gH is the coset
	*/
	denominator[0] = x
	for i := 1; i < n; i++ {
		denominator[i].MulByElement(&denominator[i-1], &domain.GeneratorInv)
	}

	for i := 0; i < n; i++ {
		denominator[i].Sub(&denominator[i], &one)
		if denominator[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly
			var res fext.Element
			fext.FromBase(&res, &poly[i])
			return res
		}
	}

	/*
		Then, we compute the sum between the inverse of the denominator
		and the poly

		\sum_{x \in H}\frac{P(gx)}{D_x}
	*/
	denominator = fext.BatchInvertE4(denominator)
	res := vectorext.ScalarProdByElement(denominator, poly)

	/*
		Then multiply the res by a factor \frac{g^{1 - n}X^n -g}{n}
	*/
	var factor fext.Element
	factor.Exp(x, big.NewInt(int64(n)))
	factor.Sub(&factor, &one)
	factor.MulByElement(&factor, &domain.CardinalityInv)
	res.Mul(&res, &factor)

	return res
}

// Batch version of Interpolate
func BatchInterpolate(polys [][]field.Element, x fext.Element, oncoset ...bool) []fext.Element {

	results := make([]fext.Element, len(polys))
	poly := polys[0]

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	n := len(poly)

	domain := fft.NewDomain(uint64(n))
	denominator := make([]fext.Element, n)

	one := fext.One()

	if len(oncoset) > 0 && oncoset[0] {
		x.MulByElement(&x, &domain.FrMultiplicativeGenInv)
	}

	/*
		First, we compute the denominator,

		D_x = \frac{X}{x} - g for x \in H
			where H is the subgroup of the roots of unity (not the coset)
			and g a field element such that gH is the coset
	*/
	denominator[0] = x
	for i := 1; i < n; i++ {
		denominator[i].MulByElement(&denominator[i-1], &domain.GeneratorInv)
	}

	for i := 0; i < n; i++ {
		denominator[i].Sub(&denominator[i], &one)

		if denominator[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly

			for k := range polys {
				fext.FromBase(&results[k], &polys[k][i])
			}
			return results
		}
	}

	/*
		Then, we compute the sum between the inverse of the denominator
		and the poly

		\sum_{x \in H}\frac{P(gx)}{D_x}
	*/
	denominator = fext.BatchInvertE4(denominator)

	// Precompute the value of x^n once outside the loop
	var factor fext.Element
	factor.Exp(x, big.NewInt(int64(n)))
	factor.Sub(&factor, &one)
	factor.MulByElement(&factor, &domain.CardinalityInv)

	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {

			// Compute the scalar product.
			res := vectorext.ScalarProdByElement(denominator, polys[k])

			// Multiply res with factor.
			res.Mul(&res, &factor)

			// Store the result.
			results[k] = res
		}
	})

	return results
}

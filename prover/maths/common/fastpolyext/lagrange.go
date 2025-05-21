package fastpolyext

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// EvaluateLagrangeOnFext computes ∑_i L_i(x), i.e. evaluates p interpreted as a polynomial in Lagrange form
func EvaluateLagrange(poly []fext.Element, x fext.Element, oncoset ...bool) fext.Element {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	// TODO handle the oncoset properly, using options
	if len(oncoset) > 0 && oncoset[0] {
		g := fft.GeneratorFullMultiplicativeGroup()
		x.MulByElement(&x, &g)
	}

	size := len(poly)
	omega, err := fft.Generator(uint64(size))
	if err != nil {
		// TODO handle that properly
		panic(err)
	}

	var accw, one, extomega fext.Element
	one.SetOne()
	accw.SetOne()
	fext.FromBase(&extomega, &omega)
	dens := make([]fext.Element, size) // [x-1, x-ω, x-ω², ...]
	for i := 0; i < size; i++ {
		dens[i].Sub(&x, &accw)
		accw.Mul(&accw, &extomega)
	}
	fext.BatchInvert(dens) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	var xn fext.Element
	xn.Exp(x, big.NewInt(int64(size))).Sub(&xn, &one) // xⁿ-1
	var nInv fext.Element
	fext.SetInt64(&nInv, int64(size)).Inverse(&nInv)
	xn.Mul(&xn, &nInv) // 1/n * (xⁿ-1)

	var res, li fext.Element
	accw.SetOne()
	for i := 0; i < size; i++ {
		li.Mul(&accw, &xn).Mul(&li, &dens[i]) // ωⁱ/n * ( xⁿ-1)/(x-ωⁱ)
		li.Mul(&li, &poly[i])                 // pᵢ *  ωⁱ/n * ( xⁿ-1)/(x-ωⁱ)
		res.Add(&res, &li)
		accw.Mul(&accw, &extomega)
	}

	return res

}

// Batch version of Interpolate
func BatchInterpolate(polys [][]fext.Element, x fext.Element, oncoset ...bool) []fext.Element {

	results := make([]fext.Element, len(polys))
	poly := polys[0]

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	n := len(poly)

	domain := fft.NewDomain(uint64(n))
	denominator := make([]fext.Element, n)

	one := fext.One()

	var wrappedFrMultiplicativeGenInv fext.Element
	fext.FromBase(&wrappedFrMultiplicativeGenInv, &domain.FrMultiplicativeGenInv)

	if len(oncoset) > 0 && oncoset[0] {
		x.Mul(&x, &wrappedFrMultiplicativeGenInv)
	}

	/*
		First, we compute the denominator,

		D_x = \frac{X}{x} - g for x \in H
			where H is the subgroup of the roots of unity (not the coset)
			and g a field element such that gH is the coset
	*/
	var wrappedGeneratorInv fext.Element
	fext.FromBase(&wrappedGeneratorInv, &domain.GeneratorInv)

	denominator[0] = x
	for i := 1; i < n; i++ {
		denominator[i].Mul(&denominator[i-1], &wrappedGeneratorInv)
	}

	for i := 0; i < n; i++ {
		denominator[i].Sub(&denominator[i], &one)

		if denominator[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly

			for k := range polys {
				results[k] = polys[k][i]
			}

			return results
		}
	}

	/*
		Then, we compute the sum between the inverse of the denominator
		and the poly

		\sum_{x \in H}\frac{P(gx)}{D_x}
	*/
	denominator = fext.BatchInvert(denominator)

	// Precompute the value of x^n once outside the loop
	xN := new(fext.Element).Exp(x, big.NewInt(int64(n)))

	// Precompute the value of domain.CardinalityInv outside the loop
	var wrappedCardinalityInv fext.Element
	fext.FromBase(&wrappedCardinalityInv, &domain.CardinalityInv)

	// Compute factor as (x^n - 1) * (1 / domain.Cardinality).
	factor := new(fext.Element).Sub(xN, &one)
	factor.Mul(factor, &wrappedCardinalityInv)

	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {

			// Compute the scalar product.
			res := vectorext.ScalarProd(polys[k], denominator)

			// Multiply res with factor.
			res.Mul(&res, factor)

			// Store the result.
			results[k] = res
		}
	})

	return results
}

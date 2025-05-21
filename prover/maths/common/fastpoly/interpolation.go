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

// EvaluateLagrange computes ∑_i L_i(x), i.e. evaluates p interpreted as a polynomial in Lagrange form
func EvaluateLagrange(poly []field.Element, x fext.Element, oncoset ...bool) fext.Element {

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
		res.Add(&res, &li)
		accw.Mul(&accw, &extomega)
	}

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

		D_x = \frac{X}{x} - g for x ∈ H
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

		∑_{x ∈ H}\frac{P(gx)}{D_x}
	*/
	denominator = fext.BatchInvert(denominator)

	// Precompute the value of xⁿ once outside the loop
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

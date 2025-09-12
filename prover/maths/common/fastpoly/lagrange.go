package fastpoly

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// EvaluateLagrangeMixed computes ∑_i L_i(x), i.e. evaluates p interpreted as a polynomial in Lagrange form, and x lives in the extension
func EvaluateLagrangeMixed(poly []field.Element, x fext.Element, oncoset ...bool) fext.Element {

	if len(oncoset) > 0 && oncoset[0] {
		genFr := fft.GeneratorFullMultiplicativeGroup()
		genFr.Inverse(&genFr)
		x.MulByElement(&x, &genFr)
	}

	res, err := vortex.EvalBasePolyLagrange(poly, x)
	if err != nil {
		panic(err)
	}

	return res
}

// EvaluateLagrange computes ∑_i L_i(x), i.e. evaluates p interpreted as a polynomial in Lagrange form, and x lives in the extension
func EvaluateLagrange(poly []field.Element, x field.Element, oncoset ...bool) field.Element {
	var xExt fext.Element
	fext.SetFromBase(&xExt, &x)
	res := EvaluateLagrangeMixed(poly, xExt, oncoset...)
	return res.B0.A0
}

// BatchEvaluateLagrangeMixed batch version of EvaluateLagrangeOnFext
func BatchEvaluateLagrangeMixed(polys [][]field.Element, x fext.Element, oncoset ...bool) []fext.Element {
	if len(polys) == 0 {
		return []fext.Element{}
	}

	n := len(polys[0])
	for i := range polys {
		if len(polys[i]) != n {
			return []fext.Element{}
		}
	}

	if !utils.IsPowerOfTwo(n) {
		return []fext.Element{}
	}

	var (
		denominators = make([]fext.Element, n)
		one          field.Element
		generator, _ = fft.Generator(uint64(n))
		generatorInv = new(field.Element).Inverse(&generator)
		cardInv      field.Element
		results      = make([]fext.Element, len(polys))
	)

	one.SetOne()
	cardInv.SetUint64(uint64(n))
	cardInv.Inverse(&cardInv)

	if len(oncoset) > 0 && oncoset[0] {
		domain := fft.NewDomain(uint64(n), fft.WithCache())
		x.MulByElement(&x, &domain.FrMultiplicativeGenInv)
	}

	// The denominator is constructed as:
	// 		D_x = x/ω^i - 1 for i = 0, 1, ..., n-1
	// 	where ω is the generator of the subgroup of roots of unity
	denominators[0] = x
	for i := 1; i < n; i++ {
		denominators[i].MulByElement(&denominators[i-1], generatorInv)
	}

	for i := 0; i < n; i++ {
		// This subtracts a base field element from a field extension element.
		denominators[i].B0.A0.Sub(&denominators[i].B0.A0, &one)
		if denominators[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly
			for k := range polys {
				fext.SetFromBase(&results[k], &polys[k][i])
			}
			return results
		}
	}

	/*
	   Then, we compute the sum between the inverse of the denominator
	   and the poly
	   \sum_{i=0}^{n-1} \frac{P(ω^i)}{D_i}
	*/
	denominators = fext.BatchInvert(denominators)

	var factor fext.Element
	// Precompute the value of x^n once outside the loop
	factor.Exp(x, big.NewInt(int64(n)))
	var extOne fext.Element
	extOne.SetOne()
	factor.Sub(&factor, &extOne)
	factor.MulByElement(&factor, &cardInv)

	// Compute results in parallel
	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {
			// Compute the scalar product between polynomial coefficients and denominators
			var res fext.Element
			for i := 0; i < n; i++ {
				res = vectorext.ScalarProdByElement(denominators, polys[k])
			}
			// Multiply res with factor.
			res.Mul(&res, &factor)
			// Store the result.
			results[k] = res
		}
	})

	return results
}

// BatchEvaluateLagrange batch evalute polys at x, where polys are in Lagrange basis
func BatchEvaluateLagrange(polys [][]field.Element, x field.Element, oncoset ...bool) []field.Element {
	var xExt fext.Element
	fext.SetFromBase(&xExt, &x)
	resExt := BatchEvaluateLagrangeMixed(polys, xExt, oncoset...)
	res := make([]field.Element, len(polys))
	for i := 0; i < len(resExt); i++ {
		res[i].Set(&resExt[i].B0.A0)
	}
	return res
}

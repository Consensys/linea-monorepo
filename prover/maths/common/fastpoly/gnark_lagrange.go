package fastpoly

import (
	"slices"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

// EvaluateLagrangeGnark a polynomial in lagrange basis on a gnark circuit
func EvaluateLagrangeGnark(api frontend.API, poly []frontend.Variable, x frontend.Variable) frontend.Variable {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	size := len(poly)
	omega, err := fft.Generator(uint64(size))
	if err != nil {
		// TODO handle that properly
		panic(err)
	}

	var accw, one field.Element
	one.SetOne()
	accw.SetOne()
	dens := make([]frontend.Variable, size) // [x-1, x-ω, x-ω², ...]
	for i := 0; i < size; i++ {
		dens[i] = api.Sub(x, accw)
		accw.Mul(&accw, &omega)
	}
	invdens := make([]frontend.Variable, size) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	for i := 0; i < size; i++ {
		invdens[i] = api.Inverse(dens[i])
	}

	var tmp frontend.Variable
	tmp = gnarkutil.Exp(api, x, size)
	tmp = api.Sub(tmp, one) // xⁿ-1
	li := api.Inverse(size)
	li = api.Mul(tmp, li) // 1/n * (xⁿ-1)

	var res frontend.Variable
	res = 0
	for i := 0; i < size; i++ {
		li = api.Mul(li, invdens[i])
		tmp = api.Mul(li, poly[i]) // pᵢ *  ωⁱ/n * ( xⁿ-1)/(x-ωⁱ)
		res = api.Add(res, tmp)
		li = api.Mul(li, dens[i])
		li = api.Mul(li, omega)
	}

	return res

}

// BatchEvaluateLagrangeGnark evaluates a batch of polynomials in Lagrange basis
// on a gnark circuit. The evaluation point is common to all polynomials.
// The implementation relies on the barycentric interpolation formula and
// leverages
func BatchEvaluateLagrangeGnark(api frontend.API, polys [][]frontend.Variable, x frontend.Variable) []frontend.Variable {

	if len(polys) == 0 {
		return []frontend.Variable{}
	}

	var (
		sizes            = listOfDifferentSizes(polys)
		maxSize          = sizes[len(sizes)-1]
		xNs              = raiseToPowersOfTwos(api, x, sizes)
		powersOfOmegaInv = powerVectorOfOmegaInv(maxSize)

		// innerProductTerms lists a sequence of term computed as
		// \frac{1}{x \omega^{-i} - 1}. These will be used to compute
		// an inner-product between the polys to obtain an unscaled
		// result. The unscaled result will then be scaled by (x^n - 1)/n
		// to obtain the result of the Lagrange interpolation.
		innerProductTerms = make([]frontend.Variable, maxSize)

		// scalingTerms lists a sequence of term computed as (x^n - 1)/n
		// where n stands for all the different sizes of the polys in
		// increasing order. (as found in "sizes"). The used to derive
		// the final result of the evaluations.
		scalingTerms = make([]frontend.Variable, len(sizes))

		// res stores the final result of the interpolation
		res = make([]frontend.Variable, len(polys))
	)

	for i := range innerProductTerms {
		innerProductTerms[i] = api.Mul(powersOfOmegaInv[i], x)
		innerProductTerms[i] = api.Sub(innerProductTerms[i], 1)
		innerProductTerms[i] = api.Inverse(innerProductTerms[i])
	}

	for i, n := range sizes {
		var nField field.Element
		nField.SetInt64(int64(n))
		scalingTerms[i] = api.Sub(xNs[i], 1)
		scalingTerms[i] = api.Div(scalingTerms[i], n)
	}

	for i := range polys {

		var (
			poly          = polys[i]
			n             = len(poly)
			scalingFactor frontend.Variable
		)

		for j := range sizes {
			if sizes[j] == n {
				scalingFactor = scalingTerms[j]
			}
		}

		// That would be completely unexpectedly wrong if that happened since
		// the sizes are deduplicated and sorted from the sizes found in poly
		// in the first place.
		if scalingFactor == nil {
			utils.Panic("could not find scaling factor for poly of size %v", n)
		}

		yUnscaled := frontend.Variable(0)

		for k := 0; k < n; k++ {

			// This optimization saves constraints when dealing with a polynomial
			// that is precomputed and raised as a Proof object in the wizard. When
			// that happens, it will contain constant values (very often actually)
			// and when they contain zeroes we can just skip the related
			// computation. The saving was not negligible in practice.
			if isConstantZeroGnarkVariable(api, poly[k]) {
				continue
			}

			tmp := api.Mul(poly[k], innerProductTerms[k*maxSize/n])
			yUnscaled = api.Add(tmp, yUnscaled)
		}

		res[i] = api.Mul(yUnscaled, scalingFactor)
	}

	return res
}

// BatchEvaluateLagrangeGnarkMixed evaluates a batch of polynomials in Lagrange basis
// on a gnark circuit. The evaluation point is common to all polynomials.
// The implementation relies on the barycentric interpolation formula and
// leverages
func BatchEvaluateLagrangeGnarkMixed(api frontend.API, polys [][]frontend.Variable, x gnarkfext.Element) []gnarkfext.Element {

	if len(polys) == 0 {
		return []gnarkfext.Element{}
	}

	var (
		sizes            = listOfDifferentSizes(polys)
		maxSize          = sizes[len(sizes)-1]
		xNs              = raiseToPowersOfTwosExt(api, x, sizes)
		powersOfOmegaInv = powerVectorOfOmegaInv(maxSize)
		one              gnarkfext.Element

		// innerProductTerms lists a sequence of term computed as
		// \frac{1}{x \omega^{-i} - 1}. These will be used to compute
		// an inner-product between the polys to obtain an unscaled
		// result. The unscaled result will then be scaled by (x^n - 1)/n
		// to obtain the result of the Lagrange interpolation.
		innerProductTerms = make([]gnarkfext.Element, maxSize)

		// scalingTerms lists a sequence of term computed as (x^n - 1)/n
		// where n stands for all the different sizes of the polys in
		// increasing order. (as found in "sizes"). The used to derive
		// the final result of the evaluations.
		scalingTerms = make([]gnarkfext.Element, len(sizes))

		// res stores the final result of the interpolation
		res = make([]gnarkfext.Element, len(polys))
	)
	one.SetOne()
	for i := range innerProductTerms {
		innerProductTerms[i].MulByFp(api, x, powersOfOmegaInv[i])
		innerProductTerms[i].Sub(api, innerProductTerms[i], one)
		innerProductTerms[i].Inverse(api, innerProductTerms[i])
	}

	for i, n := range sizes {
		var nField field.Element
		nField.SetInt64(int64(n))
		scalingTerms[i].Sub(api, xNs[i], one)
		scalingTerms[i].DivByBase(api, scalingTerms[i], n)
	}

	for i := range polys {

		var (
			poly          = polys[i]
			n             = len(poly)
			scalingFactor gnarkfext.Element
			yUnscaled     gnarkfext.Element
			tmp           gnarkfext.Element
		)

		for j := range sizes {
			if sizes[j] == n {
				scalingFactor = scalingTerms[j]
			}
		}

		// That would be completely unexpectedly wrong if that happened since
		// the sizes are deduplicated and sorted from the sizes found in poly
		// in the first place.
		if scalingFactor.B0.A0 == nil {
			utils.Panic("could not find scaling factor for poly of size %v", n)
		}

		yUnscaled.SetZero()

		for k := 0; k < n; k++ {

			// This optimization saves constraints when dealing with a polynomial
			// that is precomputed and raised as a Proof object in the wizard. When
			// that happens, it will contain constant values (very often actually)
			// and when they contain zeroes we can just skip the related
			// computation. The saving was not negligible in practice.
			if isConstantZeroGnarkVariable(api, poly[k]) {
				continue
			}

			tmp.MulByFp(api, innerProductTerms[k*maxSize/n], poly[k])
			yUnscaled.Add(api, tmp, yUnscaled)
		}

		res[i].Mul(api, yUnscaled, scalingFactor)
	}

	return res
}

// listOfDifferentSizes returns the list of sizes of the different polys
// vectors. The returned slices is deduplicated and sorted in ascending order.
func listOfDifferentSizes(polys [][]frontend.Variable) []int {

	sizes := make([]int, 0, len(polys))
	for _, poly := range polys {
		sizes = append(sizes, len(poly))
	}

	slices.Sort(sizes)
	sizes = slices.Compact(sizes)
	sizes = slices.Clip(sizes)
	return sizes
}

// raiseToPowersOfTwos returns x raised to the powers contained in ns. For
// instance, if ns = [2, 4, 8], then it returns [x, x^2, x^4, x^8]. It assumes
// that ns is sorted in ascending order and deduplicated and non-zero and are
// powers of two.
func raiseToPowersOfTwos(api frontend.API, x frontend.Variable, ns []int) []frontend.Variable {

	var (
		res   = make([]frontend.Variable, len(ns))
		curr  = x
		currN = 1
	)

	for i, n := range ns {

		if n <= 0 || !utils.IsPowerOfTwo(n) {
			utils.Panic("ns should be sorted in ascending order and deduplicated and non-zero and be powers of two, was %v", n)
		}

		for currN < n {
			curr = api.Mul(curr, curr)
			currN *= 2
		}

		res[i] = curr
	}

	return res
}

// raiseToPowersOfTwosExt returns x raised to the powers contained in ns. For
// instance, if ns = [2, 4, 8], then it returns [x, x^2, x^4, x^8]. It assumes
// that ns is sorted in ascending order and deduplicated and non-zero and are
// powers of two.
func raiseToPowersOfTwosExt(api frontend.API, x gnarkfext.Element, ns []int) []gnarkfext.Element {

	var (
		res   = make([]gnarkfext.Element, len(ns))
		curr  = x
		currN = 1
	)

	for i, n := range ns {

		if n <= 0 || !utils.IsPowerOfTwo(n) {
			utils.Panic("ns should be sorted in ascending order and deduplicated and non-zero and be powers of two, was %v", n)
		}

		for currN < n {
			curr.Mul(api, curr, curr)
			currN *= 2
		}

		res[i] = curr
	}

	return res
}

// powerVector returns w raised to the powers contained up to n starting from
// 0. For instance, if n = 4, then it returns [1, w^-1, w^-2, w^-3]. Where w
// is the inverse of the generator of the subgroup of root of unity of order 4.
func powerVectorOfOmegaInv(n int) []frontend.Variable {

	var (
		resField = field.One()
		res      = make([]frontend.Variable, n)
		w, _     = fft.Generator(uint64(n))
	)

	w.Inverse(&w)

	for i := 0; i < n; i++ {
		res[i] = resField
		resField.Mul(&resField, &w)
	}

	return res
}

// isConstantZeroGnarkVariable returns true if the variable is a constant equal to zero
func isConstantZeroGnarkVariable(api frontend.API, p frontend.Variable) bool {
	c, isC := api.Compiler().ConstantValue(p)
	return isC && c.IsInt64() && c.Int64() == 0
}

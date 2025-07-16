package fastpoly

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
	"slices"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

// Evaluate a polynomial in lagrange basis on a gnark circuit
func InterpolateGnark(api frontend.API, poly []frontend.Variable, x frontend.Variable) frontend.Variable {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	// When the poly is of length 1 it means it is a constant polynomial and its
	// evaluation is trivial.
	if len(poly) == 1 {
		return poly[0]
	}

	n := len(poly)
	domain := fft.NewDomain(n)
	one := field.One()

	// Test that x is not a root of unity. In the other case, we would
	// have to divide by zero. In practice this constraint is not necessary
	// (because the division constraint would be non-satisfiable anyway)
	// But doing an explicit check clarifies the need.
	xN := gnarkutil.Exp(api, x, n)
	api.AssertIsDifferent(xN, 1)

	// Compute the term-wise summand of the interpolation formula.
	// This will allow the gnark solver to process the expensive
	// inverses in parallel.
	terms := make([]frontend.Variable, n)

	// omegaMinI carries the domain's inverse root of unity generator raised to
	// the power I in the following loop. It is initialized with omega**0 = 1.
	omegaI := frontend.Variable(1)

	for i := 0; i < n; i++ {

		if i > 0 {
			omegaI = api.Mul(omegaI, domain.GeneratorInv)
		}

		// If the current term is the constant zero, we continue without generating
		// constraints. As a result, 'terms' may contain nil elements. Therefore,
		// we will need need to remove them later
		if c, isC := api.Compiler().ConstantValue(poly[i]); isC && c.IsInt64() && c.Int64() == 0 {
			continue
		}

		xOmegaN := api.Mul(x, omegaI)
		terms[i] = api.Sub(xOmegaN, 1)
		// No point doing a batch inverse in a circuit
		terms[i] = api.Inverse(terms[i])
		terms[i] = api.Mul(terms[i], poly[i])
	}

	nonNilTerms := make([]frontend.Variable, 0, len(terms))
	for i := range terms {
		if terms[i] == nil {
			continue
		}

		nonNilTerms = append(nonNilTerms, terms[i])
	}

	// Then sum all the terms
	var res frontend.Variable

	switch {
	case len(nonNilTerms) == 0:
		res = 0
	case len(nonNilTerms) == 1:
		res = nonNilTerms[0]
	case len(nonNilTerms) == 2:
		res = api.Add(nonNilTerms[0], nonNilTerms[1])
	default:
		res = api.Add(nonNilTerms[0], nonNilTerms[1], nonNilTerms[2:]...)
	}

	/*
		Then multiply the res by a factor \frac{g^{1 - n}X^n -g}{n}
	*/
	factor := xN
	factor = api.Sub(factor, one)
	factor = api.Mul(factor, domain.CardinalityInv)
	res = api.Mul(res, factor)

	return res
}

// BatchInterpolateGnark evaluates a batch of polynomials in Lagrange basis
// on a gnark circuit. The evaluation point is common to all polynomials.
// The implementation relies on the barycentric interpolation formula and
// leverages
func BatchInterpolateGnark(api frontend.API, polys [][]frontend.Variable, x frontend.Variable) []frontend.Variable {

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

func BatchInterpolateGnarkExt(api gnarkfext.API, polys [][]gnarkfext.Variable, x gnarkfext.Variable) []gnarkfext.Variable {

	if len(polys) == 0 {
		return []gnarkfext.Variable{}
	}

	var (
		sizes            = listOfDifferentSizesExt(polys)
		maxSize          = sizes[len(sizes)-1]
		xNs              = raiseToPowersOfTwosExt(api.Inner, x, sizes)
		powersOfOmegaInv = powerVectorOfOmegaInvExt(maxSize)

		// innerProductTerms lists a sequence of term computed as
		// \frac{1}{x \omega^{-i} - 1}. These will be used to compute
		// an inner-product between the polys to obtain an unscaled
		// result. The unscaled result will then be scaled by (x^n - 1)/n
		// to obtain the result of the Lagrange interpolation.
		innerProductTerms = make([]gnarkfext.Variable, maxSize)

		// scalingTerms lists a sequence of term computed as (x^n - 1)/n
		// where n stands for all the different sizes of the polys in
		// increasing order. (as found in "sizes"). The used to derive
		// the final result of the evaluations.
		scalingTerms = make([]gnarkfext.Variable, len(sizes))

		// res stores the final result of the interpolation
		res = make([]gnarkfext.Variable, len(polys))
	)

	for i := range innerProductTerms {
		innerProductTerms[i] = api.Mul(powersOfOmegaInv[i], x)
		innerProductTerms[i] = api.Sub(innerProductTerms[i], gnarkfext.NewFromBase(field.One()))
		innerProductTerms[i] = api.Inverse(innerProductTerms[i])
	}

	for i, n := range sizes {
		var nField fext.Element
		nField.SetInt64(int64(n))
		scalingTerms[i] = api.Sub(xNs[i], gnarkfext.NewFromBase(field.One()))
		scalingTerms[i] = api.Div(scalingTerms[i], gnarkfext.ExtToVariable(nField))
	}

	for i := range polys {

		var (
			poly          = polys[i]
			n             = len(poly)
			scalingFactor gnarkfext.Variable
		)

		for j := range sizes {
			if sizes[j] == n {
				scalingFactor = scalingTerms[j]
			}
		}

		// That would be completely unexpectedly wrong if that happened since
		// the sizes are deduplicated and sorted from the sizes found in poly
		// in the first place.
		if scalingFactor.A0 == nil || scalingFactor.A1 == nil {
			utils.Panic("could not find scaling factor for poly of size %v", n)
		}

		yUnscaled := frontend.Variable(0)
		yUnscaledExt := gnarkfext.NewFromBase(yUnscaled)

		for k := 0; k < n; k++ {

			// This optimization saves constraints when dealing with a polynomial
			// that is precomputed and raised as a Proof object in the wizard. When
			// that happens, it will contain constant values (very often actually)
			// and when they contain zeroes we can just skip the related
			// computation. The saving was not negligible in practice.
			if isConstantZeroGnarkVariableExt(api, poly[k]) {
				continue
			}

			tmp := api.Mul(poly[k], innerProductTerms[k*maxSize/n])
			yUnscaledExt = api.Add(tmp, yUnscaledExt)
		}

		res[i] = api.Mul(yUnscaledExt, scalingFactor)
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

func listOfDifferentSizesExt(polys [][]gnarkfext.Variable) []int {

	sizes := make([]int, 0, len(polys))
	for _, poly := range polys {
		sizes = append(sizes, len(poly))
	}

	slices.Sort(sizes)
	sizes = slices.Compact(sizes)
	sizes = slices.Clip(sizes)
	return sizes
}

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

// raiseToPowersOfTwos returns x raised to the powers contained in ns. For
// instance, if ns = [2, 4, 8], then it returns [x, x^2, x^4, x^8]. It assumes
// that ns is sorted in ascending order and deduplicated and non-zero and are
// powers of two.
func raiseToPowersOfTwosExt(api frontend.API, x gnarkfext.Variable, ns []int) []gnarkfext.Variable {

	outerAPI := gnarkfext.API{Inner: api}

	var (
		res   = make([]gnarkfext.Variable, len(ns))
		curr  = x
		currN = 1
	)

	for i, n := range ns {

		if n <= 0 || !utils.IsPowerOfTwo(n) {
			utils.Panic("ns should be sorted in ascending order and deduplicated and non-zero and be powers of two, was %v", n)
		}

		for currN < n {
			curr = outerAPI.Mul(curr, curr)
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
		w        = fft.GetOmega(n)
	)

	w.Inverse(&w)

	for i := 0; i < n; i++ {
		res[i] = resField
		resField.Mul(&resField, &w)
	}

	return res
}

func powerVectorOfOmegaInvExt(n int) []gnarkfext.Variable {

	var (
		resField = fext.One()
		res      = make([]gnarkfext.Variable, n)
		w        = fft.GetOmega(n)
	)

	w.Inverse(&w)

	for i := 0; i < n; i++ {
		res[i] = gnarkfext.ExtToVariable(resField)
		resField.MulByBase(&resField, &w)
	}

	return res
}

// isConstantZeroGnarkVariable returns true if the variable is a constant equal to zero
func isConstantZeroGnarkVariable(api frontend.API, p frontend.Variable) bool {
	c, isC := api.Compiler().ConstantValue(p)
	return isC && c.IsInt64() && c.Int64() == 0
}

func isConstantZeroGnarkVariableExt(api gnarkfext.API, p gnarkfext.Variable) bool {
	return isConstantZeroGnarkVariable(api.Inner, p.A0) && isConstantZeroGnarkVariable(api.Inner, p.A1)
}

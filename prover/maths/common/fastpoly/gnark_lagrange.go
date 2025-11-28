package fastpoly

import (
	"slices"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

// EvaluateLagrangeGnarkMixed a polynomial in lagrange basis on a gnark circuit
func EvaluateLagrangeGnarkMixed(api frontend.API, poly []zk.WrappedVariable, x gnarkfext.E4Gen) gnarkfext.E4Gen {

	e4Api, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	size := len(poly)
	omega, err := fft.Generator(uint64(size))
	if err != nil {
		// TODO handle that properly
		panic(err)
	}

	dens := make([]gnarkfext.E4Gen, size)

	omega.Inverse(&omega)
	wInvOmega := gnarkfext.NewE4GenFromBase(omega.String())
	dens[0] = x
	for i := 1; i < size; i++ {
		dens[i] = *e4Api.Mul(&dens[i-1], &wInvOmega)
	}

	wOne := gnarkfext.NewE4GenFromBase(1)
	for i := 0; i < size; i++ {
		dens[i] = *e4Api.Sub(&dens[i], &wOne)
		dens[i] = *e4Api.Inverse(&dens[i])
	}

	res := gnarkfext.NewE4GenFromBase(0)
	for i := 0; i < size; i++ {
		tmp := e4Api.MulByFp(&dens[i], poly[i])
		res = *e4Api.Add(&res, tmp)
	}

	var tmp gnarkfext.E4Gen
	tmp = gnarkutil.ExpExt(api, x, size)
	tmp = *e4Api.Sub(&tmp, &wOne) // xⁿ-1
	var invSize field.Element
	invSize.SetUint64(uint64(size)).Inverse(&invSize)
	wInvSize := zk.ValueOf(invSize.String())
	tmp = *e4Api.MulByFp(&tmp, wInvSize)
	res = *e4Api.Mul(&res, &tmp)

	return res

}

// EvaluateLagrangeGnark a polynomial in lagrange basis on a gnark circuit
func EvaluateLagrangeGnark(api frontend.API, poly []zk.WrappedVariable, x zk.WrappedVariable) zk.WrappedVariable {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

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
	wAccW := zk.ValueOf(accw.String())
	dens := make([]zk.WrappedVariable, size) // [x-1, x-ω, x-ω², ...]
	for i := 0; i < size; i++ {
		dens[i] = apiGen.Sub(x, wAccW)
		accw.Mul(&accw, &omega)
	}
	invdens := make([]zk.WrappedVariable, size) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	for i := 0; i < size; i++ {
		invdens[i] = apiGen.Inverse(dens[i])
	}

	var tmp zk.WrappedVariable
	wOne := zk.ValueOf(1)
	tmp = gnarkutil.Exp(api, x, size)
	tmp = apiGen.Sub(tmp, wOne) // xⁿ-1
	var invSize field.Element
	invSize.SetUint64(uint64(size))
	invSize.Inverse(&invSize)
	li := zk.ValueOf(invSize.String())
	li = apiGen.Mul(tmp, li) // 1/n * (xⁿ-1)

	res := zk.ValueOf(0)
	wOmega := zk.ValueOf(omega.String())
	for i := 0; i < size; i++ {
		li = apiGen.Mul(li, invdens[i])
		tmp = apiGen.Mul(li, poly[i]) // pᵢ *  ωⁱ/n * ( xⁿ-1)/(x-ωⁱ)
		res = apiGen.Add(res, tmp)
		li = apiGen.Mul(li, dens[i])
		li = apiGen.Mul(li, wOmega)
	}

	return res

}

// BatchEvaluateLagrangeGnark evaluates a batch of polynomials in Lagrange basis
// on a gnark circuit. The evaluation point is common to all polynomials.
// The implementation relies on the barycentric interpolation formula and
// leverages
func BatchEvaluateLagrangeGnark(api frontend.API, polys [][]zk.WrappedVariable, x zk.WrappedVariable) []zk.WrappedVariable {

	if len(polys) == 0 {
		return []zk.WrappedVariable{}
	}

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
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
		innerProductTerms = make([]zk.WrappedVariable, maxSize)

		// scalingTerms lists a sequence of term computed as (x^n - 1)/n
		// where n stands for all the different sizes of the polys in
		// increasing order. (as found in "sizes"). The used to derive
		// the final result of the evaluations.
		scalingTerms = make([]zk.WrappedVariable, len(sizes))

		// res stores the final result of the interpolation
		res = make([]zk.WrappedVariable, len(polys))
	)

	wOne := zk.ValueOf(1)
	for i := range innerProductTerms {
		innerProductTerms[i] = apiGen.Mul(powersOfOmegaInv[i], x)
		innerProductTerms[i] = apiGen.Sub(innerProductTerms[i], wOne)
		innerProductTerms[i] = apiGen.Inverse(innerProductTerms[i])
	}

	for i, n := range sizes {
		var nField field.Element
		nField.SetInt64(int64(n))
		scalingTerms[i] = apiGen.Sub(xNs[i], wOne)
		wn := zk.ValueOf(n)
		scalingTerms[i] = apiGen.Div(scalingTerms[i], wn)
	}

	for i := range polys {

		var (
			poly          = polys[i]
			n             = len(poly)
			scalingFactor zk.WrappedVariable
		)

		for j := range sizes {
			if sizes[j] == n {
				scalingFactor = scalingTerms[j]
			}
		}

		// That would be completely unexpectedly wrong if that happened since
		// the sizes are deduplicated and sorted from the sizes found in poly
		// in the first place.
		if scalingFactor.V == nil && len(scalingFactor.EV.Limbs) == 0 {
			utils.Panic("could not find scaling factor for poly of size %v", n)
		}

		yUnscaled := zk.ValueOf(0)

		for k := 0; k < n; k++ {

			// This optimization saves constraints when dealing with a polynomial
			// that is precomputed and raised as a Proof object in the wizard. When
			// that happens, it will contain constant values (very often actually)
			// and when they contain zeroes we can just skip the related
			// computation. The saving was not negligible in practice.
			if isConstantZeroGnarkVariable(api, poly[k]) {
				continue
			}

			tmp := apiGen.Mul(poly[k], innerProductTerms[k*maxSize/n])
			yUnscaled = apiGen.Add(tmp, yUnscaled)
		}

		res[i] = apiGen.Mul(yUnscaled, scalingFactor)
	}

	return res
}

// BatchEvaluateLagrangeGnarkMixed evaluates a batch of polynomials in Lagrange basis
// on a gnark circuit. The evaluation point is common to all polynomials.
// The implementation relies on the barycentric interpolation formula and
// leverages
func BatchEvaluateLagrangeGnarkMixed(api frontend.API, polys [][]zk.WrappedVariable, x gnarkfext.E4Gen) []gnarkfext.E4Gen {
	if len(polys) == 0 {
		return []gnarkfext.E4Gen{}
	}

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	var (
		sizes            = listOfDifferentSizes(polys)
		maxSize          = sizes[len(sizes)-1]
		xNs              = raiseToPowersOfTwosExt(api, x, sizes)
		powersOfOmegaInv = powerVectorOfOmegaInv(maxSize)

		// innerProductTerms lists a sequence of term computed as
		// \frac{1}{x \omega^{-i} - 1}. These will be used to compute
		// an inner-product between the polys to obtain an unscaled
		// result. The unscaled result will then be scaled by (x^n - 1)/n
		// to obtain the result of the Lagrange interpolation.
		innerProductTerms = make([]gnarkfext.E4Gen, maxSize)

		// scalingTerms lists a sequence of term computed as (x^n - 1)/n
		// where n stands for all the different sizes of the polys in
		// increasing order. (as found in "sizes"). The used to derive
		// the final result of the evaluations.
		scalingTerms = make([]gnarkfext.E4Gen, len(sizes))

		// res stores the final result of the interpolation
		res = make([]gnarkfext.E4Gen, len(polys))
	)
	e4one := ext4.One()
	for i := range innerProductTerms {
		innerProductTerms[i] = *ext4.MulByFp(&x, powersOfOmegaInv[i])
		innerProductTerms[i] = *ext4.Sub(&innerProductTerms[i], e4one)
		innerProductTerms[i] = *ext4.Inverse(&innerProductTerms[i])
	}

	for i, n := range sizes {
		var nField field.Element
		nField.SetInt64(int64(n))
		wn := zk.ValueOf(n)
		scalingTerms[i] = *ext4.Sub(&xNs[i], e4one)
		scalingTerms[i] = *ext4.DivByBase(&xNs[i], wn)
	}

	for i := range polys {

		var (
			poly          = polys[i]
			n             = len(poly)
			scalingFactor gnarkfext.E4Gen
			yUnscaled     gnarkfext.E4Gen
			tmp           gnarkfext.E4Gen
		)

		for j := range sizes {
			if sizes[j] == n {
				scalingFactor = scalingTerms[j]
			}
		}

		// That would be completely unexpectedly wrong if that happened since
		// the sizes are deduplicated and sorted from the sizes found in poly
		// in the first place.
		if scalingFactor.B0.A0.IsEmpty() {
			utils.Panic("could not find scaling factor for poly of size %v", n)
		}

		yUnscaled = *ext4.Zero()

		for k := 0; k < n; k++ {

			// This optimization saves constraints when dealing with a polynomial
			// that is precomputed and raised as a Proof object in the wizard. When
			// that happens, it will contain constant values (very often actually)
			// and when they contain zeroes we can just skip the related
			// computation. The saving was not negligible in practice.
			if isConstantZeroGnarkVariable(api, poly[k]) {
				continue
			}

			tmp = *ext4.MulByFp(&innerProductTerms[k*maxSize/n], poly[k])
			yUnscaled = *ext4.Add(&tmp, &yUnscaled)
		}

		res[i] = *ext4.Mul(&yUnscaled, &scalingFactor)
	}

	return res
}

// listOfDifferentSizes returns the list of sizes of the different polys
// vectors. The returned slices is deduplicated and sorted in ascending order.
func listOfDifferentSizes(polys [][]zk.WrappedVariable) []int {

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
func raiseToPowersOfTwos(api frontend.API, x zk.WrappedVariable, ns []int) []zk.WrappedVariable {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	var (
		res   = make([]zk.WrappedVariable, len(ns))
		curr  = x
		currN = 1
	)

	for i, n := range ns {

		if n <= 0 || !utils.IsPowerOfTwo(n) {
			utils.Panic("ns should be sorted in ascending order and deduplicated and non-zero and be powers of two, was %v", n)
		}

		for currN < n {
			curr = apiGen.Mul(curr, curr)
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
func raiseToPowersOfTwosExt(api frontend.API, x gnarkfext.E4Gen, ns []int) []gnarkfext.E4Gen {

	var (
		res   = make([]gnarkfext.E4Gen, len(ns))
		curr  = x
		currN = 1
	)

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	for i, n := range ns {

		if n <= 0 || !utils.IsPowerOfTwo(n) {
			utils.Panic("ns should be sorted in ascending order and deduplicated and non-zero and be powers of two, was %v", n)
		}

		for currN < n {
			curr = *ext4.Mul(&curr, &curr)
			currN *= 2
		}

		res[i] = curr
	}

	return res
}

// powerVector returns w raised to the powers contained up to n starting from
// 0. For instance, if n = 4, then it returns [1, w^-1, w^-2, w^-3]. Where w
// is the inverse of the generator of the subgroup of root of unity of order 4.
func powerVectorOfOmegaInv(n int) []zk.WrappedVariable {

	var (
		resField = field.One()
		res      = make([]zk.WrappedVariable, n)
		w, _     = fft.Generator(uint64(n))
	)

	w.Inverse(&w)

	for i := 0; i < n; i++ {
		res[i] = zk.ValueOf(resField.String())
		resField.Mul(&resField, &w)
	}

	return res
}

// isConstantZeroGnarkVariable returns true if the variable is a constant equal to zero
func isConstantZeroGnarkVariable(api frontend.API, p zk.WrappedVariable) bool {
	c, isC := api.Compiler().ConstantValue(p.AsNative())
	return isC && c.IsInt64() && c.Int64() == 0
}

package fastpolyext

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext" // Assuming gnarkfext has the Exp function
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// EvaluateLagrangeGnark evaluates a polynomial in lagrange basis on a gnark circuit
func EvaluateLagrangeGnark[T zk.Element](api frontend.API, poly []gnarkfext.E4Gen[T], x gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T] {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	size := len(poly)
	omega, err := fft.Generator(uint64(size))
	if err != nil {
		// TODO handle that properly
		panic(err)
	}

	e4Api, err := gnarkfext.NewExt4[T](api)
	if err != nil {
		panic(err)
	}

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	var accw, one field.Element
	one.SetOne()
	accw.SetOne()
	dens := make([]gnarkfext.E4Gen[T], size) // [x-1, x-ω, x-ω², ...]
	for i := 0; i < size; i++ {
		dens[i] = x
		dens[i].B0.A0 = *apiGen.Sub(&x.B0.A0, zk.ValueOf[T](accw)) // accw lives in the base field
		accw.Mul(&accw, &omega)
	}
	invdens := make([]gnarkfext.E4Gen[T], size) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	for i := 0; i < size; i++ {
		invdens[i] = *e4Api.Inverse(&dens[i])
	}

	tmp := e4Api.Exp(&x, big.NewInt(int64(size)))
	tmp.B0.A0 = *apiGen.Sub(&tmp.B0.A0, zk.ValueOf[T](one)) // xⁿ-1
	li := apiGen.Inverse(zk.ValueOf[T](size))
	liExt := gnarkfext.Lift[T](li)
	liExt = e4Api.Mul(tmp, liExt) // 1/n * (xⁿ-1)

	res := gnarkfext.NewFromBase[T](0)
	for i := 0; i < size; i++ {
		liExt = e4Api.Mul(liExt, &invdens[i])
		tmp = e4Api.Mul(liExt, &poly[i])
		res = e4Api.Add(res, tmp)
		liExt = e4Api.Mul(liExt, &dens[i])
		liExt = e4Api.MulByFp(liExt, zk.ValueOf[T](omega))
	}

	return *res

}

// BatchEvaluateLagrangeGnark evaluates a batch of polynomials in Lagrange basis
// on a gnark circuit. The evaluation point is common to all polynomials.
// The implementation relies on the barycentric interpolation formula and
// leverages
// func BatchEvaluateLagrangeGnark(api frontend.API, polys [][]gnarkfext.E4Gen[T], x gnarkfext.E4Gen[T]) []gnarkfext.E4Gen[T] {

// 	if len(polys) == 0 {
// 		return []gnarkfext.E4Gen[T]{}
// 	}

// 	var (
// 		sizes            = listOfDifferentSizes(polys)
// 		maxSize          = sizes[len(sizes)-1]
// 		xNs              = raiseToPowersOfTwosExt(api, x, sizes)
// 		powersOfOmegaInv = powerVectorOfOmegaInv[T](maxSize)
// 		one              gnarkfext.E4Gen[T]

// 		// innerProductTerms lists a sequence of term computed as
// 		// \frac{1}{x \omega^{-i} - 1}. These will be used to compute
// 		// an inner-product between the polys to obtain an unscaled
// 		// result. The unscaled result will then be scaled by (x^n - 1)/n
// 		// to obtain the result of the Lagrange interpolation.
// 		innerProductTerms = make([]gnarkfext.E4Gen[T], maxSize)

// 		// scalingTerms lists a sequence of term computed as (x^n - 1)/n
// 		// where n stands for all the different sizes of the polys in
// 		// increasing order. (as found in "sizes"). The used to derive
// 		// the final result of the evaluations.
// 		scalingTerms = make([]gnarkfext.E4Gen[T], len(sizes))

// 		// res stores the final result of the interpolation
// 		res = make([]gnarkfext.E4Gen[T], len(polys))
// 	)
// 	one.SetOne()
// 	for i := range innerProductTerms {
// 		innerProductTerms[i].MulByFp(api, x, powersOfOmegaInv[i])
// 		innerProductTerms[i].Sub(api, innerProductTerms[i], one)
// 		innerProductTerms[i].Inverse(api, innerProductTerms[i])
// 	}

// 	for i, n := range sizes {
// 		var nField field.Element
// 		nField.SetInt64(int64(n))
// 		scalingTerms[i].Sub(api, xNs[i], one)
// 		scalingTerms[i].DivByBase(api, scalingTerms[i], n)
// 	}

// 	for i := range polys {

// 		var (
// 			poly          = polys[i]
// 			n             = len(poly)
// 			scalingFactor gnarkfext.E4Gen[T]
// 			yUnscaled     gnarkfext.E4Gen[T]
// 			tmp           gnarkfext.E4Gen[T]
// 		)

// 		for j := range sizes {
// 			if sizes[j] == n {
// 				scalingFactor = scalingTerms[j]
// 			}
// 		}

// 		// That would be completely unexpectedly wrong if that happened since
// 		// the sizes are deduplicated and sorted from the sizes found in poly
// 		// in the first place.
// 		if scalingFactor.B0.A0 == nil {
// 			utils.Panic("could not find scaling factor for poly of size %v", n)
// 		}

// 		yUnscaled.SetZero()

// 		for k := 0; k < n; k++ {

// 			// This optimization saves constraints when dealing with a polynomial
// 			// that is precomputed and raised as a Proof object in the wizard. When
// 			// that happens, it will contain constant values (very often actually)
// 			// and when they contain zeroes we can just skip the related
// 			// computation. The saving was not negligible in practice.
// 			if isConstantZeroGnarkVariable(api, poly[k]) {
// 				continue
// 			}

// 			tmp.Mul(api, innerProductTerms[k*maxSize/n], poly[k])
// 			yUnscaled.Add(api, tmp, yUnscaled)
// 		}

// 		res[i].Mul(api, yUnscaled, scalingFactor)
// 	}

// 	return res
// }

// listOfDifferentSizes returns the list of sizes of the different polys
// vectors. The returned slices is deduplicated and sorted in ascending order.
// func listOfDifferentSizes(polys [][]gnarkfext.E4Gen[T]) []int {

// 	sizes := make([]int, 0, len(polys))
// 	for _, poly := range polys {
// 		sizes = append(sizes, len(poly))
// 	}

// 	slices.Sort(sizes)
// 	sizes = slices.Compact(sizes)
// 	sizes = slices.Clip(sizes)
// 	return sizes
// }

// raiseToPowersOfTwosExt returns x raised to the powers contained in ns. For
// instance, if ns = [2, 4, 8], then it returns [x, x^2, x^4, x^8]. It assumes
// that ns is sorted in ascending order and deduplicated and non-zero and are
// powers of two.
// func raiseToPowersOfTwosExt(api frontend.API, x gnarkfext.E4Gen[T], ns []int) []gnarkfext.E4Gen[T] {

// 	var (
// 		res   = make([]gnarkfext.E4Gen[T], len(ns))
// 		curr  = x
// 		currN = 1
// 	)

// 	for i, n := range ns {

// 		if n <= 0 || !utils.IsPowerOfTwo(n) {
// 			utils.Panic("ns should be sorted in ascending order and deduplicated and non-zero and be powers of two, was %v", n)
// 		}

// 		for currN < n {
// 			curr.Mul(api, curr, curr)
// 			currN *= 2
// 		}

// 		res[i] = curr
// 	}

// 	return res
// }

// powerVector returns w raised to the powers contained up to n starting from
// 0. For instance, if n = 4, then it returns [1, w^-1, w^-2, w^-3]. Where w
// is the inverse of the generator of the subgroup of root of unity of order 4.
func powerVectorOfOmegaInv[T zk.Element](n int) []T {

	var (
		resField = field.One()
		res      = make([]T, n)
		w, _     = fft.Generator(uint64(n))
	)

	w.Inverse(&w)

	for i := 0; i < n; i++ {
		res[i] = *zk.ValueOf[T](resField)
		resField.Mul(&resField, &w)
	}

	return res
}

// isConstantZeroGnarkVariable returns true if the variable is a constant equal to zero
func isConstantZeroGnarkVariable[T zk.Element](api frontend.API, p T) bool {
	c, isC := api.Compiler().ConstantValue(p)
	return isC && c.IsInt64() && c.Int64() == 0
}

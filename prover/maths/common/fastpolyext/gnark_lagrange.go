package fastpolyext

import (
	"math/big"
	"slices"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext" // Assuming gnarkfext has the Exp function
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

// EvaluateLagrangeGnark evaluates a polynomial in lagrange basis on a gnark circuit
func EvaluateLagrangeGnark(api frontend.API, poly []gnarkfext.E4Gen, x gnarkfext.E4Gen) gnarkfext.E4Gen {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	e4Api, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
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

	wOne := *e4Api.One()
	for i := 0; i < size; i++ {
		dens[i] = *e4Api.Sub(&dens[i], &wOne)
		dens[i] = *e4Api.Inverse(&dens[i])
	}

	res := *e4Api.Zero()
	for i := 0; i < size; i++ {
		tmp := *e4Api.Mul(&dens[i], &poly[i])
		res = *e4Api.Add(&res, &tmp)
	}

	var tmp gnarkfext.E4Gen
	tmp = gnarkutil.ExpExt(api, x, size)
	tmp = *e4Api.Sub(&tmp, &wOne) // xⁿ-1
	var invSize field.Element
	invSize.SetUint64(uint64(size)).Inverse(&invSize)
	bInvSize := big.NewInt(0).SetUint64(invSize.Uint64())
	tmp = *e4Api.MulConst(&tmp, bInvSize)
	res = *e4Api.Mul(&res, &tmp)

	return res

}

// BatchEvaluateLagrangeGnark evaluates a batch of polynomials in Lagrange basis
// on a gnark circuit. The evaluation point is common to all polynomials.
// The implementation relies on the barycentric interpolation formula and
// leverages
func BatchEvaluateLagrangeGnark(api frontend.API, polys [][]gnarkfext.E4Gen, x gnarkfext.E4Gen) []gnarkfext.E4Gen {

	if len(polys) == 0 {
		return []gnarkfext.E4Gen{}
	}

	e4Api, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	sizes := listOfDifferentSizes(polys)
	maxSize := sizes[len(sizes)-1]
	xNs := raiseToPowersOfTwosExt(api, x, sizes)
	powersOfOmegaInv := powerVectorOfOmegaInv(maxSize)
	one := *e4Api.One()

	innerProductTerms := make([]gnarkfext.E4Gen, maxSize)
	scalingTerms := make([]gnarkfext.E4Gen, len(sizes))

	// res stores the final result of the interpolation
	res := make([]gnarkfext.E4Gen, len(polys))
	omegaInv := big.NewInt(0)
	for i := range innerProductTerms {
		omegaInv.SetUint64(powersOfOmegaInv[i])
		innerProductTerms[i] = *e4Api.MulConst(&x, omegaInv)
		innerProductTerms[i] = *e4Api.Sub(&innerProductTerms[i], &one)
		innerProductTerms[i] = *e4Api.Inverse(&innerProductTerms[i])
	}

	for i, n := range sizes {
		var nField field.Element
		nField.SetInt64(int64(n))
		scalingTerms[i] = *e4Api.Sub(&xNs[i], &one)
		wn := zk.ValueOf(n)
		scalingTerms[i] = *e4Api.DivByBase(&scalingTerms[i], wn)
	}

	for i := range polys {

		poly := polys[i]
		n := len(poly)
		var scalingFactor gnarkfext.E4Gen
		var tmp gnarkfext.E4Gen

		for j := range sizes {
			if sizes[j] == n {
				scalingFactor = scalingTerms[j]
			}
		}

		if scalingFactor.B0.A0.IsEmpty() {
			utils.Panic("could not find scaling factor for poly of size %v", n)
		}

		yUnscaled := *e4Api.Zero()
		for k := 0; k < n; k++ {
			// TOOD @thomas restore this optim
			// if isConstantZeroGnarkVariable(api, poly[k]) {
			// 	continue
			// }

			tmp = *e4Api.Mul(&innerProductTerms[k*maxSize/n], &poly[k])
			yUnscaled = *e4Api.Add(&tmp, &yUnscaled)
		}

		res[i] = *e4Api.Mul(&yUnscaled, &scalingFactor)
	}

	return res
}

// listOfDifferentSizes returns the list of sizes of the different polys
// vectors. The returned slices is deduplicated and sorted in ascending order.
func listOfDifferentSizes(polys [][]gnarkfext.E4Gen) []int {

	sizes := make([]int, 0, len(polys))
	for _, poly := range polys {
		sizes = append(sizes, len(poly))
	}

	slices.Sort(sizes)
	sizes = slices.Compact(sizes)
	sizes = slices.Clip(sizes)
	return sizes
}

// raiseToPowersOfTwosExt returns x raised to the powers contained in ns. For
// instance, if ns = [2, 4, 8], then it returns [x, x^2, x^4, x^8]. It assumes
// that ns is sorted in ascending order and deduplicated and non-zero and are
// powers of two.
func raiseToPowersOfTwosExt(api frontend.API, x gnarkfext.E4Gen, ns []int) []gnarkfext.E4Gen {

	res := make([]gnarkfext.E4Gen, len(ns))
	curr := x
	currN := 1

	e4Api, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	for i, n := range ns {

		if n <= 0 || !utils.IsPowerOfTwo(n) {
			utils.Panic("ns should be sorted in ascending order and deduplicated and non-zero and be powers of two, was %v", n)
		}

		for currN < n {
			curr = *e4Api.Mul(&curr, &curr)
			currN *= 2
		}

		res[i] = curr
	}

	return res
}

// powerVector returns w raised to the powers contained up to n starting from
// 0. For instance, if n = 4, then it returns [1, w^-1, w^-2, w^-3]. Where w
// is the inverse of the generator of the subgroup of root of unity of order 4.
func powerVectorOfOmegaInv(n int) []uint64 {

	var (
		resField = field.One()
		res      = make([]uint64, n)
		w, _     = fft.Generator(uint64(n))
	)

	w.Inverse(&w)

	for i := 0; i < n; i++ {
		res[i] = uint64(resField.Uint64())
		resField.Mul(&resField, &w)
	}

	return res
}

package field

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

type canExt interface {
	Lift() Ext
}

// UnivariateEvalCoeff evaluates the univariate polynomial p at x. p is
// interpreted as a polynomial in coefficient form and p[i] is the coefficient
// of x^i
func UnivariateEvalCoeff[F anyF](pol []F, x canExt) Ext {

	var res Ext
	liftedX := x.Lift()
	var _u F

	switch any(_u).(type) {
	case Fr:
		pol := *unsafeCast[[]F, []Fr](&pol)
		for i := len(pol) - 1; i >= 0; i-- {
			res.Mul(&res, &liftedX)
			res.AddFr(&res, &pol[i])
		}
		return res
	case Ext:
		pol := *unsafeCast[[]F, []Ext](&pol)
		for i := len(pol) - 1; i >= 0; i-- {
			res.Mul(&res, &liftedX)
			res.Add(&res, &pol[i])
		}
		return res
	default:
		// anyF is defined as a union type of Fr and Ext
		panic("unreachable")
	}
}

// UnivariateEvalLagrange evaluates poly in Lagrange form at point X
func UnivariateEvalLagrange[F anyF](pol []F, x_ canExt, oncoset ...bool) Ext {

	x := x_.Lift()

	if len(oncoset) > 0 && oncoset[0] {
		genFr := Fr(fft.GeneratorFullMultiplicativeGroup())
		genFr.Inverse(&genFr)
		x.MulByElement(&x, &genFr)
	}

	if !utils.IsPowerOfTwo(len(pol)) {
		utils.Panic("only support powers of two but poly has length %v", len(pol))
	}

	// Here we can call the function optimized by gnark-crypto
	var _f F
	switch any(_f).(type) {
	case Fr:
		poly := *unsafeCast[[]F, []koalabear.Element](&pol)
		x := extensions.E4(x)
		res, err := vortex.EvalBasePolyLagrange(poly, x)
		if err != nil {
			panic(err)
		}
		return Ext(res)

	case Ext:
		poly := *unsafeCast[[]F, []extensions.E4](&pol)
		x := extensions.E4(x)
		res, err := vortex.EvalFextPolyLagrange(poly, x)
		if err != nil {
			panic(err)
		}
		return Ext(res)

	default:
		// anyF is defined as a union type of Fr and Ext
		panic("unreachable")
	}
}

// UnivariateBatchEvaluateLagrange batch version of EvaluateLagrangeOnFext
func UnivariateBatchEvaluateLagrange(polys []any, x_ canExt, oncoset ...bool) []Ext {

	// The function could be made faster in case x was Fr but this case in
	// unexpected outside of testing.
	x := x_.Lift()

	// TODO should we check that the polynomials are of the same size ??
	results := make([]Ext, len(polys))

	size, notSameSizeErr := utils.AllReturnEqual(VecLen, polys)
	if notSameSizeErr != nil {
		utils.Panic("length mismatch: %v", notSameSizeErr)
	}

	if !utils.IsPowerOfTwo(size) {
		utils.Panic("only support powers of two but poly has length %v", size)
	}

	omega_, err := fft.Generator(uint64(size))
	if err != nil {
		// TODO handle that properly
		panic(err)
	}

	omega := Fr(omega_)

	domain := fft.NewDomain(uint64(size))

	if len(oncoset) > 0 && oncoset[0] {
		x.MulByElement(&x, unsafeCast[koalabear.Element, Fr](&domain.FrMultiplicativeGenInv))
	}

	lagrangeAtX := make([]Ext, size)
	var accw, one, extomega Ext
	one.SetOne()
	accw.SetOne()
	extomega.SetFr(&omega)
	dens := make([]Ext, size) // [x-1, x-ω, x-ω², ...]
	for i := 0; i < size; i++ {
		dens[i].Sub(&x, &accw)
		accw.Mul(&accw, &extomega)
	}
	invdens := BatchInvertExt(dens) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	var tmp Ext
	tmp.Exp(x, big.NewInt(int64(size))).Sub(&tmp, &one) // xⁿ-1
	lagrangeAtX[0].SetFromInt(int64(size))
	lagrangeAtX[0].Inverse(&lagrangeAtX[0])
	lagrangeAtX[0].Mul(&tmp, &lagrangeAtX[0])        // 1/n * (xⁿ-1)
	lagrangeAtX[0].Mul(&lagrangeAtX[0], &invdens[0]) // 1/n * (xⁿ-1)/(x-1)

	for i := 1; i < size; i++ {
		lagrangeAtX[i].Mul(&lagrangeAtX[i-1], &dens[i-1]).
			Mul(&lagrangeAtX[i], &invdens[i]).
			Mul(&lagrangeAtX[i], &extomega)
	}

	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {
			res := InnerProduct(lagrangeAtX, polys[k])
			results[k] = res.AsExt()
		}
	})
	return results
}

/*
Evaluation x^n - 1 for x on a coset of domain of size N

	N is the size of the larger domain
	n Is the size of the smaller domain

The result is (in theory) of size N, but it has a periodicity
of r = N/n. Thus, we only return the "r" first entries

Largely inspired from gnark's
https://github.com/ConsenSys/gnark/blob/8bc13b200cb9aa1ec74e4f4807e2e97fc8d8396f/internal/backend/bls12-377/plonk/prove.go#L734
*/
func EvalXnMinusOneOnACoset(n, N int) []Fr {
	/*
		Sanity-checks on the sizes of the elements
	*/
	if !utils.IsPowerOfTwo(n) || !utils.IsPowerOfTwo(N) {
		utils.Panic("Both n %v and N %v should be powers of two", n, N)
	}

	if N < n {
		utils.Panic("N %v is smaller than n %v", N, n)
	}

	// memoized
	nBigInt := big.NewInt(int64(n))

	res := make([]Fr, N/n)
	res[0].SetUint64(field.MultiplicativeGen)
	res[0].Exp(res[0], nBigInt)

	t_, err := fft.Generator(uint64(N))
	if err != nil {
		panic(err)
	}

	t := Fr(t_)
	t.Exp(t, nBigInt)

	for i := 1; i < N/n; i++ {
		res[i].Mul(&res[i-1], &t)
	}

	one := One()
	for i := 0; i < N/n; i++ {
		res[i].Sub(&res[i], &one)
	}

	return res
}

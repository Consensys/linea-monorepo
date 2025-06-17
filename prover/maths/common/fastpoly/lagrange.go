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
	fext.FromBase(&xExt, &x)
	res := EvaluateLagrangeMixed(poly, xExt, oncoset...)
	return res.B0.A0
}

// BatchEvaluateLagrangeMixed batch version of EvaluateLagrangeOnFext
func BatchEvaluateLagrangeMixed(polys [][]field.Element, x fext.Element, oncoset ...bool) []fext.Element {

	// TODO should we check that the polynomials are of the same size ??
	results := make([]fext.Element, len(polys))
	poly := polys[0]

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	size := len(poly)
	omega, err := fft.Generator(uint64(size))
	if err != nil {
		// TODO handle that properly
		panic(err)
	}

	domain := fft.NewDomain(uint64(len(poly)))

	if len(oncoset) > 0 && oncoset[0] {
		x.MulByElement(&x, &domain.FrMultiplicativeGenInv)
	}

	lagrangeAtX := make([]fext.Element, size)
	var accw, one, extomega fext.Element
	one.SetOne()
	accw.SetOne()
	fext.FromBase(&extomega, &omega)
	dens := make([]fext.Element, size) // [x-1, x-ω, x-ω², ...]
	for i := 0; i < size; i++ {
		dens[i].Sub(&x, &accw)
		accw.Mul(&accw, &extomega)
	}
	invdens := fext.BatchInvert(dens) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	var tmp fext.Element
	tmp.Exp(x, big.NewInt(int64(size))).Sub(&tmp, &one) // xⁿ-1
	fext.SetInt64(&lagrangeAtX[0], int64(size))
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
			results[k] = vectorext.ScalarProdByElement(lagrangeAtX, polys[k])
		}
	})
	return results
}

// BatchEvaluateLagrange batch evalute polys at x, where polys are in Lagrange basis
func BatchEvaluateLagrange(polys [][]field.Element, x field.Element, oncoset ...bool) []field.Element {
	var xExt fext.Element
	fext.FromBase(&xExt, &x)
	resExt := BatchEvaluateLagrangeMixed(polys, xExt, oncoset...)
	res := make([]field.Element, len(polys))
	for i := 0; i < len(resExt); i++ {
		res[i].Set(&resExt[i].B0.A0)
	}
	return res
}

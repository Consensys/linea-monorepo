package fastpolyext

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// EvaluateLagrange computes ∑_i L_i(x), i.e. evaluates p interpreted as a polynomial in Lagrange form
func EvaluateLagrange(poly []fext.Element, x fext.Element, oncoset ...bool) fext.Element {

	if len(oncoset) > 0 && oncoset[0] {
		genFr := fft.GeneratorFullMultiplicativeGroup()
		genFr.Inverse(&genFr)
		x.MulByElement(&x, &genFr)
	}

	res, err := vortex.EvalFextPolyLagrange(poly, x)
	// TODO handle that properly
	if err != nil {
		panic(err)
	}

	return res

}

// BatchEvaluateLagrangeOnFext batch version of EvaluateLagrangeOnFext
func BatchEvaluateLagrange(polys [][]fext.Element, x fext.Element, oncoset ...bool) []fext.Element {

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

	if len(oncoset) > 0 && oncoset[0] {
		genFr := fft.GeneratorFullMultiplicativeGroup()
		genFr.Inverse(&genFr)
		x.MulByElement(&x, &genFr)
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
			results[k] = vectorext.ScalarProd(lagrangeAtX, polys[k])
		}
	})
	return results
}

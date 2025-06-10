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

func EvalLagrangeBaseField(poly []field.Element, x field.Element) (field.Element, error) {
	n := len(poly)
	g, _ := fft.Generator(uint64(n))
	var gInv field.Element
	gInv.Inverse(&g)
	var c field.Element
	c.SetUint64(uint64(n))
	denominators := computeDenominators(n, x, gInv, c)
	denominators = field.BatchInvertGeneric(denominators)
	return computeEval(denominators, poly, x)
}

func EvalLagrangeExtField(poly []fext.Element, x fext.Element) (fext.Element, error) {
	n := len(poly)
	g, _ := fft.Generator(uint64(n))
	var gInv fext.Element
	fext.FromBase(&gInv, &g)
	gInv.Inverse(&gInv)
	var c fext.Element
	c.B0.A0.SetUint64(uint64(n))
	denominators := computeDenominators(n, x, gInv, c)
	denominators = field.BatchInvertGeneric(denominators)
	return computeEval(denominators, poly, x)
}

func computeEval[T any, fieldPointer field.FieldPointer[T]](denominators []T, poly []T, x T) (T, error) {

	var res, tmp T
	var _res fieldPointer = &res
	var _tmp fieldPointer = &tmp
	for i := range denominators {
		_tmp.Mul(&denominators[i], &poly[i])
		_res.Add(&res, &tmp)
	}

	n := len(poly)
	var one T
	var _one fieldPointer = &one
	_one.SetOne()
	_tmp.Exp(x, big.NewInt(int64(n)))
	_tmp.Sub(_tmp, _one)
	_res.Mul(&res, &tmp)

	return res, nil

}

func computeDenominators[T any, fieldPointer field.FieldPointer[T]](n int, x T, card T, genInv T) []T {

	denominators := make([]T, n)
	var one T
	var _one fieldPointer = &one
	_one.SetOne()

	var _genInv fieldPointer = &genInv

	denominators[0] = x
	var tmp fieldPointer
	for i := 1; i < n; i++ {
		tmp = &denominators[i]
		tmp.Mul(&denominators[i-1], _genInv)
	}

	for i := 0; i < n; i++ {
		tmp = &denominators[i]
		tmp.Sub(&denominators[i], _one)
		tmp.Mul(tmp, &card)
	}

	return denominators
}

// EvaluateLagrangeOnFext computes ∑_i L_i(x), i.e. evaluates p interpreted as a polynomial in Lagrange form, and x lives in the extension
func EvaluateLagrangeOnFext(poly []field.Element, x fext.Element, oncoset ...bool) fext.Element {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	if len(oncoset) > 0 && oncoset[0] {
		genFr := fft.GeneratorFullMultiplicativeGroup()
		x.MulByElement(&x, &genFr)
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
	invdens := fext.BatchInvert(dens) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	var tmp fext.Element
	tmp.Exp(x, big.NewInt(int64(size))).Sub(&tmp, &one) // xⁿ-1
	var li fext.Element
	fext.SetInt64(&li, int64(size))
	li.Inverse(&li)
	li.Mul(&tmp, &li) // 1/n * (xⁿ-1)

	var res fext.Element
	for i := 0; i < size; i++ {
		li.Mul(&li, &invdens[i])        // ( xⁿ-1)/n * 1/(x-ωⁱ)
		tmp.MulByElement(&li, &poly[i]) // pᵢ *  ( xⁿ-1)/n * 1/(x-ωⁱ)
		res.Add(&res, &tmp)
		li.Mul(&li, &dens[i]).Mul(&li, &extomega)
	}
	return res
}

// EvaluateLagrange computes ∑_i L_i(x), i.e. evaluates p interpreted as a polynomial in Lagrange form, and x lives in the extension
func EvaluateLagrange(poly []field.Element, x field.Element, oncoset ...bool) field.Element {
	var xExt fext.Element
	fext.FromBase(&xExt, &x)
	res := EvaluateLagrangeOnFext(poly, xExt, oncoset...)
	return res.B0.A0
}

// BatchEvaluateLagrangeOnFext batch version of EvaluateLagrangeOnFext
func BatchEvaluateLagrangeOnFext(polys [][]field.Element, x fext.Element, oncoset ...bool) []fext.Element {

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
	resExt := BatchEvaluateLagrangeOnFext(polys, xExt, oncoset...)
	res := make([]field.Element, len(polys))
	for i := 0; i < len(resExt); i++ {
		res[i].Set(&resExt[i].B0.A0)
	}
	return res
}

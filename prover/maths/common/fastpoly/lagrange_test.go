package fastpoly

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/require"
)

func randomPoly(size int) []field.Element {
	res := make([]field.Element, size)
	for i := 0; i < size; i++ {
		res[i].SetRandom()
	}
	return res
}

func TestEvaluateLagrangeFext(t *testing.T) {

	size := 64
	domain := fft.NewDomain(uint64(size), fft.WithoutPrecompute())
	p := randomPoly(size)
	pLagrange := make([]field.Element, size)

	var x fext.Element
	x.SetRandom()
	u := poly.EvalMixed(p, x)

	/*
		Test without coset
	*/
	copy(pLagrange, p)
	domain.FFT(pLagrange, fft.DIF)
	fft.BitReverse(pLagrange)
	v := EvaluateLagrangeMixed(pLagrange, x)
	require.Equal(t, u.String(), v.String())

	/*
		Test with coset
	*/
	copy(pLagrange, p)
	domain.FFT(pLagrange, fft.DIF, fft.OnCoset())
	fft.BitReverse(pLagrange)
	vOnCoset := EvaluateLagrangeMixed(pLagrange, x, true)
	require.Equal(t, u.String(), vOnCoset.String())
}

func TestBatchEvaluateLagrangeOnFext(t *testing.T) {

	sizePoly := 64
	nbPoly := 20

	// sample a bunch of polynomials
	polys := make([][]field.Element, nbPoly)
	for i := 0; i < nbPoly; i++ {
		polys[i] = randomPoly(sizePoly)
	}

	// sample a random point
	x := fext.RandomElement()

	Eval := make([]fext.Element, nbPoly)
	for i := 0; i < nbPoly; i++ {
		Eval[i] = poly.EvalMixed(polys[i], x)
	}
	d := fft.NewDomain(uint64(sizePoly), fft.WithoutPrecompute())

	/*
		Test without coset
	*/
	onRoots := make([][]field.Element, nbPoly)

	for i := 0; i < nbPoly; i++ {
		onRoots[i] = vector.DeepCopy(polys[i])

		d.FFT(onRoots[i], fft.DIF)
		fft.BitReverse(onRoots[i])
	}

	// compute lagrange eval
	lagEvalExt := BatchEvaluateLagrangeMixed(onRoots, x)

	// check the result
	for i := 0; i < nbPoly; i++ {
		require.Equal(t, Eval[i].String(), lagEvalExt[i].String())
	}

	/*
		Test with coset
	*/
	onCosets := make([][]field.Element, nbPoly)

	for i := 0; i < nbPoly; i++ {
		onCosets[i] = vector.DeepCopy(polys[i])

		d.FFT(onCosets[i], fft.DIF, fft.OnCoset())
		fft.BitReverse(onCosets[i])
	}

	// compute lagrange eval
	lagEvalExtcoset := BatchEvaluateLagrangeMixed(onCosets, x, true)

	// check the result
	for i := 0; i < nbPoly; i++ {
		require.Equal(t, Eval[i].String(), lagEvalExtcoset[i].String())
	}
}

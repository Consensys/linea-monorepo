package fastpolyext

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/stretchr/testify/require"
)

func randomPoly(size int) []fext.Element {
	res := make([]fext.Element, size)
	for i := 0; i < size; i++ {
		res[i].SetRandom()
	}
	return res
}

func TestEvaluateLagrange(t *testing.T) {

	size := 16
	domain := fft.NewDomain(uint64(size), fft.WithoutPrecompute())
	p := randomPoly(size)
	pLagrange := make([]fext.Element, size)

	var x fext.Element
	x.SetRandom()
	u := polyext.Eval(p, x)

	/*
		Test without coset
	*/
	copy(pLagrange, p)
	domain.FFTExt(pLagrange, fft.DIF)
	fft.BitReverse(pLagrange)
	v := EvaluateLagrange(pLagrange, x)
	require.Equal(t, u.String(), v.String())

	/*
		Test with coset
	*/
	copy(pLagrange, p)
	domain.FFTExt(pLagrange, fft.DIF, fft.OnCoset())
	fft.BitReverse(pLagrange)
	vOnCoset := EvaluateLagrange(pLagrange, x, true)
	require.Equal(t, u.String(), vOnCoset.String())
}

func TestBatchEvaluateLagrangeOnFext(t *testing.T) {

	sizePoly := 16
	nbPoly := 20

	// sample a bunch of polynomials
	polys := make([][]fext.Element, nbPoly)
	for i := 0; i < nbPoly; i++ {
		polys[i] = randomPoly(sizePoly)
	}

	// sample a random point
	x := fext.RandomElement()

	Eval := make([]fext.Element, nbPoly)
	for i := 0; i < nbPoly; i++ {
		Eval[i] = polyext.Eval(polys[i], x)
	}
	d := fft.NewDomain(uint64(sizePoly), fft.WithoutPrecompute())

	/*
		Test without coset
	*/
	onRoots := make([][]fext.Element, nbPoly)

	for i := 0; i < nbPoly; i++ {
		onRoots[i] = vectorext.DeepCopy(polys[i])

		d.FFTExt(onRoots[i], fft.DIF)
		fft.BitReverse(onRoots[i])
	}

	// compute lagrange eval
	lagEvalExt := BatchEvaluateLagrange(onRoots, x)

	// check the result
	for i := 0; i < nbPoly; i++ {
		require.Equal(t, Eval[i].String(), lagEvalExt[i].String())
	}

	/*
		Test with coset
	*/
	onCosets := make([][]fext.Element, nbPoly)

	for i := 0; i < nbPoly; i++ {
		onCosets[i] = vectorext.DeepCopy(polys[i])

		d.FFTExt(onCosets[i], fft.DIF, fft.OnCoset())
		fft.BitReverse(onCosets[i])
	}

	// compute lagrange eval
	lagEvalExtcoset := BatchEvaluateLagrange(onCosets, x, true)

	// check the result
	for i := 0; i < nbPoly; i++ {
		require.Equal(t, Eval[i].String(), lagEvalExtcoset[i].String())
	}
}

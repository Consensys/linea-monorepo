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

func TestEvaluateLagrange(t *testing.T) {

	size := 64
	domain := fft.NewDomain(uint64(size))
	p := randomPoly(size)
	pLagrange := make([]field.Element, size)
	copy(pLagrange, p)
	domain.FFT(pLagrange, fft.DIF)
	fft.BitReverse(pLagrange)

	var x field.Element
	x.SetRandom()

	u := poly.Eval(p, x)
	v := EvaluateLagrange(pLagrange, x)

	tt := u.Equal(&v)
	if !tt {
		t.Fatal("Evaluate Lagrange failed")
	}

}

func TestEvaluateLagrangeFext(t *testing.T) {

	size := 64
	domain := fft.NewDomain(uint64(size))
	p := randomPoly(size)
	pLagrange := make([]field.Element, size)

	/*
		Test without coset
	*/
	copy(pLagrange, p)
	domain.FFT(pLagrange, fft.DIF)
	fft.BitReverse(pLagrange)

	var x fext.Element
	x.SetRandom()

	u := poly.EvalOnExtField(p, x)
	v := EvaluateLagrangeOnFext(pLagrange, x)

	require.Equal(t, u.String(), v.String())

	/*
		Test with coset
	*/
	copy(pLagrange, p)
	domain.FFT(pLagrange, fft.DIF, fft.OnCoset())
	fft.BitReverse(pLagrange)
	vOnCoset := EvaluateLagrangeOnFext(pLagrange, x, true)
	require.Equal(t, u.String(), vOnCoset.String())
}

func TestBatchEvaluateLagrangeOnFext(t *testing.T) {

	n := 4
	randPoly := vector.ForTest(1, 2, 3, 4)
	randPoly2 := vector.ForTest(5, 6, 7, 8)
	x := fext.NewElement(51, 1, 3, 4)

	expectedY := poly.EvalOnExtField(randPoly, x)
	expectedY2 := poly.EvalOnExtField(randPoly2, x)
	domain := fft.NewDomain(uint64(n))

	/*
		Test without coset
	*/
	onRoots := vector.DeepCopy(randPoly)
	onRoots2 := vector.DeepCopy(randPoly2)
	polys := make([][]field.Element, 2)
	polys[0] = onRoots
	polys[1] = onRoots2

	domain.FFT(polys[0], fft.DIF)
	domain.FFT(polys[1], fft.DIF)
	fft.BitReverse(polys[0])
	fft.BitReverse(polys[1])

	yOnRoots := BatchEvaluateLagrangeOnFext(polys, x)
	require.Equal(t, expectedY.String(), yOnRoots[0].String())
	require.Equal(t, expectedY2.String(), yOnRoots[1].String())

	/*
		Test with coset
	*/
	onCoset := vector.DeepCopy(randPoly)
	onCoset2 := vector.DeepCopy(randPoly2)
	onCosets := make([][]field.Element, 2)
	onCosets[0] = onCoset
	onCosets[1] = onCoset2

	domain.FFT(onCosets[0], fft.DIF, fft.OnCoset())
	domain.FFT(onCosets[1], fft.DIF, fft.OnCoset())
	fft.BitReverse(onCosets[0])
	fft.BitReverse(onCosets[1])

	yOnCosets := BatchEvaluateLagrangeOnFext(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())

}

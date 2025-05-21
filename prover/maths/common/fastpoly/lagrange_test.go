package fastpoly

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/stretchr/testify/require"
)

func prettyprint(p []field.Element) {
	for i := 0; i < len(p)-1; i++ {
		fmt.Printf("%s*x**%d+", p[i].String(), i)
	}
	fmt.Printf("%s*x**%d\n", p[len(p)-1].String(), len(p)-1)
}

func randomPoly(size int) []field.Element {
	res := make([]field.Element, size)
	for i := 0; i < size; i++ {
		res[i].SetRandom()
	}
	return res
}

func evalCanonical(p []field.Element, x field.Element) field.Element {
	var res field.Element
	for i := 0; i < len(p); i++ {
		res.Mul(&res, &x).Add(&res, &p[len(p)-1-i])
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

	u := evalCanonical(p, x)
	v := EvaluateLagrange(pLagrange, x)

	tt := u.Equal(&v)
	if !tt {
		t.Fatal("Evaluate Lagrange failed")
	}

}

func TestBatchInterpolation(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 2, 3, 4)
	randPoly2 := vector.ForTest(5, 6, 7, 8)

	randpolyext := make([]fext.Element, len(randPoly))
	for i := 0; i < len(randPoly); i++ {
		fext.FromBase(&randpolyext[i], &randPoly[i])

	}

	randpolyext2 := make([]fext.Element, len(randPoly2))
	for i := 0; i < len(randPoly2); i++ {
		fext.FromBase(&randpolyext2[i], &randPoly2[i])

	}

	x := fext.NewElement(1, 2, 3, 4)

	expectedY := polyext.Eval(randpolyext, x)
	expectedY2 := polyext.Eval(randpolyext2, x)
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

	yOnRoots := BatchInterpolate(polys, x)
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

	yOnCosets := BatchInterpolate(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())

}

// edge-case : x is a root of unity of the domain. In this case, we can just return
// the associated value for poly
func TestBatchInterpolationRootOfUnity(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 2, 3, 4)
	randPoly2 := vector.ForTest(5, 6, 7, 8)

	randpolyext := make([]fext.Element, len(randPoly))
	for i := 0; i < len(randPoly); i++ {
		fext.FromBase(&randpolyext[i], &randPoly[i])

	}

	randpolyext2 := make([]fext.Element, len(randPoly2))
	for i := 0; i < len(randPoly2); i++ {
		fext.FromBase(&randpolyext2[i], &randPoly2[i])

	}

	// define x as a root of unity
	x := fext.One()

	expectedY := polyext.Eval(randpolyext, x)
	expectedY2 := polyext.Eval(randpolyext2, x)
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

	yOnRoots := BatchInterpolate(polys, x)
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

	yOnCosets := BatchInterpolate(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())

}

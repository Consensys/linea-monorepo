package fastpolyext

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/stretchr/testify/require"
)

func TestInterpolation(t *testing.T) {
	n := 4
	randPoly := vectorext.ForTest(1, 2, 3, 4)
	x := fext.NewElement(51, 0, 0, 0)
	expectedY := polyext.Eval(randPoly, x)
	domain := fft.NewDomain(uint64(n))

	/*
		Test without coset
	*/
	onRoots := vectorext.DeepCopy(randPoly)
	domain.FFTExt(onRoots, fft.DIF)

	fft.BitReverse(onRoots)
	yOnRoots := Interpolate(onRoots, x)
	require.Equal(t, expectedY.String(), yOnRoots.String())

	/*
		Test with coset
	*/
	onCoset := vectorext.DeepCopy(randPoly)
	domain.FFTExt(onCoset, fft.DIF, fft.OnCoset())
	fft.BitReverse(onCoset)
	yOnCoset := Interpolate(onCoset, x, true)
	require.Equal(t, expectedY.String(), yOnCoset.String())

}

func TestBatchInterpolation(t *testing.T) {
	n := 4
	randPoly := vectorext.ForTest(1, 2, 3, 4)
	randPoly2 := vectorext.ForTest(5, 6, 7, 8)
	x := fext.NewElement(51, 0, 0, 0)

	expectedY := polyext.Eval(randPoly, x)
	expectedY2 := polyext.Eval(randPoly2, x)
	domain := fft.NewDomain(uint64(n))

	/*
		Test without coset
	*/
	onRoots := vectorext.DeepCopy(randPoly)
	onRoots2 := vectorext.DeepCopy(randPoly2)
	polys := make([][]fext.Element, 2)
	polys[0] = onRoots
	polys[1] = onRoots2

	domain.FFTExt(polys[0], fft.DIF)
	domain.FFTExt(polys[1], fft.DIF)
	fft.BitReverse(polys[0])
	fft.BitReverse(polys[1])

	yOnRoots := BatchInterpolate(polys, x)
	require.Equal(t, expectedY.String(), yOnRoots[0].String())
	require.Equal(t, expectedY2.String(), yOnRoots[1].String())

	/*
		Test with coset
	*/
	onCoset := vectorext.DeepCopy(randPoly)
	onCoset2 := vectorext.DeepCopy(randPoly2)
	onCosets := make([][]fext.Element, 2)
	onCosets[0] = onCoset
	onCosets[1] = onCoset2

	domain.FFTExt(onCosets[0], fft.DIF, fft.OnCoset())
	domain.FFTExt(onCosets[1], fft.DIF, fft.OnCoset())
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
	randPoly := vectorext.ForTest(1, 2, 3, 4)
	randPoly2 := vectorext.ForTest(5, 6, 7, 8)

	// define x as a root of unity
	x := fext.One()

	expectedY := polyext.Eval(randPoly, x)
	expectedY2 := polyext.Eval(randPoly2, x)
	domain := fft.NewDomain(uint64(n))

	/*
		Test without coset
	*/
	onRoots := vectorext.DeepCopy(randPoly)
	onRoots2 := vectorext.DeepCopy(randPoly2)
	polys := make([][]fext.Element, 2)
	polys[0] = onRoots
	polys[1] = onRoots2

	domain.FFTExt(polys[0], fft.DIF)
	domain.FFTExt(polys[1], fft.DIF)
	fft.BitReverse(polys[0])
	fft.BitReverse(polys[1])

	yOnRoots := BatchInterpolate(polys, x)
	require.Equal(t, expectedY.String(), yOnRoots[0].String())
	require.Equal(t, expectedY2.String(), yOnRoots[1].String())

	/*
		Test with coset
	*/
	onCoset := vectorext.DeepCopy(randPoly)
	onCoset2 := vectorext.DeepCopy(randPoly2)
	onCosets := make([][]fext.Element, 2)
	onCosets[0] = onCoset
	onCosets[1] = onCoset2

	domain.FFTExt(onCosets[0], fft.DIF, fft.OnCoset())
	domain.FFTExt(onCosets[1], fft.DIF, fft.OnCoset())
	fft.BitReverse(onCosets[0])
	fft.BitReverse(onCosets[1])

	yOnCosets := BatchInterpolate(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())

}

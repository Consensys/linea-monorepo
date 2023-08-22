package fastpoly

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/stretchr/testify/require"
)

func TestInterpolation(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 2, 3, 4)
	x := field.NewElement(51)
	expectedY := poly.EvalUnivariate(randPoly, x)
	domain := fft.NewDomain(n).WithCoset()

	/*
		Test without coset
	*/
	onRoots := vector.DeepCopy(randPoly)
	domain.FFT(onRoots, fft.DIF)

	fft.BitReverse(onRoots)
	yOnRoots := Interpolate(onRoots, x)
	require.Equal(t, expectedY.String(), yOnRoots.String())

	/*
		Test with coset
	*/
	onCoset := vector.DeepCopy(randPoly)
	domain.FFT(onCoset, fft.DIF, true)
	fft.BitReverse(onCoset)
	yOnCoset := Interpolate(onCoset, x, true)
	require.Equal(t, expectedY.String(), yOnCoset.String())

}

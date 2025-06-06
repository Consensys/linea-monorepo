//go:build !race

package smartvectors

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateLagrange(t *testing.T) {

	var x fext.Element
	x.SetRandom()
	size := 64
	poly := make([]field.Element, size)
	for i := 0; i < size; i++ {
		poly[i].SetRandom()
	}

	d := fft.NewDomain(64)
	polyLagrange := make([]field.Element, size)
	copy(polyLagrange, poly)
	d.FFT(polyLagrange, fft.DIF)
	fft.BitReverse(polyLagrange)

	var evalCan fext.Element
	var tmp fext.Element
	for i := size - 1; i >= 0; i-- {
		fext.FromBase(&tmp, &poly[i])
		evalCan.Mul(&evalCan, &x)
		evalCan.Add(&evalCan, &tmp)
	}

	var evalLag fext.Element
	polyLagrangeSv := NewRegular(polyLagrange)
	evalLag = EvaluateLagrangeOnFext(polyLagrangeSv, x)

	if !evalLag.Equal(&evalCan) {
		t.Fatal("error")
	}

}

func TestRuffini(t *testing.T) {

	testCases := []struct {
		q           field.Element
		p           SmartVector
		expectedQuo SmartVector
		expectedRem field.Element
	}{
		{
			q:           field.NewElement(1),
			p:           ForTest(3, 0, 1),
			expectedQuo: ForTest(1, 1),
			expectedRem: field.NewElement(4),
		},
		{
			// 3 = 0 * (X - 1) + 3
			q:           field.NewElement(1),
			p:           ForTest(3),
			expectedQuo: NewConstant(field.Zero(), 1),
			expectedRem: field.NewElement(3),
		},
	}

	for _, testCase := range testCases {

		quo, rem := RuffiniQuoRem(testCase.p, testCase.q)
		require.Equal(t, testCase.expectedQuo.Pretty(), quo.Pretty())
		require.Equal(t, testCase.expectedRem.String(), rem.String())
	}

}

func TestFuzzPolynomial(t *testing.T) {

	for i := 0; i < FuzzIteration; i++ {

		// We reuse the test-case generator of lincomb but we only
		// use the first generated edge-case for each. The fact that
		// we use two test-cases gives a and b of different length.
		tcaseA := newTestBuilder(2 * i).NewTestCaseForLinComb()
		tcaseB := newTestBuilder(2*i + 1).NewTestCaseForLinComb()

		success := t.Run(fmt.Sprintf("fuzz-poly-%v", i), func(t *testing.T) {

			a := tcaseA.svecs[0]
			b := tcaseB.svecs[0]

			// Try interpolating by zero (should return the first element)
			var zeroExt fext.Element
			zeroExt.SetZero()
			xa := EvaluateLagrangeOnFext(a, zeroExt)
			expecteda0 := a.GetExt(0)
			assert.Equal(t, xa.String(), expecteda0.String())

			// Get a random x to use as an evaluation point to check polynomial
			// identities
			var x fext.Element
			x.SetRandom()
			aX := EvalCoeffOnFext(a, x)
			bX := EvalCoeffOnFext(b, x)

			// Get the evaluations of a-n, b-a, a+b
			var aSubBx, bSubAx, aPlusBx fext.Element
			aSubBx.Sub(&aX, &bX)
			bSubAx.Sub(&bX, &aX)
			aPlusBx.Add(&aX, &bX)

			// And evaluate the corresponding polynomials to compare
			// with the above values
			aSubb := PolySub(a, b)
			bSuba := PolySub(b, a)
			aPlusb := PolyAdd(a, b)
			bPlusa := PolyAdd(b, a)

			aSubBxActual := EvalCoeffOnFext(aSubb, x)
			bSubAxActual := EvalCoeffOnFext(bSuba, x)
			aPlusbxActual := EvalCoeffOnFext(aPlusb, x)
			bPlusaxActual := EvalCoeffOnFext(bPlusa, x)

			t.Logf(
				"Len of a %v, b %v, a+b %v, a-b %v, b-a %v",
				a.Len(), b.Len(), aPlusb.Len(), aSubb.Len(), bSuba.Len(),
			)

			require.Equal(t, aSubBx.String(), aSubBxActual.String())
			require.Equal(t, bSubAx.String(), bSubAxActual.String())
			require.Equal(t, aPlusBx.String(), aPlusbxActual.String())
			require.Equal(t, aPlusBx.String(), bPlusaxActual.String())
		})

		if !success {
			t.Logf("TEST CASE %v\n", i)
			t.FailNow()
		}
	}
}

func TestBivariatePolynomial(t *testing.T) {

	testCases := []struct {
		v         SmartVector
		x         field.Element
		y         field.Element
		numCoeffX int
		res       field.Element
	}{
		{
			v:         ForTest(1, 2, 3, 4),
			x:         field.NewElement(2),
			y:         field.NewElement(3),
			numCoeffX: 2,
			res:       field.NewElement(38),
		},
	}

	for i, testCase := range testCases {

		t.Run(fmt.Sprintf("case-%v", i), func(t *testing.T) {
			res := EvalCoeffBivariate(
				testCase.v,
				testCase.x,
				testCase.numCoeffX,
				testCase.y,
			)

			require.Equal(t, testCase.res.String(), res.String())
		})
	}

}

func TestBatchInterpolationWithConstantVector(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 1, 1, 1)
	randPoly2 := vector.ForTest(2, 2, 2, 2)
	x := fext.NewElement(51, 1, 2, 3)

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

	yOnRoots := fastpoly.BatchEvaluateLagrangeOnFext(polys, x)
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

	yOnCosets := fastpoly.BatchEvaluateLagrangeOnFext(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())
}

func TestBatchEvaluateLagrangeOnFextOnlyConstantVector(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 1, 1, 1)
	randPoly2 := vector.ForTest(2, 2, 2, 2)
	x := fext.NewElement(51, 1, 2, 3)

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

	yOnRoots := fastpoly.BatchEvaluateLagrangeOnFext(polys, x)
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

	yOnCosets := fastpoly.BatchEvaluateLagrangeOnFext(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())
}

// three vectors to see if range check and continue statement
// for edge cases works as expected
func TestBatchInterpolationThreeVectors(t *testing.T) {
	n := 4
	randPoly := vector.ForTest(1, 2, 3, 4)
	randPoly2 := vector.ForTest(1, 1, 1, 1)
	randPoly3 := vector.ForTest(1, 2, 3, 4)

	x := fext.NewElement(51, 1, 2, 3)

	expectedY := poly.EvalOnExtField(randPoly, x)
	expectedY2 := poly.EvalOnExtField(randPoly2, x)
	expectedY3 := poly.EvalOnExtField(randPoly3, x)
	domain := fft.NewDomain(uint64(n))

	/*
		Test without coset
	*/
	onRoots := vector.DeepCopy(randPoly)
	onRoots2 := vector.DeepCopy(randPoly2)
	onRoots3 := vector.DeepCopy(randPoly3)
	polys := make([][]field.Element, 3)
	polys[0] = onRoots
	polys[1] = onRoots2
	polys[2] = onRoots3

	domain.FFT(polys[0], fft.DIF)
	domain.FFT(polys[1], fft.DIF)
	domain.FFT(polys[2], fft.DIF)
	fft.BitReverse(polys[0])
	fft.BitReverse(polys[1])
	fft.BitReverse(polys[2])

	yOnRoots := fastpoly.BatchEvaluateLagrangeOnFext(polys, x)
	require.Equal(t, expectedY.String(), yOnRoots[0].String())
	require.Equal(t, expectedY2.String(), yOnRoots[1].String())
	require.Equal(t, expectedY3.String(), yOnRoots[2].String())

	/*
		Test with coset
	*/
	onCoset := vector.DeepCopy(randPoly)
	onCoset2 := vector.DeepCopy(randPoly2)
	onCoset3 := vector.DeepCopy(randPoly3)
	onCosets := make([][]field.Element, 3)
	onCosets[0] = onCoset
	onCosets[1] = onCoset2
	onCosets[2] = onCoset3

	domain.FFT(onCosets[0], fft.DIF, fft.OnCoset())
	domain.FFT(onCosets[1], fft.DIF, fft.OnCoset())
	domain.FFT(onCosets[2], fft.DIF, fft.OnCoset())
	fft.BitReverse(onCosets[0])
	fft.BitReverse(onCosets[1])
	fft.BitReverse(onCosets[2])

	yOnCosets := fastpoly.BatchEvaluateLagrangeOnFext(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())
	require.Equal(t, expectedY3.String(), yOnCosets[2].String())

}

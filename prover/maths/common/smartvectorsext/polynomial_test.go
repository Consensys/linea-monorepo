//go:build !race

package smartvectorsext

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
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
	polyLagrangeSv := smartvectors.NewRegular(polyLagrange)
	evalLag = EvaluateLagrange(polyLagrangeSv, x)

	if !evalLag.Equal(&evalCan) {
		t.Fatal("error")
	}

}

func TestRuffini(t *testing.T) {

	testCases := []struct {
		q           fext.Element
		p           smartvectors.SmartVector
		expectedQuo smartvectors.SmartVector
		expectedRem fext.Element
	}{
		{
			q:           fext.NewElement(1, 0, 0, 0),
			p:           ForTestFromQuads(3, 0, 0, 0, 1, 0, 0, 0),
			expectedQuo: ForTestFromQuads(1, 0, 0, 0),
			expectedRem: fext.NewElement(4, 0, 0, 0),
		},
		{
			// 3 = 0 * (X - 1) + 3
			q:           fext.NewElement(1, 0, 0, 0),
			p:           ForTestFromQuads(3, 0, 0, 0),
			expectedQuo: NewConstantExt(fext.Zero(), 1),
			expectedRem: fext.NewElement(3, 0, 0, 0),
		},
		{
			// (3 -α^2 - 3 α) + (2 + α) x + (- 1 - α^2 - 2 α)x^2 + (α +1)x^3 =
			// (x-(1+alpha))(x^2*(1+alpha)+(2+alpha))+5
			// alpha is a square root used to build the extension field,
			// i.e. alpha = v, alpha^2=u,
			q:           fext.NewElement(1, 0, 1, 0),
			p:           ForTestFromQuads(3, -1, -3, 0, 2, 0, 1, 0, -1, -1, -2, 0, 1, 0, 1, 0),
			expectedQuo: ForTestFromQuads(2, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 0),
			expectedRem: fext.NewElement(5, 0, 0, 0),
		},
	}

	for _, testCase := range testCases {

		quo, rem := RuffiniQuoRem(testCase.p, testCase.q)
		require.Equal(t, testCase.expectedQuo.Pretty(), quo.Pretty())
		require.Equal(t, testCase.expectedRem.String(), rem.String())
	}

}

func TestFuzzPolynomial(t *testing.T) {

	for i := 0; i < smartvectors.FuzzIteration; i++ {

		// We reuse the test-case generator of lincomb but we only
		// use the first generated edge-case for each. The fact that
		// we use two test-cases gives a and b of different length.
		tcaseA := newTestBuilder(2 * i).NewTestCaseForLinComb()
		tcaseB := newTestBuilder(2*i + 1).NewTestCaseForLinComb()

		success := t.Run(fmt.Sprintf("fuzz-poly-%v", i), func(t *testing.T) {

			a := tcaseA.svecs[0]
			b := tcaseB.svecs[0]

			// Try interpolating by one (should return the first element)
			var zeroExt fext.Element
			zeroExt.SetZero()
			xa := EvaluateLagrange(a, zeroExt)
			expecteda0 := a.GetExt(0)
			assert.Equal(t, xa.String(), expecteda0.String())

			// Get a random x to use as an evaluation point to check polynomial
			// identities
			var x fext.Element
			x.SetRandom()
			aX := EvalCoeff(a, x)
			bX := EvalCoeff(b, x)

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

			aSubBxActual := EvalCoeff(aSubb, x)
			bSubAxActual := EvalCoeff(bSuba, x)
			aPlusbxActual := EvalCoeff(aPlusb, x)
			bPlusaxActual := EvalCoeff(bPlusa, x)

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
		v         smartvectors.SmartVector
		x         fext.Element
		y         fext.Element
		numCoeffX int
		res       fext.Element
	}{
		{
			// P(X) = P1(X)+Y*P2(X) = (1+4)+(3+8)Y = 5+11Y = 5+33 = 38
			v:         ForTestExt(1, 2, 3, 4),
			x:         fext.NewElement(2, 0, 0, 0),
			y:         fext.NewElement(3, 0, 0, 0),
			numCoeffX: 2,
			res:       fext.NewElement(38, 0, 0, 0),
		},
		{
			// P(X) = P1(X)+Y*P2(X)
			v:         ForTestFromQuads(1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 0, 0, 0),
			x:         fext.NewElement(2, 0, 0, 0),
			y:         fext.NewElement(3, 1, 2, 0),
			numCoeffX: 2,
			res:       fext.NewElement(38, 11, 22, 0),
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
	randPoly := vectorext.ForTest(1, 2, 3, 4)
	randPoly2 := vectorext.ForTest(1, 1, 1, 1)

	x := fext.NewElement(51, 1, 2, 3)

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

	yOnRoots := fastpolyext.BatchEvaluateLagrange(polys, x)
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

	yOnCosets := fastpolyext.BatchEvaluateLagrange(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())

}

func TestBatchEvaluateLagrangeOnlyConstantVector(t *testing.T) {
	n := 4
	randPoly := vectorext.ForTest(1, 1, 1, 1)
	randPoly2 := vectorext.ForTest(2, 2, 2, 2)
	x := fext.NewElement(51, 1, 2, 3)

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

	yOnRoots := fastpolyext.BatchEvaluateLagrange(polys, x)
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

	yOnCosets := fastpolyext.BatchEvaluateLagrange(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())
}

// three vectors to see if range check and continue statement
// for edge cases works as expected
func TestBatchInterpolationThreeVectors(t *testing.T) {
	n := 4
	randPoly := vectorext.ForTest(1, 2, 3, 4)
	randPoly2 := vectorext.ForTest(1, 1, 1, 1)
	randPoly3 := vectorext.ForTest(1, 2, 3, 4)

	x := fext.NewElement(51, 1, 2, 3)
	expectedY := polyext.Eval(randPoly, x)
	expectedY2 := polyext.Eval(randPoly2, x)
	expectedY3 := polyext.Eval(randPoly3, x)
	domain := fft.NewDomain(uint64(n))
	/*
		Test without coset
	*/
	onRoots := vectorext.DeepCopy(randPoly)
	onRoots2 := vectorext.DeepCopy(randPoly2)
	onRoots3 := vectorext.DeepCopy(randPoly3)
	polys := make([][]fext.Element, 3)
	polys[0] = onRoots
	polys[1] = onRoots2
	polys[2] = onRoots3

	domain.FFTExt(polys[0], fft.DIF)
	domain.FFTExt(polys[1], fft.DIF)
	domain.FFTExt(polys[2], fft.DIF)
	fft.BitReverse(polys[0])
	fft.BitReverse(polys[1])
	fft.BitReverse(polys[2])

	yOnRoots := fastpolyext.BatchEvaluateLagrange(polys, x)
	require.Equal(t, expectedY.String(), yOnRoots[0].String())
	require.Equal(t, expectedY2.String(), yOnRoots[1].String())
	require.Equal(t, expectedY3.String(), yOnRoots[2].String())

	/*
		Test with coset
	*/
	onCoset := vectorext.DeepCopy(randPoly)
	onCoset2 := vectorext.DeepCopy(randPoly2)
	onCoset3 := vectorext.DeepCopy(randPoly3)
	onCosets := make([][]fext.Element, 3)
	onCosets[0] = onCoset
	onCosets[1] = onCoset2
	onCosets[2] = onCoset3

	domain.FFTExt(onCosets[0], fft.DIF, fft.OnCoset())
	domain.FFTExt(onCosets[1], fft.DIF, fft.OnCoset())
	domain.FFTExt(onCosets[2], fft.DIF, fft.OnCoset())
	fft.BitReverse(onCosets[0])
	fft.BitReverse(onCosets[1])
	fft.BitReverse(onCosets[2])

	yOnCosets := fastpolyext.BatchEvaluateLagrange(onCosets, x, true)
	require.Equal(t, expectedY.String(), yOnCosets[0].String())
	require.Equal(t, expectedY2.String(), yOnCosets[1].String())
	require.Equal(t, expectedY3.String(), yOnCosets[2].String())

}

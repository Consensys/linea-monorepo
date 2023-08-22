//go:build !race

package smartvectors

import (
	"fmt"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuffini(t *testing.T) {
	q := field.NewElement(1)
	p := ForTest(3, 0, 1)
	expected := ForTest(1, 1)

	res, rem := RuffiniQuoRem(p, q)

	require.Equal(t, "4", rem.String())
	require.Equal(t, expected.Pretty(), res.Pretty())
}

func TestFuzzPolynomial(t *testing.T) {

	for i := 0; i < FUZZ_ITERATION; i++ {

		// We reuse the test-case generator of lincomb but we only
		// use the first generated edge-case for each. The fact that
		// we use two test-cases gives a and b of different length.
		tcaseA := NewTestBuilder(2 * i).NewTestCaseForLinComb()
		tcaseB := NewTestBuilder(2*i + 1).NewTestCaseForLinComb()

		success := t.Run(fmt.Sprintf("fuzz-poly-%v", i), func(t *testing.T) {

			a := tcaseA.svecs[0]
			b := tcaseB.svecs[0]

			// Try interpolating by one (should return the first element)
			xa := Interpolate(a, field.One())
			expecteda0 := a.Get(0)
			assert.Equal(t, xa.String(), expecteda0.String())

			// Get a random x to use as an evaluation point to check polynomial
			// identities
			var x field.Element
			x.SetRandom()
			aX := EvalCoeff(a, x)
			bX := EvalCoeff(b, x)

			// Get the evaluations of a-n, b-a, a+b
			var aSubBx, bSubAx, aPlusBx field.Element
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

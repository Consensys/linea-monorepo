//go:build !race

package smartvectors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuzzProduct(t *testing.T) {

	for i := 0; i < FUZZ_ITERATION; i++ {
		tcase := NewTestBuilder(i).NewTestCaseForProd()

		success := t.Run(tcase.name, func(t *testing.T) {
			actualProd := Product(tcase.coeffs, tcase.svecs)
			require.Equal(t, tcase.expectedValue.Pretty(), actualProd.Pretty(), "product failed")

			// And let us do it a second time for idemnpotency
			actualProd = Product(tcase.coeffs, tcase.svecs)
			require.Equal(t, tcase.expectedValue.Pretty(), actualProd.Pretty(), "product failed")
		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}

}

func TestFuzzLinComb(t *testing.T) {
	for i := 0; i < FUZZ_ITERATION; i++ {
		tcase := NewTestBuilder(i).NewTestCaseForLinComb()

		success := t.Run(tcase.name, func(t *testing.T) {

			actualLinComb := LinComb(tcase.coeffs, tcase.svecs)
			require.Equal(t, tcase.expectedValue.Pretty(), actualLinComb.Pretty(), "linear combination failed")

			// And a second time for idemnpotency
			actualLinComb = LinComb(tcase.coeffs, tcase.svecs)
			require.Equal(t, tcase.expectedValue.Pretty(), actualLinComb.Pretty(), "linear combination failed")
		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

func TestFuzzPolyEval(t *testing.T) {
	for i := 0; i < FUZZ_ITERATION; i++ {
		tcase := NewTestBuilder(i).NewTestCaseForPolyEval()

		success := t.Run(tcase.name, func(t *testing.T) {

			actualRes := PolyEval(tcase.svecs, tcase.evaluationPoint)
			require.Equal(t, tcase.expectedValue.Pretty(), actualRes.Pretty(), "linear combination failed")

			// and a second time to ensure idemnpotency
			actualRes = PolyEval(tcase.svecs, tcase.evaluationPoint)
			require.Equal(t, tcase.expectedValue.Pretty(), actualRes.Pretty(), "linear combination failed")

		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

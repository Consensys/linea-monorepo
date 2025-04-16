//go:build !fuzzlight

package smartvectors

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuzzProductExt(t *testing.T) {

	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForProdExt()

		success := t.Run(tcase.name, func(t *testing.T) {
			actualProd := ProductExt(tcase.coeffs, tcase.svecs)
			require.Equal(t, tcase.expectedValue.Pretty(), actualProd.Pretty(), "product failed")

			// And let us do it a second time for idempotency
			actualProd = ProductExt(tcase.coeffs, tcase.svecs)
			require.Equal(t, tcase.expectedValue.Pretty(), actualProd.Pretty(), "product failed")
		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}

}

func TestFuzzLinCombExt(t *testing.T) {
	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForLinCombExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			actualLinComb := LinCombExt(tcase.coeffs, tcase.svecs)
			require.Equal(t, tcase.expectedValue.Pretty(), actualLinComb.Pretty(), "linear combination failed")

			// And a second time for idempotency
			actualLinComb = LinCombExt(tcase.coeffs, tcase.svecs)
			require.Equal(t, tcase.expectedValue.Pretty(), actualLinComb.Pretty(), "linear combination failed")
		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

func TestFuzzPolyEvalExt(t *testing.T) {
	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForPolyEvalExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			actualRes := PolyEvalExt(tcase.svecs, tcase.evaluationPoint)
			require.Equal(t, tcase.expectedValue.Pretty(), actualRes.Pretty(), "linear combination failed")

			// and a second time to ensure idempotency
			actualRes = PolyEvalExt(tcase.svecs, tcase.evaluationPoint)
			require.Equal(t, tcase.expectedValue.Pretty(), actualRes.Pretty(), "linear combination failed")

		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

func TestFuzzProductWithPoolExt(t *testing.T) {

	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForProdExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			pool := mempool.CreateFromSyncPool(tcase.svecs[0].Len())

			t.Logf("TEST CASE %v\n", tcase.String())

			prodWithPool := ProductExt(tcase.coeffs, tcase.svecs, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), prodWithPool.Pretty(), "product with pool failed")

			// And let us do it a second time for idempotency
			prodWithPool = ProductExt(tcase.coeffs, tcase.svecs, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), prodWithPool.Pretty(), "product with pool failed")
		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}

}

func TestFuzzProductWithPoolCompareExt(t *testing.T) {

	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForProdExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			pool := mempool.CreateFromSyncPool(tcase.svecs[0].Len())

			t.Logf("TEST CASE %v\n", tcase.String())

			// ProductExt() with pool
			prodWithPool := ProductExt(tcase.coeffs, tcase.svecs, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), prodWithPool.Pretty(), "ProductExt() with pool failed")

			// ProductExt() without pool
			prod := ProductExt(tcase.coeffs, tcase.svecs)

			// check if ProductExt() with pool = ProductExt() without pool
			require.Equal(t, prodWithPool.Pretty(), prod.Pretty(), "ProductExt() w/ and w/o pool are different")
		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}

}

func TestFuzzLinCombWithPoolExt(t *testing.T) {

	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForLinCombExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			pool := mempool.CreateFromSyncPool(tcase.svecs[0].Len())

			t.Logf("TEST CASE %v\n", tcase.String())

			linCombWithPool := LinCombExt(tcase.coeffs, tcase.svecs, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), linCombWithPool.Pretty(), "LinCombExt() with pool failed")

			// And let us do it a second time for idempotency
			linCombWithPool = LinCombExt(tcase.coeffs, tcase.svecs, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), linCombWithPool.Pretty(), "LinCombExt() with pool failed")
		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

func TestFuzzLinCombWithPoolCompareExt(t *testing.T) {

	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForLinCombExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			pool := mempool.CreateFromSyncPool(tcase.svecs[0].Len())

			t.Logf("TEST CASE %v\n", tcase.String())

			// LinCombExt() with pool
			linCombWithPool := LinCombExt(tcase.coeffs, tcase.svecs, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), linCombWithPool.Pretty(), "LinCombExt() with pool failed")

			// LinCombExt() without pool
			linComb := LinCombExt(tcase.coeffs, tcase.svecs)

			// check if LinCombExt() with pool = LinCombExt() without pool
			require.Equal(t, linCombWithPool.Pretty(), linComb.Pretty(), "LinCombExt() w/ and w/o pool are different")
		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

func TestOpBasicEdgeCasesExt(t *testing.T) {

	two := fext.NewElement(2, fieldPaddingInt())
	eight := new(fext.Element).Exp(two, big.NewInt(3))

	testCases := []struct {
		explainer   string
		inputs      []SmartVector
		expectedRes SmartVector
		fn          func(...SmartVector) SmartVector
	}{
		{
			explainer: "full-covering windows and a constant",
			inputs: []SmartVector{
				NewConstantExt(two, 16),
				LeftPaddedExt(vectorext.Repeat(two, 12), two, 16),
				RightPaddedExt(vectorext.Repeat(two, 12), two, 16),
			},
			expectedRes: NewRegularExt(vectorext.Repeat(fext.NewElement(6, 3*fieldPaddingInt()), 16)),
			fn:          AddExt,
		},
		{
			explainer: "full-covering windows and a constant (mul)",
			inputs: []SmartVector{
				NewConstantExt(two, 16),
				LeftPaddedExt(vectorext.Repeat(two, 12), two, 16),
				RightPaddedExt(vectorext.Repeat(two, 12), two, 16),
			},
			expectedRes: NewRegularExt(vectorext.Repeat(*eight, 16)),
			fn:          MulExt,
		},
		{
			explainer: "full-covering windows, a regular and a constant",
			inputs: []SmartVector{
				NewConstantExt(two, 16),
				LeftPaddedExt(vectorext.Repeat(two, 12), two, 16),
				RightPaddedExt(vectorext.Repeat(two, 12), two, 16),
				NewRegularExt(vectorext.Repeat(two, 16)),
			},
			expectedRes: NewRegularExt(vectorext.Repeat(fext.NewElement(8, 4*fieldPaddingInt()), 16)),
			fn:          AddExt,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("case-%v", i), func(t *testing.T) {
			t.Logf("test-case details: %v", testCase.explainer)
			res := testCase.fn(testCase.inputs...).(*PooledExt)
			actual := NewRegularExt(res.RegularExt)
			require.Equal(t, testCase.expectedRes, actual, "expectedRes=%v\nres=%v", testCase.expectedRes.Pretty(), res.Pretty())
		})
	}
}

func TestInnerProductExt(t *testing.T) {
	a := ForTestFromPairs(1, 1, 2, 1, 1, 1, 2, 1, 1, 1)
	b := ForTestFromPairs(1, 1, -1, 1, 2, 1, -1, 1, 2, 1)
	sum := new(fext.Element).SetInt64Pair(int64(1+5*fext.RootPowers[1]), 10)

	testCases := []struct {
		a, b SmartVector
		y    fext.Element
	}{
		{
			a: a,
			b: b,
			y: *sum,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("case-%v", i), func(t *testing.T) {
			y := InnerProductExt(testCase.a, testCase.b)
			assert.Equal(t, testCase.y.String(), y.String())
		})
	}
}

func TestScalarMulExt(t *testing.T) {
	testCases := []struct {
		a SmartVector
		b fext.Element
		y SmartVector
	}{
		{
			a: ForTestExt(1, 2, 1, 2, 1),
			b: fext.NewElement(3, 1),
			y: ForTestFromPairs(3, 1, 6, 2, 3, 1, 6, 2, 3, 1),
		},
		{
			a: ForTestExt(1, 2, 1, 2, 1),
			b: fext.NewElement(3, 0),
			y: ForTestExt(3, 6, 3, 6, 3),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("case-%v", i), func(t *testing.T) {
			y := ScalarMulExt(testCase.a, testCase.b)
			assert.Equal(t, testCase.y.Pretty(), y.Pretty())
		})
	}
}

func TestFuzzPolyEvalWithPoolExt(t *testing.T) {
	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForPolyEvalExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			pool := mempool.CreateFromSyncPool(tcase.svecs[0].Len())

			// PolyEvalExt() with pool
			polyEvalWithPool := PolyEvalExt(tcase.svecs, tcase.evaluationPoint, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), polyEvalWithPool.Pretty(), "linear combination with pool failed")

			// and a second time to ensure idempotency
			polyEvalWithPool = PolyEvalExt(tcase.svecs, tcase.evaluationPoint, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), polyEvalWithPool.Pretty(), "linear combination with pool failed")

		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

func TestFuzzPolyEvalWithPoolCompareExt(t *testing.T) {
	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForPolyEvalExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			pool := mempool.CreateFromSyncPool(tcase.svecs[0].Len())

			// PolyEvalExt() with pool
			polyEvalWithPool := PolyEvalExt(tcase.svecs, tcase.evaluationPoint, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), polyEvalWithPool.Pretty(), "PolyEvalExt() with pool failed")

			// PolyEvalExt() without pool
			polyEval := PolyEvalExt(tcase.svecs, tcase.evaluationPoint)

			// check if PolyEvalExt() with pool = PolyEvalExt() without pool
			require.Equal(t, polyEvalWithPool.Pretty(), polyEval.Pretty(), "PolyEvalExt() w/ and w/o pool are different")

		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

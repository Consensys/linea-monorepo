package smartvectors

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/stretchr/testify/require"
)

func TestFuzzProductExt(t *testing.T) {

	for i := 0; i < 1; i++ {
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

func TestFuzzLinearCombinationExt(t *testing.T) {
	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForLinearCombinationExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			actualRes := LinearCombinationExt(tcase.svecs, tcase.evaluationPoint)
			require.Equal(t, tcase.expectedValue.Pretty(), actualRes.Pretty(), "linear combination failed")

			// and a second time to ensure idempotency
			actualRes = LinearCombinationExt(tcase.svecs, tcase.evaluationPoint)
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

func TestFuzzPolyEvalWithPoolExt(t *testing.T) {
	for i := 0; i < FuzzIteration; i++ {
		tcase := newTestBuilderExt(i).NewTestCaseForLinearCombinationExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			pool := mempool.CreateFromSyncPool(tcase.svecs[0].Len())

			// LinearCombinationExt() with pool
			polyEvalWithPool := LinearCombinationExt(tcase.svecs, tcase.evaluationPoint, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), polyEvalWithPool.Pretty(), "linear combination with pool failed")

			// and a second time to ensure idempotency
			polyEvalWithPool = LinearCombinationExt(tcase.svecs, tcase.evaluationPoint, pool)
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
		tcase := newTestBuilderExt(i).NewTestCaseForLinearCombinationExt()

		success := t.Run(tcase.name, func(t *testing.T) {

			pool := mempool.CreateFromSyncPool(tcase.svecs[0].Len())

			// LinearCombinationExt() with pool
			polyEvalWithPool := LinearCombinationExt(tcase.svecs, tcase.evaluationPoint, pool)
			require.Equal(t, tcase.expectedValue.Pretty(), polyEvalWithPool.Pretty(), "LinearCombinationExt() with pool failed")

			// LinearCombinationExt() without pool
			polyEval := LinearCombinationExt(tcase.svecs, tcase.evaluationPoint)

			fmt.Printf("tcase.svecs=%s", tcase.svecs)
			// check if LinearCombinationExt() with pool = LinearCombinationExt() without pool
			require.Equal(t, polyEvalWithPool.Pretty(), polyEval.Pretty(), "LinearCombinationExt() w/ and w/o pool are different")

		})

		if !success {
			t.Logf("TEST CASE %v\n", tcase.String())
			t.FailNow()
		}
	}
}

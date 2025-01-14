package polyext_test

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"

	"github.com/stretchr/testify/require"
)

func TestEvalUnivariate(t *testing.T) {
	// Just a simple test vector
	testVec := []fext.Element{
		fext.NewElement(1, 0),
		fext.NewElement(2, 0),
		fext.NewElement(5, 0),
		fext.NewElement(12, 0),
	}

	x := fext.NewElement(17, 0)

	y := polyext.EvalUnivariate(testVec, x)

	require.Equal(t, y, fext.NewElement(60436, 0))

}

func TestMul(t *testing.T) {

	t.Run("same-size", func(t *testing.T) {
		var (
			a        = vectorext.ForTest(1, 1)
			b        = vectorext.ForTest(-1, 1)
			expected = vectorext.ForTest(-1, 0, 1)
			res      = polyext.Mul(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("a-is-smaller", func(t *testing.T) {
		var (
			a        = vectorext.ForTest(1, 1)
			b        = vectorext.ForTest(-1, 0, 1)
			expected = vectorext.ForTest(-1, -1, 1, 1)
			res      = polyext.Mul(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("a-is-larger", func(t *testing.T) {
		var (
			a        = vectorext.ForTest(-1, 0, 1)
			b        = vectorext.ForTest(1, 1)
			expected = vectorext.ForTest(-1, -1, 1, 1)
			res      = polyext.Mul(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("a-is-empty", func(t *testing.T) {
		var (
			a        = vectorext.ForTest()
			b        = vectorext.ForTest(1, 1)
			expected = vectorext.ForTest()
			res      = polyext.Mul(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("b-is-empty", func(t *testing.T) {
		var (
			a        = vectorext.ForTest(1, 1)
			b        = vectorext.ForTest()
			expected = vectorext.ForTest()
			res      = polyext.Mul(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})
}

func TestAdd(t *testing.T) {

	t.Run("same-size", func(t *testing.T) {
		var (
			a        = vectorext.ForTest(1, 1)
			b        = vectorext.ForTest(-1, 1)
			expected = vectorext.ForTest(0, 2)
			res      = polyext.Add(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("a-is-smaller", func(t *testing.T) {
		var (
			a        = vectorext.ForTest(1, 1)
			b        = vectorext.ForTest(1, 1, 1, 1)
			expected = vectorext.ForTest(2, 2, 1, 1)
			res      = polyext.Add(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("a-is-larger", func(t *testing.T) {
		var (
			a        = vectorext.ForTest(1, 1, 1, 1)
			b        = vectorext.ForTest(1, 1)
			expected = vectorext.ForTest(2, 2, 1, 1)
			res      = polyext.Add(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("a-is-empty", func(t *testing.T) {
		var (
			a        = vectorext.ForTest()
			b        = vectorext.ForTest(1, 1)
			expected = vectorext.ForTest(1, 1)
			res      = polyext.Add(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("b-is-empty", func(t *testing.T) {
		var (
			a        = vectorext.ForTest(1, 1)
			b        = vectorext.ForTest()
			expected = vectorext.ForTest(1, 1)
			res      = polyext.Add(a, b)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})
}

func TestScalarMul(t *testing.T) {

	t.Run("normal-vec", func(t *testing.T) {
		var (
			vec      = vectorext.ForTest(1, 2, 3, 4)
			x        = fext.NewElement(2, 0)
			expected = vectorext.ForTest(2, 4, 6, 8)
			res      = polyext.ScalarMul(vec, x)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

	t.Run("empty-vec", func(t *testing.T) {
		var (
			vec      = vectorext.ForTest()
			x        = fext.NewElement(2, 0)
			expected = vectorext.ForTest()
			res      = polyext.ScalarMul(vec, x)
		)
		require.Equal(t, vectorext.Prettify(expected), vectorext.Prettify(res))
	})

}

func TestEvaluateLagrangeAnyDomain(t *testing.T) {

	t.Run("single-point-domain", func(t *testing.T) {
		var (
			domain   = []fext.Element{fext.Element{field.Zero(), field.Zero()}}
			x        = fext.NewElement(42, 0)
			ys       = polyext.EvaluateLagrangesAnyDomain(domain, x)
			expected = vectorext.ForTest(1)
		)
		require.Equal(t, expected, ys)
	})

	t.Run("many-point-domain", func(t *testing.T) {
		var (
			// the first lagrange poly is 1-X and the second one is X
			domain   = vectorext.ForTest(0, 1)
			x        = fext.NewElement(42, 0)
			ys       = polyext.EvaluateLagrangesAnyDomain(domain, x)
			expected = vectorext.ForTest(-41, 42)
		)
		require.Equal(t, expected, ys)
	})

}

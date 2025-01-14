package vectorext_test

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectors(t *testing.T) {

	var (
		// a, b and x are very common vectors in all the tests
		a = vectorext.ForTest(1, 2, 3, 4, 5)
		b = vectorext.ForTest(3, 4, 5, 6, 7)
		x = fext.NewElement(2, 0)

		// aBAndXMustNotChange asserts that a and b did not change as this is
		// a very common check in all the sub-tests.
		aBAndXMustNotChange = func(t *testing.T) {
			require.Equal(t, vectorext.ForTest(1, 2, 3, 4, 5), a, "a must not change")
			require.Equal(t, vectorext.ForTest(3, 4, 5, 6, 7), b, "b must not change")
			require.Equal(t, "2+0*u", x.String(), "x must not change")
		}
	)

	t.Run("DeepCopy", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		assert.Equal(t, a, c, "the deep copied vector must be equal")
		c[0] = fext.NewElement(40, 0)
		aBAndXMustNotChange(t)
	})

	t.Run("ScalarMul", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.ScalarMul(c, b, x)
		assert.Equal(t, vectorext.ForTest(6, 8, 10, 12, 14), c, "c must be equal to 2*b")
		vectorext.ScalarMul(c, c, x)
		assert.Equal(t, vectorext.ForTest(12, 16, 20, 24, 28), c, "c must be equal to 2*b")
		aBAndXMustNotChange(t)
	})

	t.Run("ScalarProd", func(t *testing.T) {
		c := vectorext.ScalarProd(a, b)
		assert.Equal(t, "85+0*u", c.String())
		aBAndXMustNotChange(t)
	})

	t.Run("Rand", func(t *testing.T) {
		c, d := vectorext.Rand(5), vectorext.Rand(5)
		assert.NotEqual(t, c, d, "Rand should not return twice the same value")
	})

	t.Run("MulElementWise", func(t *testing.T) {
		c := vectorext.DeepCopy(b)
		vectorext.MulElementWise(c, b, a)
		assert.Equal(t, vectorext.ForTest(3, 8, 15, 24, 35), c)

		c = vectorext.DeepCopy(b)
		vectorext.MulElementWise(c, c, a)
		assert.Equal(t, vectorext.ForTest(3, 8, 15, 24, 35), c)

		aBAndXMustNotChange(t)
	})

	t.Run("Prettify", func(t *testing.T) {
		assert.Equal(t, "[1+0*u, 2+0*u, 3+0*u, 4+0*u, 5+0*u]", vectorext.Prettify(a))
		aBAndXMustNotChange(t)
	})

	t.Run("Reverse", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.Reverse(c)
		assert.Equal(t, vectorext.ForTest(5, 4, 3, 2, 1), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Repeat", func(t *testing.T) {
		c := vectorext.Repeat(x, 5)
		assert.Equal(t, vectorext.ForTest(2, 2, 2, 2, 2), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Add", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.Add(c, a, b)
		assert.Equal(t, vectorext.ForTest(4, 6, 8, 10, 12), c)

		c = vectorext.DeepCopy(a)
		vectorext.Add(c, c, b)
		assert.Equal(t, vectorext.ForTest(4, 6, 8, 10, 12), c)

		c = vectorext.DeepCopy(a)
		vectorext.Add(c, a, b, a)
		assert.Equal(t, vectorext.ForTest(5, 8, 11, 14, 17), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Sub", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.Sub(c, b, a)
		assert.Equal(t, vectorext.ForTest(2, 2, 2, 2, 2), c)

		c = vectorext.DeepCopy(a)
		vectorext.Sub(c, b, c)
		assert.Equal(t, vectorext.ForTest(2, 2, 2, 2, 2), c)

		c = vectorext.DeepCopy(b)
		vectorext.Sub(c, c, a)
		assert.Equal(t, vectorext.ForTest(2, 2, 2, 2, 2), c)
		aBAndXMustNotChange(t)
	})

	t.Run("ZeroPad", func(t *testing.T) {
		c := vectorext.ZeroPad(a, 7)
		assert.Equal(t, vectorext.ForTest(1, 2, 3, 4, 5, 0, 0), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Interleave", func(t *testing.T) {
		c := vectorext.Interleave(a, b)
		assert.Equal(t, vectorext.ForTest(1, 3, 2, 4, 3, 5, 4, 6, 5, 7), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Fill", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.Fill(c, x)
		assert.Equal(t, vectorext.ForTest(2, 2, 2, 2, 2), c)
		aBAndXMustNotChange(t)
	})

	t.Run("PowerVec", func(t *testing.T) {
		c := vectorext.PowerVec(x, 5)
		assert.Equal(t, vectorext.ForTest(1, 2, 4, 8, 16), c)
	})

	t.Run("IntoGnarkAssignment", func(t *testing.T) {
		c := vectorext.IntoGnarkAssignment(a)
		for i := range c {
			first := c[i].A0.(field.Element)
			second := c[i].A1.(field.Element)
			assert.Equal(t, a[i].A0.String(), first.String())
			assert.Equal(t, a[i].A1.String(), second.String())
		}
		aBAndXMustNotChange(t)
	})

}

func TestReverse(t *testing.T) {
	vec := []fext.Element{
		fext.NewElement(0, 0),
		fext.NewElement(1, 0),
		fext.NewElement(2, 0),
	}
	vectorext.Reverse(vec)
	require.Equal(t, vec[0], fext.NewElement(2, 0))
	require.Equal(t, vec[1], fext.NewElement(1, 0))
	require.Equal(t, vec[2], fext.NewElement(0, 0))
}

func TestScalarProd(t *testing.T) {
	require.Equal(t,
		vectorext.ScalarProd(
			vectorext.ForTest(1, 2, 3, 4),
			vectorext.ForTest(1, 2, 3, 4),
		),
		fext.NewElement(30, 0),
	)
}

func TestForTest(t *testing.T) {
	testcase := vectorext.ForTest(-1, 0, 1)
	require.Equal(t, testcase[0].String(), "-1+0*u")
	require.Equal(t, testcase[1].String(), "0+0*u")
	require.Equal(t, testcase[2].String(), "1+0*u")
}

package vectorext_test

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectors(t *testing.T) {

	var (
		// a, b and x are very common vectors in all the tests
		list_a = []int{1, 2, 3, 4, 5, 6, 7, 8}
		list_b = []int{3, 4, 5, 6, 7, 8, 9, 10}

		a = vectorext.ForTestFromQuads(list_a...)
		b = vectorext.ForTestFromQuads(list_b...)
		x = fext.NewFromUint(2, 0, 0, 0)

		// aBAndXMustNotChange asserts that a and b did not change as this is
		// a very common check in all the sub-tests.
		aBAndXMustNotChange = func(t *testing.T) {
			require.Equal(t, vectorext.ForTestFromQuads(list_a...), a, "a must not change")
			require.Equal(t, vectorext.ForTestFromQuads(list_b...), b, "b must not change")
			require.Equal(t, "2+0*u+(0+0*u)*v", x.String(), "x must not change")
		}
	)

	t.Run("DeepCopy", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		assert.Equal(t, a, c, "the deep copied vector must be equal")
		c[0] = fext.NewFromUint(0, 0, 40, 0)
		aBAndXMustNotChange(t)
	})

	t.Run("ScalarMul", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.ScalarMul(c, b, x)
		assert.Equal(t, vectorext.ForTestFromQuads(6, 8, 10, 12, 14, 16, 18, 20), c, "c must be equal to 2*b")
		vectorext.ScalarMul(c, c, x)
		assert.Equal(t, vectorext.ForTestFromQuads(12, 16, 20, 24, 28, 32, 36, 40), c, "c must be equal to 4*b")
		aBAndXMustNotChange(t)
	})

	t.Run("ScalarProd", func(t *testing.T) {
		z1 := vectorext.ForTestCalculateQuadProduct(list_a[:4], list_b[0:4])
		z2 := vectorext.ForTestCalculateQuadProduct(list_a[4:8], list_b[4:8])
		for i := 0; i < len(z1); i++ {
			z2[i] = z1[i] + z2[i]
		}
		c := vectorext.ScalarProd(a, b)
		assert.Equal(t, fmt.Sprintf("%d+%d*u+(%d+%d*u)*v", z2[0], z2[1], z2[2], z2[3]), c.String())
		aBAndXMustNotChange(t)
	})

	t.Run("Rand", func(t *testing.T) {
		c, d := vectorext.Rand(5), vectorext.Rand(5)
		assert.NotEqual(t, c, d, "Rand should not return twice the same value")
	})

	t.Run("MulElementWise", func(t *testing.T) {
		c := vectorext.DeepCopy(b)
		vectorext.MulElementWise(c, b, a)

		z := vectorext.ForTestCalculateQuadProduct(list_a[:4], list_b[0:4])
		z = append(z, vectorext.ForTestCalculateQuadProduct(list_a[4:8], list_b[4:8])...)
		assert.Equal(t, vectorext.ForTestFromQuads(
			z...,
		), c)

		c = vectorext.DeepCopy(b)
		vectorext.MulElementWise(c, c, a)
		assert.Equal(t, vectorext.ForTestFromQuads(
			z...,
		), c)

		aBAndXMustNotChange(t)
	})

	t.Run("Prettify", func(t *testing.T) {
		assert.Equal(t, "[1+2*u+(3+4*u)*v, 5+6*u+(7+8*u)*v]", vectorext.Prettify(a))
		aBAndXMustNotChange(t)
	})

	t.Run("Reverse", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.Reverse(c)
		// we invert the order of the pairs, but not the order inside the pairs as that
		// would lead to different field extensions
		assert.Equal(t, vectorext.ForTestFromQuads(5, 6, 7, 8, 1, 2, 3, 4), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Constant", func(t *testing.T) {
		y := fext.NewFromUint(1, 2, 3, 4)
		c := vectorext.Repeat(y, 2)
		assert.Equal(t, vectorext.ForTestFromQuads(1, 2, 3, 4, 1, 2, 3, 4), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Add", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.Add(c, a, b)
		assert.Equal(t, vectorext.ForTestFromQuads(4, 6, 8, 10, 12, 14, 16, 18), c)

		c = vectorext.DeepCopy(a)
		vectorext.Add(c, c, b)
		assert.Equal(t, vectorext.ForTestFromQuads(4, 6, 8, 10, 12, 14, 16, 18), c)

		c = vectorext.DeepCopy(a)
		vectorext.Add(c, a, b, a)
		assert.Equal(t, vectorext.ForTestFromQuads(5, 8, 11, 14, 17, 20, 23, 26), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Sub", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.Sub(c, b, a)
		assert.Equal(t, vectorext.ForTestFromQuads(2, 2, 2, 2, 2, 2, 2, 2), c)

		c = vectorext.DeepCopy(a)
		vectorext.Sub(c, b, c)
		assert.Equal(t, vectorext.ForTestFromQuads(2, 2, 2, 2, 2, 2, 2, 2), c)

		c = vectorext.DeepCopy(b)
		vectorext.Sub(c, c, a)
		assert.Equal(t, vectorext.ForTestFromQuads(2, 2, 2, 2, 2, 2, 2, 2), c)
		aBAndXMustNotChange(t)
	})

	t.Run("ZeroPad", func(t *testing.T) {
		c := vectorext.ZeroPad(a, 3)
		assert.Equal(t, vectorext.ForTestFromQuads(1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 0), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Interleave", func(t *testing.T) {
		c := vectorext.Interleave(a, b)
		assert.Equal(t, vectorext.ForTestFromQuads(1, 2, 3, 4, 3, 4, 5, 6, 5, 6, 7, 8, 7, 8, 9, 10), c)
		aBAndXMustNotChange(t)
	})

	t.Run("Fill", func(t *testing.T) {
		c := vectorext.DeepCopy(a)
		vectorext.Fill(c, x)
		assert.Equal(t, vectorext.ForTestFromQuads(2, 0, 0, 0, 2, 0, 0, 0), c)
		aBAndXMustNotChange(t)
	})

	t.Run("PowerVec", func(t *testing.T) {
		c := vectorext.PowerVec(x, 3)
		assert.Equal(t, vectorext.ForTestFromQuads(1, 0, 0, 0, 2, 0, 0, 0, 4, 0, 0, 0), c)
	})

	// t.Run("IntoGnarkAssignment", func(t *testing.T) {
	// 	c := vectorext.IntoGnarkAssignment(a)
	// 	var tmp fext.Element
	// 	for i := range c {
	// 		tmp.B0.A0 = c[i].B0.A0.(field.Element)
	// 		tmp.B0.A1 = c[i].B0.A1.(field.Element)
	// 		tmp.B1.A0 = c[i].B1.A0.(field.Element)
	// 		tmp.B1.A1 = c[i].B1.A1.(field.Element)
	// 		assert.Equal(t, a[i].B0.A0.String(), tmp.B0.A0.String())
	// 		assert.Equal(t, a[i].B0.A1.String(), tmp.B0.A1.String())
	// 		assert.Equal(t, a[i].B1.A0.String(), tmp.B1.A0.String())
	// 		assert.Equal(t, a[i].B1.A1.String(), tmp.B1.A1.String())
	// 	}
	// 	aBAndXMustNotChange(t)
	// })

}

func TestReverse(t *testing.T) {
	vec := []fext.Element{
		fext.NewFromUint(7, 0, 0, 0),
		fext.NewFromUint(5, 6, 0, 0),
		fext.NewFromUint(1, 2, 3, 4),
	}
	vectorext.Reverse(vec)
	require.Equal(t, vec[0], fext.NewFromUint(1, 2, 3, 4))
	require.Equal(t, vec[1], fext.NewFromUint(5, 6, 0, 0))
	require.Equal(t, vec[2], fext.NewFromUint(7, 0, 0, 0))
}

func TestScalarProd(t *testing.T) {
	require.Equal(t,
		vectorext.ScalarProd(
			vectorext.ForTest(1, 2, 3, 4),
			vectorext.ForTest(1, 2, 3, 4),
		),
		fext.NewFromUint(30, 0, 0, 0),
	)
}

func TestForTest(t *testing.T) {
	testcase := vectorext.ForTest(-1, 0, 1)
	require.Equal(t, testcase[0].String(), "-1+0*u+(0+0*u)*v")
	require.Equal(t, testcase[1].String(), "0+0*u+(0+0*u)*v")
	require.Equal(t, testcase[2].String(), "1+0*u+(0+0*u)*v")
}

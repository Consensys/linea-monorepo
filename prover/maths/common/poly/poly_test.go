package poly_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"

	"github.com/stretchr/testify/require"
)

func TestEvaluation(t *testing.T) {
	// Just a simple test vector
	testVec := []field.Element{
		field.NewElement(1),
		field.NewElement(2),
		field.NewElement(5),
		field.NewElement(12),
	}

	x := field.NewElement(17)

	y := poly.EvalUnivariate(testVec, x)

	require.Equal(t, y, field.NewElement(60436))

}

func TestMul(t *testing.T) {
	a := vector.ForTest(1, 1)
	b := vector.ForTest(-1, 1)
	expected := vector.ForTest(-1, 0, 1)

	res := poly.Mul(a, b)

	require.Equal(t, vector.Prettify(expected), vector.Prettify(res))
}

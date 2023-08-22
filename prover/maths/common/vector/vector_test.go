package vector_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"

	"github.com/stretchr/testify/require"
)

func TestReverse(t *testing.T) {
	vec := []field.Element{field.NewElement(0), field.NewElement(1), field.NewElement(2)}
	vector.Reverse(vec)
	require.Equal(t, vec[0], field.NewElement(2))
	require.Equal(t, vec[1], field.NewElement(1))
	require.Equal(t, vec[2], field.NewElement(0))
}

func TestScalarProd(t *testing.T) {
	require.Equal(t,
		vector.ScalarProd(
			vector.ForTest(1, 2, 3, 4),
			vector.ForTest(1, 2, 3, 4),
		),
		field.NewElement(30),
	)
}

func TestForTest(t *testing.T) {
	testcase := vector.ForTest(-1, 0, 1)
	require.Equal(t, testcase[0].String(), "-1")
}

func TestMarshal(t *testing.T) {
	// For small values
	v := vector.ForTest(0, 1, 2, 3, 4, 5)
	vUnmarshalled := vector.Unmarshal(vector.Marshal(v), 6)
	require.Equal(t, vector.Prettify(v), vector.Prettify(vUnmarshalled))

	// For large numbers : the invert will produce larger numbers
	v = field.BatchInvert(v)
	vUnmarshalled = vector.Unmarshal(vector.Marshal(v), 6)
	require.Equal(t, vector.Prettify(v), vector.Prettify(vUnmarshalled))
}

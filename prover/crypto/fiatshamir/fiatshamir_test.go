package fiatshamir

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

func TestFiatShamirSafeguardUpdate(t *testing.T) {

	fs := hashtypes.Poseidon2()

	a := RandomFext(fs)
	b := RandomFext(fs)

	// Two consecutive call to fs do not return the same result
	require.NotEqual(t, a.String(), b.String())
}

func TestBatchUpdates(t *testing.T) {

	fs := hashtypes.Poseidon2()
	Update(fs, field.NewElement(2))
	Update(fs, field.NewElement(2))
	Update(fs, field.NewElement(1))
	expectedVal := RandomFext(fs)

	t.Run("for a variadic call", func(t *testing.T) {
		fs := hashtypes.Poseidon2()
		Update(fs, field.NewElement(2), field.NewElement(2), field.NewElement(1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for slice of field elements", func(t *testing.T) {
		fs := hashtypes.Poseidon2()
		UpdateVec(fs, vector.ForTest(2, 2, 1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for multi-slice of field elements", func(t *testing.T) {
		fs := hashtypes.Poseidon2()
		UpdateVec(fs, vector.ForTest(2, 2), vector.ForTest(1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for a smart-vector", func(t *testing.T) {
		sv := smartvectors.RightPadded(vector.ForTest(2, 2), field.NewElement(1), 3)
		fs := hashtypes.Poseidon2()
		UpdateSV(fs, sv)
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

}

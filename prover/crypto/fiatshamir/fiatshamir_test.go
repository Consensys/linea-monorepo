package fiatshamir

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFiatShamirSafeguardUpdate(t *testing.T) {

	fs := NewMiMCFiatShamir()

	a := fs.RandomField()
	b := fs.RandomField()

	// Two consecutive call to fs do not return the same result
	require.NotEqual(t, a.String(), b.String())
}

func TestFiatShamirRandomVec(t *testing.T) {

	for _, testCase := range randomIntVecTestCases {
		testName := fmt.Sprintf("%v-integers-of-%v-bits", testCase.NumIntegers, testCase.IntegerBitSize)
		t.Run(testName, func(t *testing.T) {

			fs := NewMiMCFiatShamir()
			fs.Update(field.NewElement(420))

			var oldState, newState field.Element

			if err := oldState.SetBytesCanonical(fs.hasher.Sum(nil)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			a := fs.RandomManyIntegers(testCase.NumIntegers, 1<<testCase.IntegerBitSize)
			t.Logf("the generated vector= %v", a)

			if err := newState.SetBytesCanonical(fs.hasher.Sum(nil)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Len(t, a, testCase.NumIntegers)
			assert.NoError(t, errIfHasDuplicate(a...))

			if testCase.ShouldUpdate {
				assert.NotEqualf(t, oldState.String(), newState.String(), "the state was not updated (%v)", oldState.String())
			} else {
				assert.Equal(t, oldState.String(), newState.String(), "the state was updated")
			}
		})
	}
}

func TestBatchUpdates(t *testing.T) {

	fs := NewMiMCFiatShamir()
	fs.Update(field.NewElement(2))
	fs.Update(field.NewElement(2))
	fs.Update(field.NewElement(1))
	expectedVal := fs.RandomField()

	t.Run("for a variadic call", func(t *testing.T) {
		fs := NewMiMCFiatShamir()
		fs.Update(field.NewElement(2), field.NewElement(2), field.NewElement(1))
		actualValue := fs.RandomField()
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for slice of field elements", func(t *testing.T) {
		fs := NewMiMCFiatShamir()
		fs.UpdateVec(vector.ForTest(2, 2, 1))
		actualValue := fs.RandomField()
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for multi-slice of field elements", func(t *testing.T) {
		fs := NewMiMCFiatShamir()
		fs.UpdateVec(vector.ForTest(2, 2), vector.ForTest(1))
		actualValue := fs.RandomField()
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for a smart-vector", func(t *testing.T) {
		sv := smartvectors.RightPadded(vector.ForTest(2, 2), field.NewElement(1), 3)
		fs := NewMiMCFiatShamir()
		fs.UpdateSV(sv)
		actualValue := fs.RandomField()
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

}

// errIfHasDuplicate returns an error if a duplicate is found in the caller's slice
// indicating the positions and the value corresponding to the first duplicate.
func errIfHasDuplicate[T comparable](slice ...T) error {
	m := map[T]int{}
	for i, f := range slice {
		if prevPos, ok := m[f]; ok {
			return fmt.Errorf("found a duplicate value: %v at positions %v and %v", f, prevPos, i)
		}

		m[f] = i
	}
	return nil
}

// Test vectors that are used to test the RandomIntVec feature
var randomIntVecTestCases = []struct {
	ShouldUpdate   bool
	NumIntegers    int
	IntegerBitSize int
}{
	{
		ShouldUpdate:   false,
		NumIntegers:    0,
		IntegerBitSize: 16,
	},
	{
		ShouldUpdate:   true,
		NumIntegers:    1,
		IntegerBitSize: 16,
	},
	{
		ShouldUpdate:   true,
		NumIntegers:    15,
		IntegerBitSize: 16,
	},
	{
		ShouldUpdate:   true,
		NumIntegers:    16,
		IntegerBitSize: 16,
	},
	{
		ShouldUpdate:   true,
		NumIntegers:    17,
		IntegerBitSize: 16,
	},
	{
		ShouldUpdate:   true,
		NumIntegers:    45,
		IntegerBitSize: 16,
	},
}

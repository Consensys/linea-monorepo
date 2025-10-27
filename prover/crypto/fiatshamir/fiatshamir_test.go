package fiatshamir

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFiatShamirSafeguardUpdate(t *testing.T) {

	fs := poseidon2.Poseidon2()

	a := RandomFext(fs)
	b := RandomFext(fs)

	// Two consecutive call to fs do not return the same result
	require.NotEqual(t, a.String(), b.String())
}

func TestFiatShamirRandomVec(t *testing.T) {

	for _, testCase := range randomIntVecTestCases {
		testName := fmt.Sprintf("%v-integers-of-%v-bits", testCase.NumIntegers, testCase.IntegerBitSize)
		t.Run(testName, func(t *testing.T) {

			fs := poseidon2.Poseidon2()
			Update(fs, field.NewElement(420))

			var oldState, newState field.Element
			d := types.HashToBytes32(fs.SumElement())
			var ss [32]byte
			copy(ss[:], d[:])

			if err := oldState.SetBytesCanonical(ss[:4]); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			a := RandomManyIntegers(fs, testCase.NumIntegers, 1<<testCase.IntegerBitSize)
			t.Logf("the generated vector= %v", a)

			d = types.HashToBytes32(fs.SumElement())
			copy(ss[:], d[:])
			if err := newState.SetBytesCanonical(ss[:4]); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Len(t, a, testCase.NumIntegers)
			assert.NoError(t, errIfHasDuplicate(a...))

			if testCase.ShouldUpdate {
				errStr := fmt.Sprintf("the state was not updated (%v)", oldState.String())
				if oldState.Equal(&newState) {
					t.Fatal(errStr)
				}
			} else {
				errStr := "the state was updated"
				if !oldState.Equal(&newState) {
					t.Fatal(errStr)
				}
			}
		})
	}
}
func TestBatchUpdates(t *testing.T) {

	fs := poseidon2.Poseidon2()
	Update(fs, field.NewElement(2))
	Update(fs, field.NewElement(2))
	Update(fs, field.NewElement(1))
	expectedVal := RandomFext(fs)

	t.Run("for a variadic call", func(t *testing.T) {
		fs := poseidon2.Poseidon2()
		Update(fs, field.NewElement(2), field.NewElement(2), field.NewElement(1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for slice of field elements", func(t *testing.T) {
		fs := poseidon2.Poseidon2()
		UpdateVec(fs, vector.ForTest(2, 2, 1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for multi-slice of field elements", func(t *testing.T) {
		fs := poseidon2.Poseidon2()
		UpdateVec(fs, vector.ForTest(2, 2), vector.ForTest(1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for a smart-vector", func(t *testing.T) {
		sv := smartvectors.RightPadded(vector.ForTest(2, 2), field.NewElement(1), 3)
		fs := poseidon2.Poseidon2()
		UpdateSV(fs, sv)
		actualValue := RandomFext(fs)
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

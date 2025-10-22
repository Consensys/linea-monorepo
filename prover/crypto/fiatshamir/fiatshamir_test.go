package fiatshamir

import (
	"fmt"
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
	fs := hashtypes.Poseidon2()

	a := RandomFext(fs)
	b := RandomFext(fs)

	// Two consecutive call to fs do not return the same result
	require.NotEqual(t, a.String(), b.String())
}

// func TestFiatShamirRandomVec(t *testing.T) {

// 	for _, testCase := range randomIntVecTestCases {
// 		testName := fmt.Sprintf("%v-integers-of-%v-bits", testCase.NumIntegers, testCase.IntegerBitSize)
// 		t.Run(testName, func(t *testing.T) {

// 			fs := hashtypes.Poseidon2()
// 			Update(fs, field.NewElement(420))

// 			var oldState, newState field.Element

// 			ss := fs.Sum(nil)
// 			if err := oldState.SetBytesCanonical(ss[:4]); err != nil {
// 				t.Fatalf("unexpected error: %v", err)
// 			}

// 			a := RandomManyIntegers(fs, testCase.NumIntegers, 1<<testCase.IntegerBitSize)
// 			t.Logf("the generated vector= %v", a)

// 			ss = fs.Sum(nil)
// 			if err := newState.SetBytesCanonical(ss[:4]); err != nil {
// 				t.Fatalf("unexpected error: %v", err)
// 			}

// 			assert.Len(t, a, testCase.NumIntegers)
// 			assert.NoError(t, errIfHasDuplicate(a...))

// 			if testCase.ShouldUpdate {
// 				errStr := fmt.Sprintf("the state was not updated (%v)", oldState.String())
// 				if oldState.Equal(&newState) {
// 					t.Fatal(errStr)
// 				}
// 			} else {
// 				errStr := "the state was updated"
// 				if !oldState.Equal(&newState) {
// 					t.Fatal(errStr)
// 				}
// 			}
// 		})
// 	}
// }

func TestBatchUpdates(t *testing.T) {

	fs := hashtypes.Poseidon2()
	fs := hashtypes.Poseidon2()
	Update(fs, field.NewElement(2))
	Update(fs, field.NewElement(2))
	Update(fs, field.NewElement(1))
	expectedVal := RandomFext(fs)

	t.Run("for a variadic call", func(t *testing.T) {
		fs := hashtypes.Poseidon2()
		fs := hashtypes.Poseidon2()
		Update(fs, field.NewElement(2), field.NewElement(2), field.NewElement(1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for slice of field elements", func(t *testing.T) {
		fs := hashtypes.Poseidon2() // Updated to use hashtypes.Poseidon2()
		UpdateVec(fs, vector.ForTest(2, 2, 1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for multi-slice of field elements", func(t *testing.T) {
		fs := hashtypes.Poseidon2()
		fs := hashtypes.Poseidon2()
		UpdateVec(fs, vector.ForTest(2, 2), vector.ForTest(1))
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

	t.Run("for a smart-vector", func(t *testing.T) {
		sv := smartvectors.RightPadded(vector.ForTest(2, 2), field.NewElement(1), 3)
		fs := hashtypes.Poseidon2()
		fs := hashtypes.Poseidon2()
		UpdateSV(fs, sv)
		actualValue := RandomFext(fs)
		assert.Equal(t, expectedVal.String(), actualValue.String())
	})

}

// TestSamplingFromSeed tests that sampling from seed is consistent, does not forget
// bytes from the seed and does not depends on the state.
func TestSamplingFromSeed(t *testing.T) {

	t.Run("dependence-on-name", func(t *testing.T) {

		totalSize := 1000
		name := ""
		resMap := map[field.Element]struct{}{}

		var seed, initialState field.Element
		seed.SetRandom()
		initialState.SetRandom()

		fs := hashtypes.Poseidon2()
		fs.SetState(initialState.Marshal())

		for i := 0; i < totalSize; i++ {

			x := RandomFieldFromSeed(fs, seed, name)
			if _, found := resMap[x]; found {
				t.Errorf("found a collision for i=%v", i)
			}

			resMap[x] = struct{}{}
			name += "a"
		}
	})

	t.Run("non-dependance-curr-state", func(t *testing.T) {

		var s1, s2, seed field.Element
		seed.SetRandom()
		s1.SetRandom()
		s2.SetRandom()

		name := "string-name"
		fs1 := hashtypes.Poseidon2() // Updated to use hashtypes.Poseidon2()
		fs2 := hashtypes.Poseidon2()

		fs1.SetState(s1.Marshal())
		fs2.SetState(s2.Marshal())

		y1 := RandomFieldFromSeed(fs1, seed, name)
		y2 := RandomFieldFromSeed(fs2, seed, name)

		if y1 != y2 {
			t.Errorf("starting from different state does not give the same results")
		}
	})

	t.Run("does-not-modify-state", func(t *testing.T) {

		var initialState, seed field.Element
		initialState.SetRandom()
		seed.SetRandom()

		name := "ddqsdjqskljd"
		fs := hashtypes.Poseidon2()

		oldState := initialState.Marshal()
		fs.SetState(oldState)
		oldState = fs.State() // we read the state again because padding might happen in setState

		RandomFieldFromSeed(fs, seed, name)

		newState := fs.State()

		errStr := "state was modified"

		for i := 0; i < len(newState); i++ {
			if newState[i] != oldState[i] {
				t.Fatal(errStr)
			}
		}
	})

	t.Run("is-repeatable", func(t *testing.T) {

		var initialState, seed field.Element
		initialState.SetRandom()
		seed.SetRandom()
		name := "ddqsdjqskljd"
		fs := hashtypes.Poseidon2()

		bInitialState := make([]byte, fs.BlockSize())
		copy(bInitialState, initialState.Marshal())
		fs.SetState(initialState.Marshal())

		y1 := RandomFieldFromSeed(fs, seed, name)
		y2 := RandomFieldFromSeed(fs, seed, name)

		if y1 != y2 {
			t.Errorf("state was modified")
		}
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

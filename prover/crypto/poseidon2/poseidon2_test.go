package poseidon2_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var rng = rand.New(utils.NewRandSource(0)) // #nosec G404

// Test different input sizes to ensure consistency with the Merkle-Damgard construction.
var testCases = []struct {
	name        string
	numElements int
}{
	{"SmallInput", 10},
	{"LargeInput", 512},
	{"SingleBlock", 8},
}

// This test ensures that the Poseidon2BlockCompression function is correctly implemented and produces the same output as
// the hashtypes.Poseidon2(), which uses Write and Sum methods to get the final hash output
//
// We hash and compress one Octuplet at a time
func TestPoseidon2BlockCompression(t *testing.T) {

	for i := 0; i < 100; i++ {
		var state field.Octuplet
		var input field.Octuplet

		var inputBytes [32]byte
		for i := 0; i < 8; i++ {
			startIndex := i * 4
			input[i] = field.PseudoRand(rng)
			valBytes := input[i].Bytes()
			copy(inputBytes[startIndex:startIndex+4], valBytes[:])
		}

		// Compute hash using the Poseidon2BlockCompression.
		h := poseidon2.Poseidon2BlockCompression(state, input)

		// Compute hash using the NewMerkleDamgardHasher implementation.
		merkleHasher := hashtypes.Poseidon2()
		merkleHasher.Reset()
		merkleHasher.Write(inputBytes[:]) // write one 32 bytes (equivalent to one Octuplet)
		newBytes := merkleHasher.Sum(nil)

		var result field.Octuplet
		for i := 0; i < 8; i++ {
			startIndex := i * 4
			segment := newBytes[startIndex : startIndex+4]
			var newElement koalabear.Element
			newElement.SetBytes(segment)
			result[i] = newElement
			require.Equal(t, result[i].String(), h[i].String())

		}

	}
}

// This test ensures that the Poseidon2Sponge function is correctly implemented and produces the same output as
// the hashtypes.Poseidon2(), which uses Write and Sum methods to get the final hash output
// We write and compress the 'whole slice'
func TestPoseidon2SpongeConsistency(t *testing.T) {
	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < 10; i++ {
				// Generate random input.
				input := make([]field.Element, tc.numElements)
				inputBytes := make([]byte, 0, tc.numElements*field.Bytes)
				for j := 0; j < tc.numElements; j++ {
					input[j] = field.PseudoRand(rng)
					bytes := input[j].Bytes()
					inputBytes = append(inputBytes, bytes[:]...)
				}

				fmt.Printf("len inputBytes: %d\n", len(inputBytes))
				// Compute hash using the Poseidon2Sponge function.
				state := poseidon2.Poseidon2Sponge(input)

				// Compute hash using the reference Merkle-Damgard hasher.
				merkleHasher := hashtypes.Poseidon2()
				merkleHasher.Reset()
				merkleHasher.Write(inputBytes[:])
				newBytes := merkleHasher.Sum(nil)

				var result field.Octuplet
				for i := 0; i < 8; i++ {
					startIndex := i * 4
					segment := newBytes[startIndex : startIndex+4]
					var newElement koalabear.Element
					newElement.SetBytes(segment)
					result[i] = newElement
					require.Equal(t, result[i].String(), state[i].String())
				}

			}
		})
	}
}

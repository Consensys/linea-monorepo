package poseidon2_test

import (
	"encoding/hex"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
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

// TestPoseidonTestVectors ensures that the test vectors that we use for
// besu-native are working.
func TestPoseidonTestVectors(t *testing.T) {

	type testVec struct {
		name      string
		inputHex  string
		outputHex string
	}

	testVectors := []testVec{
		{
			name:      "empty",
			inputHex:  "0x",
			outputHex: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:      "just-one-byte",
			inputHex:  "0x01",
			outputHex: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:      "4-bytes",
			inputHex:  "0x01020304",
			outputHex: "0x1d1620185895cf146bed58674aa2156f72d95ae2644cf35e7a80f8bb7bbe8f5d",
		},
		{
			name:      "few-bytes",
			inputHex:  "0x0102040405050606090c0201",
			outputHex: "0x1f617ad3102d502a2852fe482e2ce6c24bb4ecb31dbcfd0f79249fa1576649a8",
		},
		{
			name:      "32-zero-bytes",
			inputHex:  "0x0000000000000000000000000000000000000000000000000000000000000000",
			outputHex: "0x0656ab853b3f52840362a8177e217b630c3f876b11e848365145aa24220647fc",
		},
		{
			name:      "31-zero-and-1-byte",
			inputHex:  "0x0000000000000000000000000000000000000000000000000000000000000001",
			outputHex: "0x532d760a239087f458d23ee549db4a5771815a387616ec5f31be90fd690886a5",
		},
		{
			name:      "long-string",
			inputHex:  "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000700000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000009000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000b000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000d000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000f",
			outputHex: "0x254c857251520cbd40981dd74c2b3ee345acf16978e701324181926236278aaa",
		},
	}

	for _, tc := range testVectors {
		t.Run(tc.name, func(t *testing.T) {
			input, e := hex.DecodeString(strings.TrimPrefix(tc.inputHex, "0x"))
			if e != nil {
				t.Fatal(e)
			}

			h := poseidon2.Poseidon2()
			h.Reset()
			h.Write(input)
			out := h.Sum(nil)

			outHex := "0x" + hex.EncodeToString(out)
			require.Equal(t, tc.outputHex, outHex)
		})
	}

}

// This test ensures that the Poseidon2BlockCompression function is correctly implemented and produces the same output as
// the poseidon2.Poseidon2(), which uses Write and Sum methods to get the final hash output
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
		h := vortex.CompressPoseidon2(state, input)

		// Compute hash using the NewMerkleDamgardHasher implementation.
		merkleHasher := poseidon2.Poseidon2()
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
// the poseidon2.Poseidon2(), which uses Write and Sum methods to get the final hash output
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

				// Compute hash using the Poseidon2Sponge function.
				state := poseidon2.Poseidon2Sponge(input)

				// Compute hash using the reference Merkle-Damgard hasher.
				merkleHasher := poseidon2.Poseidon2()
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

func TestFieldHasher(t *testing.T) {
	assert := require.New(t)

	h1 := poseidon2.Poseidon2()
	h2 := poseidon2.Poseidon2()
	randInputs := make(field.Vector, 10)
	randInputs.MustSetRandom()

	// test Write + Sum
	for _, elem := range randInputs {
		h1.Write(elem.Marshal())
	}
	dgst1 := h1.Sum(nil)
	var dgst1Byte32 types.Bytes32
	copy(dgst1Byte32[:], dgst1[:])

	// test WriteElement + SumElement
	h2.WriteElements(randInputs)
	dgst2 := h2.SumElement()
	assert.Equal(types.Bytes32ToOctuplet(dgst1Byte32), dgst2, "hashes do not match")

}

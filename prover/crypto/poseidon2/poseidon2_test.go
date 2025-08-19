package poseidon2_test

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func TestPoseidon2Sponge(t *testing.T) {
	// This test ensures that the Poseidon2Sponge function is correctly implemented and produces the same output as
	// the poseidon2.NewMerkleDamgardHasher(), which uses Write and Sum methods to get the final hash output

	// We write and compress one Element at a time
	var rng = rand.New(utils.NewRandSource(0))
	for i := 0; i < 1; i++ {

		var state field.Octuplet
		var input [512]field.Element

		elementNumber := len(input)

		merkleHasher := poseidon2.NewPoseidon2()
		merkleHasher.Reset()
		for i := 0; i < elementNumber; i++ {
			input[i] = field.PseudoRand(rng)
			inputBytes := input[i].Bytes()
			merkleHasher.Write(inputBytes[:]) // Write one Element at a time
		}
		state = poseidon2.Poseidon2Sponge(input[:])

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
}

func TestPoseidon2BlockCompression(t *testing.T) {
	// This test ensures that the Poseidon2BlockCompression function is correctly implemented and produces the same output as
	// the poseidon2.NewMerkleDamgardHasher(), which uses Write and Sum methods to get the final hash output

	// We hash and compress one Octuplet at a time
	var rng = rand.New(utils.NewRandSource(0))

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

		h := poseidon2.Poseidon2BlockCompression(state, input)

		merkleHasher := poseidon2.NewPoseidon2()
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

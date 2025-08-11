package poseidon2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

func TestPoseidon2TwoBlockCompressions(t *testing.T) {
	// This test ensures that the CompressPoseidon2 function is correctly implemented and produces the same output as
	// the poseidon2.NewMerkleDamgardHasher(), which uses Write and Sum methods to get the final hash output

	for i := 0; i < 100; i++ {
		var zero, blockOne, blocktwo [8]field.Element

		var input [16]field.Element

		var inputBytesone, inputbytestwo [32]byte
		for i := 0; i < 16; i++ {

			if i < 8 {
				startIndex := i * 4
				input[i].SetRandom()
				blockOne[i] = input[i]
				valBytes := input[i].Bytes()
				copy(inputBytesone[startIndex:startIndex+4], valBytes[:])

			} else {
				startIndex := i*4 - 32
				input[i].SetRandom()
				blocktwo[i-8] = input[i]
				valBytes := input[i].Bytes()
				copy(inputbytestwo[startIndex:startIndex+4], valBytes[:])
			}

		}
		// Compute the hash using the iterative chaining method calling BlockCompressionMekle.
		state := poseidon2.BlockCompressionMekle(zero, blockOne)
		h := poseidon2.BlockCompressionMekle(state, blocktwo)

		// Compute the hash using the standard `Write`/`Sum` interface.
		merkleHasher := poseidon2.NewPoseidon2()
		merkleHasher.Reset()
		merkleHasher.Write(inputBytesone[:])
		merkleHasher.Write(inputbytestwo[:])

		newBytes := merkleHasher.Sum(nil)

		var result [8]field.Element

		for i := 0; i < 8; i++ {
			startIndex := i * 4
			segment := newBytes[startIndex : startIndex+4]
			var newElement field.Element
			newElement.SetBytes(segment)
			result[i] = newElement
			require.Equal(t, result[i].String(), h[i].String())
		}

	}
}

func TestPoseidon2BlockCompression(t *testing.T) {
	// This test ensures that the CompressPoseidon2 function is correctly implemented and produces the same output as
	// the poseidon2.NewMerkleDamgardHasher(), which uses Write and Sum methods to get the final hash output

	for i := 0; i < 100; i++ {
		var zero [8]koalabear.Element
		var input [8]koalabear.Element

		var inputBytes [32]byte
		for i := 0; i < 8; i++ {
			startIndex := i * 4
			input[i].SetRandom()
			valBytes := input[i].Bytes()
			copy(inputBytes[startIndex:startIndex+4], valBytes[:])
		}

		h := poseidon2.BlockCompressionMekle(zero, input)

		merkleHasher := poseidon2.NewPoseidon2()
		merkleHasher.Reset()
		merkleHasher.Write(inputBytes[:])
		newBytes := merkleHasher.Sum(nil)

		var result [8]koalabear.Element // Array to store the 8 reconstructed Elements

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

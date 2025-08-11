package poseidon2

import (
	"hash"

	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

const (
	blockSize = 8
)

// NewPoseidon2 wraps [poseidon2.NewMerkleDamgardHasher], this is used to limit the number of gnark-crypto imports.
func NewPoseidon2() hash.Hash {
	return poseidon2.NewMerkleDamgardHasher()
}

// TODO@yao Sponge and Merkle

// BlockCompressionMekle applies the Poseidon2 block compression function to a given block
// over a given state. This what is run under the hood by the Poseidon2 hash function
func BlockCompressionMekle(oldState, block [blockSize]field.Element) (newState [blockSize]field.Element) {
	res := vortex.Hash{}
	var x [2 * blockSize]field.Element
	copy(x[:], oldState[:])
	copy(x[8:], block[:])

	// Create a buffer to hold the feed-forward input.
	copy(res[:], x[8:])
	if err := poseidon2.NewPermutation(16, 6, 21).Permutation(x[:]); err != nil {
		panic(err)
	}

	for i := range res {
		res[i].Add(&res[i], &x[8+i])
	}
	return res
}

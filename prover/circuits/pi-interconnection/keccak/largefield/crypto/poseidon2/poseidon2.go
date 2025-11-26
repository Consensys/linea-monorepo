package poseidon2

import (
	"hash"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
	"github.com/consensys/gnark/frontend"
	poseidon2permutation "github.com/consensys/gnark/std/permutation/poseidon2"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/maths/field"
)

// NewPoseidon2 wraps [poseidon2.NewMerkleDamgardHasher], this is used to limit the number of gnark-crypto imports.
func NewPoseidon2() hash.Hash {
	return poseidon2.NewMerkleDamgardHasher()
}

// BlockCompression applies the Poseidon2 block compression function to a given block
// over a given state. This what is run under the hood by the Poseidon2 hash function
// in Miyaguchi-Preneel mode.
func BlockCompression(oldState, block field.Element) (newState field.Element) {
	compressor := poseidon2.NewDefaultPermutation()
	res, err := compressor.Compress(oldState.Marshal(), block.Marshal())
	if err != nil {
		panic(err)
	}
	if err := newState.SetBytesCanonical(res); err != nil {
		panic(err)
	}
	return
}

// GnarkBlockCompression applies the Poseidon2 permutation to a given block within
// a gnark circuit and mirrors exactly [BlockCompression].
func GnarkBlockCompression(api frontend.API, oldState, block frontend.Variable) (newState frontend.Variable) {
	compressor, err := poseidon2permutation.NewPoseidon2(api)
	if err != nil {
		panic(err)
	}
	return compressor.Compress(oldState, block)
}

// HashVec hashes a vector of field elements
func HashVec(v []field.Element) (h field.Element) {
	state := NewPoseidon2()
	for i := range v {
		vBytes := v[i].Bytes()
		state.Write(vBytes[:])
	}
	h.SetBytes(state.Sum(nil))
	return
}

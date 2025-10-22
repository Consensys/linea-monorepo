package hashtypes

import (
	"hash"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
)

// Poseidon2Hasher is a concrete implementation of the FieldHasher interface using the Poseidon2 hash function.
type Poseidon2Hasher struct {
	// By embedding hash.Hash, Poseidon2Hasher satisfies the standard Go hash interface,
	// allowing it to be used with standard library packages that expect a hash.Hash.
	hash.Hash
	maxValue Bytes32 // the maximal value obtainable with that hasher

	// Sponge construction state
	h    field.Octuplet
	data []field.Element // data to hash
}

// ///// Implementation for the FieldHasher interface ///////
func (p Poseidon2Hasher) SumElements(xs []field.Element) field.Octuplet {
	return poseidon2.Poseidon2Sponge(xs)
}

func (p Poseidon2Hasher) MaxBytes32() Bytes32 {
	return p.maxValue
}

// ///// Constructor for Poseidon2Hasher /////
func Poseidon2() Poseidon2Hasher {
	var maxVal field.Octuplet
	for i := range maxVal {
		maxVal[i] = field.NewFromString("-1")
	}
	return Poseidon2Hasher{
		Hash:     gnarkposeidon2.NewMerkleDamgardHasher(),
		maxValue: HashToBytes32(maxVal),
		h:        field.Octuplet{},
		data:     make([]field.Element, 0),
	}
}

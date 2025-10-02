package hashtypes

import (
	"hash"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
)

// Create a new FieldHasher

// FieldHasher defines an interface for hashing operations that work directly on field elements,
type FieldHasher interface {
	// TODO@yao: consider width
	MaxBytes32() Bytes32
	// FieldHash takes a slice of field elements and returns a fixed-size hash output.
	// It's typically implemented using a sponge construction.
	FieldHash(xs []field.Element) field.Octuplet
	// BlockCompression is a function used in Merkle trees to combine two child hashes
	// (left, right) into a single parent hash.
	BlockCompression(left, right field.Octuplet) field.Octuplet
}

// Poseidon2Hasher is a concrete implementation of the FieldHasher interface using the Poseidon2 hash function.
type Poseidon2Hasher struct {
	// By embedding hash.Hash, Poseidon2Hasher satisfies the standard Go hash interface,
	// allowing it to be used with standard library packages that expect a hash.Hash.
	hash.Hash
	maxValue Bytes32 // the maximal value obtainable with that hasher
}

// ///// Implementation for the FieldHasher interface ///////
func (p Poseidon2Hasher) FieldHash(xs []field.Element) field.Octuplet {
	return poseidon2.Poseidon2Sponge(xs)
}

func (p Poseidon2Hasher) BlockCompression(left, right field.Octuplet) field.Octuplet {
	return poseidon2.Poseidon2BlockCompression(left, right)
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
	}
}

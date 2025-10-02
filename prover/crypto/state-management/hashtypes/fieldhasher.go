package hashtypes

import (
	"hash"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
)

// ----- FieldHasher Interface ----- //

// FieldHasher defines the interface for any field-element-based hashing algorithm
type FieldHasher interface {
	FieldHash(xs []field.Element) field.Octuplet // Hashes field elements into an Octuplet (e.g., Poseidon sponge)
	MaxBytes32() Bytes32                         // Returns the maximum representable value as Bytes32
}

// ----- Hasher Struct ----- //

// Hasher is a generic hashing implementation that takes a backend FieldHasher, and a standard hash.Hash
type Hasher struct {
	hash.Hash             // Any byte-stream hash implementation (e.g., Merkle-Damgard)
	Backend   FieldHasher // Backend implementation of field-element-based hashing
	MaxByte32 Bytes32     // Precomputed maximum field value from the backend
}

// FieldHash delegates the field hashing operation to the backend FieldHasher
func (f Hasher) FieldHash(xs []field.Element) field.Octuplet {
	return f.Backend.FieldHash(xs)
}

// MaxBytes32 delegates the retrieval of max field value to the backend FieldHasher
func (f Hasher) MaxBytes32() Bytes32 {
	return f.MaxByte32
}

// Constructor for Hasher
func NewHasher(backend FieldHasher, hasher hash.Hash) Hasher {
	return Hasher{
		Hash:      hasher,
		Backend:   backend,
		MaxByte32: backend.MaxBytes32(),
	}
}

// ----- Poseidon2 Implementation ----- //

// Poseidon2FieldHasher is a specific implementation of FieldHasher using the Poseidon2 function
type Poseidon2FieldHasher struct {
	maxValue Bytes32
}

// FieldHash computes the Poseidon2 Sponge hash based on field elements
func (p Poseidon2FieldHasher) FieldHash(xs []field.Element) field.Octuplet {
	return poseidon2.Poseidon2Sponge(xs)
}

// MaxBytes32 returns the maximal field value that Poseidon2 can work with
func (p Poseidon2FieldHasher) MaxBytes32() Bytes32 {
	return p.maxValue
}

// Constructor for Poseidon2FieldHasher, later be used as backend in Poseidon2 Hasher
func NewPoseidon2Backend() Poseidon2FieldHasher {
	var maxVal field.Octuplet // This stores the maximal value for each element
	for i := range maxVal {
		maxVal[i] = field.NewFromString("-1") // Initialize max field value (field modulus - 1)
	}
	return Poseidon2FieldHasher{
		maxValue: HashToBytes32(maxVal), // Convert the max value Octuplet to Bytes32 representation
	}
}

// Constructor for Poseidon2
func Poseidon2() Hasher {
	poseidonBackend := NewPoseidon2Backend()
	fieldHasher := NewHasher(poseidonBackend, gnarkposeidon2.NewMerkleDamgardHasher())
	return fieldHasher
}

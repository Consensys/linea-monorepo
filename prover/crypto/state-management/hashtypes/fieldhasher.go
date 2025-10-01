package hashtypes

import (
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

const HashSize = 8

type FieldHasher interface {
	Hash(xs []field.Element) [HashSize]field.Element
}

// Create a new Poseidon2 hasher
type poseidon2Hasher struct {
	// TODO@yao: consider width
}

func (p poseidon2Hasher) Hash(xs []field.Element) [HashSize]field.Element {
	return poseidon2.Poseidon2Sponge(xs)
}

func Poseidon2() FieldHasher {
	return poseidon2Hasher{}
}

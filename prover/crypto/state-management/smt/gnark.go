package smt

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
)

// Merkle proofs
type GnarkProof struct {
	Path     frontend.Variable
	Siblings []frontend.Variable
}

// Returns the root as an output of the Merkle verification
func GnarkRecoverRoot(
	api frontend.API,
	proof GnarkProof,
	leaf frontend.Variable,
	h hash.FieldHasher) frontend.Variable {

	current := leaf
	nbBits := len(proof.Siblings)
	b := api.ToBinary(proof.Path, nbBits)
	for i := 0; i < len(proof.Siblings); i++ {
		h.Reset()
		left := api.Select(b[i], proof.Siblings[i], current)
		right := api.Select(b[i], current, proof.Siblings[i])
		h.Write(left, right)
		current = h.Sum()
	}

	return current
}

func GnarkVerifyMerkleProof(
	api frontend.API,
	proof GnarkProof,
	leaf frontend.Variable,
	root frontend.Variable,
	h hash.FieldHasher) {

	r := GnarkRecoverRoot(api, proof, leaf, h)
	api.AssertIsEqual(root, r)

}

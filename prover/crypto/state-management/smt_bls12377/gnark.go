package smt_bls12377

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
)

// GnarkProof mirrors [Proof] in a gnark circuit.
type GnarkProof struct {
	Path     frontend.Variable
	Siblings []frontend.Variable
}

// GnarkRecoverRoot is as [RecoverRoot] in a gnark circuit. The provided
// position is range-checked against the number of siblings in the Merkle proof
// object.
func GnarkRecoverRoot(
	api frontend.API,
	proof GnarkProof,
	leaf frontend.Variable,
	h poseidon2_bls12377.GnarkMDHasher) frontend.Variable {

	current := leaf
	nbBits := len(proof.Siblings)
	api.Println("Path:", proof.Path)

	b := api.ToBinary(proof.Path, nbBits)
	api.Println("bits:", b)
	for i := 0; i < len(proof.Siblings); i++ {
		h.Reset()
		left := api.Select(b[i], proof.Siblings[i], current)
		right := api.Select(b[i], current, proof.Siblings[i])
		api.Println("Left state:", left)
		api.Println("Right state:", right)

		h.Write(left, right)
		current = h.Sum()
		api.Println("Current state:", current)

	}

	return current
}

// GnarkVerifyMerkleProof asserts the validity of a [GnarkProof] against a root.
func GnarkVerifyMerkleProof(
	api frontend.API,
	proof GnarkProof,
	leaf frontend.Variable,
	root frontend.Variable,
	h poseidon2_bls12377.GnarkMDHasher) {

	api.Println("root circuit:", root)
	api.Println("leaf circuit:", leaf)
	api.Println("proof circuit", proof)

	r := GnarkRecoverRoot(api, proof, leaf, h)
	api.Println("Computed root:", r)
	api.Println("Expected root:", root)
	api.AssertIsEqual(root, r)
}

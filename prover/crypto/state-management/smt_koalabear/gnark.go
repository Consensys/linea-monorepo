package smt_koalabear

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// GnarkProof mirrors [Proof] in a gnark circuit.
type GnarkProof struct {
	Path     frontend.Variable
	Siblings []zk.Octuplet
}

// selectOcuplet if b=1, returns l else return r
func selectOcuplet(b frontend.Variable, l, r zk.Octuplet) zk.Octuplet {

}

// GnarkRecoverRoot computes the root form the proof and the leaf
func GnarkRecoverRoot(
	api frontend.API,
	proof GnarkProof,
	leaf zk.Octuplet,
	h poseidon2.GnarkHasher) zk.Octuplet {

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

// GnarkVerifyMerkleProof asserts the validity of a [GnarkProof] against a root.
func GnarkVerifyMerkleProof(
	api frontend.API,
	proof GnarkProof,
	leaf frontend.Variable,
	root frontend.Variable,
	h hash.FieldHasher) {

	r := GnarkRecoverRoot(api, proof, leaf, h)

	api.AssertIsEqual(root, r)
}

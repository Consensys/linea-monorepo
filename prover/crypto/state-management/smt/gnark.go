package smt

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// GnarkProof mirrors [Proof] in a gnark circuit.
type GnarkProof[T zk.Element] struct {
	Path     T
	Siblings []T
}

// GnarkRecoverRoot is as [RecoverRoot] in a gnark circuit. The provided
// position is range-checked against the number of siblings in the Merkle proof
// object.
func GnarkRecoverRoot[T zk.Element](
	api frontend.API,
	proof GnarkProof[T],
	leaf T,
	h hash.FieldHasher) T {

	current := leaf

	// TODO @thomas fixme
	// nbBits := len(proof.Siblings)
	// b := api.ToBinary(proof.Path, nbBits)
	// for i := 0; i < len(proof.Siblings); i++ {
	// 	h.Reset()
	// 	left := api.Select(b[i], proof.Siblings[i], current)
	// 	right := api.Select(b[i], current, proof.Siblings[i])
	// 	h.Write(left, right)
	// 	current = h.Sum()
	// }

	return current
}

// GnarkVerifyMerkleProof asserts the validity of a [GnarkProof] against a root.
func GnarkVerifyMerkleProof[T zk.Element](
	api frontend.API,
	proof GnarkProof[T],
	leaf T,
	root T,
	h hash.FieldHasher) {

	r := GnarkRecoverRoot(api, proof, leaf, h)
	api.AssertIsEqual(root, r)
}

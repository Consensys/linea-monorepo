package smt

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// GnarkProof mirrors [Proof] in a gnark circuit.
type GnarkProof struct {
	Path     zk.WrappedVariable
	Siblings []zk.WrappedVariable
}

// GnarkRecoverRoot is as [RecoverRoot] in a gnark circuit. The provided
// position is range-checked against the number of siblings in the Merkle proof
// object.
func GnarkRecoverRoot(
	api frontend.API,
	proof GnarkProof,
	leaf zk.WrappedVariable,
	h hash.FieldHasher) zk.WrappedVariable {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	current := leaf
	nbBits := len(proof.Siblings)
	b := apiGen.ToBinary(&proof.Path, nbBits)
	for i := 0; i < len(proof.Siblings); i++ {
		h.Reset()
		left := apiGen.Select(b[i], &proof.Siblings[i], &current)
		right := apiGen.Select(b[i], &current, &proof.Siblings[i])
		h.Write(left, right)
		tmp := h.Sum()
		current = zk.WrapFrontendVariable(tmp)
	}

	return current
}

// GnarkVerifyMerkleProof asserts the validity of a [GnarkProof] against a root.
func GnarkVerifyMerkleProof(
	api frontend.API,
	proof GnarkProof,
	leaf zk.WrappedVariable,
	root zk.WrappedVariable,
	h hash.FieldHasher) {

	r := GnarkRecoverRoot(api, proof, leaf, h)
	api.AssertIsEqual(root, r)
}

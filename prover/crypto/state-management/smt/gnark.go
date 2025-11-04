package smt

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// GnarkProof mirrors [Proof] in a gnark circuit.
type GnarkProof struct {
	Path     zk.WrappedVariable
	Siblings []poseidon2.GHash
}

// GnarkRecoverRoot is as [RecoverRoot] in a gnark circuit. The provided
// position is range-checked against the number of siblings in the Merkle proof
// object.
func GnarkRecoverRoot(
	api frontend.API,
	proof GnarkProof,
	leaf poseidon2.GHash,
	h poseidon2.GnarkHasher) poseidon2.GHash {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	current := leaf
	nbBits := len(proof.Siblings)
	b := apiGen.ToBinary(proof.Path, nbBits)
	for i := 0; i < len(proof.Siblings); i++ {
		h.Reset()
		var left, right poseidon2.GHash
		if b[i] == 1 {
			left = proof.Siblings[i]
			right = current

		} else {
			left = current
			right = proof.Siblings[i]
		}

		slices := make([]zk.WrappedVariable, 16)
		copy(slices[0:8], left[:])
		copy(slices[8:16], right[:])
		h.Write(slices...)
		current = h.Sum()
	}

	return current
}

// GnarkVerifyMerkleProof asserts the validity of a [GnarkProof] against a root.
func GnarkVerifyMerkleProof(
	api frontend.API,
	proof GnarkProof,
	leaf poseidon2.GHash,
	root poseidon2.GHash,
	h poseidon2.GnarkHasher) {

	r := GnarkRecoverRoot(api, proof, leaf, h)
	api.AssertIsEqual(root, r)
}

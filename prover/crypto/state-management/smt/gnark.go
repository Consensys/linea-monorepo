package smt

import (
	"fmt"

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
	apiGen zk.GenericApi,
	proof GnarkProof,
	leaf poseidon2.GHash,
	h poseidon2.GnarkHasher) poseidon2.GHash {

	current := leaf
	nbBits := len(proof.Siblings)
	b := apiGen.ToBinary(proof.Path, nbBits)

	fmt.Printf("nbBits=%d\n", nbBits)
	fmt.Printf("len(proof.Siblings)=%d\n", len(b))
	for i := 0; i < len(proof.Siblings); i++ {
		h.Reset()
		var left, right poseidon2.GHash

		for j := 0; j < 8; j++ {
			left[j] = apiGen.Select(b[i], proof.Siblings[i][j], current[j])
			right[j] = apiGen.Select(b[i], current[j], proof.Siblings[i][j])
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
	apiGen zk.GenericApi,
	proof GnarkProof,
	leaf poseidon2.GHash,
	root poseidon2.GHash,
	h poseidon2.GnarkHasher) {

	// check the result
	r := GnarkRecoverRoot(apiGen, proof, leaf, h)

	for i := 0; i < 8; i++ {
		apiGen.AssertIsEqual(root[i], r[i])
	}
}

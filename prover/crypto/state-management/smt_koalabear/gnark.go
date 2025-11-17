package smt_koalabear

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
)

// GnarkProof mirrors [Proof] in a gnark circuit.
type GnarkProof struct {
	Path     frontend.Variable
	Siblings []poseidon2_koalabear.Octuplet
}

// selectOcuplet if b=1, returns l else return r
func selectOcuplet(api frontend.API, b frontend.Variable, l, r poseidon2_koalabear.Octuplet) poseidon2_koalabear.Octuplet {
	var res poseidon2_koalabear.Octuplet
	for i := 0; i < 8; i++ {
		res[i] = api.Select(b, l[i], r[i])
	}
	return res
}

// GnarkRecoverRoot computes the root form the proof and the leaf
func GnarkRecoverRoot(
	api frontend.API,
	proof GnarkProof,
	leaf poseidon2_koalabear.Octuplet,
	h poseidon2_koalabear.GnarkMDHasher) poseidon2_koalabear.Octuplet {

	current := leaf
	nbBits := len(proof.Siblings)
	b := api.ToBinary(proof.Path, nbBits)
	for i := 0; i < len(proof.Siblings); i++ {
		h.Reset()
		left := selectOcuplet(api, b[i], proof.Siblings[i], current)
		right := selectOcuplet(api, b[i], current, proof.Siblings[i])
		h.WriteOctuplet(left, right)
		current = h.Sum()
	}

	return current
}

// GnarkVerifyMerkleProof asserts the validity of a [GnarkProof] against a root.
func GnarkVerifyMerkleProof(
	api frontend.API,
	proof GnarkProof,
	leaf poseidon2_koalabear.Octuplet,
	root poseidon2_koalabear.Octuplet,
	h poseidon2_koalabear.GnarkMDHasher) {

	r := GnarkRecoverRoot(api, proof, leaf, h)

	for i := 0; i < 8; i++ {
		api.AssertIsEqual(r[i], root[i])
	}
}

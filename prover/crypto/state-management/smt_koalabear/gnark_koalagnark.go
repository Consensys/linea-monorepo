package smt_koalabear

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
)

// KoalagnarkGnarkProof mirrors [Proof] using koalagnark types for emulated circuits.
type KoalagnarkGnarkProof struct {
	Path     koalagnark.Element
	Siblings []poseidon2_koalabear.KoalagnarkOctuplet
}

// koalagnarkSelectOctuplet: if b=1, returns l else return r
func koalagnarkSelectOctuplet(koalaAPI *koalagnark.API, b frontend.Variable, l, r poseidon2_koalabear.KoalagnarkOctuplet) poseidon2_koalabear.KoalagnarkOctuplet {
	var res poseidon2_koalabear.KoalagnarkOctuplet
	for i := 0; i < 8; i++ {
		res[i] = koalaAPI.Select(b, l[i], r[i])
	}
	return res
}

// KoalagnarkRecoverRoot computes the Merkle root from the proof and the leaf
// using koalagnark types (works in both native and emulated circuits).
func KoalagnarkRecoverRoot(
	koalaAPI *koalagnark.API,
	proof KoalagnarkGnarkProof,
	leaf poseidon2_koalabear.KoalagnarkOctuplet,
) poseidon2_koalabear.KoalagnarkOctuplet {

	h := poseidon2_koalabear.NewKoalagnarkMDHasher(koalaAPI.Frontend())

	current := leaf
	nbBits := len(proof.Siblings)
	b := koalaAPI.ToBinary(proof.Path, nbBits)

	for i := 0; i < len(proof.Siblings); i++ {
		h.Reset()
		left := koalagnarkSelectOctuplet(koalaAPI, b[i], proof.Siblings[i], current)
		right := koalagnarkSelectOctuplet(koalaAPI, b[i], current, proof.Siblings[i])
		h.WriteOctuplet(left, right)
		current = h.Sum()
	}

	return current
}

// KoalagnarkVerifyMerkleProof asserts the validity of a [KoalagnarkGnarkProof]
// against a root using koalagnark types.
func KoalagnarkVerifyMerkleProof(
	koalaAPI *koalagnark.API,
	proof KoalagnarkGnarkProof,
	leaf poseidon2_koalabear.KoalagnarkOctuplet,
	root poseidon2_koalabear.KoalagnarkOctuplet,
) {
	r := KoalagnarkRecoverRoot(koalaAPI, proof, leaf)
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(r[i], root[i])
	}
}

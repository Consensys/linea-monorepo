package smt_koalabear

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
)

// GnarkProof mirrors [Proof] in a gnark circuit.
type GnarkProof struct {
	Path     frontend.Variable
	Siblings []poseidon2_koalabear.GnarkOctuplet
}

// KoalagnarkProof mirrors [Proof] but uses koalagnark-based octuplets.
type KoalagnarkProof struct {
	Path     frontend.Variable
	Siblings []poseidon2_koalabear.KoalagnarkOctuplet
}

// selectOcuplet if b=1, returns l else return r
func selectOcuplet(api frontend.API, b frontend.Variable, l, r poseidon2_koalabear.GnarkOctuplet) poseidon2_koalabear.GnarkOctuplet {
	var res poseidon2_koalabear.GnarkOctuplet
	for i := 0; i < 8; i++ {
		res[i] = api.Select(b, l[i], r[i])
	}
	return res
}

// GnarkRecoverRoot computes the root form the proof and the leaf
func GnarkRecoverRoot(
	api frontend.API,
	proof GnarkProof,
	leaf poseidon2_koalabear.GnarkOctuplet) (poseidon2_koalabear.GnarkOctuplet, error) {

	h, err := poseidon2_koalabear.NewGnarkMDHasher(api)
	if err != nil {
		return poseidon2_koalabear.GnarkOctuplet{}, err
	}

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

	return current, nil
}

// GnarkVerifyMerkleProof asserts the validity of a [GnarkProof] against a root.
func GnarkVerifyMerkleProof(
	api frontend.API,
	proof GnarkProof,
	leaf poseidon2_koalabear.GnarkOctuplet,
	root poseidon2_koalabear.GnarkOctuplet) error {

	r, err := GnarkRecoverRoot(api, proof, leaf)
	if err != nil {
		return err
	}

	for i := 0; i < 8; i++ {
		api.AssertIsEqual(r[i], root[i])
	}
	return nil
}

// selectKoalagnarkOctuplet if b=1, returns l else return r
func selectKoalagnarkOctuplet(api *koalagnark.API, b frontend.Variable, l, r poseidon2_koalabear.KoalagnarkOctuplet) poseidon2_koalabear.KoalagnarkOctuplet {
	var res poseidon2_koalabear.KoalagnarkOctuplet
	for i := 0; i < 8; i++ {
		res[i] = api.Select(b, l[i], r[i])
	}
	return res
}

// KoalagnarkRecoverRoot computes the root from the proof and the leaf using koalagnark elements.
func KoalagnarkRecoverRoot(
	api frontend.API,
	proof KoalagnarkProof,
	leaf poseidon2_koalabear.KoalagnarkOctuplet,
) poseidon2_koalabear.KoalagnarkOctuplet {
	koalaAPI := koalagnark.NewAPI(api)
	h := poseidon2_koalabear.NewKoalagnarkMDHasher(api)

	current := leaf
	nbBits := len(proof.Siblings)
	if nbBits == 0 {
		return current
	}
	b := api.ToBinary(proof.Path, nbBits)
	for i := 0; i < len(proof.Siblings); i++ {
		h.Reset()
		left := selectKoalagnarkOctuplet(koalaAPI, b[i], proof.Siblings[i], current)
		right := selectKoalagnarkOctuplet(koalaAPI, b[i], current, proof.Siblings[i])
		h.WriteOctuplet(left, right)
		current = h.Sum()
	}

	return current
}

// KoalagnarkVerifyMerkleProof asserts the validity of a [KoalagnarkProof] against a root.
func KoalagnarkVerifyMerkleProof(
	api frontend.API,
	proof KoalagnarkProof,
	leaf poseidon2_koalabear.KoalagnarkOctuplet,
	root poseidon2_koalabear.KoalagnarkOctuplet,
) error {
	koalaAPI := koalagnark.NewAPI(api)
	r := KoalagnarkRecoverRoot(api, proof, leaf)
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(r[i], root[i])
	}
	return nil
}

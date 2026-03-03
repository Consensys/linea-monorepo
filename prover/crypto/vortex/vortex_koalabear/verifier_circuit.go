package vortex_koalabear

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
)

// GnarkCheckColumnInclusionNoSis checks the merkle proof opening
// (merkleProofs[i][j], root[i]) for columns[i][j].
// The leaves are poseidon2_koalabear(columns[i][j]).
// Uses koalagnark types for correct emulated-mode support.
func GnarkCheckColumnInclusionNoSis(api frontend.API, columns [][][]koalagnark.Element,
	merkleProofs [][]smt_koalabear.KoalagnarkGnarkProof, roots []poseidon2_koalabear.KoalagnarkOctuplet) {

	koalaAPI := koalagnark.NewAPI(api)

	for i := 0; i < len(roots); i++ {

		h := poseidon2_koalabear.NewKoalagnarkMDHasher(api)

		for j := 0; j < len(columns[i]); j++ {

			// compute leaf = poseidon2_koalabear(columns[i][j]))
			h.Write(columns[i][j]...)
			leaf := h.Sum()
			h.Reset()

			// check merkle proof
			smt_koalabear.KoalagnarkVerifyMerkleProof(koalaAPI, merkleProofs[i][j], leaf, roots[i])
		}
	}
}

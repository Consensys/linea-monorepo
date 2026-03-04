package vortex_bn254

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bn254"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bn254"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
)

// GnarkCheckColumnInclusionNoSis verifies BN254 Merkle proofs in a gnark circuit.
// The leaves are computed as poseidon2_bn254(columns[i][j]).
// columns[i][j] are KoalaBear elements (as koalagnark.Element).
// roots[i] are BN254 field elements (as frontend.Variable).
func GnarkCheckColumnInclusionNoSis(api frontend.API, columns [][][]koalagnark.Element,
	merkleProofs [][]smt_bn254.GnarkProof, roots []frontend.Variable) error {

	for i := 0; i < len(roots); i++ {
		h, err := poseidon2_bn254.NewGnarkMDHasher(api)
		if err != nil {
			return err
		}

		for j := 0; j < len(columns[i]); j++ {
			// compute leaf = poseidon2_bn254(columns[i][j])
			h.WriteWVs(columns[i][j]...)
			leaf := h.Sum()
			h.Reset()

			err = smt_bn254.GnarkVerifyMerkleProof(api, merkleProofs[i][j], leaf, roots[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

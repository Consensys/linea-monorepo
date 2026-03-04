package vortex_bn254

import (
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bn254"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bn254"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// CheckColumnInclusion verifies BN254 Merkle proofs for column openings.
// The leaves are poseidon2_bn254(columns[i][j]).
func CheckColumnInclusion(columns [][][]field.Element,
	merkleProofs [][]smt_bn254.Proof, commitments []Commitment) error {

	h := poseidon2_bn254.NewMDHasher()
	for i := 0; i < len(commitments); i++ {
		for j := 0; j < len(columns[i]); j++ {
			h.Reset()
			h.WriteKoalabearElements(columns[i][j]...)
			leaf := h.SumElement()

			err := smt_bn254.Verify(&merkleProofs[i][j], leaf, commitments[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Verify performs the full Vortex verification with BN254 Merkle trees.
func Verify(params *Params, proof *vortex.OpeningProof, vi *vortex.VerifierInput, commitments []Commitment, merkleProofs [][]smt_bn254.Proof) error {
	err := vortex.CheckIsCodeWord(params.RsParams, proof.LinearCombination)
	if err != nil {
		return err
	}

	err = vortex.CheckLinComb(proof.LinearCombination, vi.EntryList, vi.Alpha, proof.Columns)
	if err != nil {
		return err
	}

	err = vortex.CheckStatement(proof.LinearCombination, vi.Ys, vi.X, vi.Alpha)
	if err != nil {
		return err
	}

	err = CheckColumnInclusion(proof.Columns, merkleProofs, commitments)
	return err
}

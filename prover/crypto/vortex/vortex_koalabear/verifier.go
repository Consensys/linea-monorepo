package vortex_koalabear

import (
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Check the merkle proof opening (merkleProofs[i][j], root[i]) for columns[i][j].
// The leaves are poseidon2_koalabear(sis(columns[i][j]))
func CheckColumnInclusionSis(sis *ringsis.Key, columns [][][]field.Element,
	merkleProofs [][]smt_koalabear.Proof, commitments []Commitment) error {

	sisHash := make([]field.Element, sis.OutputSize())
	// var leaf fr.Element
	h := poseidon2_koalabear.NewMDHasher()
	for i := 0; i < len(commitments); i++ {

		for j := 0; j < len(columns[i]); j++ {

			// compute leaf = poseidon2_koalabear(columns[i][j]))
			sis.SisGnarkCrypto.Hash(columns[i][j], sisHash)
			h.Reset()
			h.WriteElements(sisHash...)
			leaf := h.SumElement()

			// check merkle proof
			err := smt_koalabear.Verify(&merkleProofs[i][j], leaf, commitments[i])
			if err != nil {
				return err
			}

		}
	}
	return nil
}

// Check the merkle proof opening (merkleProofs[i][j], root[i]) for columns[i][j].
// The leaves are poseidon2_koalabear(columns[i][j])
func CheckColumnInclusionNoSis(columns [][][]field.Element,
	merkleProofs [][]smt_koalabear.Proof, commitments []Commitment) error {

	for i := 0; i < len(commitments); i++ {

		h := poseidon2_koalabear.NewMDHasher()

		for j := 0; j < len(columns[i]); j++ {

			// compute leaf = poseidon2_koalabear(columns[i][j]))
			h.WriteElements(columns[i][j]...)
			leaf := h.SumElement()
			h.Reset()

			// check merkle proof
			err := smt_koalabear.Verify(&merkleProofs[i][j], leaf, commitments[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type config struct {
	withSis bool
}

type VerifierOption func(c *config) error

func WithSis() VerifierOption {
	return func(c *config) error {
		c.withSis = true
		return nil
	}
}

func VerifySIS(params *Params, proof *vortex.OpeningProof, vi *vortex.VerifierInput, commitments []Commitment, merkleProofs [][]smt_koalabear.Proof) error {

	err := VerifyCommon(params, proof, vi)
	if err != nil {
		return err
	}

	err = CheckColumnInclusionSis(params.Key, proof.Columns,
		merkleProofs, commitments)

	return err
}

func Verify(params *Params, proof *vortex.OpeningProof, vi *vortex.VerifierInput, commitments []Commitment, merkleProofs [][]smt_koalabear.Proof, option ...VerifierOption) error {

	var c config
	for _, opts := range option {
		err := opts(&c)
		if err != nil {
			return err
		}
	}

	err := VerifyCommon(params, proof, vi)
	if err != nil {
		return err
	}

	if c.withSis {
		err = CheckColumnInclusionSis(params.Key, proof.Columns,
			merkleProofs, commitments)
	} else {
		err = CheckColumnInclusionNoSis(proof.Columns,
			merkleProofs, commitments)
	}

	return err
}

func VerifyCommon(params *Params, proof *vortex.OpeningProof, vi *vortex.VerifierInput) error {

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

	return err

}

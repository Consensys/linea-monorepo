package vortex_bls12377

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Gnark params
type GParams struct {
	Key         *ringsis.Key
	HasherFunc  func(frontend.API) (poseidon2_bls12377.GnarkMDHasher, error)
	NoSisHasher func(frontend.API) (poseidon2_bls12377.GnarkMDHasher, error)
}

func (p *GParams) HasNoSisHasher() bool {
	return p.NoSisHasher != nil
}

// Check the merkle proof opening (merkleProofs[i][j], root[i]) for columns[i][j].
// The leaves are poseidon2_bls12377(sis(columns[i][j]))
func CheckColumnInclusionSis(sis *ringsis.Key, columns [][][]field.Element,
	merkleProofs [][]smt_bls12377.Proof, commitments []Commitment) error {

	sisHash := make([]field.Element, sis.OutputSize())
	// var leaf fr.Element
	h := poseidon2_bls12377.NewMDHasher()
	for i := 0; i < len(commitments); i++ {

		for j := 0; j < len(columns[i]); j++ {

			// compute leaf = poseidon2_bls12377(columns[i][j]))
			sis.SisGnarkCrypto.Hash(columns[i][j], sisHash)
			h.Reset()
			h.WriteKoalabearElements(sisHash...)
			leaf := h.SumElement()

			// check merkle proof
			err := smt_bls12377.Verify(&merkleProofs[i][j], leaf, commitments[i])
			if err != nil {
				return err
			}

		}
	}
	return nil
}

// Check the merkle proof opening (merkleProofs[i][j], root[i]) for columns[i][j].
// The leaves are poseidon2_bls12377(columns[i][j])
func CheckColumnInclusionNoSis(columns [][][]field.Element,
	merkleProofs [][]smt_bls12377.Proof, commitments []Commitment) error {

	for i := 0; i < len(commitments); i++ {

		h := poseidon2_bls12377.NewMDHasher()

		for j := 0; j < len(columns[i]); j++ {

			// compute leaf = poseidon2_bls12377(columns[i][j]))
			h.WriteKoalabearElements(columns[i][j]...)
			leaf := h.SumElement()
			h.Reset()

			// check merkle proof
			err := smt_bls12377.Verify(&merkleProofs[i][j], leaf, commitments[i])
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

func VerifySIS(params *Params, proof *vortex.OpeningProof, vi *vortex.VerifierInput, commitments []Commitment, merkleProofs [][]smt_bls12377.Proof) error {

	err := VerifyCommon(params, proof, vi)
	if err != nil {
		return err
	}

	err = CheckColumnInclusionSis(params.Key, proof.Columns,
		merkleProofs, commitments)

	return err
}

func Verify(params *Params, proof *vortex.OpeningProof, vi *vortex.VerifierInput, commitments []Commitment, merkleProofs [][]smt_bls12377.Proof, option ...VerifierOption) error {

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

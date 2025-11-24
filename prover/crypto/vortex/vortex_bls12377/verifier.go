package vortex

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

var (
	// ErrInvalidVerifierInputs flags a verifier input with invalid dimensions.
	ErrInvalidVerifierInputs = errors.New("invalid verification input")
)

// VerifierInputs represents the statement made by the prover in the opening
// protocol of Vortex. It stands for univariate evaluation as desribed in the
// paper. The struct stands for the input of the checks of the verifier, it
// does not cover the construction of the random coins.
//
// Thus, the caller is responsible for handling the construction of RandomCoin
// and EntryList as they are the random coins. It can be done by sampling them
// at random in the interactive setting or using the Fiat-Shamir heuristic.
//
// In our settings, the caller is a function in a framework managing the random
// coins as Vortex is used a sub-protocol of a larger protocol.
type VerifierInputs struct {
	vortex.AlgebraicCheckInputs

	// Merkle checks, it uses below parameters and EntryList from AlgebraicCheckInputs
	BLS12_377_Params GnarkParams
	// MerkleProofs store a list of [smt.Proof] (Merkle proofs) allegedly
	// attesting the membership of the columns in the commitment tree.
	//
	// MerkleProofs[i][j] corresponds to the Merkle proof attesting the j-th
	// column of the i-th commitment root hash.
	MerkleProofs [][]smt_bls12377.Proof
	// MerkleRoots are the commitment to verify the opening for
	MerkleRoots []types.Bytes32
	// Flag indicating if the SIS hash is replaced for the particular round
	// the default behavior is to use the SIS hash function along with the
	// Poseidon2 hash function
	IsSISReplacedByPoseidon2 []bool
}

// VerifyOpening verifies a Vortex opening proof, see [VerifierInputs] for
// more details. The function returns an error on failure. The function validates
// the dimensions of the items in the proof and returns an error if they are
// inconsistent with the parameters. If the provided parameters are invalid
// themselves the function may panic.
func VerifyOpening(v *VerifierInputs) error {

	var (
		numCommitments = len(v.MerkleRoots)
		numEntries     = len(v.AlgebraicCheckInputs.EntryList)
		proof          = v.OpeningProof
		merkleProofs   = v.MerkleProofs
	)

	if numCommitments == 0 {
		return ErrInvalidVerifierInputs
	}

	if len(v.Ys) != numCommitments ||
		len(proof.Columns) != numCommitments ||
		len(merkleProofs) != numCommitments ||
		proof.LinearCombination.Len() != v.Koalabear_Params.NumEncodedCols() {
		return ErrInvalidVerifierInputs
	}

	for i := range v.MerkleRoots {

		if len(proof.Columns[i]) != numEntries ||
			len(merkleProofs[i]) != numEntries ||
			len(v.Ys[i]) == 0 ||
			len(v.Ys[i]) > v.Koalabear_Params.MaxNbRows {
			return ErrInvalidVerifierInputs
		}

		for j := range v.AlgebraicCheckInputs.EntryList {
			if len(proof.Columns[i][j]) != len(v.Ys[i]) ||
				len(proof.Columns) == 0 {
				return ErrInvalidVerifierInputs
			}
		}
	}

	if err := v.Koalabear_Params.IsCodewordExt(v.OpeningProof.LinearCombination); err != nil {
		return err
	}

	if err := v.AlgebraicCheckInputs.CheckColLinCombination(numCommitments); err != nil {
		return err
	}

	if err := v.AlgebraicCheckInputs.CheckStatement(); err != nil {
		return err
	}

	if err := v.checkColumnInclusion(); err != nil {
		return err
	}

	return nil
}

// checkColumnInclusion checks the inclusion of the opened column in their
// respective Merkle trees at the positions requested by the verifier. Returns
// an error if a proof does not pass.
func (v *VerifierInputs) checkColumnInclusion() error {

	var (
		mTreeHashConfig = &smt_bls12377.Config{
			HashFunc: v.BLS12_377_Params.GnarkMerkleHashFunc,
			Depth:    utils.Log2Ceil(v.BLS12_377_Params.NumEncodedCols()),
		}
	)

	// If IsSISReplacedByPoseidon2 is not assigned, we assign them with default false values
	if v.IsSISReplacedByPoseidon2 == nil {
		v.IsSISReplacedByPoseidon2 = make([]bool, len(v.MerkleRoots))
	}

	for i := 0; i < len(v.MerkleRoots); i++ {
		for j := 0; j < len(v.AlgebraicCheckInputs.EntryList); j++ {

			var (
				// Selected columns #j contained in the commitment #i.
				selectedSubCol = v.OpeningProof.Columns[i][j]
				leaf           types.Bytes32
				entry          = v.AlgebraicCheckInputs.EntryList[j]
				root           = v.MerkleRoots[i]
				mProof         = v.MerkleProofs[i][j]
			)

			if !v.IsSISReplacedByPoseidon2[i] {
				panic("the SIS hash function is not supported in BLS12-377 vortex verifier")
			} else {
				// We assume that HashFunc (to be used for Merkle Tree) and NoSisHashFunc()
				// (to be used for in place of SIS hash) are the same i.e. the Poseidon2 hash function
				hasher := smt_bls12377.Poseidon2()
				hasher.Reset()
				colBytes := EncodeKoalabearsToBytes(selectedSubCol)

				hasher.Write(colBytes[:])
				leaf = types.AsBytes32(hasher.Sum(nil))

			}
			if j == 0 {
				var rootElem fr.Element
				fmt.Printf("root non-circuit=%v\n", rootElem.SetBytes(root[:]))
			}
			// Check the Merkle-proof for the obtained leaf
			ok := mProof.Verify(mTreeHashConfig, leaf, root)
			if !ok {
				return fmt.Errorf("merkle proof failed for com #%v and entry %v (mProof.path=%v)", i, j, mProof.Path)
			}

			// And check that the Merkle proof is related to the correct entry
			if mProof.Path != entry {
				return fmt.Errorf("expected the Merkle proof to hold for position %v but was %v", entry, mProof.Path)
			}
		}
	}

	return nil
}

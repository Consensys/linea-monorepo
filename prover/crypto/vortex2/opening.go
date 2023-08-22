package vortex2

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
)

// Opening proof for a Vortex commitment
type Proof struct {
	// columns on against which the linear combination is checked
	// (the i-th entry is the EntryList[i]-th column). The columns may
	// as well be dispatched in several matrices.
	// Columns [i][j][k] returns the k-th entry of the j-th selected
	// column of the i-th commitment
	Columns [][][]field.Element

	// Linear combination of the rows of the polynomial P written as a square matrix
	LinearCombination smartvectors.SmartVector

	// Optional, Merkle-proof of opening of the Columns. Only for Merkle-mode
	MerkleProofs [][]smt.Proof
}

// The committedSV are not in reed-solomon form (i.e. they are the original ones)
// the result is the linear combination in reed-solomon form.
func (params *Params) OpenWithLC(committedSV []smartvectors.SmartVector, randomCoin field.Element) *Proof {
	proof := Proof{}

	// Compute the linear combination
	linComb := make([]field.Element, params.NbColumns)
	parallel.ExecuteChunky(len(linComb), func(start, stop int) {
		subTask := make([]smartvectors.SmartVector, 0, len(committedSV))
		for i := range committedSV {
			subTask = append(subTask, committedSV[i].SubVector(start, stop))
		}
		// Collect the result in the larger slice at the end
		subResult := smartvectors.PolyEval(subTask, randomCoin)
		subResult.WriteInSlice(linComb[start:stop])
	})

	linCombSV := smartvectors.NewRegular(linComb)
	proof.LinearCombination = params.rsEncode(linCombSV)
	return &proof
}

// Complete the proof with the columns chosen in entryList (implictly by the verifier)
func (proof *Proof) WithEntryList(committedMatrices []CommittedMatrix, entryList []int) {
	selectedColumns := make([][][]field.Element, len(committedMatrices))

	for i := range committedMatrices {
		selectedColumns[i] = make([][]field.Element, len(entryList))
		for j := range entryList {
			col := make([]field.Element, len(committedMatrices[i]))
			for k := range committedMatrices[i] {
				col[k] = committedMatrices[i][k].Get(entryList[j])
			}
			selectedColumns[i][j] = col
		}
	}

	proof.Columns = selectedColumns
}

// Complete the proof with the opened columns Merkle proofs
func (p *Proof) WithMerkleProof(trees []*smt.Tree, entryList []int) {

	numTrees := len(trees)
	proofs := make([][]smt.Proof, numTrees)

	// Generate the proofs for each tree and each entry
	for treeID, tree := range trees {
		proofs[treeID] = make([]smt.Proof, len(entryList))
		for k, entry := range entryList {
			proofs[treeID][k] = tree.Prove(entry)
		}
	}

	p.MerkleProofs = proofs
}

/*
Verify an opening proof
  - commitments : the commitments for all rounds
  - proof : the batched evaluation proof
  - x : statement : the evaluation point
  - ys : statement : the claimed evaluations
  - randomCoin : verifier randomness to compute the
    linear combination of the rows
  - entryList : verifier randomness : the columns to select
*/
func (params *Params) VerifyOpening(commitments []Commitment, proof *Proof,
	x field.Element, ys [][]field.Element, randomCoin field.Element,
	entryList []int,
) error {

	// Sanity-check
	if params.IsMerkleMode() {
		utils.Panic("Called `VerifyMerkle` for params in `MerkleMode` or remove the `WithMerkleMode`")
	}

	// Starts by running the generic checks
	selectedColsHashes, err := params.verifyCommon(proof, x, ys, randomCoin, entryList)
	if err != nil {
		return err
	}

	// Size of a Merkle digest
	hashSize := 1 // The right value if
	if !params.HasSisReplacement() {
		hashSize = params.Key.Degree
	}

	// Check that the computed hashes are consistent with the commitment
	// This works for both the SIS and the no-SIS settings
	for i, com := range commitments {
		for j, entry := range entryList {
			if !vector.Equal(
				com[entry*hashSize:(entry+1)*hashSize],
				selectedColsHashes[i][j],
			) {
				return fmt.Errorf("inconsistent digest fol column %v", entry)
			}
		}
	}

	return nil
}

/*
verify an opening proof in Merkle-mode
  - merkleRoots : the commitments for all rounds
  - proof : the batched evaluation proof
  - x : statement : the evaluation point
  - ys : statement : the claimed evaluations
  - randomCoin : verifier randomness to compute the linear combination of the rows
  - entryList : verifier randomness : the columns to select
*/
func (params *Params) VerifyMerkle(
	merkleRoots []hashtypes.Digest,
	proof *Proof,
	x field.Element,
	ys [][]field.Element,
	randomCoin field.Element,
	entryList []int,
) error {

	// Sanity-check
	if !params.IsMerkleMode() {
		utils.Panic("Called `VerifyMerkle` for params in non `MerkleMode`")
	}

	// Starts by running the generic checks
	selectedColsHashes, err := params.verifyCommon(proof, x, ys, randomCoin, entryList)
	if err != nil {
		return err
	}

	// Size of a Merkle digest
	hasher := params.HashFunc()
	config := smt.Config{HashFunc: func() hashtypes.Hasher { return hashtypes.Hasher{Hash: params.HashFunc()} }}

	// Check that the computed hashes are consistent with the commitment
	for i, root := range merkleRoots {
		for j, entry := range entryList {
			// Hash the SIS hash
			var leaf hashtypes.Digest

			if !params.HasSisReplacement() {
				// Hash the SIS hashes
				hasher.Reset()
				for _, x := range selectedColsHashes[i][j] {
					xBytes := x.Bytes()
					hasher.Write(xBytes[:])
				}
				copy(leaf[:], hasher.Sum(nil))
			} else {
				// Pass the column hashes directly
				if len(selectedColsHashes[i][j]) > 1 {
					panic("unexpected number of hashes")
				}
				leaf = selectedColsHashes[i][j][0].Bytes()
			}

			// Check the Merkle-proof for the obtained leaf
			ok := proof.MerkleProofs[i][j].Verify(&config, leaf, root)
			if !ok {
				return fmt.Errorf("merkle proof failed for com #%v and entry %v", i, j)
			}

			// And check that the Merkle proof is related to the correct entry
			if proof.MerkleProofs[i][j].Path != entry {
				return fmt.Errorf("expected the Merkle proof to hold for position %v but was %v", entry, proof.MerkleProofs[i][j].Path)
			}
		}
	}

	return nil

}

// verify generics all the common checks for the Vortex verifier. It is used
// both in Vanilla mode and in Merkle mode
func (params *Params) verifyCommon(
	proof *Proof,
	x field.Element,
	ys [][]field.Element,
	randomCoin field.Element,
	entryList []int,
) (selectedColSisDigests [][][]field.Element, err error) {

	// The linear combination should be a correct codeword
	if err := params.isCodeword(proof.LinearCombination); err != nil {
		return nil, fmt.Errorf("incorrect linear combination : %v", err)
	}

	// Check the consistency of Ys and proof.Linear combination
	Yjoined := utils.Join(ys...)
	alphaY := smartvectors.Interpolate(proof.LinearCombination, x)
	alphaYProme := poly.EvalUnivariate(Yjoined, randomCoin)
	if alphaY != alphaYProme {
		return nil, fmt.Errorf("RowLincomb and Y are inconsistent")
	}

	// Size of the hash of 1 column
	numRounds := len(ys)

	// Test that the opened columns are consistent with the commitments
	// and with the linear combination
	selectedColSisDigests = make([][][]field.Element, numRounds)
	for j, selectedColID := range entryList {

		// Will carry the concatenation of the columns for the same entry j
		fullCol := []field.Element{}

		for i := range selectedColSisDigests {
			if j == 0 {
				selectedColSisDigests[i] = make([][]field.Element, len(entryList))
			}

			// Entries of the selected columns #j contained in the commitment #i.
			selectedSubCol := proof.Columns[i][j]
			fullCol = append(fullCol, selectedSubCol...)

			// Check consistency between the opened column and the commitment
			if !params.HasSisReplacement() {
				// Use the provided SIS hasher
				selectedColSisDigests[i][j] = params.Key.Hash(selectedSubCol)
			} else {
				// Use the provided SIS function has a substitution
				hasher := params.NoSisHashFunc()
				hasher.Reset()
				for k := range selectedSubCol {
					xBytes := selectedSubCol[k].Bytes()
					hasher.Write(xBytes[:])
				}
				digestBytes := hasher.Sum(nil)
				var digestF field.Element
				digestF.SetBytes(digestBytes[:])
				// To be symmetrical with the return of the SIS hasher
				// we return the column hash as a slice of a single element
				// containing the hash.
				selectedColSisDigests[i][j] = []field.Element{digestF}
			}
		}

		// Check the linear combination is consistent with the opened column
		y := poly.EvalUnivariate(fullCol, randomCoin)
		if y != proof.LinearCombination.Get(selectedColID) {
			other := proof.LinearCombination.Get(selectedColID)
			return nil, fmt.Errorf("the linear combination is inconsistent %v : %v", y.String(), other.String())
		}
	}

	return selectedColSisDigests, nil
}

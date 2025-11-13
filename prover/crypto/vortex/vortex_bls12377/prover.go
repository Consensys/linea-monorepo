package vortex

import (
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// OpeningProof represents an opening proof for a Vortex commitment. The proof
// possibly relates to several commitments simultaneously. This corresponds to
// the batch settings.
type OpeningProof struct {
	// Columns against which the linear combination is checked (the i-th
	// entry is the EntryList[i]-th column). The columns may as well be
	// dispatched in several matrices. Columns [i][j][k] returns the k-th entry
	// of the j-th selected column of the i-th commitment
	Columns [][][]field.Element

	// Linear combination of the Reed-Solomon encoded polynomials to open.
	LinearCombination smartvectors.SmartVector
}

// Complete completes the proof adding the columns pointed by entryList
// (implictly the random positions pointed to by the verifier).
func (proof *OpeningProof) Complete(
	entryList []int,
	committedMatrices []vortex.EncodedMatrix,
	trees []*smt_bls12377.Tree,
) [][]smt_bls12377.Proof {

	if len(entryList) == 0 {
		utils.Panic("empty entry list")
	}

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

	numTrees := len(trees)
	proofs := make([][]smt_bls12377.Proof, numTrees)

	// Generate the proofs for each tree and each entry
	for treeID, tree := range trees {
		if tree == nil {
			utils.Panic("tree is nil")
		}
		proofs[treeID] = make([]smt_bls12377.Proof, len(entryList))
		for k, entry := range entryList {
			var err error
			proofs[treeID][k], err = tree.Prove(entry)
			if err != nil {
				utils.Panic("invalid entry leaf: %v", err.Error())
			}
		}
	}

	proof.Columns = selectedColumns

	return proofs
}

package vortex_bn254

import (
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bn254"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// SelectColumnsAndMerkleProofs selects columns from the encoded matrix
// pointed to by entryList and generates BN254 Merkle proofs.
func SelectColumnsAndMerkleProofs(
	proof *vortex.OpeningProof,
	entryList []int,
	committedMatrices []EncodedMatrix,
	trees []*smt_bn254.Tree,
) [][]smt_bn254.Proof {

	if len(entryList) == 0 {
		utils.Panic("empty entry list")
	}

	proof.Columns = make([][][]field.Element, len(committedMatrices))

	for i := range committedMatrices {
		proof.Columns[i] = make([][]field.Element, len(entryList))
		for j := range entryList {
			col := make([]field.Element, len(committedMatrices[i]))
			for k := range committedMatrices[i] {
				col[k] = committedMatrices[i][k].Get(entryList[j])
			}
			proof.Columns[i][j] = col
		}
	}

	numTrees := len(trees)
	proofs := make([][]smt_bn254.Proof, numTrees)

	for treeID, tree := range trees {
		if tree == nil {
			utils.Panic("tree is nil")
		}
		proofs[treeID] = make([]smt_bn254.Proof, len(entryList))
		for k, entry := range entryList {
			proofs[treeID][k] = tree.MustProve(entry)
		}
	}

	return proofs
}

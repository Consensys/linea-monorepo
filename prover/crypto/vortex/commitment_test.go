package vortex

import (
	"crypto/sha256"
	"hash"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/require"
)

func copyVi(dst, src *VerifierInputs) {

	dst.Params = src.Params
	dst.MerkleRoots = make([]types.Bytes32, len(src.MerkleRoots))
	copy(dst.MerkleRoots, src.MerkleRoots)
	dst.X.Set(&src.X)
	dst.Ys = make([][]fext.Element, len(src.Ys))
	for i := 0; i < len(dst.Ys); i++ {
		dst.Ys[i] = make([]fext.Element, len(src.Ys[i]))
		copy(dst.Ys[i], src.Ys[i])
	}
	dst.OpeningProof = src.OpeningProof // keep the same proof, since the Merkle proofs are tested in a separate package
	dst.RandomCoin.Set(&src.RandomCoin)
	dst.EntryList = make([]int, len(src.EntryList))
	copy(dst.EntryList, src.EntryList)

}

func TestProver(t *testing.T) {

	// Define test cases in a slice of structs
	tests := []struct {
		name          string
		hasher_Merkle func() hash.Hash
		hasher_Column func() hash.Hash
	}{
		{
			name:          "With SHA256 for both Merkle and Column",
			hasher_Merkle: sha256.New,
			hasher_Column: sha256.New, // use sha256 for column hash
		},
		{
			name:          "With SHA256 for Merkle and SIS for Column",
			hasher_Merkle: sha256.New,
			hasher_Column: nil, // use SIS for column hash
		},
	}

	for _, tc := range tests {
		params := NewParams(2, 1<<4, 1<<10, ringsis.StdParams, tc.hasher_Merkle, tc.hasher_Column)
		x := fext.RandomElement()
		randomCoin := fext.RandomElement()
		entryList := []int{1, 7, 5, 6, 4, 5, 1, 2}
		NbPolysPerCommitment := []int{20} //, 32, 32, 32}
		nbCommitments := len(NbPolysPerCommitment)
		polySize := params.NbColumns

		// create polynomials, and the yis
		polyLists := make([][]smartvectors.SmartVector, nbCommitments)
		yLists := make([][]fext.Element, nbCommitments)
		for i := range polyLists {
			polys := make([]smartvectors.SmartVector, NbPolysPerCommitment[i])
			ys := make([]fext.Element, NbPolysPerCommitment[i])
			for j := range polys {
				polys[j] = smartvectors.Rand(polySize)
				ys[j] = smartvectors.EvaluateLagrangeMixed(polys[j], x)
			}
			polyLists[i] = polys
			yLists[i] = ys
		}

		// commit
		roots := make([]types.Bytes32, nbCommitments)
		trees := make([]*smt.Tree, nbCommitments)
		committedMatrices := make([]EncodedMatrix, nbCommitments)
		for i := range trees {
			committedMatrices[i], trees[i], _ = params.Commit(polyLists[i])
			roots[i] = trees[i].Root
		}

		// open
		proof := params.Open(utils.Join(polyLists...), randomCoin)
		proof.Complete(entryList, committedMatrices, trees)

		// verify
		vi := VerifierInputs{
			Params:       *params,
			MerkleRoots:  roots,
			X:            x,
			Ys:           yLists,
			OpeningProof: *proof,
			RandomCoin:   randomCoin,
			EntryList:    entryList,
		}
		err := VerifyOpening(&vi)
		if err != nil {
			t.Fatal(err)
		}

		// tamper the proof and check that the verification fails
		var viTampered VerifierInputs

		// wrong point for SZ
		copyVi(&viTampered, &vi)
		viTampered.X.SetRandom()
		err = VerifyOpening(&viTampered)
		require.Error(t, err)

		// wrong evaulation of a row
		for i := 0; i < len(vi.Ys); i++ {
			for j := 0; j < len(vi.Ys[i]); j++ {
				copyVi(&viTampered, &vi)
				viTampered.Ys[i][j].SetRandom()
				err = VerifyOpening(&viTampered)
				require.Error(t, err)
			}
		}

		// wrong random coin
		copyVi(&viTampered, &vi)
		viTampered.RandomCoin.SetRandom()
		err = VerifyOpening(&viTampered)
		require.Error(t, err)

	}

}

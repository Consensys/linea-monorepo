package vortex

import (
	"crypto/sha256"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/require"
)

// testParams is a corpus of valid parameters for Vortex
var testParams = []*Params{
	//NewParams(2, 1<<4, 32, ringsis.StdParams, poseidon2.NewPoseidon2),
	NewParams(2, 1<<4, 32, ringsis.StdParams, sha256.New), //TODO@yao change to poseidon2
	//NewParams(2, 1<<4, 32, ringsis.StdParams, mimc.NewMiMC).RemoveSis(mimc.NewMiMC),
	//NewParams(4, 1<<3, 32, ringsis.StdParams, mimc.NewMiMC),
	//NewParams(4, 1<<3, 32, ringsis.StdParams, mimc.NewMiMC).RemoveSis(mimc.NewMiMC),
}

func TestLlocal(t *testing.T) {

	params := NewParams(2, 1<<4, 32, ringsis.StdParams, sha256.New)
	x := fext.RandomElement()
	randomCoin := fext.RandomElement()
	entryList := []int{1, 7, 5, 6, 4, 5, 1, 2}
	NbPolysPerCommitment := []int{2}
	nbCommitments := len(NbPolysPerCommitment)
	polySize := params.NbColumns
	NumOpenedColumns := 4

	// create polynomials
	polyLists := make([][]smartvectors.SmartVector, nbCommitments)
	yLists := make([][]fext.Element, nbCommitments)
	for i := range polyLists {
		polys := make([]smartvectors.SmartVector, NbPolysPerCommitment[i])
		ys := make([]fext.Element, NbPolysPerCommitment[i])
		for j := range polys {
			polys[j] = smartvectors.Rand(polySize)
			ys[j] = smartvectors.EvaluateLagrangeOnFext(polys[j], x)
		}
		polyLists[i] = polys
		yLists[i] = ys
	}
	for i := 0; i < len(yLists); i++ {
		for j := 0; j < len(yLists[i]); j++ {
			fmt.Printf("%s, ", yLists[i][j].String())
			fmt.Println("")
		}
		fmt.Println("-")
	}

	// commit
	roots := make([]types.Bytes32, nbCommitments)
	trees := make([]*smt.Tree, nbCommitments)
	committedMatrices := make([]EncodedMatrix, nbCommitments)
	for i := range trees {
		committedMatrices[i], trees[i], _ = params.CommitMerkleWithSIS(polyLists[i])
		roots[i] = trees[i].Root
	}

	// open
	proof := params.Open(utils.Join(polyLists...), randomCoin)
	proof.Complete(entryList[:NumOpenedColumns], committedMatrices, trees)
}

func TestProver(t *testing.T) {

	x := fext.RandomElement()
	randomCoin := fext.RandomElement()
	entryList := []int{1, 7, 5, 6, 4, 5, 1, 2}

	// the testCases are applied over all those of [testParams]
	testCases := []struct {
		Explainer            string
		NbPolysPerCommitment []int
		NumOpenedColumns     int
		// ChangeAssignmentSize takes params.NbColumns and returns a possibly
		// different value corresponding to the size of the assignment that
		// the testCase provides to the prover. If nil, then this is equivalent
		// to `f(n) -> n`.
		ChangeAssignmentSize func(int) int
		MustPanic            bool
	}{
		{
			Explainer:            "1 matrix commitment with one poly",
			NbPolysPerCommitment: []int{1},
			NumOpenedColumns:     4,
		},
		// {
		// 	Explainer:             "1 matrix commitment with several polys",
		// 	NbPolysPerCommitment: []int{3},
		// 	NumOpenedColumns:      4,
		// },
		// {
		// 	Explainer:             "1 matrix commitment with several polys",
		// 	NbPolysPerCommitment: []int{3, 3},
		// 	NumOpenedColumns:      8,
		// },
		// {
		// 	Explainer:             "1 matrix commitment with several polys",
		// 	NbPolysPerCommitment: []int{1, 15},
		// 	NumOpenedColumns:      8,
		// },
		// {
		// 	Explainer:             "too many rows",
		// 	NbPolysPerCommitment: []int{1, 105},
		// 	NumOpenedColumns:      8,
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "no commitment",
		// 	NbPolysPerCommitment: []int{},
		// 	NumOpenedColumns:      8,
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "1 commitment but zero rows",
		// 	NbPolysPerCommitment: []int{0},
		// 	NumOpenedColumns:      8,
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "Several commitment but none have rows",
		// 	NbPolysPerCommitment: []int{0, 0, 0},
		// 	NumOpenedColumns:      8,
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "Several commitment but none have rows",
		// 	NbPolysPerCommitment: []int{0, 0, 0},
		// 	NumOpenedColumns:      8,
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "Empty entry list",
		// 	NbPolysPerCommitment: []int{5, 6},
		// 	NumOpenedColumns:      0,
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "the polys are twice too large",
		// 	NbPolysPerCommitment: []int{3, 3},
		// 	NumOpenedColumns:      8,
		// 	ChangeAssignmentSize:  func(i int) int { return 2 * i },
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "the polys are twice to small",
		// 	NbPolysPerCommitment: []int{3, 3},
		// 	NumOpenedColumns:      8,
		// 	ChangeAssignmentSize:  func(i int) int { return i / 2 },
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "the polys are twice to small",
		// 	NbPolysPerCommitment: []int{3, 3},
		// 	NumOpenedColumns:      8,
		// 	ChangeAssignmentSize:  func(i int) int { return i + 1 },
		// 	MustPanic:             true,
		// },
		// {
		// 	Explainer:             "the polys are twice to small",
		// 	NbPolysPerCommitment: []int{3, 3},
		// 	NumOpenedColumns:      8,
		// 	ChangeAssignmentSize:  func(i int) int { return i - 1 },
		// 	MustPanic:             true,
		// },
	}

	for i := range testParams {
		params := testParams[i]
		for j := range testCases {

			t.Run(fmt.Sprintf("params-%v-case-%v", i, j), func(t *testing.T) {

				testCase := testCases[j]

				t.Logf("params=%++v test-case=%++v", params, testCase)

				nbCommitments := len(testCase.NbPolysPerCommitment)
				effPolySize := params.NbColumns
				polyLists := make([][]smartvectors.SmartVector, nbCommitments)
				yLists := make([][]fext.Element, nbCommitments)
				roots := make([]types.Bytes32, nbCommitments)
				trees := make([]*smt.Tree, nbCommitments)

				if testCase.ChangeAssignmentSize != nil {
					effPolySize = testCase.ChangeAssignmentSize(effPolySize)
				}

				for i := range polyLists {
					// Polynomials to commit to
					polys := make([]smartvectors.SmartVector, testCase.NbPolysPerCommitment[i])
					ys := make([]fext.Element, testCase.NbPolysPerCommitment[i])
					for j := range polys {
						polys[j] = smartvectors.Rand(effPolySize)

						// effPolySize is messed with and is not a power of 2,
						// the interpolation algorithm will panic as this counts
						// as invalid inputs.
						if utils.IsPowerOfTwo(effPolySize) {
							ys[j] = smartvectors.EvaluateLagrangeOnFext(polys[j], x)
						} else {
							ys[j].SetRandom()
						}
					}
					polyLists[i] = polys
					yLists[i] = ys
				}

				// Importantly, the defer must be declared right before calling
				// the tested code or it will occult potential mistakes in the
				// test.
				if testCase.MustPanic {
					defer func() {
						if r := recover(); r != nil {
							t.Logf("Panicked with message = %v and stacktraces %v", r, string(debug.Stack()))
							return
						}

						t.Fatalf("The test did not panic")
					}()
				}

				// Commits to it
				committedMatrices := make([]EncodedMatrix, nbCommitments)
				for i := range trees {
					committedMatrices[i], trees[i], _ = params.CommitMerkleWithSIS(polyLists[i])
					roots[i] = trees[i].Root
				}

				// Generate the proof
				proof := params.Open(utils.Join(polyLists...), randomCoin)
				proof.Complete(entryList[:testCase.NumOpenedColumns], committedMatrices, trees)

				// Check the proof
				err := VerifyOpening(
					&VerifierInputs{
						Params:       *params,
						MerkleRoots:  roots,
						X:            x,
						Ys:           yLists,
						OpeningProof: *proof,
						RandomCoin:   randomCoin,
						EntryList:    entryList[:testCase.NumOpenedColumns],
					})

				require.NoError(t, err)
			})
		}
	}
}

func TestVerifierNegative(t *testing.T) {

	var (
		numPolysPerCommitmentCorpus = [][]int{
			{1},
			{1, 3},
			{3, 1, 15},
		}
		params = []*Params{
			NewParams(2, 8, 17, ringsis.StdParams, sha256.New), //TODO@yao change to poseidon2
			//NewParams(2, 8, 17, ringsis.StdParams, mimc.NewMiMC).RemoveSis(mimc.NewMiMC),
		}

		statementMutatorCorpus = []struct {
			Explainer string
			Func      func(*VerifierInputs) bool
		}{
			{
				Explainer: "Increment the first y",
				Func: func(v *VerifierInputs) bool {
					one := fext.One()
					v.Ys[0][0].Add(&v.Ys[0][0], &one)
					return true
				},
			},
			{
				Explainer: "Swap the two first y in the first slice",
				Func: func(v *VerifierInputs) bool {
					if len(v.Ys[0]) < 2 {
						return false
					}
					v.Ys[0][1], v.Ys[0][0] = v.Ys[0][0], v.Ys[0][1]
					return true
				},
			},
			{
				Explainer: "Swap the two slices in Y",
				Func: func(v *VerifierInputs) bool {
					if len(v.Ys) < 2 {
						return false
					}
					v.Ys[0], v.Ys[1] = v.Ys[1], v.Ys[0]
					return true
				},
			},
			{
				Explainer: "Move the last entry of Ys[0] to the beginning of Ys[1]",
				Func: func(v *VerifierInputs) bool {
					if len(v.Ys) < 2 {
						return false
					}
					y := v.Ys[0][len(v.Ys[0])-1]
					v.Ys[0] = v.Ys[0][:len(v.Ys[0])-1]
					v.Ys[1] = append([]fext.Element{y}, v.Ys[1]...)
					return true
				},
			},
			{
				Explainer: "Bump the X value",
				Func: func(v *VerifierInputs) bool {
					one := fext.One()
					v.X.Add(&v.X, &one)
					return true
				},
			},
			{
				Explainer: "Pop the first Y",
				Func: func(v *VerifierInputs) bool {
					v.Ys[0] = v.Ys[0][1:]
					return true
				},
			},
		}
		/*
			proofMutatorCorpus = []struct {
				Explainer string
				Func      func(v *VerifierInputs) bool
			}{
				{
					Explainer: "Swap two first entryLists",
					Func: func(v *VerifierInputs) bool {
						v.EntryList[0], v.EntryList[1] = v.EntryList[1], v.EntryList[0]
						return true
					},
				},
				{
					Explainer: "Cut the first entry",
					Func: func(v *VerifierInputs) bool {
						v.EntryList = v.EntryList[1:]
						return true
					},
				},
				{
					Explainer: "Cut the last entry",
					Func: func(v *VerifierInputs) bool {
						v.EntryList = v.EntryList[:len(v.EntryList)-1]
						return true
					},
				},
				{
					Explainer: "Add an extra entry",
					Func: func(v *VerifierInputs) bool {
						v.EntryList = append(v.EntryList, 0)
						return true
					},
				},
				{
					Explainer: "Swap two roots",
					Func: func(v *VerifierInputs) bool {
						if len(v.MerkleRoots) < 2 {
							return false
						}
						v.MerkleRoots[0], v.MerkleRoots[1] = v.MerkleRoots[1], v.MerkleRoots[0]
						return true
					},
				},
				{
					Explainer: "Remove the first root",
					Func: func(v *VerifierInputs) bool {
						if len(v.MerkleRoots) < 1 {
							return false
						}
						v.MerkleRoots = v.MerkleRoots[1:]
						return true
					},
				},
				{
					Explainer: "Add an extra root",
					Func: func(v *VerifierInputs) bool {
						if len(v.MerkleRoots) < 1 {
							return false
						}
						v.MerkleRoots = append(v.MerkleRoots, v.MerkleRoots[0])
						return true
					},
				},
				{
					Explainer: "Swap two positions in the linear combination",
					Func: func(v *VerifierInputs) bool {
						lc := v.OpeningProof.LinearCombination.IntoRegVecSaveAllocExt()
						lc[0], lc[1] = lc[1], lc[0]
						lc_ := smartvectors.RegularExt(lc)
						v.OpeningProof.LinearCombination = &lc_
						return true
					},
				},
				{
					Explainer: "Overwrite a position in the linear combination",
					Func: func(v *VerifierInputs) bool {
						lc := v.OpeningProof.LinearCombination.IntoRegVecSaveAllocExt()
						lc[0] = lc[1]
						lc_ := smartvectors.RegularExt(lc)
						v.OpeningProof.LinearCombination = &lc_
						return true
					},
				},
				{
					Explainer: "Swap two Merkle proofs",
					Func: func(v *VerifierInputs) bool {
						mps := v.OpeningProof.MerkleProofs
						mps[0][0], mps[0][1] = mps[0][1], mps[0][0]
						v.OpeningProof.MerkleProofs = mps
						return true
					},
				},
				{
					Explainer: "Set the first entry to a very large number",
					Func: func(v *VerifierInputs) bool {
						v.EntryList[0] = 10000
						return true
					},
				},
				{
					Explainer: "Mess with a Merkle proof path",
					Func: func(v *VerifierInputs) bool {
						mps := v.OpeningProof.MerkleProofs
						mps[0][0].Path = 5
						return true
					},
				},
			}
		*/
		generateVerifierInputs = func(
			params *Params,
			numPolyPerCommitment []int,
		) *VerifierInputs {
			var (
				x             = fext.RandomElement()
				randomCoin    = fext.RandomElement()
				entryList     = []int{1, 2, 3, 4, 5, 6, 7, 8}
				nbCommitments = len(numPolyPerCommitment)
				effPolySize   = params.NbColumns
				polyLists     = make([][]smartvectors.SmartVector, nbCommitments)
				yLists        = make([][]fext.Element, nbCommitments)
				roots         = make([]types.Bytes32, nbCommitments)
				trees         = make([]*smt.Tree, nbCommitments)
			)
			for i := range polyLists {
				// Polynomials to commit to
				polys := make([]smartvectors.SmartVector, numPolyPerCommitment[i])
				ys := make([]fext.Element, numPolyPerCommitment[i])
				for j := range polys {
					polys[j] = smartvectors.Rand(effPolySize)

					// effPolySize is messed with and is not a power of 2,
					// the interpolation algorithm will panic as this counts
					// as invalid inputs.
					if utils.IsPowerOfTwo(effPolySize) {
						ys[j] = smartvectors.EvaluateLagrangeOnFext(polys[j], x)
					} else {
						ys[j].SetRandom()
					}
				}
				polyLists[i] = polys
				yLists[i] = ys
			}
			// Commits to it
			committedMatrices := make([]EncodedMatrix, nbCommitments)
			for i := range trees {
				committedMatrices[i], trees[i], _ = params.CommitMerkleWithSIS(polyLists[i])
				roots[i] = trees[i].Root
			}
			// Generate the proof
			proof := params.Open(utils.Join(polyLists...), randomCoin)
			proof.Complete(entryList, committedMatrices, trees)
			return &VerifierInputs{
				Params:       *params,
				MerkleRoots:  roots,
				X:            x,
				Ys:           yLists,
				OpeningProof: *proof,
				RandomCoin:   randomCoin,
				EntryList:    entryList,
			}
		}
	)

	for iParams := range params {
		for iNumPoly := range numPolysPerCommitmentCorpus {
			for iMut := range statementMutatorCorpus {

				t.Run(
					fmt.Sprintf("statement-mutation-%v-%v_%v", iParams, iNumPoly, iMut),
					func(t *testing.T) {

						// It's important to regenerate the entry every time as
						// they will be mutated every time.
						v := generateVerifierInputs(
							params[iParams],
							numPolysPerCommitmentCorpus[iNumPoly],
						)
						ok := statementMutatorCorpus[iMut].Func(v)
						if !ok {
							return
						}

						// Check the proof
						err := VerifyOpening(v)

						require.Error(t, err)

					},
				)
			}
		}
	}
	/*
		for iParams := range params {
			for iNumPoly := range numPolysPerCommitmentCorpus {
				for iMut := range proofMutatorCorpus {
					t.Run(
						fmt.Sprintf("proof-mutation-%v-%v_%v", iParams, iNumPoly, iMut),
						func(t *testing.T) {

							// It's important to regenerate the entry every time as
							// they will be mutated every time.
							v := generateVerifierInputs(
								params[iParams],
								numPolysPerCommitmentCorpus[iNumPoly],
							)

							ok := proofMutatorCorpus[iMut].Func(v)
							if !ok {
								return
							}

							// Check the proof
							err := VerifyOpening(v)
							require.Error(t, err)
						},
					)
				}
			}
		}
	*/
}

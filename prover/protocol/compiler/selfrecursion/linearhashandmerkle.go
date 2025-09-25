package selfrecursion

import (
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/merkle"
	poseidon2W "github.com/consensys/linea-monorepo/prover/protocol/dedicated/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// LinearHashAndMerkle verifies the following things:
// 1. The Poseidon2 hash of the SIS digest for the sis rounds are correctly computed
// 2. The Poseidon2 hash of the selected columns for the non SIS rounds are correctly computed
// 3. The Merkle proofs are correctly verified for both SIS and non SIS rounds
// 4. The leaves of the SIS and non SIS rounds are consistent with the Merkle tree leaves
// used for the Merkle proof verification.
func (ctx *SelfRecursionCtx) LinearHashAndMerkle() {
	roundQ := ctx.Columns.Q.Round()

	// Calculate round parameters
	//
	// numRound denotes the total number of commitment rounds
	// including SIS and non SIS rounds
	numRound := ctx.VortexCtx.NumCommittedRounds()
	if ctx.VortexCtx.IsNonEmptyPrecomputed() {
		numRound += 1
	}
	// Next we consider the number of rounds for which we apply the SIS hash
	numRoundSis := ctx.VortexCtx.NumCommittedRoundsSis()
	// We increase numRoundSis by 1 if sis is applied to the precomputed
	if ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		numRoundSis += 1
	}
	// The number of non SIS rounds is the difference
	// between the total number of rounds and the number of SIS rounds
	// It is after considering the precomputed round
	numRoundNonSis := numRound - numRoundSis

	// CalculateSISParameters computes SIS-related parameters
	sisNumHash := ctx.VortexCtx.NbColsToOpen() * numRoundSis
	sisColSize := ctx.VortexCtx.SisParams.OutputSize()
	sisColChunks := (sisColSize + blockSize - 1) / blockSize
	sisTotalChunks := sisColChunks * sisNumHash
	sisNumRowToHash := utils.NextPowerOfTwo(sisTotalChunks)
	sisNumRowExpectedHash := utils.NextPowerOfTwo(sisNumHash)
	concatDhQSizeUnpadded := sisColSize * sisNumHash
	concatDhQSize := utils.NextPowerOfTwo(concatDhQSizeUnpadded)

	// CalculateMerkleParameters computes Merkle tree parameters
	// The leaves are computed for both SIS and non SIS rounds
	numHash := ctx.VortexCtx.NbColsToOpen() * numRound
	numRowExpectedHash := utils.NextPowerOfTwo(numHash)

	// The leaves size for non SIS rounds
	nonSisNumHash := ctx.VortexCtx.NbColsToOpen() * numRoundNonSis

	// Register Merkle-related columns
	ctx.Columns.MerkleProofPositions = ctx.Comp.InsertCommit(roundQ, ctx.merklePositionsName(), numRowExpectedHash)
	for i := 0; i < blockSize; i++ {
		ctx.Columns.MerkleProofsLeaves[i] = ctx.Comp.InsertCommit(roundQ, ctx.merkleLeavesName(i), numRowExpectedHash)
		ctx.Columns.MerkleRoots[i] = ctx.Comp.InsertCommit(roundQ, ctx.merkleRootsName(i), numRowExpectedHash)
	}

	// Register SIS-related columns if needed
	// We commit to the below columns only if SIS is applied to any of the rounds including precomputed
	if ctx.VortexCtx.NumCommittedRoundsSis() > 0 || ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		ctx.Columns.ConcatenatedDhQ = ctx.Comp.InsertCommit(roundQ, ctx.concatenatedDhQ(), concatDhQSize)
		for j := 0; j < blockSize; j++ {
			ctx.Columns.SisToHash[j] = ctx.Comp.InsertCommit(roundQ, ctx.sisToHash(j), sisNumRowToHash)
			ctx.Columns.SisRoundLeaves[j] = make([]ifaces.Column, 0, numRoundSis)
			for i := 0; i < numRoundSis; i++ {
				// Register the SIS round leaves
				ctx.Columns.SisRoundLeaves[j] = append(ctx.Columns.SisRoundLeaves[j], ctx.Comp.InsertCommit(
					roundQ, ctx.sisRoundLeavesName(i, j), ctx.VortexCtx.NbColsToOpen()))
			}
		}
	}

	// Register the linear hash columns for the non sis rounds
	var (
		nonSisNumRowExpectedHash int
		nonSisNumRowsToHash      []int
	)
	if numRoundNonSis > 0 {
		// Register the linear hash columns for non sis rounds
		// If SIS is not applied to the precomputed, we consider
		// it to be the first non sis round
		for j := 0; j < blockSize; j++ {
			ctx.Poseidon2MetaData.NonSisLeaves[j] = make([]ifaces.Column, 0, numRoundNonSis)
		}
		ctx.Poseidon2MetaData.NonSiSToHash = make([][blockSize]ifaces.Column, 0, numRoundNonSis)
		ctx.Poseidon2MetaData.ColChunks = make([]int, 0, numRoundNonSis)
		nonSisNumRowExpectedHash, nonSisNumRowsToHash = ctx.registerPoseidon2MetaDataForNonSisRounds(numRoundNonSis, roundQ)
	}

	ctx.Comp.RegisterProverAction(roundQ, &LinearHashMerkleProverAction{
		Ctx:                   ctx,
		ConcatDhQSize:         concatDhQSize,
		SisNumRowExpectedHash: sisNumRowExpectedHash,
		SisTotalChunks:        sisTotalChunks,
		SisNumRowToHash:       sisNumRowToHash,
		SisNumHash:            sisNumHash,
		NumRowExpectedHash:    numRowExpectedHash,
		NumHash:               numHash,
		NumOpenedCol:          ctx.VortexCtx.NbColsToOpen(),
		NonSisNumHash:         nonSisNumHash,
		NumNonSisRound:        numRoundNonSis,
		NumSisRound:           numRoundSis,
		NonSISExpectedHash:    nonSisNumRowExpectedHash,
		NonSISToHash:          nonSisNumRowsToHash,
	})

	depth := utils.Log2Ceil(ctx.VortexCtx.NumEncodedCols())

	// The Merkle proof verification is for both sis and non sis rounds
	merkle.MerkleProofCheck(ctx.Comp, ctx.merkleProofVerificationName(), depth, numHash, ctx.Columns.MerkleProofPositions,
		ctx.Columns.MerkleProofs, ctx.Columns.MerkleRoots, ctx.Columns.MerkleProofsLeaves)

	// The below linear hash verification is for only sis rounds
	if ctx.VortexCtx.NumCommittedRoundsSis() > 0 || ctx.VortexCtx.IsSISAppliedToPrecomputed() {

		var cleanSisLeaves [blockSize][]ifaces.Column
		var stackedSisLeaves [blockSize]dedicated.StackedColumn
		var expectedHash [blockSize]ifaces.Column

		for j := 0; j < blockSize; j++ {
			cleanSisLeaves[j] = make([]ifaces.Column, 0, numRoundSis)
			for i := 0; i < numRoundSis; i++ {
				cleanSisLeaves[j] = append(cleanSisLeaves[j], ctx.Columns.SisRoundLeaves[j][i])
			}
			// We stack the sis round leaves
			stackedSisLeaves[j] = dedicated.StackColumn(ctx.Comp, cleanSisLeaves[j])
			// Register the prover action for the stacked column
			ctx.Comp.RegisterProverAction(roundQ, &dedicated.StackedColumn{
				Column: stackedSisLeaves[j].Column,
				Source: cleanSisLeaves[j],
			})
			expectedHash[j] = stackedSisLeaves[j].Column

		}

		poseidon2W.CheckLinearHash(ctx.Comp, ctx.linearHashVerificationName(), sisColChunks, ctx.Columns.SisToHash,
			sisNumHash, expectedHash)
	}

	// Register the linear hash verification for the non sis rounds
	for i := 0; i < numRoundNonSis; i++ {
		var expectedHash [blockSize]ifaces.Column
		for j := 0; j < blockSize; j++ {
			expectedHash[j] = ctx.Poseidon2MetaData.NonSisLeaves[j][i]
		}
		//map ctx.Poseidon2MetaData.ToHashSizes[i] to chunks like above
		poseidon2W.CheckLinearHash(ctx.Comp, ctx.nonSisRoundLinearHashVerificationName(i), ctx.Poseidon2MetaData.ColChunks[i], ctx.Poseidon2MetaData.NonSiSToHash[i],
			ctx.VortexCtx.NbColsToOpen(), expectedHash)
	}

	// leafConsistency imposes fixed permutation constraints between the sis
	// and non sis rounds leaves with that of the merkle tree leaves.
	ctx.leafConsistency(roundQ)
}

// registerPoseidon2MetaDataForNonSisRounds registers the metadata for the
// for linear hash verification for the non SIS rounds
// and return the poseidon2HashColumnSize
// and the preimage column sizes per non sis round
func (ctx *SelfRecursionCtx) registerPoseidon2MetaDataForNonSisRounds(
	numRoundNonSis int, round int) (int, []int) {
	// Compute the concatenated hashes and preimages sizes
	var (
		numhash            = ctx.VortexCtx.NbColsToOpen()
		numRowExpectedHash = utils.NextPowerOfTwo(numhash)
		numRowsToHash      = make([]int, 0, numRoundNonSis)
	)

	// Consider the precomputed polynomials
	if ctx.VortexCtx.IsNonEmptyPrecomputed() && !ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		colSize := len(ctx.VortexCtx.Items.Precomputeds.PrecomputedColums)
		colChunks := (colSize + blockSize - 1) / blockSize
		totalChunks := colChunks * numhash
		numRowToHash := utils.NextPowerOfTwo(totalChunks)

		// Leaves are the expected hashes
		for j := 0; j < blockSize; j++ {
			ctx.Poseidon2MetaData.NonSisLeaves[j] = append(ctx.Poseidon2MetaData.NonSisLeaves[j],
				ctx.Comp.InsertCommit(
					round,
					ctx.expectedNonSisPrecomputedHashes(j),
					numRowExpectedHash,
				))
		}

		var hashPreimages [blockSize]ifaces.Column
		for j := 0; j < blockSize; j++ {
			hashPreimages[j] = ctx.Comp.InsertCommit(
				round,
				ctx.nonSisPrecomputedToHash(j),
				numRowToHash,
			)
		}
		ctx.Poseidon2MetaData.NonSiSToHash = append(ctx.Poseidon2MetaData.NonSiSToHash,
			hashPreimages)
		numRowsToHash = append(numRowsToHash, numRowToHash)
		ctx.Poseidon2MetaData.ColChunks = append(ctx.Poseidon2MetaData.ColChunks, colChunks)
	}

	// Next, consider only the non SIS rounds

	// firstNonEmptyRound and isPastFirstNonEmptyRound are used to determine the
	// first non-empty round for Vortex in the wizard. This is needed for
	// uniformity of the columns across different distributed segments. In
	// particular, we want the name of the columns to be the same regardless of
	// the starting round number.
	firstNonEmptyRound := 0
	isPastFirstNonEmptyRound := false

	for i := range ctx.VortexCtx.RoundStatus {

		if !isPastFirstNonEmptyRound && ctx.VortexCtx.RoundStatus[i] == vortex.IsEmpty {
			firstNonEmptyRound = i + 1
			continue
		}

		isPastFirstNonEmptyRound = true

		if ctx.VortexCtx.RoundStatus[i] == vortex.IsOnlyPoseidon2Applied {
			colSize := ctx.VortexCtx.GetNumPolsForNonSisRounds(i)
			colChunks := (colSize + blockSize - 1) / blockSize
			totalChunks := colChunks * numhash
			numRowToHash := utils.NextPowerOfTwo(totalChunks)

			for j := 0; j < blockSize; j++ {
				ctx.Poseidon2MetaData.NonSisLeaves[j] = append(
					ctx.Poseidon2MetaData.NonSisLeaves[j],
					ctx.Comp.InsertCommit(
						round,
						ctx.nonSisLeaves(i-firstNonEmptyRound, j),
						numRowExpectedHash,
					))

			}

			var hashPreimages [blockSize]ifaces.Column
			for j := 0; j < blockSize; j++ {
				hashPreimages[j] = ctx.Comp.InsertCommit(
					round,
					ctx.nonSisToHash(i-firstNonEmptyRound, j),
					numRowToHash,
				)
			}
			ctx.Poseidon2MetaData.NonSiSToHash = append(ctx.Poseidon2MetaData.NonSiSToHash, hashPreimages)
			ctx.Poseidon2MetaData.ColChunks = append(ctx.Poseidon2MetaData.ColChunks, colChunks)
			numRowsToHash = append(numRowsToHash, numRowToHash)
		} else {
			continue
		}
	}

	return numRowExpectedHash, numRowsToHash
}

func (ctx *SelfRecursionCtx) leafConsistency(round int) {
	// Fixed permutation constraint between the SIS and non SIS leaves
	// and the Merkle leaves
	// cleanLeaves = (nonSisLeaves || sisLeaves) is checked to be identical to
	// the Merkle leaves.
	var cleanLeaves [blockSize][]ifaces.Column
	var stackedCleanLeaves [blockSize]dedicated.StackedColumn
	var s [blockSize][]field.Element
	for i := 0; i < blockSize; i++ {
		if len(ctx.Poseidon2MetaData.NonSisLeaves[i]) > 0 {
			cleanLeaves[i] = append(cleanLeaves[i], ctx.Poseidon2MetaData.NonSisLeaves[i]...)
		}
		if ctx.VortexCtx.NumCommittedRoundsSis() > 0 || ctx.VortexCtx.IsSISAppliedToPrecomputed() {
			cleanLeaves[i] = append(cleanLeaves[i], ctx.Columns.SisRoundLeaves[i]...)
		}
		stackedCleanLeaves[i] = dedicated.StackColumn(ctx.Comp, cleanLeaves[i])

		// Register prover action for the stacked column
		ctx.Comp.RegisterProverAction(round, &dedicated.StackedColumn{
			Column: stackedCleanLeaves[i].Column,
			Source: cleanLeaves[i],
		})

		// Next we compute the identity permutation
		s[i] = make([]field.Element, stackedCleanLeaves[i].Column.Size())
		for j := range s[i] {
			s[i][j].SetInt64(int64(j))
		}
		s_smart := smartvectors.NewRegular(s[i])

		// Insert the fixed permutation constraint.
		// Here we assume that the number of opened columns for Vortex
		// is a power of two. If that is not the case, the below
		// constraint is supposed to fail
		ctx.Comp.InsertFixedPermutation(
			round,
			ctx.leafConsistencyName(i),
			[]smartvectors.SmartVector{s_smart},
			[]ifaces.Column{stackedCleanLeaves[i].Column},
			[]ifaces.Column{ctx.Columns.MerkleProofsLeaves[i]},
		)
	}

}

// Implements the prover action interface
type LinearHashMerkleProverAction struct {
	Ctx                   *SelfRecursionCtx
	ConcatDhQSize         int
	SisNumRowExpectedHash int
	SisTotalChunks        int
	NumRowExpectedHash    int
	NumHash               int
	SisNumRowToHash       int
	SisNumHash            int
	NumOpenedCol          int
	NonSisNumHash         int
	NumNonSisRound        int
	NumSisRound           int
	NonSISExpectedHash    int
	NonSISToHash          []int
}

// linearHashMerkleProverActionBuilder builds the assignment parameters
// of the prover action for the linear hash and merkle
type linearHashMerkleProverActionBuilder struct {
	// Contains the concatenated sis hashes of the selected columns
	// for the sis round matrices
	ConcatDhQ []field.Element

	SisToHash [blockSize][]field.Element
	// The leaves of the merkle tree (both sis and non sis)
	MerkleLeaves [blockSize][]field.Element // extend to blockSize here
	// The positions of the leaves in the merkle tree
	MerklePositions []field.Element
	// The roots of the merkle tree
	MerkleRoots [blockSize][]field.Element // extend to blockSize here
	// The merkle proofs are aligned as (non sis, sis).
	// Hence we need to align leaves, position and roots
	// in the same way. Meaning we need to store them separately
	// and append them later
	MerkleSisLeaves    [blockSize][]field.Element // extend to blockSize here
	MerkleSisPositions []field.Element
	MerkleSisRoots     [blockSize][]field.Element // extend to blockSize here
	// Now the non sis round values
	MerkleNonSisLeaves    [blockSize][]field.Element // extend to blockSize here
	MerkleNonSisPositions []field.Element
	MerkleNonSisRoots     [blockSize][]field.Element // extend to blockSize here
	// The leaves of the sis round matrices
	SisLeaves [][blockSize][]field.Element // extend to blockSize here
	// The leaves of the non sis round matrices.
	// NonSisLeaves[i][j] is leaf of the jth selected column
	// for the ith non sis round matrix
	NonSisLeaves [][blockSize][]field.Element // extend to blockSize here
	// The Poseidon2 hash pre images of the
	// non sis round matrices, NonSisHashPreimages[i][j]
	// is the jth selected column for the ith non sis round matrix
	NonSisHashPreimages [][]field.Element // no need to extend to blockSize here
	// the size of the sis hash digest
	SisHashSize int
	// the number of opened/selected columns
	// per round
	NumOpenedCol int
	// total number of rounds
	TotalNumRounds int
	SisTotalChunks int
	// the committed round offset
	CommittedRound int
}

// newLinearHashMerkleProverActionBuilder returns an empty
// linearHashMerkleProverActionBuilder
func newLinearHashMerkleProverActionBuilder(a *LinearHashMerkleProverAction) *linearHashMerkleProverActionBuilder {
	lmp := linearHashMerkleProverActionBuilder{}
	lmp.ConcatDhQ = make([]field.Element, a.SisNumHash*a.Ctx.VortexCtx.SisParams.OutputSize())
	lmp.MerklePositions = make([]field.Element, 0, a.NumHash)
	lmp.MerkleSisPositions = make([]field.Element, 0, a.SisNumHash)
	lmp.MerkleNonSisPositions = make([]field.Element, 0, a.NonSisNumHash)
	lmp.NonSisHashPreimages = make([][]field.Element, 0, a.NumNonSisRound)
	lmp.SisHashSize = a.Ctx.VortexCtx.SisParams.OutputSize()
	lmp.NumOpenedCol = a.Ctx.VortexCtx.NbColsToOpen()
	lmp.TotalNumRounds = a.Ctx.VortexCtx.MaxCommittedRound
	lmp.CommittedRound = 0
	lmp.SisTotalChunks = a.SisTotalChunks

	for i := 0; i < blockSize; i++ {
		lmp.SisToHash[i] = make([]field.Element, a.SisTotalChunks)
		lmp.MerkleLeaves[i] = make([]field.Element, 0, a.NumHash)
		lmp.MerkleRoots[i] = make([]field.Element, 0, a.NumHash)
		lmp.MerkleSisLeaves[i] = make([]field.Element, 0, a.SisNumHash)
		lmp.MerkleSisRoots[i] = make([]field.Element, 0, a.SisNumHash)
		lmp.MerkleNonSisLeaves[i] = make([]field.Element, 0, a.NonSisNumHash)
		lmp.MerkleNonSisRoots[i] = make([]field.Element, 0, a.NonSisNumHash)
	}
	lmp.SisLeaves = make([][blockSize][]field.Element, 0, a.NumSisRound)
	lmp.NonSisLeaves = make([][blockSize][]field.Element, 0, a.NumNonSisRound)
	return &lmp
}

// Run implements the prover action for the linear hash and merkle
func (a *LinearHashMerkleProverAction) Run(run *wizard.ProverRuntime) {
	openingIndices := run.GetRandomCoinIntegerVec(a.Ctx.Coins.Q.Name)
	lmp := newLinearHashMerkleProverActionBuilder(a)

	// Handle the precomputed round
	if a.Ctx.VortexCtx.IsNonEmptyPrecomputed() {
		processPrecomputedRound(a, lmp, run, openingIndices)
	}

	// Handle the SIS and non SIS rounds
	processRound(a, lmp, run, openingIndices)

	numCommittedRound := a.Ctx.VortexCtx.NumCommittedRounds()
	if a.Ctx.VortexCtx.IsNonEmptyPrecomputed() {
		numCommittedRound += 1
	}

	if lmp.CommittedRound != numCommittedRound {
		utils.Panic("Committed rounds %v does not match the total number of committed rounds %v", lmp.CommittedRound, numCommittedRound)
	}

	// Append non sis and sis round leaves, roots, and positions
	for i := 0; i < blockSize; i++ {
		lmp.MerkleLeaves[i] = append(lmp.MerkleNonSisLeaves[i], lmp.MerkleSisLeaves[i]...)
		lmp.MerkleRoots[i] = append(lmp.MerkleNonSisRoots[i], lmp.MerkleSisRoots[i]...)
	}
	lmp.MerklePositions = append(lmp.MerkleNonSisPositions, lmp.MerkleSisPositions...)

	// Assign columns using IDs from ctx.Columns
	run.AssignColumn(a.Ctx.Columns.MerkleProofPositions.GetColID(), smartvectors.RightZeroPadded(lmp.MerklePositions, a.NumRowExpectedHash))

	for i := 0; i < blockSize; i++ {
		run.AssignColumn(a.Ctx.Columns.MerkleProofsLeaves[i].GetColID(), smartvectors.RightZeroPadded(lmp.MerkleLeaves[i], a.NumRowExpectedHash))
		run.AssignColumn(a.Ctx.Columns.MerkleRoots[i].GetColID(), smartvectors.RightZeroPadded(lmp.MerkleRoots[i], a.NumRowExpectedHash))
		// The below assignments are only done if SIS is applied to any of the rounds
		if a.Ctx.VortexCtx.NumCommittedRoundsSis() > 0 || a.Ctx.VortexCtx.IsSISAppliedToPrecomputed() {
			// Assign the concatenated SIS hashes
			run.AssignColumn(a.Ctx.Columns.SisToHash[i].GetColID(), smartvectors.RightZeroPadded(lmp.SisToHash[i], a.SisNumRowToHash))
			for j := 0; j < a.NumSisRound; j++ {
				// Assign the SIS round leaves
				run.AssignColumn(a.Ctx.Columns.SisRoundLeaves[i][j].GetColID(), smartvectors.NewRegular(lmp.SisLeaves[j][i]))
			}
		}
	}
	if a.Ctx.VortexCtx.NumCommittedRoundsSis() > 0 || a.Ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		run.AssignColumn(a.Ctx.Columns.ConcatenatedDhQ.GetColID(), smartvectors.RightZeroPadded(lmp.ConcatDhQ, a.ConcatDhQSize))

	}

	// Assign the hash values and preimages for the non SIS rounds
	for i := 0; i < a.NumNonSisRound; i++ {
		var th [blockSize][]field.Element
		colSize := len(lmp.NonSisHashPreimages[i])
		colChunks := (colSize + blockSize - 1) / blockSize
		for j := 0; j < blockSize; j++ {
			th[j] = make([]field.Element, colChunks)
		}

		th = poseidon2W.PrepareToHashWitness(th, lmp.NonSisHashPreimages[i], 0)

		for j := 0; j < blockSize; j++ {
			run.AssignColumn(a.Ctx.Poseidon2MetaData.NonSisLeaves[j][i].GetColID(), smartvectors.RightZeroPadded(lmp.NonSisLeaves[i][j], a.NonSISExpectedHash))

			run.AssignColumn(a.Ctx.Poseidon2MetaData.NonSiSToHash[i][j].GetColID(), smartvectors.RightZeroPadded(th[j], a.NonSISToHash[i]))

		}
	}
}

// processPrecomputedRound processes the precomputed polynomials
// assignments for the linear hash and merkle tree prover action
func processPrecomputedRound(
	a *LinearHashMerkleProverAction,
	lmp *linearHashMerkleProverActionBuilder,
	run *wizard.ProverRuntime,
	openingIndices []int,
) {
	// The merkle root for the precomputed round
	var rootPrecomp [blockSize]field.Element
	for i := 0; i < blockSize; i++ {
		precompRootSv := run.GetColumn(a.Ctx.Columns.PrecompRoot[i].GetColID())
		rootPrecomp[i] = precompRootSv.IntoRegVecSaveAlloc()[0]
	}
	if a.Ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		precompColSisHash := a.Ctx.VortexCtx.Items.Precomputeds.DhWithMerkle
		var precompSisLeaves [blockSize][]field.Element
		for j := 0; j < blockSize; j++ {
			precompSisLeaves[j] = make([]field.Element, 0, len(openingIndices))
		}
		for i, selectedCol := range openingIndices {
			srcStart := selectedCol * lmp.SisHashSize
			sisHash := precompColSisHash[srcStart : srcStart+lmp.SisHashSize]
			destStart := i * lmp.SisHashSize
			// Allocate sisHash to SisToHash columns

			chunks := (lmp.SisHashSize + blockSize - 1) / blockSize
			tohashDestStart := i * chunks
			lmp.SisToHash = poseidon2W.PrepareToHashWitness(lmp.SisToHash, sisHash, tohashDestStart)
			copy(lmp.ConcatDhQ[destStart:destStart+lmp.SisHashSize], sisHash)

			leaf := poseidon2.Poseidon2Sponge(sisHash)
			for j := 0; j < blockSize; j++ {
				lmp.MerkleSisLeaves[j] = append(lmp.MerkleSisLeaves[j], leaf[j])
				precompSisLeaves[j] = append(precompSisLeaves[j], leaf[j])
				lmp.MerkleSisRoots[j] = append(lmp.MerkleSisRoots[j], rootPrecomp[j])
			}
			lmp.MerkleSisPositions = append(lmp.MerkleSisPositions, field.NewElement(uint64(selectedCol)))

		}
		lmp.SisLeaves = append(lmp.SisLeaves, precompSisLeaves)
		lmp.CommittedRound++
		lmp.TotalNumRounds++
	} else {
		numhash := lmp.NumOpenedCol
		precompColPoseidon2Hash := a.Ctx.VortexCtx.Items.Precomputeds.DhWithMerkle
		var precompPoseidon2ExpectedHash [blockSize][]field.Element
		for j := 0; j < blockSize; j++ {
			precompPoseidon2ExpectedHash[j] = make([]field.Element, 0, numhash)
		}
		precompPoseidon2ToHash := make([]field.Element, 0, lmp.NumOpenedCol*len(a.Ctx.VortexCtx.Items.Precomputeds.PrecomputedColums))
		for _, selectedCol := range openingIndices {
			srcStart := selectedCol * blockSize
			// Poseidon2 hash is [8]field.Element
			var leaf [blockSize]field.Element
			poseidon2Hash := precompColPoseidon2Hash[srcStart : srcStart+blockSize]
			copy(leaf[:], poseidon2Hash)

			poseidon2Preimage := a.Ctx.VortexCtx.GetPrecomputedSelectedCol(selectedCol)
			// Also compute the leaf from the column
			// to check sanity
			leaf_ := poseidon2.Poseidon2Sponge(poseidon2Preimage)
			// Sanity check
			// The leaf computed from the precomputed column
			if leaf != leaf_ {
				utils.Panic("Poseidon2 hash of the precomputed column %v does not match the leaf %v", leaf_, leaf)
			}
			for j := 0; j < blockSize; j++ {

				lmp.MerkleNonSisLeaves[j] = append(lmp.MerkleNonSisLeaves[j], leaf[j])
				lmp.MerkleNonSisRoots[j] = append(lmp.MerkleNonSisRoots[j], rootPrecomp[j])
				precompPoseidon2ExpectedHash[j] = append(precompPoseidon2ExpectedHash[j], leaf[j])

			}
			lmp.MerkleNonSisPositions = append(lmp.MerkleNonSisPositions, field.NewElement(uint64(selectedCol)))
			precompPoseidon2ToHash = append(precompPoseidon2ToHash, poseidon2Preimage...)
		}
		// Append the hash values and preimages
		lmp.NonSisLeaves = append(lmp.NonSisLeaves, precompPoseidon2ExpectedHash)
		lmp.NonSisHashPreimages = append(lmp.NonSisHashPreimages, precompPoseidon2ToHash)

		lmp.CommittedRound++
		lmp.TotalNumRounds++
	}
}

// processRound processes the round assignements
// for the linear hash and merkle tree prover action
func processRound(
	a *LinearHashMerkleProverAction,
	lmp *linearHashMerkleProverActionBuilder,
	run *wizard.ProverRuntime,
	openingIndices []int,
) {
	// If there are non SIS rounds, we need to fetch the
	// non SIS opened columns
	var (
		nonSisOpenedCols     [][][]field.Element
		nonSisOpenedColsName string
		nonSisRoundCount     = 0
		sisRoundCount        = 0
	)
	if a.NumNonSisRound > 0 {
		nonSisOpenedColsName = a.Ctx.VortexCtx.SelectedColumnNonSISName()
		nonSisOpenedColsSV, found := run.State.TryGet(nonSisOpenedColsName)
		if !found {
			utils.Panic("nonSisOpenedColsName %v not found", nonSisOpenedColsName)
		}
		nonSisOpenedCols = nonSisOpenedColsSV.([][][]field.Element)
		// Note nonSisOpenedCols contains the precomputed columns also if
		// SIS is not applied to the precomputed.
		// However, we already have it at the time of processing
		// the precomputed round, so we need to exclude it
		if a.Ctx.VortexCtx.IsNonEmptyPrecomputed() && !a.Ctx.VortexCtx.IsSISAppliedToPrecomputed() {
			nonSisOpenedCols = nonSisOpenedCols[1:]
		}
	}
	// If SIS is applied to the precomputed, we need to
	// increase the sisRoundCount by 1
	if a.Ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		sisRoundCount++
	}

	// The SIS and non SIS rounds are processed
	numRound := lmp.TotalNumRounds
	// We need to decrease the number of rounds by 1
	// as precomputed round is considered seperately
	if a.Ctx.VortexCtx.IsNonEmptyPrecomputed() {
		numRound -= 1
	}
	// for j := 0; j < blockSize; j++ {
	// 	lmp.SisToHash[j] = make([]field.Element, 0, lmp.SisTotalChunks*(lmp.TotalNumRounds+1))
	// }
	for round := 0; round <= numRound; round++ {
		if a.Ctx.VortexCtx.RoundStatus[round] == vortex.IsSISApplied {
			colSisHashName := a.Ctx.VortexCtx.SisHashName(round)
			colSisHashSV, found := run.State.TryGet(colSisHashName)
			if !found {
				utils.Panic("colSisHashName %v not found", colSisHashName)
			}

			var rooth field.Octuplet
			for j := 0; j < blockSize; j++ {
				precompRootSv := run.GetColumn(a.Ctx.Columns.Rooth[round][j].GetColID())
				rooth[j] = precompRootSv.IntoRegVecSaveAlloc()[0]
			}

			colSisHash := colSisHashSV.([]field.Element)

			var sisRoundLeaves [blockSize][]field.Element
			for j := 0; j < blockSize; j++ {
				sisRoundLeaves[j] = make([]field.Element, 0, lmp.NumOpenedCol)
			}

			for i, selectedCol := range openingIndices {
				srcStart := selectedCol * lmp.SisHashSize
				sisHash := colSisHash[srcStart : srcStart+lmp.SisHashSize]
				destStart := sisRoundCount*lmp.NumOpenedCol*lmp.SisHashSize + i*lmp.SisHashSize
				copy(lmp.ConcatDhQ[destStart:destStart+lmp.SisHashSize], sisHash)

				chunks := (lmp.SisHashSize + blockSize - 1) / blockSize
				tohashDestStart := sisRoundCount*lmp.NumOpenedCol*chunks + i*chunks
				lmp.SisToHash = poseidon2W.PrepareToHashWitness(lmp.SisToHash, sisHash, tohashDestStart)

				leaf := poseidon2.Poseidon2Sponge(sisHash)

				for j := 0; j < blockSize; j++ {
					lmp.MerkleSisLeaves[j] = append(lmp.MerkleSisLeaves[j], leaf[j])
					sisRoundLeaves[j] = append(sisRoundLeaves[j], leaf[j])
					lmp.MerkleSisRoots[j] = append(lmp.MerkleSisRoots[j], rooth[j])
				}
				lmp.MerkleSisPositions = append(lmp.MerkleSisPositions, field.NewElement(uint64(selectedCol)))
			}
			// Append the sis leaves
			lmp.SisLeaves = append(lmp.SisLeaves, sisRoundLeaves)
			sisRoundCount++
			run.State.TryDel(colSisHashName)
			lmp.CommittedRound++

		} else if a.Ctx.VortexCtx.RoundStatus[round] == vortex.IsOnlyPoseidon2Applied {
			// Fetch the Poseidon2 hash values
			colNoSisHashName := a.Ctx.VortexCtx.NoSisHashName(round)
			colPoseidon2HashSV, found := run.State.TryGet(colNoSisHashName)
			if !found {
				utils.Panic("colNoSisHashName %v not found", colNoSisHashName)
			}
			colPoseidon2Hash := colPoseidon2HashSV.([]field.Element)

			// Fetch the root for the round
			var rooth field.Octuplet
			for j := 0; j < blockSize; j++ {
				precompRootSv := run.GetColumn(a.Ctx.Columns.Rooth[round][j].GetColID())
				rooth[j] = precompRootSv.IntoRegVecSaveAlloc()[0]
			}

			var poseidon2HashValues [blockSize][]field.Element
			for j := 0; j < blockSize; j++ {
				poseidon2HashValues[j] = make([]field.Element, 0, lmp.NumOpenedCol)
			}
			poseidon2HashPreimages := make([]field.Element, 0, lmp.NumOpenedCol*a.Ctx.VortexCtx.GetNumPolsForNonSisRounds(round))
			for i, selectedCol := range openingIndices {
				srcStart := selectedCol * blockSize
				// Poseidon2 hash is [8]field.Element
				var leaf [blockSize]field.Element
				poseidon2Hash := colPoseidon2Hash[srcStart : srcStart+blockSize]
				copy(leaf[:], poseidon2Hash)
				poseidon2Preimage := nonSisOpenedCols[nonSisRoundCount][i]
				// Also compute the leaf from the column
				// to check sanity
				leaf_ := poseidon2.Poseidon2Sponge(poseidon2Preimage)

				if leaf != leaf_ {
					utils.Panic("Poseidon2 hash of the non SIS column %v does not match the leaf %v", leaf_, leaf)
				}
				for j := 0; j < blockSize; j++ {
					lmp.MerkleNonSisLeaves[j] = append(lmp.MerkleNonSisLeaves[j], leaf[j])
					lmp.MerkleNonSisRoots[j] = append(lmp.MerkleNonSisRoots[j], rooth[j])
					poseidon2HashValues[j] = append(poseidon2HashValues[j], leaf[j])
				}
				lmp.MerkleNonSisPositions = append(lmp.MerkleNonSisPositions, field.NewElement(uint64(selectedCol)))
				poseidon2HashPreimages = append(poseidon2HashPreimages, poseidon2Preimage...)
			}
			// Append the hash values and preimages
			lmp.NonSisLeaves = append(lmp.NonSisLeaves, poseidon2HashValues)
			lmp.NonSisHashPreimages = append(lmp.NonSisHashPreimages, poseidon2HashPreimages)
			run.State.TryDel(colNoSisHashName)
			lmp.CommittedRound++
			nonSisRoundCount++
		} else if a.Ctx.VortexCtx.RoundStatus[round] == vortex.IsEmpty {
			continue
		}
	}
	// We can delete the non SIS opened columns now
	run.State.TryDel(nonSisOpenedColsName)
}

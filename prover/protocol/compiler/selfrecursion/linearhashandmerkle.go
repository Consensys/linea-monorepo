package selfrecursion

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/merkle"
	mimcW "github.com/consensys/linea-monorepo/prover/protocol/dedicated/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// linearHashAndMerkle verifies the following things:
// 1. The MiMC hash of the SIS digest for the sis rounds are correctly computed
// 2. The MiMC hash of the selected columns for the non SIS rounds are correctly computed
// 3. The Merkle proof is correctly verified for both SIS and non SIS rounds
// 4. The leaves of the SIS and non SIS rounds are consistent with the Merkle tree leaves
func (ctx *SelfRecursionCtx) linearHashAndMerkle() {
	roundQ := ctx.Columns.Q.Round()
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
	numRoundNonSis := numRound - numRoundSis
	if numRoundSis == 0 {
		utils.Panic("SIS is not applied to any round, we don't support this case")
	}

	// The total SIS hash length = size of a single SIS hash *
	// total number of SIS hash per SIS round * number of SIS rounds
	concatDhQSizeUnpadded := ctx.VortexCtx.SisParams.OutputSize() * ctx.VortexCtx.NbColsToOpen() * numRoundSis
	concatDhQSize := utils.NextPowerOfTwo(concatDhQSizeUnpadded)
	// The leaves are computed for both SIS and non SIS rounds
	leavesSizeUnpadded := ctx.VortexCtx.NbColsToOpen() * numRound
	leavesSize := utils.NextPowerOfTwo(leavesSizeUnpadded)
	// The leaves for SIS rounds
	sisRoundLeavesSizeUnpadded := ctx.VortexCtx.NbColsToOpen() * numRoundSis
	sisRoundLeavesSize := utils.NextPowerOfTwo(sisRoundLeavesSizeUnpadded)

	ctx.Columns.ConcatenatedDhQ = ctx.comp.InsertCommit(roundQ, ctx.concatenatedDhQ(), concatDhQSize)
	ctx.Columns.MerkleProofsLeaves = ctx.comp.InsertCommit(roundQ, ctx.merkleLeavesName(), leavesSize)
	ctx.Columns.MerkleProofPositions = ctx.comp.InsertCommit(roundQ, ctx.merklePositionssName(), leavesSize)
	ctx.Columns.MerkleRoots = ctx.comp.InsertCommit(roundQ, ctx.merkleRootsName(), leavesSize)
	ctx.Columns.SisRoundLeaves = ctx.comp.InsertCommit(roundQ, ctx.sisRoundLeavesName(), sisRoundLeavesSize)

	// Register the linear hash columns for non sis rounds
	// If SIS is not applied to the precomputed, we consider
	// it to be the first non sis round
	ctx.MIMCMetaData.NonSisLeaves = make([]ifaces.Column, 0, numRoundNonSis)
	ctx.MIMCMetaData.ConcatenatedHashPreimages = make([]ifaces.Column, 0, numRoundNonSis)
	ctx.MIMCMetaData.ToHashSizes = make([]int, 0, numRoundNonSis)

	// Register the linear hash columns for the non sis rounds
	var (
		mimcHashColumnSize      int
		mimcPreimageColumnsSize []int
	)
	if numRoundNonSis > 0 {
		mimcHashColumnSize, mimcPreimageColumnsSize = ctx.registerMiMCMetaDataForNonSisRounds(numRoundNonSis, roundQ)
	}

	ctx.comp.RegisterProverAction(roundQ, &linearHashMerkleProverAction{
		ctx:                        ctx,
		concatDhQSize:              concatDhQSize,
		leavesSize:                 leavesSize,
		leavesSizeUnpadded:         leavesSizeUnpadded,
		sisRoundLeavesSize:         sisRoundLeavesSize,
		sisRoundLeavesSizeUnpadded: sisRoundLeavesSizeUnpadded,
		numNonSisRound:             numRoundNonSis,
		hashValuesSize:             mimcHashColumnSize,
		hashPreimagesSize:          mimcPreimageColumnsSize,
	})

	depth := utils.Log2Ceil(ctx.VortexCtx.NumEncodedCols())

	// The Merkle proof verification is for both sis and non sis rounds
	merkle.MerkleProofCheck(ctx.comp, ctx.merkleProofVerificationName(), depth, leavesSizeUnpadded,
		ctx.Columns.MerkleProofs, ctx.Columns.MerkleRoots, ctx.Columns.MerkleProofsLeaves, ctx.Columns.MerkleProofPositions)

	// The linear hash verification is for only sis rounds
	mimcW.CheckLinearHash(ctx.comp, ctx.linearHashVerificationName(), ctx.Columns.ConcatenatedDhQ,
		ctx.VortexCtx.SisParams.OutputSize(), sisRoundLeavesSizeUnpadded, ctx.Columns.SisRoundLeaves)

	// Register the linear hash verification for the non sis rounds
	for i := 0; i < numRoundNonSis; i++ {
		mimcW.CheckLinearHash(ctx.comp, ctx.nonSisRoundLinearHashVerificationName(i), ctx.MIMCMetaData.ConcatenatedHashPreimages[i],
			ctx.MIMCMetaData.ToHashSizes[i], ctx.VortexCtx.NbColsToOpen(), ctx.MIMCMetaData.NonSisLeaves[i])
	}

	// leafConsistency imposes lookup constraints between the sis
	// and non sis rounds leaves with that of the merkle tree leaves.
	ctx.leafConsistency(roundQ)
}

// registerMiMCMetaDataForNonSisRounds registers the metadata for the
// for linear hash verification for the non SIS rounds
// and return the mimcHashColumnSize
// and the preimage column sizes per non sis round
func (ctx *SelfRecursionCtx) registerMiMCMetaDataForNonSisRounds(
	numRoundNonSis int, roundQ int) (int, []int) {
	// Compute the concatenated hashes and preimages sizes
	var (
		mimcHashColumnSize      = utils.NextPowerOfTwo(ctx.VortexCtx.NbColsToOpen())
		mimcPreimageColumnsSize = make([]int, 0, numRoundNonSis)
	)

	// Consider the precomputed polynomials
	if ctx.VortexCtx.IsNonEmptyPrecomputed() && ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		precompPreimageSize := utils.NextPowerOfTwo(
			ctx.VortexCtx.NbColsToOpen() *
				len(ctx.VortexCtx.Items.Precomputeds.PrecomputedColums))

		ctx.MIMCMetaData.NonSisLeaves = append(ctx.MIMCMetaData.NonSisLeaves,
			ctx.comp.InsertCommit(
				roundQ,
				ctx.concatenatedPrecomputedHashes(),
				mimcHashColumnSize,
			))

		ctx.MIMCMetaData.ConcatenatedHashPreimages = append(ctx.MIMCMetaData.ConcatenatedHashPreimages,
			ctx.comp.InsertCommit(
				roundQ,
				ctx.concatenatedPrecomputedPreimages(),
				precompPreimageSize,
			))
		mimcPreimageColumnsSize = append(mimcPreimageColumnsSize, precompPreimageSize)
		ctx.MIMCMetaData.ToHashSizes = append(ctx.MIMCMetaData.ToHashSizes, len(ctx.VortexCtx.Items.Precomputeds.PrecomputedColums))
	}
	// Next, consider only the non SIS rounds
	for i := 0; i < ctx.VortexCtx.NumCommittedRounds(); i++ {
		if ctx.VortexCtx.RoundStatus[i] == vortex.IsOnlyMiMCApplied {

			roundPreimageSize := utils.NextPowerOfTwo(
				ctx.VortexCtx.NbColsToOpen() *
					ctx.VortexCtx.GetNumPolsForNonSisRounds(i))

			ctx.MIMCMetaData.NonSisLeaves = append(
				ctx.MIMCMetaData.NonSisLeaves,
				ctx.comp.InsertCommit(
					roundQ,
					ctx.concatenatedMiMCHashes(i),
					mimcHashColumnSize,
				))

			ctx.MIMCMetaData.ConcatenatedHashPreimages = append(ctx.MIMCMetaData.ConcatenatedHashPreimages, ctx.comp.InsertCommit(
				roundQ,
				ctx.concatenatedMIMCPreimages(i),
				roundPreimageSize,
			))

			ctx.MIMCMetaData.ToHashSizes = append(ctx.MIMCMetaData.ToHashSizes, ctx.VortexCtx.GetNumPolsForNonSisRounds(i))
			mimcPreimageColumnsSize = append(mimcPreimageColumnsSize, roundPreimageSize)
		} else {
			continue
		}
	}
	return mimcHashColumnSize, mimcPreimageColumnsSize
}

func (ctx *SelfRecursionCtx) leafConsistency(round int) {
	// Lookup constrains between the SIS leaves
	// and the Merkle leaves
	ctx.comp.InsertInclusion(
		round,
		ctx.sisRoundAndMerkleLeavesInclusion(),
		[]ifaces.Column{
			ctx.Columns.MerkleProofsLeaves,
		},
		[]ifaces.Column{
			ctx.Columns.SisRoundLeaves,
		},
	)
	// Lookup constrains between the non SIS leaves
	// and the Merkle leaves
	for i := 0; i < len(ctx.MIMCMetaData.NonSisLeaves); i++ {
		ctx.comp.InsertInclusion(
			round,
			ctx.nonSisRoundAndMerkleLeavesInclusion(i),
			[]ifaces.Column{
				ctx.Columns.MerkleProofsLeaves,
			},
			[]ifaces.Column{
				ctx.MIMCMetaData.NonSisLeaves[i],
			},
		)
	}
}

// Implements the prover action interface
type linearHashMerkleProverAction struct {
	ctx                        *SelfRecursionCtx
	concatDhQSize              int
	leavesSize                 int
	leavesSizeUnpadded         int
	sisRoundLeavesSize         int
	sisRoundLeavesSizeUnpadded int
	numNonSisRound             int
	hashValuesSize             int
	hashPreimagesSize          []int
}

// linearHashMerkleProverActionBuilder builds the assignment parameters
// of the prover action for the linear hash and merkle
type linearHashMerkleProverActionBuilder struct {
	// Contains the concatenated sis hashes of the selected columns
	// for the sis round matrices
	concatDhQ []field.Element
	// The leaves of the merkle tree (both sis and non sis)
	merkleLeaves []field.Element
	// The positions of the leaves in the merkle tree
	merklePositions []field.Element
	// The roots of the merkle tree
	merkleRoots []field.Element
	// The leaves of the sis round matrices
	sisLeaves []field.Element
	// The leaves of the non sis round matrices.
	// nonSisLeaves[i][j] is leaf of the jth selected column
	// for the ith non sis round matrix
	nonSisLeaves [][]field.Element
	// The MiMC hash pre images of the
	// non sis round matrices, nonSisHashPreimages[i][j]
	// is the jth selected column for the ith non sis round matrix
	nonSisHashPreimages [][]field.Element
	// the size of the sis hash digest
	sisHashSize int
	// the number of opened/selected columns
	// per round
	numOpenedCol int
	// total number of rounds
	totalNumRounds int
	// the committed round offset
	committedRound int
}

// newLinearHashMerkleProverActionBuilder returns an empty
// linearHashMerkleProverActionBuilder
func newLinearHashMerkleProverActionBuilder(a *linearHashMerkleProverAction) *linearHashMerkleProverActionBuilder {
	lmp := linearHashMerkleProverActionBuilder{}
	lmp.concatDhQ = make([]field.Element, a.sisRoundLeavesSizeUnpadded*a.ctx.VortexCtx.SisParams.OutputSize())
	lmp.merkleLeaves = make([]field.Element, a.leavesSizeUnpadded)
	lmp.merklePositions = make([]field.Element, a.leavesSizeUnpadded)
	lmp.merkleRoots = make([]field.Element, a.leavesSizeUnpadded)
	lmp.sisLeaves = make([]field.Element, 0, a.sisRoundLeavesSizeUnpadded)
	lmp.nonSisLeaves = make([][]field.Element, 0, a.numNonSisRound)
	lmp.nonSisHashPreimages = make([][]field.Element, 0, a.numNonSisRound)
	lmp.sisHashSize = a.ctx.VortexCtx.SisParams.OutputSize()
	lmp.numOpenedCol = a.ctx.VortexCtx.NbColsToOpen()
	// For some reason, using a.ctx.comp.NumRounds() here does not work well here.
	lmp.totalNumRounds = a.ctx.VortexCtx.MaxCommittedRound
	lmp.committedRound = 0
	return &lmp
}

// Run implements the prover action for the linear hash and merkle
func (a *linearHashMerkleProverAction) Run(run *wizard.ProverRuntime) {
	openingIndices := run.GetRandomCoinIntegerVec(a.ctx.Coins.Q.Name)
	lmp := newLinearHashMerkleProverActionBuilder(a)

	// Handle the precomputed round
	if a.ctx.VortexCtx.IsNonEmptyPrecomputed() {
		processPrecomputedRound(a, lmp, run, openingIndices)
	}

	// Handle the SIS and non SIS rounds
	processRound(a, lmp, run, openingIndices)

	numCommittedRound := a.ctx.VortexCtx.NumCommittedRounds()
	if a.ctx.VortexCtx.IsNonEmptyPrecomputed() {
		numCommittedRound += 1
	}

	if lmp.committedRound != numCommittedRound {
		utils.Panic("Committed rounds %v does not match the total number of committed rounds %v", lmp.committedRound, numCommittedRound)
	}

	// Assign columns using IDs from ctx.Columns
	run.AssignColumn(a.ctx.Columns.ConcatenatedDhQ.GetColID(), smartvectors.RightZeroPadded(lmp.concatDhQ, a.concatDhQSize))
	run.AssignColumn(a.ctx.Columns.MerkleProofsLeaves.GetColID(), smartvectors.RightZeroPadded(lmp.merkleLeaves, a.leavesSize))
	run.AssignColumn(a.ctx.Columns.MerkleProofPositions.GetColID(), smartvectors.RightZeroPadded(lmp.merklePositions, a.leavesSize))
	run.AssignColumn(a.ctx.Columns.MerkleRoots.GetColID(), smartvectors.RightZeroPadded(lmp.merkleRoots, a.leavesSize))
	run.AssignColumn(a.ctx.Columns.SisRoundLeaves.GetColID(), smartvectors.RightZeroPadded(lmp.sisLeaves, a.sisRoundLeavesSize))

	// Assign the hash values and preimages for the non SIS rounds
	for i := 0; i < a.numNonSisRound; i++ {
		run.AssignColumn(a.ctx.MIMCMetaData.NonSisLeaves[i].GetColID(), smartvectors.RightZeroPadded(lmp.nonSisLeaves[i], a.hashValuesSize))
		run.AssignColumn(a.ctx.MIMCMetaData.ConcatenatedHashPreimages[i].GetColID(), smartvectors.RightZeroPadded(lmp.nonSisHashPreimages[i], a.hashPreimagesSize[i]))
	}
}

// processPrecomputedRound processes the precomputed polynomials
// assignments for the linear hash and merkle tree prover action
func processPrecomputedRound(
	a *linearHashMerkleProverAction,
	lmp *linearHashMerkleProverActionBuilder,
	run *wizard.ProverRuntime,
	openingIndices []int,
) {
	// The merkle root for the precomputed round
	rootPrecomp := a.ctx.Columns.precompRoot.GetColAssignment(run).Get(0)
	if a.ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		precompColSisHash := a.ctx.VortexCtx.Items.Precomputeds.DhWithMerkle
		for i, selectedCol := range openingIndices {
			srcStart := selectedCol * lmp.sisHashSize
			destStart := i * lmp.sisHashSize
			sisHash := precompColSisHash[srcStart : srcStart+lmp.sisHashSize]
			copy(lmp.concatDhQ[destStart:destStart+lmp.sisHashSize], sisHash)
			leaf := mimc.HashVec(sisHash)
			insertAt := i
			lmp.merkleLeaves[insertAt] = leaf
			lmp.sisLeaves = append(lmp.sisLeaves, leaf)
			lmp.merkleRoots[insertAt] = rootPrecomp
			lmp.merklePositions[insertAt].SetInt64(int64(selectedCol))
		}
		lmp.committedRound++
		lmp.totalNumRounds++
	} else if a.ctx.VortexCtx.IsNonEmptyPrecomputed() {
		precompColMiMCHash := a.ctx.VortexCtx.Items.Precomputeds.DhWithMerkle
		precompMimcHashValues := make([]field.Element, 0, lmp.numOpenedCol)
		precompMimcHashPreimages := make([]field.Element, 0, lmp.numOpenedCol*len(a.ctx.VortexCtx.Items.Precomputeds.PrecomputedColums))
		for i, selectedCol := range openingIndices {
			srcStart := selectedCol
			// MiMC hash is a single value
			mimcHash := precompColMiMCHash[srcStart : srcStart+1]
			leaf := mimcHash[0]
			mimcPreimage := a.ctx.VortexCtx.GetPrecomputedSelectedCol(selectedCol)
			insertAt := i
			lmp.merkleLeaves[insertAt] = leaf
			lmp.merkleRoots[insertAt] = rootPrecomp
			lmp.merklePositions[insertAt].SetInt64(int64(selectedCol))
			precompMimcHashValues = append(precompMimcHashValues, leaf)
			precompMimcHashPreimages = append(precompMimcHashPreimages, mimcPreimage...)
		}
		// Append the hash values and preimages
		lmp.nonSisLeaves = append(lmp.nonSisLeaves, precompMimcHashValues)
		lmp.nonSisHashPreimages = append(lmp.nonSisHashPreimages, precompMimcHashPreimages)
		lmp.committedRound++
		lmp.totalNumRounds++
	}
}

// processRound processes the round assignements
// for the linear hash and merkle tree prover action
func processRound(
	a *linearHashMerkleProverAction,
	lmp *linearHashMerkleProverActionBuilder,
	run *wizard.ProverRuntime,
	openingIndices []int,
) {
	for round := 0; round <= lmp.totalNumRounds; round++ {
		if a.ctx.VortexCtx.RoundStatus[round] == vortex.IsSISApplied {
			colSisHashName := a.ctx.VortexCtx.SisHashName(round)
			colSisHashSV, found := run.State.TryGet(colSisHashName)
			if !found {
				utils.Panic("colSisHashName %v not found", colSisHashName)
			}

			rooth := a.ctx.Columns.Rooth[round].GetColAssignment(run).Get(0)
			colSisHash := colSisHashSV.([]field.Element)

			for i, selectedCol := range openingIndices {
				srcStart := selectedCol * lmp.sisHashSize
				destStart := lmp.committedRound*lmp.numOpenedCol*lmp.sisHashSize + i*lmp.sisHashSize
				sisHash := colSisHash[srcStart : srcStart+lmp.sisHashSize]
				copy(lmp.concatDhQ[destStart:destStart+lmp.sisHashSize], sisHash)
				leaf := mimc.HashVec(sisHash)
				insertAt := lmp.committedRound*lmp.numOpenedCol + i
				lmp.merkleLeaves[insertAt] = leaf
				lmp.sisLeaves = append(lmp.sisLeaves, leaf)
				lmp.merkleRoots[insertAt] = rooth
				lmp.merklePositions[insertAt].SetInt64(int64(selectedCol))
			}

			run.State.TryDel(colSisHashName)
			lmp.committedRound++
		} else if a.ctx.VortexCtx.RoundStatus[round] == vortex.IsOnlyMiMCApplied {
			// Fetch the MiMC hash values
			colMimcHashName := a.ctx.VortexCtx.MIMCHashName(round)
			colMimcHashSV, found := run.State.TryGet(colMimcHashName)
			if !found {
				utils.Panic("colMimcHashName %v not found", colMimcHashName)
			}
			colMimcHash := colMimcHashSV.([]field.Element)
			// Fetch the MiMC preimages
			nonSisOpenedColsName := a.ctx.VortexCtx.SelectedColumnNonSISName()
			nonSisOpenedColsSV, found := run.State.TryGet(nonSisOpenedColsName)
			if !found {
				utils.Panic("nonSisOpenedColsName %v not found", nonSisOpenedColsName)
			}
			nonSisOpenedCols := nonSisOpenedColsSV.([][][]field.Element)
			// Note nonSisOpenedCols contains the precomputed columns also if
			// SIS is applied to the precomputed.
			// However, we already have it, so we need to exclude it
			if a.ctx.VortexCtx.IsSISAppliedToPrecomputed() {
				nonSisOpenedCols = nonSisOpenedCols[1:]
			}
			// Fetch the root for the round
			rooth := a.ctx.Columns.Rooth[round].GetColAssignment(run).Get(0)
			mimcHashValues := make([]field.Element, 0, lmp.numOpenedCol)
			mimcHashPreimages := make([]field.Element, 0, lmp.numOpenedCol*a.ctx.VortexCtx.GetNumPolsForNonSisRounds(round))
			for i, selectedCol := range openingIndices {
				srcStart := selectedCol
				// MiMC hash is a single value
				mimcHash := colMimcHash[srcStart : srcStart+1]
				mimcPreimage := nonSisOpenedCols[round][i]
				leaf := mimcHash[0]
				insertAt := lmp.committedRound*lmp.numOpenedCol + i
				lmp.merkleLeaves[insertAt] = leaf
				lmp.merkleRoots[insertAt] = rooth
				lmp.merklePositions[insertAt].SetInt64(int64(selectedCol))
				mimcHashValues = append(mimcHashValues, leaf)
				mimcHashPreimages = append(mimcHashPreimages, mimcPreimage...)
			}
			// Append the hash values and preimages
			lmp.nonSisLeaves = append(lmp.nonSisLeaves, mimcHashValues)
			lmp.nonSisHashPreimages = append(lmp.nonSisHashPreimages, mimcHashPreimages)
			run.State.TryDel(colMimcHashName)
			run.State.TryDel(nonSisOpenedColsName)
			lmp.committedRound++

		} else if a.ctx.VortexCtx.RoundStatus[round] == vortex.IsEmpty {
			continue
		}
	}
}

package selfrecursion

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Here we develop the functionality that allows us to check that the root hashes
// given as commitments and the one used for Merkle proof verification are consistent
func (ctx *SelfRecursionCtx) RootHashGlue() {

	// Get the list of the root hashes (without the non-appended ones)
	// Insert precomputed roots
	var (
		rootHashSis     []ifaces.Column
		rootHashNonsis  []ifaces.Column
		rootHashesClean []ifaces.Column
	)
	if ctx.VortexCtx.IsNonEmptyPrecomputed() {
		precompRoots := ctx.Columns.PrecompRoot
		if precompRoots == nil {
			utils.Panic("Precomputed root should not be nil! That's because, we are in commit to precomputed mode.")
		}
		if ctx.VortexCtx.IsSISAppliedToPrecomputed() {
			rootHashSis = append(rootHashSis, precompRoots)
		} else {
			rootHashNonsis = append(rootHashNonsis, precompRoots)
		}
	}

	for round, rh := range ctx.Columns.Rooth {
		if ctx.VortexCtx.RoundStatus[round] == vortex.IsSISApplied {
			rootHashSis = append(rootHashSis, rh)
		} else if ctx.VortexCtx.RoundStatus[round] == vortex.IsOnlyMiMCApplied {
			rootHashNonsis = append(rootHashNonsis, rh)
		} else if ctx.VortexCtx.RoundStatus[round] == vortex.IsEmpty {
			continue
		}
	}
	rootHashesClean = append(rootHashNonsis, rootHashSis...)

	numCommittedRound := ctx.VortexCtx.NumCommittedRounds()
	// numCommittedRound increses by 1 if we commit to the precomputeds
	if ctx.VortexCtx.IsNonEmptyPrecomputed() {
		numCommittedRound += 1
	}

	if len(rootHashesClean) != numCommittedRound {
		utils.Panic(
			"unexpected %v != %v",
			len(rootHashesClean),
			numCommittedRound,
		)
	}

	// Number of distinct roots
	numRoots := len(rootHashesClean)
	// Number of active roots in MerkleRoots
	nbOpenCol := ctx.VortexCtx.NbColsToOpen()
	numActiveRoot := nbOpenCol * numCommittedRound
	// Length of MerkleRoots = numActiveRoot + padding
	totalRoots := ctx.Columns.MerkleRoots.Size()

	rootHashVecParts := utils.RightPadWith(
		rootHashesClean,
		utils.NextPowerOfTwo(len(rootHashesClean)),
		verifiercol.NewConstantCol(field.Zero(), 1, ""),
	)

	numRootsPadded := len(rootHashVecParts)

	rootHashVecParts = utils.RepeatSlice(
		rootHashVecParts,
		totalRoots/len(rootHashVecParts),
	)

	rootHashVec := verifiercol.NewConcatTinyColumns(
		ctx.Comp,
		len(rootHashVecParts),
		field.Element{}, // note: that this will be ditched by the function
		rootHashVecParts...,
	)

	if rootHashVec.Size() != ctx.Columns.MerkleRoots.Size() {
		utils.Panic("unexpected lengths %v expected %v", rootHashVec.Size(), ctx.Columns.MerkleRoots.Size())
	}

	// If MerkleRoots is correct, then there is a permutation we can
	// make to audit it.
	s := make([]field.Element, 2*totalRoots)

	// Group all the indexes for each distinct value. The last one is
	// for the padding.
	groups := make([][]int, numRoots+1)
	for i := range groups {
		groups[i] = make([]int, 0, 2*totalRoots)
	}
	groupPadding := len(groups) - 1

	for i := 0; i < totalRoots; i++ {
		// the corresponding element in rootHashVec is an actual root
		if i%numRootsPadded < numRoots {
			groups[i%numRootsPadded] = append(groups[i%numRootsPadded], i)
		} else {
			// else, the corresponding element in rootHashVec is padded
			groups[groupPadding] = append(groups[groupPadding], i)
		}

		// The corresponding element in MerkleRoots is an actual root
		if i < numActiveRoot {
			groups[i/nbOpenCol] = append(groups[i/nbOpenCol], i+totalRoots)
		} else {
			// else, the corresponding element in MerkleRoots is padded
			groups[groupPadding] = append(groups[groupPadding], i+totalRoots)
		}
	}

	// Allocate the cycles
	for _, grp := range groups {
		// We make each element point to the next one
		for i := 0; i < len(grp); i++ {
			s[grp[i]].SetUint64(uint64(grp[(i+1)%len(grp)]))
		}
	}

	// And from that, we get s1 and s2 and declare the corresponding
	// copy constraint.
	ctx.Comp.InsertFixedPermutation(
		ctx.Columns.MerkleRoots.Round(),
		ctx.rootHasGlue(),
		[]smartvectors.SmartVector{
			smartvectors.NewRegular(s[:totalRoots]),
			smartvectors.NewRegular(s[totalRoots:]),
		},
		[]ifaces.Column{
			rootHashVec,
			ctx.Columns.MerkleRoots,
		},
		[]ifaces.Column{
			rootHashVec,
			ctx.Columns.MerkleRoots,
		},
	)
}

// Ensures that openedColumns are the same as the one given in the Merkle proofs
func (ctx SelfRecursionCtx) GluePositions() {

	// The vector that the verifier trusts
	positionVec := verifiercol.NewFromIntVecCoin(
		ctx.Comp,
		ctx.Coins.Q,
		verifiercol.RightPadZeroToNextPowerOfTwo,
	)

	// The vector that the verifier wants to audit w.r.t. position vec
	merklePos := ctx.Columns.MerkleProofPositions

	sizeSmallPos := ctx.Coins.Q.Size // =nbOpenCols
	sizePositionVec := positionVec.Size()
	numCommittedRound := ctx.VortexCtx.NumCommittedRounds()
	// numCommittedRound increses by 1 if we commit to the precomputeds
	if ctx.VortexCtx.IsNonEmptyPrecomputed() {
		numCommittedRound += 1
	}
	numActive := sizeSmallPos * numCommittedRound
	totalSize := merklePos.Size()

	// repeat positionVec as many time as need to equal the length
	// of merklePos (otherwise, we can't do the fixed permutation
	// check).
	positionVec = verifiercol.NewFromAccessors(
		utils.RepeatSlice(
			positionVec.(verifiercol.FromAccessors).Accessors,
			merklePos.Size()/sizePositionVec,
		),
		field.Zero(),
		merklePos.Size(),
	)

	// If MerkleRoots is correct, then there is a permutation we can
	// make to audit it.
	s := make([]field.Element, 2*totalSize)

	// Group all the indexes for each distinct value. The last one is
	// for the padding.
	groups := make([][]int, sizeSmallPos+1)
	for i := range groups {
		groups[i] = make([]int, 0, 2*totalSize)
	}
	groupPadding := len(groups) - 1

	for i := 0; i < totalSize; i++ {
		// the corresponding element in rootHashVec is an actual root
		if i%sizePositionVec < sizeSmallPos {
			groups[i%sizePositionVec] = append(groups[i%sizePositionVec], i)
		} else {
			// else, the corresponding element in rootHashVec is padded
			groups[groupPadding] = append(groups[groupPadding], i)
		}

		// The corresponding element in MerkleRoots is an actual root
		if i < numActive {
			groups[i%sizeSmallPos] = append(groups[i%sizeSmallPos], i+totalSize)
		} else {
			// else, the corresponding element in MerkleRoots is padded
			groups[groupPadding] = append(groups[groupPadding], i+totalSize)
		}
	}

	// Allocate the cycles
	for _, grp := range groups {
		// We make each element point to the next one
		for i := 0; i < len(grp); i++ {
			s[grp[i]].SetUint64(uint64(grp[(i+1)%len(grp)]))
		}
	}

	// And from that, we get s1 and s2 and declare the corresponding
	// copy constraint.
	ctx.Comp.InsertFixedPermutation(
		ctx.Columns.MerkleProofPositions.Round(),
		ctx.positionGlue(),
		[]smartvectors.SmartVector{
			smartvectors.NewRegular(s[:totalSize]),
			smartvectors.NewRegular(s[totalSize:]),
		},
		[]ifaces.Column{
			positionVec,
			ctx.Columns.MerkleProofPositions,
		},
		[]ifaces.Column{
			positionVec,
			ctx.Columns.MerkleProofPositions,
		},
	)

}

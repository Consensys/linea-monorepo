package vortex

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

func (ctx *Ctx) Verify(vr wizard.Runtime) error {

	// The skip verification flag may be on, if the current vortex
	// context get self-recursed. In this case, the verifier does
	// not need to do anything
	if ctx.IsSelfrecursed {
		return nil
	}
	var (
		// The roots of the merkle trees. We stack the no SIS round
		// roots before the SIS roots. The precomputed root is the
		// first root of the SIS roots if SIS hash is applied on the
		// precomputed. Otherwise, it is the first root of the no SIS roots.
		noSisRoots = []types.Bytes32{}
		sisRoots   = []types.Bytes32{}
	)

	// Append the precomputed roots and the corresponding flag
	if ctx.IsNonEmptyPrecomputed() {
		precompRootSv := vr.GetColumn(ctx.Items.Precomputeds.MerkleRoot.GetColID()) // len 1 smart vector
		var precompRootF [8]field.Element
		for j := 0; j < 8; j++ {
			precompRootF[j] = precompRootSv.Get(j)
		}

		if ctx.IsSISAppliedToPrecomputed() {
			sisRoots = append(sisRoots, types.HashToBytes32(precompRootF))
		} else {
			noSisRoots = append(noSisRoots, types.HashToBytes32(precompRootF))
		}
	}
	// Collect all the roots: rounds by rounds
	// and append them to the sis or no sis roots
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		rootSv := vr.GetColumn(ctx.Items.MerkleRoots[round].GetColID()) // len 1 smart vector
		var rootF [8]field.Element
		for j := 0; j < 8; j++ {
			rootF[j] = rootSv.Get(j)
		}

		// Append the isSISApplied flag
		switch ctx.RoundStatus[round] {
		case IsOnlyPoseidon2Applied:
			noSisRoots = append(noSisRoots, types.HashToBytes32(rootF))
		case IsSISApplied:
			sisRoots = append(sisRoots, types.HashToBytes32(rootF))
		}
	}
	// assign the roots
	roots := append(noSisRoots, sisRoots...)

	proof := &vortex.OpeningProof{}
	randomCoin := vr.GetRandomCoinFieldExt(ctx.LinCombRandCoinName())

	// Collect the linear combination
	proof.LinearCombination = vr.GetColumn(ctx.LinCombName())

	// Collect the random entry List and the random coin
	entryList := vr.GetRandomCoinIntegerVec(ctx.RandColSelectionName())

	// Collect the opened columns and split them "by-commitment-rounds"
	proof.Columns = ctx.RecoverSelectedColumns(vr, entryList)
	x := vr.GetUnivariateParams(ctx.Query.QueryID).X

	packedMProofs := vr.GetColumn(ctx.MerkleProofName())
	proof.MerkleProofs = ctx.unpackMerkleProofs(packedMProofs, entryList)

	return vortex.VerifyOpening(&vortex.VerifierInputs{
		Params:       *ctx.VortexParams,
		MerkleRoots:  roots,
		X:            x,
		Ys:           ctx.getYs(vr),
		OpeningProof: *proof,
		RandomCoin:   randomCoin,
		EntryList:    entryList,
	})
}

// returns the number of committed rows for the given round. This takes
// into account the fact that we use shadow columns.
func (ctx *Ctx) getNbCommittedRows(round int) int {
	return ctx.CommitmentsByRounds.LenOf(round)
}

// returns the Ys as a vector
func (ctx *Ctx) getYs(vr wizard.Runtime) (ys [][]fext.Element) {

	var (
		query   = ctx.Query
		params  = vr.GetUnivariateParams(ctx.Query.QueryID)
		ysNoSIS = [][]fext.Element{}
		ysSIS   = [][]fext.Element{}
	)

	// Build an index table to efficiently lookup an alleged
	// prover evaluation from its colID.
	ysMap := make(map[ifaces.ColID]fext.Element, len(params.Ys))
	for i := range query.Pols {
		ysMap[query.Pols[i].GetColID()] = params.Ys[i]
	}

	// Also add the shadow evaluations into ysMap. Since the shadow columns
	// are full-zeroes. We know that the evaluation will also always be zero
	//
	// The sorting is necessary to ensure that the iteration below happens in
	// deterministic order over the [ShadowCols] map.
	shadowIDs := utils.SortedKeysOf(ctx.ShadowCols, func(a, b ifaces.ColID) bool {
		return a < b
	})

	for _, shadowID := range shadowIDs {
		ysMap[shadowID] = fext.Zero()
	}

	// add ys for precomputed
	if ctx.IsNonEmptyPrecomputed() {
		names := make([]ifaces.ColID, len(ctx.Items.Precomputeds.PrecomputedColums))
		for i, poly := range ctx.Items.Precomputeds.PrecomputedColums {
			names[i] = poly.GetColID()
		}
		ysPrecomputed := make([]fext.Element, len(names))
		for i, name := range names {
			ysPrecomputed[i] = ysMap[name]
		}
		// conditionally append the ysPrecomputed to the SIS or no SIS list
		if ctx.IsSISAppliedToPrecomputed() {
			ysSIS = append(ysSIS, ysPrecomputed)
		} else {
			ysNoSIS = append(ysNoSIS, ysPrecomputed)
		}
	}

	// Get the list of the polynomials rounds by rounds
	// and append them to the sis or no sis lists
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		names := ctx.CommitmentsByRounds.MustGet(round)
		ysRounds := make([]fext.Element, len(names))
		for i, name := range names {
			ysRounds[i] = ysMap[name]
		}
		// conditionally append ysRounds to the SIS or no SIS list
		if ctx.RoundStatus[round] == IsOnlyPoseidon2Applied {
			ysNoSIS = append(ysNoSIS, ysRounds)
		} else if ctx.RoundStatus[round] == IsSISApplied {
			ysSIS = append(ysSIS, ysRounds)
		}
	}
	// Append the ysNoSIS and ysSIS
	// We compute ysFull = (ysNoSIS, ysSIS)
	ysFull := append(ysNoSIS, ysSIS...)

	return ysFull
}

// Returns the opened columns from the messages. The returned columns are
// split "by-commitment-round".
func (ctx *Ctx) RecoverSelectedColumns(vr wizard.Runtime, entryList []int) [][][]field.Element {
	// We assign as below:
	// openedSubColumns = (openedSubColumnsNoSIS, openedSubColumnsSIS)
	var (
		openedSubColumnsNoSIS = [][][]field.Element{}
		openedSubColumnsSIS   = [][][]field.Element{}
	)
	// Collect the columns : first extract the full columns
	// Bear in mind that the prover messages are zero-padded
	fullSelectedCols := make([][]field.Element, len(entryList))
	for j := range entryList {
		fullSelectedCol := vr.GetColumn(ctx.SelectedColName(j))
		fullSelectedCols[j] = smartvectors.IntoRegVec(fullSelectedCol)
	}

	// Split the columns per commitment for the verification
	roundStartAt := 0

	// Process precomputed
	if ctx.IsNonEmptyPrecomputed() {
		openedPrecompCols := make([][]field.Element, len(entryList))
		numPrecomputeds := len(ctx.Items.Precomputeds.PrecomputedColums)
		for j := range entryList {
			openedPrecompCols[j] = fullSelectedCols[j][roundStartAt : roundStartAt+numPrecomputeds]
		}
		// update the start counter to ensure we do not pass twice the same row
		roundStartAt += numPrecomputeds
		// conditionally append the opened precomputed columns
		// to the SIS or no SIS list
		if ctx.IsSISAppliedToPrecomputed() {
			openedSubColumnsSIS = append(openedSubColumnsSIS, openedPrecompCols)
		} else {
			openedSubColumnsNoSIS = append(openedSubColumnsNoSIS, openedPrecompCols)
		}
	}

	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		openedSubColumnsForRound := make([][]field.Element, len(entryList))
		numRowsForRound := ctx.getNbCommittedRows(round)
		for j := range entryList {
			openedSubColumnsForRound[j] = fullSelectedCols[j][roundStartAt : roundStartAt+numRowsForRound]
		}

		// update the start counter to ensure we do not pass twice the same row
		roundStartAt += numRowsForRound
		// conditionally append the opened columns
		// to the SIS or no SIS list
		if ctx.RoundStatus[round] == IsOnlyPoseidon2Applied {
			openedSubColumnsNoSIS = append(openedSubColumnsNoSIS, openedSubColumnsForRound)
		} else if ctx.RoundStatus[round] == IsSISApplied {
			openedSubColumnsSIS = append(openedSubColumnsSIS, openedSubColumnsForRound)
		}
	}
	// Append the no SIS and SIS opened sub columns
	openedSubColumns := append(openedSubColumnsNoSIS, openedSubColumnsSIS...)

	// sanity-check : make sure we have not forgotten any column
	// We need to treat the precomputed separately if they are committed
	if roundStartAt != ctx.CommittedRowsCount {
		utils.Panic("we have a mistmatch in the row count : %v != %v", roundStartAt, ctx.CommittedRowsCount)
	}

	return openedSubColumns
}

// Evaluates explicitly the public polynomials (proof, vk, public inputs)
func (ctx *Ctx) explicitPublicEvaluation(vr wizard.Runtime) error {

	params := vr.GetUnivariateParams(ctx.Query.QueryID)

	for i, pol := range ctx.Query.Pols {

		// If the column is a VerifierDefined column, then it is
		// directly concerned by direct verification but we can
		// access its witness or status so we need a specific check.
		if _, ok := pol.(verifiercol.VerifierCol); !ok {
			status := ctx.comp.Columns.Status(pol.GetColID())
			if !status.IsPublic() {
				// then, its not concerned by direct evaluation
				continue
			}
		}

		val := pol.GetColAssignment(vr)

		y := smartvectors.EvaluateLagrangeFullFext(val, params.X)
		if y != params.Ys[i] {
			return fmt.Errorf("inconsistent evaluation")
		}
	}

	return nil
}

type shadowRowProverAction struct {
	name ifaces.ColID
	size int
}

func (a *shadowRowProverAction) Run(run *wizard.ProverRuntime) {
	run.AssignColumn(a.name, smartvectors.NewConstant(field.Zero(), a.size))
}

// A shadow row is a row filled with zeroes that we **may** add at the end of
// the rounds commitment. Its purpose is to ensure the number of "SIS limbs" in
// a row divides the degree of the ring-SIS instance.
func autoAssignedShadowRow(comp *wizard.CompiledIOP, size, round, id int) ifaces.Column {

	name := ifaces.ColIDf("VORTEX_%v_SHADOW_ROUND_%v_ID_%v", comp.SelfRecursionCount, round, id)
	col := comp.InsertCommit(round, name, size)

	comp.RegisterProverAction(round, &shadowRowProverAction{
		name: name,
		size: size,
	})

	return col
}

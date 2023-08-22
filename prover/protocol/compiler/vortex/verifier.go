package vortex

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/vortex2"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column/verifiercol"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

func (ctx *Ctx) Verify(vr *wizard.VerifierRuntime) error {

	// Evaluate explicitly the public columns
	if err := ctx.explicitPublicEvaluation(vr); err != nil {
		return err
	}

	// The skip verification flag may be on, if the current vortex
	// context get self-recursed. In this case, the verifier does
	// not need to
	if ctx.IsSelfrecursed {
		return nil
	}

	// In Merkle-Mode `commitments` is left as empty
	commitments := []vortex2.Commitment{}

	if !ctx.UseMerkleProof {
		// Collect all the commitments : rounds by rounds
		for round := 0; round <= ctx.MaxCommittedRound; round++ {
			// There are not included in the commitments so there is no
			// commitement to look for.
			if ctx.isDry(round) {
				continue
			}

			commitment := vr.GetColumn(ctx.CommitmentName(round))
			commitments = append(commitments, smartvectors.IntoRegVec(commitment))
		}
	}

	// In non-Merkle mode, this is left as empty
	roots := []hashtypes.Digest{}

	if ctx.UseMerkleProof {
		// Collect all the commitments : rounds by rounds
		for round := 0; round <= ctx.MaxCommittedRound; round++ {
			// There are not included in the commitments so there is no
			// commitement to look for.
			if ctx.isDry(round) {
				continue
			}

			rootSv := vr.GetColumn(ctx.MerkleRootName(round)) // len 1 smart vector
			rootF := rootSv.Get(0)                            // root as a field element
			roots = append(roots, hashtypes.Digest(rootF.Bytes()))
		}
	}

	proof := &vortex2.Proof{}
	randomCoin := vr.GetRandomCoinField(ctx.LinCombRandCoinName())

	// Collect the linear combination
	proof.LinearCombination = vr.GetColumn(ctx.LinCombName())

	// Collect the random entry List and the random coin
	entryList := vr.GetRandomCoinIntegerVec(ctx.RandColSelectionName())

	// Collect the opened columns and split them "by-commitment-rounds"
	proof.Columns = ctx.RecoverSelectedColumns(vr, entryList)
	x := vr.GetUnivariateParams(ctx.Query.QueryID).X

	if !ctx.UseMerkleProof {
		return ctx.VortexParams.VerifyOpening(commitments, proof, x, ctx.getYs(vr), randomCoin, entryList)
	}

	if ctx.UseMerkleProof {
		packedMProofs := vr.GetColumn(ctx.MerkleProofName())
		proof.MerkleProofs = ctx.unpackMerkleProofs(packedMProofs, entryList)
		return ctx.VortexParams.VerifyMerkle(roots, proof, x, ctx.getYs(vr), randomCoin, entryList)
	}

	panic("unreachable")
}

// returns the number of committed rows for the given round. This takes
// into account the fact that we use shadow columns.
func (ctx *Ctx) getNbCommittedRows(round int) int {
	return ctx.CommitmentsByRounds.LenOf(round)
}

// returns the Ys as a vector
func (ctx *Ctx) getYs(vr *wizard.VerifierRuntime) (ys [][]field.Element) {

	query := ctx.Query
	params := vr.GetUnivariateParams(ctx.Query.QueryID)

	// Build an index table to efficiently lookup an alleged
	// prover evaluation from its colID.
	ysMap := make(map[ifaces.ColID]field.Element, len(params.Ys))
	for i := range query.Pols {
		ysMap[query.Pols[i].GetColID()] = params.Ys[i]
	}

	// Also add the shadow evaluations into ysMap. Since the shadow columns
	// are full-zeroes. We know that the evaluation will also always be zero
	for shadowID := range ctx.ShadowCols {
		ysMap[shadowID] = field.Zero()
	}

	ys = [][]field.Element{}

	// Get the list of the polynomials
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		// again, skip the dry rounds
		if ctx.isDry(round) {
			continue
		}

		names := ctx.CommitmentsByRounds.MustGet(round)
		ysRounds := make([]field.Element, len(names))
		for i, name := range names {
			ysRounds[i] = ysMap[name]
		}

		ys = append(ys, ysRounds)
	}

	return ys
}

// Returns the opened columns from the messages. The returned columns are
// split "by-commitment-round".
func (ctx *Ctx) RecoverSelectedColumns(vr *wizard.VerifierRuntime, entryList []int) [][][]field.Element {

	// Collect the columns : first extract the full columns
	// Bear in mind that the prover messages are zero-padded
	fullSelectedCols := make([][]field.Element, len(entryList))
	for j := range entryList {
		fullSelectedCol := vr.GetColumn(ctx.SelectedColName(j))
		fullSelectedCols[j] = smartvectors.IntoRegVec(fullSelectedCol)
	}

	// Split the columns per commitment for the verification
	openedSubColumns := [][][]field.Element{}
	roundStartAt := 0

	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		// again, skip the dry rounds
		if ctx.isDry(round) {
			continue
		}

		openedSubColumnsForRound := make([][]field.Element, len(entryList))
		numRowsForRound := ctx.getNbCommittedRows(round)
		for j := range entryList {
			openedSubColumnsForRound[j] = fullSelectedCols[j][roundStartAt : roundStartAt+numRowsForRound]
		}

		// update the start counter to ensure we do not pass twice the same row
		roundStartAt += numRowsForRound
		openedSubColumns = append(openedSubColumns, openedSubColumnsForRound)
	}

	// sanity-check : make sure we have not forgotten any column
	if roundStartAt != ctx.CommittedRowsCount {
		utils.Panic("we have a mistmatch in the row count : %v != %v", roundStartAt, ctx.CommittedRowsCount)
	}

	return openedSubColumns
}

// Evaluates explicitly the public polynomials (proof, vk, public inputs)
func (ctx *Ctx) explicitPublicEvaluation(vr *wizard.VerifierRuntime) error {

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

		y := smartvectors.Interpolate(val, params.X)
		if y != params.Ys[i] {
			return fmt.Errorf("inconsistent evaluation")
		}
	}

	return nil
}

// A shadow row is a row filled with zeroes that we **may** add
// at the end of the rounds commitment. Its purpose is to ensure
// the number of "SIS limbs" in a row divides the degree of the
// ring-SIS instance.
func autoAssignedShadowRow(comp *wizard.CompiledIOP, size, round, id int) ifaces.Column {

	name := ifaces.ColIDf("VORTEX_%v_SHADOW_ROUND_%v_ID_%v", comp.SelfRecursionCount, round, id)
	col := comp.InsertCommit(round, name, size)

	comp.SubProvers.AppendToInner(round, func(assi *wizard.ProverRuntime) {
		assi.AssignColumn(name, smartvectors.NewConstant(field.Zero(), size))
	})

	return col
}

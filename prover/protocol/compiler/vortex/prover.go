package vortex

import (
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/vortex2"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Prover steps of Vortex that is run in place of committing to polynomials
func (ctx *Ctx) AssignColumn(round int) func(*wizard.ProverRuntime) {

	// Check if that is a dry round
	if ctx.isDry(round) {
		// Nothing special to do. The prover will send the polynomials
		// to verifier directly and the verifier will be able to check
		// the evaluation by himself
		return func(pr *wizard.ProverRuntime) {}
	}

	return func(pr *wizard.ProverRuntime) {
		pols := ctx.getPols(pr, round)

		if !ctx.UseMerkleProof {
			// Call Vortex in vanilla mode
			commitment, committedMatrix := ctx.VortexParams.Commit(pols)
			pr.State.InsertNew(ctx.VortexProverStateName(round), committedMatrix)
			pr.AssignColumn(ctx.CommitmentName(round), smartvectors.NewRegular(commitment))
			return
		}

		if ctx.UseMerkleProof {
			// Call Vortex in Merkle mode
			committedMatrix, tree, sisDigest := ctx.VortexParams.CommitMerkle(pols)
			pr.State.InsertNew(ctx.VortexProverStateName(round), committedMatrix)
			pr.State.InsertNew(ctx.MerkleTreeName(round), tree)

			// Only to be read by the self-recursion compiler.
			if ctx.IsSelfrecursed {
				pr.State.InsertNew(string(ctx.CommitmentName(round)), sisDigest)
			}

			// And assign the 1-sized column to contain the root
			var root field.Element
			root.SetBytes(tree.Root[:])
			pr.AssignColumn(ifaces.ColID(ctx.MerkleRootName(round)), smartvectors.NewConstant(root, 1))
			return
		}
	}
}

// Prover steps of Vortex that is run when committing to the linear combination
func (ctx *Ctx) ComputeLinearComb(pr *wizard.ProverRuntime) {

	committedSV := []smartvectors.SmartVector{}

	// Collect all the committed polynomials : round by round
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		// There are not included in the commitments so there
		// is no need to compute their linear combination.
		if ctx.isDry(round) {
			continue
		}

		pols := ctx.getPols(pr, round)
		committedSV = append(committedSV, pols...)
	}

	// And get the randomness
	randomCoinLC := pr.GetRandomCoinField(ctx.LinCombRandCoinName())

	// and compute and assign the random linear combination of the rows
	proof := ctx.VortexParams.OpenWithLC(committedSV, randomCoinLC)
	pr.AssignColumn(ctx.LinCombName(), proof.LinearCombination)
}

// Prover steps of Vortex where he opens the columns selected by the verifier
func (ctx *Ctx) OpenSelectedColumns(pr *wizard.ProverRuntime) {

	committedMatrices := []vortex2.CommittedMatrix{}

	// left at this default value in case ctx.UseMerkleTree == false
	trees := []*smt.Tree{}

	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		// There are not included in the commitments so there is no need to
		// compute their linear combination.
		if ctx.isDry(round) {
			continue
		}

		// Fetch it from the state
		committedMatrix := pr.State.MustGet(ctx.VortexProverStateName(round)).(vortex2.CommittedMatrix)
		// and delete it because it won't be needed anymore and its very heavy
		pr.State.Del(ctx.VortexProverStateName(round))
		committedMatrices = append(committedMatrices, committedMatrix)

		if ctx.UseMerkleProof {
			// Also fetches the trees from the prover state
			tree := pr.State.MustGet(ctx.MerkleTreeName(round)).(*smt.Tree)
			trees = append(trees, tree)
		}
	}

	entryList := pr.GetRandomCoinIntegerVec(ctx.RandColSelectionName())
	proof := vortex2.Proof{}
	proof.WithEntryList(committedMatrices, entryList)
	selectedCols := proof.Columns

	// The columns are split by commitment round. So we need to
	// restick them when we commit them.
	for j := range entryList {
		fullCol := []field.Element{}
		for i := range selectedCols {
			fullCol = append(fullCol, selectedCols[i][j]...)
		}

		// Converts it into a smart-vector and zero-pad it if necessary
		var assignable smartvectors.SmartVector = smartvectors.NewRegular(fullCol)
		if assignable.Len() < utils.NextPowerOfTwo(len(fullCol)) {
			assignable = smartvectors.RightZeroPadded(fullCol, utils.NextPowerOfTwo(len(fullCol)))
		}

		pr.AssignColumn(ctx.SelectedColName(j), assignable)
	}

	if ctx.UseMerkleProof {
		// Merkle mode only:
		// Amend the Vortex proof with the Merkle proofs and registers
		// the Merkle proofs in the
		proof.WithMerkleProof(trees, entryList)
		packedMProofs := ctx.packMerkleProofs(proof.MerkleProofs)
		pr.AssignColumn(ctx.MerkleProofName(), packedMProofs)
	}
}

// returns true if the round is dry (i.e, there is nothing to commit to)
func (ctx *Ctx) isDry(round int) bool {
	return ctx.CommitmentsByRounds.Len() <= round || ctx.CommitmentsByRounds.LenOf(round) == 0
}

// returns the list of all committed smartvectors for the given round
// so that we can commit to them
func (ctx *Ctx) getPols(run *wizard.ProverRuntime, round int) (pols []smartvectors.SmartVector) {
	names := ctx.CommitmentsByRounds.MustGet(round)
	pols = make([]smartvectors.SmartVector, len(names))
	for i := range names {
		pols[i] = run.Columns.MustGet(names[i])
	}
	return pols
}

// pack a list of merkle-proofs in a vector as in
// https://github.com/ConsenSys/zkevm-monorepo/issues/67
func (ctx *Ctx) packMerkleProofs(proofs [][]smt.Proof) smartvectors.SmartVector {

	depth := len(proofs[0][0].Siblings) // depth of the Merkle-tree
	res := make([]field.Element, ctx.MerkleProofSize())
	numProofWritten := 0

	// Sanity-checks

	if depth != utils.Log2Ceil(ctx.NumEncodedCols()) {
		utils.Panic(
			"expected depth to be equal to Log2(NumEncodedCols()), got %v, %v",
			depth, utils.Log2Ceil(ctx.NumEncodedCols()),
		)
	}

	if len(proofs) != ctx.NumCommittedRounds() {
		utils.Panic(
			"inconsitent proofs length %v, %v",
			len(proofs), ctx.NumCommittedRounds(),
		)
	}

	if len(proofs[0]) != ctx.NbColsToOpen() {
		utils.Panic(
			"expected proofs[0] and NbColsToOpen to be equal: %v, %v",
			len(proofs[0]), ctx.NbColsToOpen(),
		)
	}

	for i := range proofs {
		for j := range proofs[i] {
			p := proofs[i][j]
			for k := range p.Siblings {
				// The proof stores the sibling bottom-up but
				// we want to pack the proof in top-down order.
				res[numProofWritten*depth+k].SetBytes(p.Siblings[depth-1-k][:])
			}
			numProofWritten++
		}
	}

	return smartvectors.NewRegular(res)
}

// unpack a list of merkle proofs from a vector as in
// https://github.com/ConsenSys/zkevm-monorepo/issues/67
func (ctx *Ctx) unpackMerkleProofs(sv smartvectors.SmartVector, entryList []int) (proofs [][]smt.Proof) {

	depth := utils.Log2Ceil(ctx.NumEncodedCols()) // depth of the Merkle-tree
	numComs := ctx.NumCommittedRounds()
	numEntries := len(entryList)

	proofs = make([][]smt.Proof, numComs)
	curr := 0 // tracks the position in sv that we are parsing.
	for i := range proofs {
		proofs[i] = make([]smt.Proof, numEntries)
		for j := range proofs[i] {
			// initialize the proof that we are parsing
			proof := smt.Proof{
				Path:     entryList[j],
				Siblings: make([]hashtypes.Digest, depth),
			}

			// parse the siblings accounting for the fact that we
			// are inversing the order.
			for k := range proof.Siblings {
				v := sv.Get(curr)
				proof.Siblings[depth-k-1] = hashtypes.Digest(v.Bytes())
				curr++
			}

			proofs[i][j] = proof
		}
	}
	return proofs
}

package vortex

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
)

// return the name of the linear combination random coin
func (ctx *Ctx) LinCombRandCoinName() coin.Name {
	return coin.Namef("VORTEX_%v_LC_RANDOM_COIN", ctx.SelfRecursionCount)
}

// return the name of the random linear combination randomness
func (ctx *Ctx) LinCombName() ifaces.ColID {
	return ifaces.ColIDf("VORTEX_%v_ROW_LINEAR_COMBINATION", ctx.SelfRecursionCount)
}

// return the name of the linear combination
func (ctx *Ctx) RandColSelectionName() coin.Name {
	return coin.Namef("VORTEX_%v_COL_SELECTION", ctx.SelfRecursionCount)
}

// return the name of the i-th randomly selected columns
func (ctx *Ctx) SelectedColName(num int) ifaces.ColID {
	return ifaces.ColIDf("VORTEX_%v_SELECTED_COL_#%v", ctx.SelfRecursionCount, num)
}

// returns a formatted message name for the commitment of the given round
func (ctx *Ctx) CommitmentName(round int) ifaces.ColID {
	return ifaces.ColIDf("VORTEX_%v_COMMITMENT_ROUND_%v", ctx.SelfRecursionCount, round)
}

// returns the name of a prover state for a given round of Vortex
func (ctx *Ctx) VortexProverStateName(round int) string {
	return fmt.Sprintf("VORTEX_%v_PROVER_STATE_%v", ctx.SelfRecursionCount, round)
}

// returns the name of a prover state for a given round of Vortex
func (ctx *Ctx) MerkleTreeName(round int) string {
	return fmt.Sprintf("VORTEX_%v_MERKLE_TREE_%v", ctx.SelfRecursionCount, round)
}

// returns the name of the vector containing all the Merkle proofs
func (ctx *Ctx) MerkleProofName() ifaces.ColID {
	return ifaces.ColIDf("VORTEX_%v_MERKLEPROOF", ctx.SelfRecursionCount)
}

// returns the name of the vector containing all the Merkle proofs
func (ctx *Ctx) MerkleRootName(round int) ifaces.ColID {
	return ifaces.ColIDf("VORTEX_%v_MERKLEROOT_%v", ctx.SelfRecursionCount, round)
}

// returns the name of the precomputed commitment when Merkle is not applied
func (ctx *Ctx) PrecomputedCommitmentNameWithoutMerkle() ifaces.ColID {
	return ifaces.ColIDf("VORTEX_PRECOMPUTED_COMMITMENT_WITHOUT_MERKLE")
}

// returns the name of the precomputed sis digest when Merkle is applied
func (ctx *Ctx) PrecomputedSisDigestNameWithMerkle() ifaces.ColID {
	return ifaces.ColIDf("VORTEX_PRECOMPUTED_SIS_DIGEST_WITH_MERKLE_%d", ctx.SelfRecursionCount)
}

// returns the name of the precomputed Merkle root when Merkle is applied
func (ctx *Ctx) PrecomputedMerkleRootName() ifaces.ColID {
	return ifaces.ColIDf("VORTEX_PRECOMPUTED_MERKLE_ROOT_%d", ctx.SelfRecursionCount)
}

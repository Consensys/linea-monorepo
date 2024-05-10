package selfrecursion

import (
	"fmt"

	"github.com/consensys/zkevm-monorepo/prover/crypto/ringsis"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"

	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// Name of the polynomial I(x)
func (ctx *SelfRecursionCtx) iName(length int) ifaces.ColID {
	return ifaces.ColIDf("PRECOMPUTED_%v_I_%v", ctx.SelfRecursionCnt, length)
}

// Name of the aH polynomials
func (ctx *SelfRecursionCtx) ahName(key *ringsis.Key, start, length, maxSize int) ifaces.ColID {
	// Sanity-check : the chunkNo can't be off-bound
	if maxSize < start+length {
		utils.Panic("inconsistent arguments : %v + %v > %v", start, length, maxSize)
	}

	subName := ifaces.ColIDf("SISKEY_%v_%v_%v", key.LogTwoBound, key.LogTwoDegree, key.MaxNumFieldHashable())
	return ifaces.ColIDf("%v_%v_%v_%v", subName, start, length, maxSize)
}

// Name of the preimage in limb expanded. nameWhole is the name of the
// associated column without limb expansion.
func (ctx *SelfRecursionCtx) limbExpandedPreimageName(nameWhole ifaces.ColID) ifaces.ColID {
	return ifaces.ColIDf("%v_LIMB_EXPANDED_%v", nameWhole, ctx.SelfRecursionCnt)
}

// Name of the UalphaQ column
func (ctx *SelfRecursionCtx) uAlphaQName() ifaces.ColID {
	return ifaces.ColIDf("SELFRECURSION_U_ALPHA_Q_%v", ctx.SelfRecursionCnt)
}

// Name of the self-recursed inclusion query
func (ctx *SelfRecursionCtx) selectQInclusion() ifaces.QueryID {
	return ifaces.QueryIDf("SELFRECURSION_SELECT_Q_INCLUSION_%v", ctx.SelfRecursionCnt)
}

// Name of the collapse coin
func (ctx *SelfRecursionCtx) collapseCoin() coin.Name {
	return coin.Namef("SELFRECURSION_COLLAPSE_COIN_%v", ctx.SelfRecursionCnt)
}

// Name for the coeff eval for consistency check between Ualphaq
// and the preimage (left-side, over UalphaA)
func (ctx *SelfRecursionCtx) constencyUalphaQPreimageLeft() string {
	return fmt.Sprintf("SELFRECURSION_CONSISTENCY_UALPHA_PREIMAGE_LEFT_%v", ctx.SelfRecursionCnt)
}

// Name for the coeff eval for consistency check between Ualphaq
// and the preimage (right-side, over preiimages)
func (ctx *SelfRecursionCtx) constencyUalphaQPreimageRight() string {
	return fmt.Sprintf("SELFRECURSION_CONSISTENCY_UALPHA_PREIMAGE_RIGHT_%v", ctx.SelfRecursionCnt)
}

// Name of Edual
func (ctx *SelfRecursionCtx) eDual() ifaces.ColID {
	return ifaces.ColIDf("SELFRECURSION_E_DUAL_%v", ctx.SelfRecursionCnt)
}

// Name for the fold coin
func (ctx *SelfRecursionCtx) foldCoinName() coin.Name {
	return coin.Namef("SELFRECURSION_FOLD_COIN_%v", ctx.SelfRecursionCnt)
}

// Name of the inner product between between PreimageCollapseFold and ACollapseFold
func (ctx *SelfRecursionCtx) preimagesAndAmergeIP() ifaces.QueryID {
	return ifaces.QueryIDf("SELFRECURSION_PREIMAGE_A_IP_%v", ctx.SelfRecursionCnt)
}

// Name of the interpolation context for Ualpha
func (ctx *SelfRecursionCtx) interpolateUAlphaX() string {
	return fmt.Sprintf("SELFRECURSION_INTERPOLATE_UALPHA_X_%v", ctx.SelfRecursionCnt)
}

// Name of the concatenation of the DhQs
func (ctx *SelfRecursionCtx) concatenatedDhQ() ifaces.ColID {
	return ifaces.ColIDf("SELFRECURSION_CONCAT_DHQ_%v", ctx.SelfRecursionCnt)
}

// Name of the MerkleLeaves
func (ctx *SelfRecursionCtx) merkleLeavesName() ifaces.ColID {
	return ifaces.ColIDf("SELFRECURSION_MERKLE_LEAVES_%v", ctx.SelfRecursionCnt)
}

// Name of the MerklePositions
func (ctx *SelfRecursionCtx) merklePositionssName() ifaces.ColID {
	return ifaces.ColIDf("SELFRECURSION_MERKLE_POSITIONS_%v", ctx.SelfRecursionCnt)
}

// Name of the MerkleRoots
func (ctx *SelfRecursionCtx) merkleRootsName() ifaces.ColID {
	return ifaces.ColIDf("SELFRECURSION_MERKLE_ROOTS_%v", ctx.SelfRecursionCnt)
}

// Name of the Merkle proof verification
func (ctx *SelfRecursionCtx) merkleProofVerificationName() string {
	return fmt.Sprintf("SELFRECURSION_MERKLE_%v", ctx.SelfRecursionCnt)
}

// Name of the collapsed key
func (ctx *SelfRecursionCtx) aCollapsedName() string {
	return fmt.Sprintf("SELFRECURSION_ACOLLAPSE_%v", ctx.comp.SelfRecursionCount)
}

// Name of the collapsed key
func (ctx *SelfRecursionCtx) rootHasGlue() ifaces.QueryID {
	return ifaces.QueryIDf("SELFRECURSION_ROOT_HASH_GLUE_%v", ctx.comp.SelfRecursionCount)
}

// Positions glue
func (ctx *SelfRecursionCtx) positionGlue() ifaces.QueryID {
	return ifaces.QueryIDf("SELFRECURSION_POSITION_GLUE_%v", ctx.comp.SelfRecursionCount)
}

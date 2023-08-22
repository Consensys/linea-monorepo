package selfrecursion

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/sishashing"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
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

	subName := sishashing.SisKeyName(key)
	return ifaces.ColIDf("%v_%v_%v_%v", subName, start, length, maxSize)
}

// Name of the preimage in limb expanded. nameWhole is the name of the
// associated column without limb expansion.
func (ctx *SelfRecursionCtx) limbExpandedPreimageName(nameWhole ifaces.ColID) ifaces.ColID {
	return ifaces.ColIDf("%v_LIMB_EXPANDED_%v", nameWhole, ctx.SelfRecursionCnt)
}

// Name of the merge coin
func (ctx *SelfRecursionCtx) mergeCoinName() coin.Name {
	return coin.Namef("SELFRECURSION_MERGE_COIN_%v", ctx.SelfRecursionCnt)
}

// Name of the A merge column
func (ctx *SelfRecursionCtx) aMergeColName() string {
	return fmt.Sprintf("SELFRECURSION_A_MERGE_%v", ctx.SelfRecursionCnt)
}

// Name of the A merge column
func (ctx *SelfRecursionCtx) dMergeColName() string {
	return fmt.Sprintf("SELFRECURSION_D_MERGE_%v", ctx.SelfRecursionCnt)
}

// Name of the DmergeQ column
func (ctx *SelfRecursionCtx) dMergeQName() ifaces.ColID {
	return ifaces.ColIDf("SELFRECURSION_D_MERGE_Q_%v", ctx.SelfRecursionCnt)
}

// Name of the UalphaQ column
func (ctx *SelfRecursionCtx) uAlphaQName() ifaces.ColID {
	return ifaces.ColIDf("SELFRECURSION_U_ALPHA_Q_%v", ctx.SelfRecursionCnt)
}

// Name of the squash coin
func (ctx *SelfRecursionCtx) squashCoin() coin.Name {
	return coin.Namef("SELFRECURSION_SQUASH_COIN_%v", ctx.SelfRecursionCnt)
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

// Name of the inner product between between PreimageCollapseFold and AmergeFold
func (ctx *SelfRecursionCtx) preimagesAndAmergeIP() ifaces.QueryID {
	return ifaces.QueryIDf("SELFRECURSION_PREIMAGE_A_IP_%v", ctx.SelfRecursionCnt)
}

// Name of the interpolation context for Ualpha
func (ctx *SelfRecursionCtx) interpolateUAlphaX() string {
	return fmt.Sprintf("SELFRECURSION_INTERPOLATE_UALPHA_X_%v", ctx.SelfRecursionCnt)
}

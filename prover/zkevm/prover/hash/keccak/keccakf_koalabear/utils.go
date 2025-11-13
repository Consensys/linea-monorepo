package keccakfkoalabear

import (
	"strconv"
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
	common_coalabear "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
)

// decompose decomposes the non-negative integer r into digits (limbs) in the given base.
//
// The function returns a slice of uint64 holding the base-"base" representation of r,
// with the least-significant limb first (little-endian digit order). The returned slice
// has length equal to the number of non-zero limbs produced (which may be 0 when r == 0)
// and a capacity reserved for nb limbs (make([]uint64, 0, nb)).
//
// Parameters:
//   - r: the value to decompose (uint64).
//   - base: the base to use for decomposition (interpreted as uint64 internally).
//   - numLimbs: the maximum number of limbs expected; also used as the initial capacity of the result.
//   - useFullLength: if true, the function will ensure that the returned slice has exactly numLimbs elements.
//
// Behavior and guarantees:
//   - Uses repeated division and modulo to extract limbs: limb = curr % base, curr /= base.
//   - Returns a slice of limbs in little-endian order (least-significant first).
//   - If the decomposition requires more than numLimbs limbs, the function triggers utils.Panic
//     with an explanatory message ("expected %v limbs, but got %v").
//   - res may contain fewer than numLimbs limbs if r is small.
func decompose(r uint64, base int, numLimbs int, useFullLength bool) (res []uint64) {
	// It will essentially be used for chunk to slice decomposition
	if base < 2 {
		utils.Panic("base must be at least 2, got %v", base)
	}
	res = make([]uint64, 0, numLimbs)
	base64 := uint64(base)
	curr := r
	// Handle the zero case explicitly
	if curr == 0 {
		if useFullLength {
			for i := 0; i < numLimbs; i++ {
				res = append(res, 0)
			}
		} else {
			res = append(res, 0)
		}
		return res
	}
	for curr > 0 {
		limb := curr % base64
		res = append(res, limb)
		curr /= base64
	}

	if len(res) > numLimbs {
		utils.Panic("expected %v limbs, but got %v", numLimbs, len(res))
	}

	if useFullLength {
		// pad with zeros to reach numLimbs
		for len(res) < numLimbs {
			res = append(res, 0)
		}
	}

	return res
}

// numRowsKeccakSmallField returns the number of rows required to prove `numKeccakf` calls to the
// permutation function. The result is padded to the next power of 2 in order to
// satisfy the requirements of the Wizard to have only powers of 2.
func numRowsKeccakSmallField(numKeccakf int) int {
	return utils.NextPowerOfTwo(numKeccakf * common_coalabear.NumRounds)
}

// deriveNameKeccakFSmallField derive column names
func deriveNameKeccakFSmallField(mainName string, ids ...int) ifaces.ColID {
	idStr := []string{}
	for i := range ids {
		idStr = append(idStr, strconv.Itoa(ids[i]))
	}
	return ifaces.ColIDf("%v_%v_%v", "KECCAKF", mainName, strings.Join(idStr, "_"))
}

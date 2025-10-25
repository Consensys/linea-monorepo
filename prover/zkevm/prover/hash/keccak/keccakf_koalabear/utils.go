package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Decompose decomposes the non-negative integer r into digits (limbs) in the given base.
//
// The function returns a slice of uint64 holding the base-"base" representation of r,
// with the least-significant limb first (little-endian digit order). The returned slice
// has length equal to the number of non-zero limbs produced (which may be 0 when r == 0)
// and a capacity reserved for nb limbs (make([]uint64, 0, nb)).
//
// Parameters:
//  - r: the value to decompose (uint64).
//  - base: the base to use for decomposition (interpreted as uint64 internally).
//  - numLimbs: the maximum number of limbs expected; also used as the initial capacity of the result.
//
// Behavior and guarantees:
//  - Uses repeated division and modulo to extract limbs: limb = curr % base, curr /= base.
//  - Returns a slice of limbs in little-endian order (least-significant first).
//  - If the decomposition requires more than numLimbs limbs, the function triggers utils.Panic
//    with an explanatory message ("expected %v limbs, but got %v").
func Decompose(r uint64, base int, numLimbs int) (res []uint64) {
	// It will essentially be used for chunk to slice decomposition
	if base < 2 {
		utils.Panic("base must be at least 2, got %v", base)
	}
	res = make([]uint64, 0, numLimbs)
	base64 := uint64(base)
	curr := r
	for curr > 0 {
		limb := curr % base64
		res = append(res, limb)
		curr /= base64
	}

	if len(res) > numLimbs {
		utils.Panic("expected %v limbs, but got %v", numLimbs, len(res))
	}

	if len(res) < numLimbs {
		// pad with zeros to have exactly numLimbs limbs
		for len(res) < numLimbs {
			res = append(res, 0)
		}
	}

	return res
}

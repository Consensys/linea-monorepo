package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/utils"
)

func Decompose(r uint64, base int, nb int) (res []uint64) {
	// It will essentially be used for chunk to slice decomposition
	res = make([]uint64, 0, nb)
	base64 := uint64(base)
	curr := r
	for curr > 0 {
		limb := curr % base64
		res = append(res, limb)
		curr /= base64
	}

	if len(res) > nb {
		utils.Panic("expected %v limbs, but got %v", nb, len(res))
	}

	return res
}

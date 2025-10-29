package common

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

func Decompose(n []field.Element, base, nb int) [][]field.Element {

	res := make([][]field.Element, nb)
	for i := range n {
		a := DecomposeUint32(n[i].Uint64(), base, nb)
		for j := range res {
			res[j] = append(res[j], a[j])
		}
	}
	return res
}

// CleanBase takes as input a slice of uint64 numbers in base BaseChi (11) and clean the representation.
func CleanBase(in []field.Element) (out []field.Element) {
	out = make([]field.Element, len(in))
	for i := 0; i < len(in); i++ {
		// take the second bit
		a := in[i].Uint64() >> 1 & 1
		out[i] = field.NewElement(a)
	}
	return out
}

func DecomposeAndClean(n []field.Element, base, nb int) [][]field.Element {

	limbs := Decompose(n, base, nb)             // in base 'base', nb limbs
	bits := make([][]field.Element, len(limbs)) // bits representation of each limb
	for i := range limbs {
		bits[i] = CleanBase(limbs[i])
	}
	return bits
}

func BaseRecompose(n [][]field.Element, base *field.Element) (res []field.Element) {

	for j := range n[0] {

		var t []field.Element
		for i := range n {
			t = append(t, n[i][j])
		}

		res = append(res, keccakf.BaseRecompose(t, base))
	}
	return res
}

// it decompose the given field element n into the given base.
func DecomposeUint32(n uint64, base, nb int) []field.Element {
	// It will essentially be used for chunk to slice decomposition
	var (
		res    = make([]field.Element, 0, nb)
		curr   = n
		base64 = uint64(base)
	)
	for curr > 0 {
		limb := field.NewElement(uint64(curr % base64))
		res = append(res, limb)
		curr /= base64
	}

	if len(res) > nb {
		utils.Panic("expected %v limbs, but got %v", nb, len(res))
	}

	// Complete with zeroes
	for len(res) < nb {
		res = append(res, field.Zero())
	}

	return res
}

package common

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func DecomposeCol(n []field.Element, base, nb int) [][]field.Element {

	res := make([][]field.Element, nb)
	for i := range n {
		a := DecomposeU64(n[i].Uint64(), base, nb)
		for j := range res {
			res[j] = append(res[j], a[j])
		}
	}
	return res
}

// CleanBaseChi takes as input a slice of uint64 numbers in base BaseChi (11) and clean the representation.
func CleanBaseChi(in []field.Element) (out []field.Element) {
	out = make([]field.Element, len(in))
	for i := 0; i < len(in); i++ {
		// take the second bit
		a := in[i].Uint64() >> 1 & 1
		out[i] = field.NewElement(a)
	}
	return out
}

// it recompose the given columns of field elements n into the given base.
func RecomposeCols(n [][]field.Element, base []field.Element) (res []field.Element) {

	for j := range n[0] {

		var t []field.Element
		for i := range n {
			t = append(t, n[i][j])
		}

		res = append(res, RecomposeRow(t, &base[j]))
	}
	return res
}

// it recompose the given slice of field elements r into the given base.
func RecomposeRow(r []field.Element, base *field.Element) field.Element {
	// .. using the Horner method
	s := field.Zero()
	for i := len(r) - 1; i >= 0; i-- {
		s.Mul(&s, base)
		s.Add(&s, &r[i])
	}
	return s
}

// it decompose the given field element n into the given base.
func DecomposeU64(n uint64, base, nb int) []field.Element {
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

func DecomposeAndCleanFr(n field.Element, base, nb int) []field.Element {

	var (
		bits  []field.Element
		limbs = DecomposeU64(n.Uint64(), base, nb) // in base 'base', nb limbs
	)

	switch {
	case base == BaseChi:
		bits = CleanBaseChi(limbs) // bits representation of each limb
	case base == BaseTheta:
		bits = cleanBaseTheta(limbs) // bits representation of each limb
	default:
		utils.Panic("unsupported base %v for cleaning", base)
	}

	return bits
}

func cleanBaseTheta(in []field.Element) (out []field.Element) {
	out = make([]field.Element, 0, len(in))
	for _, element := range in {
		if element.Uint64()%2 == 0 {
			out = append(out, field.Zero())
		} else {
			out = append(out, field.One())
		}
	}
	return out
}

func DecomposeAndCleanCol(n []field.Element, base, nb int) [][]field.Element {

	var (
		res = make([][]field.Element, nb)
		dec = make([]field.Element, nb)
	)
	for i := range n {
		dec = DecomposeAndCleanFr(n[i], base, nb)
		for j := range dec {
			res[j] = append(res[j], dec[j])
		}
	}

	return res
}

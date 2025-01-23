package projection

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// cmptHorner computes a random Horner accumulation of the filtered elements
// starting from the last entry down to the first entry. The final value is
// stored in the last entry of the returned slice.
func cmptHorner(c, fC []field.Element, x field.Element) []field.Element {

	var (
		horner = make([]field.Element, len(c))
		prev   field.Element
	)

	for i := len(horner) - 1; i >= 0; i-- {

		if !fC[i].IsZero() && !fC[i].IsOne() {
			utils.Panic("we expected the filter to be binary")
		}

		if fC[i].IsOne() {
			prev.Mul(&prev, &x)
			prev.Add(&prev, &c[i])
		}

		horner[i] = prev
	}

	return horner
}

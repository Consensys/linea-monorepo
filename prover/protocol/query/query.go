package query

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
)

/*
Utility function used to manually check permutation and inclusion
constraints. Will return a linear combination of i-th element of
each list.
*/
func rowLinComb(alpha field.Element, i int, list []ifaces.ColAssignment) field.Element {
	var res field.Element
	for j := range list {
		res.Mul(&res, &alpha)
		x := list[j].Get(i)
		res.Add(&res, &x)
	}
	return res
}

// mustBeNaturalOrVerifierCol checks if the column is a [column.Natural]
// column or verifiercol.VerifierCol column.
func mustBeNaturalOrVerifierCol(col ifaces.Column) {
	if col.IsComposite() {
		utils.Panic("column %v should be either a Natural or a VerifierCol", col)
	}
}

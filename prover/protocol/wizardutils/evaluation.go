package wizardutils

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// returns the symbolic expression of a column obtained as a random linear combinations of differents handles
// without committing to the column itself
func RandLinCombColSymbolic(x coin.Info, hs []ifaces.Column) *symbolic.Expression {
	cols := make([]*symbolic.Expression, len(hs))
	for c := range cols {
		cols[c] = ifaces.ColumnAsVariable(hs[c])
	}
	expr := symbolic.NewPolyEval(x.AsVariable(), cols)
	return expr
}

// return the runtime assignments of a linear combination column
// that is computed on the fly from the columns stored in hs
func RandLinCombColAssignment(run *wizard.ProverRuntime, coinVal field.Element, hs []ifaces.Column) smartvectors.SmartVector {
	var colTableWit smartvectors.SmartVector
	var witnessCollapsed smartvectors.SmartVector
	x := field.One()
	witnessCollapsed = smartvectors.NewConstant(field.Zero(), hs[0].Size())
	for tableCol := range hs {
		colTableWit = hs[tableCol].GetColAssignment(run)
		witnessCollapsed = smartvectors.Add(witnessCollapsed, smartvectors.Mul(colTableWit, smartvectors.NewConstant(x, hs[0].Size())))
		x.Mul(&x, &coinVal)
	}
	return witnessCollapsed
}

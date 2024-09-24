package expr_handle

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// returns a column obtained as a random linear combinations of differents handle
func RandLinCombCol(comp *wizard.CompiledIOP, x ifaces.Accessor, hs []ifaces.Column, name ...string) ifaces.Column {
	cols := make([]*symbolic.Expression, len(hs))
	for c := range cols {
		cols[c] = ifaces.ColumnAsVariable(hs[c])
	}
	expr := symbolic.NewPolyEval(x.AsVariable(), cols)
	return ExprHandle(comp, expr, name...)
}

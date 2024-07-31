package wizardutils

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/variables"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// EvalExprColumn resolves an expression to a column assignment. The expression
// must be converted to a board prior to evaluating the expression.
//
//   - If the expression does not uses ifaces.Column as metadata, the function
//     will panic.
//
//   - If the expression contains several columns and they don't contain all
//     have the same size.
func EvalExprColumn(run *wizard.ProverRuntime, board symbolic.ExpressionBoard) smartvectors.SmartVector {

	var (
		metadata = board.ListVariableMetadata()
		inputs   = make([]smartvectors.SmartVector, len(metadata))
		length   = ExprIsOnSameLengthHandles(&board)
	)

	// Attempt to recover the size of the
	for i := range inputs {
		switch m := metadata[i].(type) {
		case ifaces.Column:
			inputs[i] = m.GetColAssignment(run)
		case coin.Info:
			v := run.GetRandomCoinField(m.Name)
			inputs[i] = smartvectors.NewConstant(v, length)
		case ifaces.Accessor:
			v := m.GetVal(run)
			inputs[i] = smartvectors.NewConstant(v, length)
		case variables.PeriodicSample:
			v := m.EvalCoset(length, 0, 1, false)
			inputs[i] = v
		case variables.X:
			v := m.EvalCoset(length, 0, 1, false)
			inputs[i] = v
		}
	}

	return board.Evaluate(inputs)
}

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

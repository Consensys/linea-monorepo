package wizardutils

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// RandLinCombColSymbolic generates a symbolic expression representing a random
// linear combination of columns (hs) using a random coin value (x). It converts
// each column to a symbolic variable and returns a polynomial expression with
// the coin value and each columns as the variables.
func RandLinCombColSymbolic(x coin.Info, hs []ifaces.Column) *symbolic.Expression {
	cols := make([]*symbolic.Expression, len(hs))
	for c := range cols {
		cols[c] = ifaces.ColumnAsVariable(hs[c])
	}
	expr := symbolic.NewPolyEval(x.AsVariable(), cols)
	return expr
}

// RandLinCombColAssignment computes the runtime assignment of a linear combination
// of columns. It takes a wizard.ProverRuntime, a random coin value, and a slice of
// columns as input. The function iteratively multiplies each column's assignment
// by a cumulative product of the coin value and adds the results to a running total,
// effectively computing a weighted sum of the columns. The weights are powers of the
// coin value. The function returns the resulting linear combination as a
// [smartvectors.SmartVector].
func RandLinCombColAssignment(run *wizard.ProverRuntime, coinVal fext.Element, hs []ifaces.Column) smartvectors.SmartVector {
	if len(hs) == 0 {
		panic("cannot compute random linear combination of zero columns")
	}

	x := fext.One()

	vColumn := make(vectorext.Vector, hs[0].Size())
	vWitness := make(vectorext.Vector, hs[0].Size())

	for tableCol := range hs {
		sv := hs[tableCol].GetColAssignment(run)
		_vColumn := vColumn
		// if sv is already a regular vector, we can avoid the copy
		if r, ok := sv.(*smartvectors.RegularExt); ok {
			_vColumn = vectorext.Vector(*r)
		} else {
			sv.WriteInSliceExt(_vColumn)
		}
		// vColumn = sv * x
		// vWitness += vColumn
		// x *= coinVal
		vColumn.ScalarMul(_vColumn, &x)
		vWitness.Add(vWitness, _vColumn)
		x.Mul(&x, &coinVal)
	}
	return smartvectors.NewRegularExt(vWitness)
}

// LinCombExpr generates a symbolic expression representing a linear combination
// the expression can be evaluated at point x over the columns hs via [column.EvalExprColumn].
func LinCombExpr(x int, hs []ifaces.Column) *symbolic.Expression {
	cols := make([]*symbolic.Expression, len(hs))
	xExpr := symbolic.NewConstant(int64(x))
	for c := range cols {
		cols[c] = ifaces.ColumnAsVariable(hs[c])
	}
	expr := symbolic.NewPolyEval(xExpr, cols)
	return expr
}

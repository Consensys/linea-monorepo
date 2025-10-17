package protocols

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// linearCombination represents the linear combination of the given columns over the powers of the given scalar.
// it is technically a polynomial evaluation at the given scalar point.
type linearCombination struct {
	// scalar for linear combination
	scalar int
	// input columns
	cols []ifaces.Column
	// output column
	CombinationRes ifaces.Column
	// size of columns
	size int
}

// LinearCombination is similar to polynomial evaluation, implementing [wizard.ProverAction].
func LinearCombination(comp *wizard.CompiledIOP, name string, r []ifaces.Column, base int) *linearCombination {
	var (
		size = r[0].Size()
		col  = comp.InsertCommit(0, ifaces.ColIDf("LINEAR_COMBINATION_RESULT_COL_%v", name), size)
	)
	// .. using the Horner method
	s := sym.NewConstant(0)
	for i := len(r) - 1; i >= 0; i-- {
		s = sym.Mul(s, base)
		s = sym.Add(s, r[i])
	}

	comp.InsertGlobal(0, ifaces.QueryIDf("LINEAR_COMBINATION_RESULT_QUERY_%v", name),
		sym.Sub(col, s),
	)

	return &linearCombination{
		scalar:         base,
		cols:           r,
		CombinationRes: col,
		size:           size,
	}
}

// Run  assign the values to the linear combination result column.
func (bc *linearCombination) Run(run *wizard.ProverRuntime) {
	var (
		s    = vector.Zero(bc.size)
		base = field.NewElement(uint64(bc.scalar))
	)
	for i := len(bc.cols) - 1; i >= 0; i-- {
		colValue := bc.cols[i].GetColAssignment(run).IntoRegVecSaveAlloc()
		vector.ScalarMul(s, s, base)
		vector.Add(s, s, colValue)
	}

	run.AssignColumn(bc.CombinationRes.GetColID(), smartvectors.NewRegular(s))
}

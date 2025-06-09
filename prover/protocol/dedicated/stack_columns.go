package dedicated

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// StackedColumn is a dedicated wizard computing a column by stacking other
// columns on top of each others.
type StackedColumn struct {
	// Column is the built column
	Column column.Natural
	// Source is the list of columns to stack.
	// If the number of columns is not a power of two,
	// we pad with const zero valued column to the next power of two.
	Source []ifaces.Column
}

// StackColumn defines and constrains a [StackedColumn] wizard element.
func StackColumn(comp *wizard.CompiledIOP, srcs []ifaces.Column) StackedColumn {

	var (
		// s is the identity permutation to be computed
		s     = make([]smartvectors.SmartVector, 0, len(srcs))
		// count is the total number of elements in the stacked column
		count = 0
		name  = fmt.Sprintf("STACKED_COLUMN_%v_%v", len(comp.Columns.AllKeys()), comp.SelfRecursionCount)
		round = 0
	)

	for i := range srcs {
		round = max(round, srcs[i].Round())
		p := make([]field.Element, srcs[i].Size())
		for j := range p {
			p[j].SetInt64(int64(count))
			count++
		}
		s = append(s, smartvectors.NewRegular(p))
	}

	if !utils.IsPowerOfTwo(count) {
		count = utils.NextPowerOfTwo(count)
	}

	col := comp.InsertCommit(round, ifaces.ColID(name), count)

	comp.InsertFixedPermutation(
		round,
		ifaces.QueryID(name)+"_CHECK",
		s,
		[]ifaces.Column{col},
		srcs,
	)

	return StackedColumn{
		Column: col.(column.Natural),
		Source: srcs,
	}
}

// Assigns assigns the stack column
func (s StackedColumn) Run(run *wizard.ProverRuntime) {

	res := make([]field.Element, 0, s.Column.Size())
	for i := range s.Source {
		a := s.Source[i].GetColAssignment(run).IntoRegVecSaveAlloc()
		res = append(res, a...)
	}

	run.AssignColumn(s.Column.ID, smartvectors.RightZeroPadded(res, s.Column.Size()))
}

package dedicated

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// StackedColumn is a dedicated wizard computing a column by stacking other
// columns on top of each others.
type StackedColumn struct {
	// Column is the built column
	Column column.Natural
	// Source is the list of columns to stack. Assumedly they are a power of
	// two number.
	Source []ifaces.Column
}

// StackColumn defines and constrains a [StackedColumn] wizard element.
func StackColumn(comp *wizard.CompiledIOP, srcs []ifaces.Column) StackedColumn {

	var (
		s     = make([]smartvectors.SmartVector, len(srcs))
		count = 0
		name  = fmt.Sprintf("STACKED_COLUMN_%v", len(comp.Columns.AllKeys()))
		round = 0
	)

	for i := range s {
		round = max(round, srcs[i].Round())
		p := make([]field.Element, srcs[i].Size())
		for j := range p {
			p[j].SetInt64(int64(count))
			count++
		}
		s[i] = smartvectors.NewRegular(p)
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

	run.AssignColumn(s.Column.ID, smartvectors.NewRegular(res))
}

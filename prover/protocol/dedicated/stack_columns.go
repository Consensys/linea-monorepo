package dedicated

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
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

	// Make the number of source column a power of two
	srcs_padded := utils.RightPadWith(
		srcs,
		utils.NextPowerOfTwo(len(srcs)),
		verifiercol.NewConstantCol(field.Zero(), 1),
	)
	var (
		s     = make([]smartvectors.SmartVector, len(srcs_padded))
		count = 0
		name  = fmt.Sprintf("STACKED_COLUMN_%v", len(comp.Columns.AllKeys()))
		round = 0
	)

	for i := range s {
		round = max(round, srcs_padded[i].Round())
		p := make([]field.Element, srcs_padded[i].Size())
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
		srcs_padded,
	)

	return StackedColumn{
		Column: col.(column.Natural),
		Source: srcs_padded,
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

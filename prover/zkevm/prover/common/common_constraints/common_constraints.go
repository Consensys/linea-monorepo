package commonconstraints

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// MustZeroWhenInactive constraints the column to cancel when inactive.
func MustZeroWhenInactive(comp *wizard.CompiledIOP, isActive ifaces.Column, cs ...ifaces.Column) {
	for _, c := range cs {
		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("%v_IS_ZERO_WHEN_INACTIVE", c.GetColID()),
			sym.Sub(c, sym.Mul(c, isActive)),
		)
	}
}

// MustBeBinary constrains the current column to be binary.
func MustBeBinary(comp *wizard.CompiledIOP, c ifaces.Column) {
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%v_MUST_BE_BINARY", c.GetColID()),
		sym.Mul(c, sym.Sub(c, 1)),
	)
}

// MustBeActivationColumns constrains all the columns of the form "IsActive" to have
// the correct form: the column is binary and it cannot transition from 0 to 1.
func MustBeActivationColumns(comp *wizard.CompiledIOP, c ifaces.Column, option ...any) {
	MustBeBinary(comp, c)

	// must have activation form where option is not zero.
	var res *sym.Expression
	if len(option) > 0 {
		res = sym.Mul(column.Shift(c, 1), option[0])
	} else {
		res = ifaces.ColumnAsVariable(column.Shift(c, 1))
	}

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%v_CANNOT_TRANSITION_FROM_0_TO_1", c.GetColID()),
		sym.Sub(res,
			sym.Mul(c, column.Shift(c, 1))),
	)
}

// MustBeMutuallyExclusiveBinaryFlags constraints all the flags to be binary
// and sum to isActive
func MustBeMutuallyExclusiveBinaryFlags(comp *wizard.CompiledIOP, isActive ifaces.Column, flags []ifaces.Column) {

	var (
		flagsNames = []string{}
		flagsAny   = []any{}
	)

	for i := range flags {
		MustBeBinary(comp, flags[i])
		flagsNames = append(flagsNames, string(flags[i].GetColID()))
		flagsAny = append(flagsAny, flags[i])
	}

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%v_ARE_MUTUALLY_EXCLUSIVE_WHEN_%v", strings.Join(flagsNames, "_"), isActive.GetColID()),
		sym.Sub(isActive, flagsAny...),
	)
}

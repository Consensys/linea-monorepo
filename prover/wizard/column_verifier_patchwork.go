package wizard

import (
	"strings"

	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

var _ Column = ColumnVerifierPatchwork{}

// ColumnVerifierPatchwork is a column obtained by concatenating together
// [Column]s and [Accessor]s values that are accessible by the verifier.
type ColumnVerifierPatchwork struct {
	components []symbolic.Metadata
	shift      int
}

func NewColumnVerifierPatchwork(v ...symbolic.Metadata) ColumnVerifierPatchwork {
	return ColumnVerifierPatchwork{
		components: v,
	}
}

func (c ColumnVerifierPatchwork) String() string {
	compString := make([]string, len(c.components))
	for i := range compString {
		compString[i] = c.components[i].String()
	}
	return "column-patchwork/" + strings.Join(compString, "||")
}

func (c ColumnVerifierPatchwork) GetAssignment(run Runtime) sv.SmartVector {

	res := []field.Element{}

	for i := range c.components {
		switch m := c.components[i].(type) {
		case Column:
			res = append(res, m.GetAssignment(run).IntoRegVecSaveAlloc()...)
		case Accessor:
			res = append(res, m.GetVal(run))
		default:
			utils.Panic("unexpected type: %T", m)
		}
	}

	return sv.NewRegular(res).RotateRight(-c.shift)
}

func (c ColumnVerifierPatchwork) GetAssignmentGnark(api frontend.API, run GnarkRuntime) []frontend.Variable {

	res := []frontend.Variable{}

	for i := range c.components {
		switch m := c.components[i].(type) {
		case Column:
			res = append(res, m.GetAssignmentGnark(api, run)...)
		case Accessor:
			res = append(res, m.GetValGnark(api, run))
		default:
			utils.Panic("unexpected type: %T", m)
		}
	}

	return shiftFrontendVarSlice(res, c.shift)
}

func (c ColumnVerifierPatchwork) Round() int {
	round := 0
	for i := range c.components {
		v := c.components[i].(interface{ Round() int })
		round = max(round, v.Round())
	}
	return round
}

func (c ColumnVerifierPatchwork) Shift(n int) Column {
	newC := c
	newC.shift += n
	return newC
}

func (c ColumnVerifierPatchwork) Size() int {
	size := 0

	for i := range c.components {
		switch m := c.components[i].(type) {
		case Column:
			size += m.Size()
		case Accessor:
			size++
		default:
			utils.Panic("unexpected type: %T", m)
		}
	}

	return size
}

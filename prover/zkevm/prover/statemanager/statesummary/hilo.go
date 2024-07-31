package statesummary

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/common"
)

// HiLoColumns represents a pair of column representing a sequence of bytes32
// element that do not fit in a single field element. The Hi columns stores the
// first 16 bytes of the column. And the Lo columns stores the last 16 bytes.
type HiLoColumns struct {
	Hi, Lo ifaces.Column
}

// newHiLoColumns returns a new HiLoColumns with initialized and unconstrained
// columns.
func newHiLoColumns(comp *wizard.CompiledIOP, size int, name string) HiLoColumns {
	return HiLoColumns{
		Hi: comp.InsertCommit(
			0,
			ifaces.ColIDf("STATE_SUMMARY_%v_HI", name),
			size,
		),
		Lo: comp.InsertCommit(
			0,
			ifaces.ColIDf("STATE_SUMMARY_%v_LO", name),
			size,
		),
	}
}

// hiLoAssignmentBuilder is a convenience structure storing the column builders
// relating to an HiLoColumns.
type hiLoAssignmentBuilder struct {
	hi, lo *common.VectorBuilder
}

// newHiLoAssignmentBuilder returns a fresh [hiLoAssignmentBuilder]
func newHiLoAssignmentBuilder(hiLo HiLoColumns) hiLoAssignmentBuilder {
	return hiLoAssignmentBuilder{
		hi: common.NewVectorBuilder(hiLo.Hi),
		lo: common.NewVectorBuilder(hiLo.Lo),
	}
}

// push pushes a row representing `fb` onto `hl`
func (hl *hiLoAssignmentBuilder) push(fb types.FullBytes32) {
	hl.hi.PushHi(fb)
	hl.lo.PushLo(fb)
}

// pushZeroes pushes a row representing 0 onto `hl`
func (hl *hiLoAssignmentBuilder) pushZeroes() {
	hl.hi.PushZero()
	hl.lo.PushZero()
}

// padAssign pads `hl` with `fb` and assigns the resulting columns into `run`.
func (hl *hiLoAssignmentBuilder) padAssign(run *wizard.ProverRuntime, fb types.FullBytes32) {
	var f field.Element
	f.SetBytes(fb[:16])
	hl.hi.PadAndAssign(run, f)
	f.SetBytes(fb[16:])
	hl.lo.PadAndAssign(run, f)
}

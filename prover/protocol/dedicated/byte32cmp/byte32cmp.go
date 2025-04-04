package byte32cmp

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

func Bytes32Cmp(
	// compiled IOP
	comp *wizard.CompiledIOP,
	// number of limbs and bit per limbs to represent a byte32 value
	numLimbs, bitPerLimbs int,
	// name of the byte32 comparison instance
	name string,
	// ColumnA, ColumnB, allegedly the column with larger and smaller values respectively
	columnA, columnB ifaces.Column,
	// activeRows works as a filter on the checking
	activeRow *symbolic.Expression,
) {
	bcp := BytesCmpCtx{}
	round := wizardutils.MaxRound(columnA, columnB)
	bcp.round = round
	bcp.columnA = columnA
	bcp.columnB = columnB
	bcp.activeRow = activeRow
	// assigns the module
	comp.RegisterProverAction(round, &bytes32CmpAssignProverAction{
		bcp:     bcp,
		columnA: columnA,
		columnB: columnB,
	})
	// We do call the assign function before define to avoid the race condition with
	// the bigrange module
	bcp.Define(comp, numLimbs, bitPerLimbs, name)
}

// bytes32CmpAssignProverAction assigns the BytesCmpCtx module for byte32 comparison.
// It implements the [wizard.ProverAction] interface.
type bytes32CmpAssignProverAction struct {
	bcp     BytesCmpCtx
	columnA ifaces.Column
	columnB ifaces.Column
}

// Run executes the assignment of the BytesCmpCtx module.
func (a *bytes32CmpAssignProverAction) Run(run *wizard.ProverRuntime) {
	colA := a.columnA.GetColAssignment(run)
	colB := a.columnB.GetColAssignment(run)
	a.bcp.assign(run, colA, colB)
}

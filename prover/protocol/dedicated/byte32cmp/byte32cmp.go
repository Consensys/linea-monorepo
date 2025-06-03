package byte32cmp

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

type bytes32CmpProverAction struct {
	bcp *BytesCmpCtx
}

func (a *bytes32CmpProverAction) Run(run *wizard.ProverRuntime) {
	colA := a.bcp.columnA.GetColAssignment(run)
	colB := a.bcp.columnB.GetColAssignment(run)
	a.bcp.assign(run, colA, colB)
}

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
	round := column.MaxRound(columnA, columnB)
	bcp.round = round
	bcp.columnA = columnA
	bcp.columnB = columnB
	bcp.activeRow = activeRow
	comp.RegisterProverAction(round, &bytes32CmpProverAction{
		// Must pass pointer here
		bcp: &bcp,
	})
	// We do call the assign function before define to avoid the race condition with
	// the bigrange module
	bcp.Define(comp, numLimbs, bitPerLimbs, name)
}

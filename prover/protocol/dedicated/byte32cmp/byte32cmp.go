package byte32cmp

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
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
	round := column.MaxRound(columnA, columnB)
	bcp.round = round
	bcp.columnA = columnA
	bcp.columnB = columnB
	bcp.activeRow = activeRow
	// assigns the module
	comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
		colA := columnA.GetColAssignment(run)
		colB := columnB.GetColAssignment(run)
		bcp.assign(run, colA, colB)
	})
	// We do call the assign function before define to avoid the race condition with
	// the bigrange module
	bcp.Define(comp, numLimbs, bitPerLimbs, name)

}

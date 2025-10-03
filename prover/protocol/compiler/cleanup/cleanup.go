package cleanup

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// Simple compilation steps which frees ignored items
func CleanUp[T zk.Element](comp *wizard.CompiledIOP[T]) {
	// Gets the last round of the comp
	lastRound := comp.NumRounds() - 1
	// Get the list of all ignored columns
	colToRemove := comp.Columns.AllKeysIgnored()

	// Register the prover action to remove unrequired data
	comp.RegisterProverAction(lastRound, &CleanupProverAction[T]{
		ColumnsToRemove: colToRemove,
	})
}

// CleanupProverAction is the action to remove ignored columns.
// It implements the [wizard.ProverAction] interface.
type CleanupProverAction[T zk.Element] struct {
	ColumnsToRemove []ifaces.ColID
}

// Run executes the cleanup by removing ignored columns from the runtime.
func (a *CleanupProverAction[T]) Run(run *wizard.ProverRuntime[T]) {
	// Remove all the ignored columns
	for _, col := range a.ColumnsToRemove {
		run.Columns.TryDel(col)
	}
}

package cleanup

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

// Simple compilation steps which frees ignored items
func CleanUp(comp *wizard.CompiledIOP) {
	// Gets the last round of the comp
	lastRound := comp.NumRounds() - 1
	// Get the list of all ignored columns
	colToRemove := comp.Columns.AllKeysIgnored()

	// The prover removes all the "now unrequired data"
	comp.SubProvers.AppendToInner(lastRound, func(run *wizard.ProverRuntime) {
		// Remove all the ignored columns
		for _, col := range colToRemove {
			run.Columns.TryDel(col)
		}
	})

}

package selfrecursionwithmerkle

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

// Apply the self-recursion transformation over a vortex compiled
func SelfRecurse(comp *wizard.CompiledIOP) {
	ctx := NewSelfRecursionCxt(comp)
	ctx.Precomputations()
	// the round-by-round commitment phase is implicit here
	ctx.RowLinearCombinationPhase()
	ctx.ColumnOpeningPhase()
	// Update the self-recursion counter
	comp.SelfRecursionCount++
}

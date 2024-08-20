package selfrecursion

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// Apply the self-recursion transformation over a vortex compiled
func SelfRecurse(comp *wizard.CompiledIOP) {

	logrus.Trace("started self-recursion compiler")
	defer logrus.Trace("finished self-recursion compiler")

	ctx := NewSelfRecursionCxt(comp)
	ctx.Precomputations()
	// the round-by-round commitment phase is implicit here
	ctx.RowLinearCombinationPhase()
	ctx.ColumnOpeningPhase()
	// Update the self-recursion counter
	comp.SelfRecursionCount++
}

package selfrecursion

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// Apply the self-recursion transformation over a vortex compiled
func SelfRecurse(comp *wizard.CompiledIOP) {

	logrus.Trace("started self-recursion compiler")
	defer logrus.Trace("finished self-recursion compiler")
	// ToDo: Split into two sub contex, namely lincombContext and columnOpeningContext
	ctx := NewSelfRecursionCxt(comp)
	// To be shifted to columnOpeningPhase
	ctx.Precomputations()
	// the round-by-round commitment phase is implicit here
	ctx.RowLinearCombinationPhase()
	ctx.ColumnOpeningPhase()
	// Update the self-recursion counter
	comp.SelfRecursionCount++
}

// RecurseOverCustomCtx applies the same compilation steps as [SelfRecurse]
// over a specified vortex compilation context.
func RecurseOverCustomCtx(comp *wizard.CompiledIOP, vortexCtx *vortex.Ctx, prefix string) {
	ctx := NewRecursionCtx(comp, vortexCtx, prefix)
	ctx.Precomputations()
	ctx.RowLinearCombinationPhase()
	ctx.ColumnOpeningPhase()
}

// SelfRecurseLinCombPhaseOnly applies the self-recursion
// compilation steps over a vortex compiled context, but only
// the linear combination phase
func SelfRecurseLinCombPhaseOnly(comp *wizard.CompiledIOP) {
	logrus.Trace("started self-recursion (lincomb phase only) compiler")
	defer logrus.Trace("finished self-recursion (lincomb phase only) compiler")
	ctx := NewSelfRecursionCxt(comp)
	ctx.RowLinearCombinationPhase()
	// Update the self-recursion counter
	comp.SelfRecursionCount++
}

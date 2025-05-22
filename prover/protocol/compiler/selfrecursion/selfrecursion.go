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

// SelfRecursionProximityCheck applies the self-recursion
// compilation steps over a vortex compiled context, but only
// the proximity check phase, e.g., the linear combination
// of the seltected preimages matches with the indices of
// UAlpha
func SelfRecursionProximityCheck(comp *wizard.CompiledIOP) {
	logrus.Trace("started self-recursion (proximity check) compiler")
	defer logrus.Trace("finished self-recursion (proximity check) compiler")
	ctx := NewSelfRecursionCxt(comp)
	// We only need to register I(X) for this step
	ctx.RegistersI()
	//   - Commits to a column containing the selected entries of
	//     the linear combination Uα: `Uα,q`
	//
	//   - Performs the following lookup constraint:
	//     `(q,Uα,q)⊂(I,Uα)`
	ctx.ColSelection()
	// Add the evaluation check
	
	// Update the self-recursion counter
	comp.SelfRecursionCount++
}

// SelfRecursionLinearHashAndMerkle applies the self-recursion
// compilation steps over a vortex compiled context, but only
// the linear hash and merkle tree phase

func SelfRecursionLinearHashAndMerkle(comp *wizard.CompiledIOP) {
	logrus.Trace("started self-recursion (linear hash and merkle) compiler")
	defer logrus.Trace("finished self-recursion (linear hash and merkle) compiler")
	ctx := NewSelfRecursionCxt(comp)
	// We only need to register I(X) for this step
	ctx.RegistersI()
	//   - Commits to a column containing the selected entries of
	//     the linear combination Uα: `Uα,q`
	//
	//   - Performs the following lookup constraint:
	//     `(q,Uα,q)⊂(I,Uα)`
	ctx.ColSelection()
	ctx.linearHashAndMerkle()
	comp.SelfRecursionCount++
}

package selfrecursionwithmerkle

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/accessors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column/verifiercol"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/functionals"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/reedsolomon"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/gnark/frontend"
)

// Linear combination phase,
//
//   - Prover sends the linear combination that he claims as a commitment Uα
//     But it's already provided during the context creation
//   - Test the RS membership using the dedicated wizard below
//   - Build a column from the alleged openings Ys
//   - Check the consistency of CoeffEval(Ys,α) = Interpolate(Uα,x) where x
//     is the opening point
func (ctx *SelfRecursionCtx) RowLinearCombinationPhase() {

	// The reed-solomon check
	reedsolomon.CheckReedSolomon(
		ctx.comp,
		ctx.VortexCtx.BlowUpFactor,
		ctx.Columns.Ualpha)

	// Create the verifier column ys
	ctx.defineYs()

	// And do the consistency check
	ctx.consistencyBetweenYsAndUalpha()
}

// Gather the alleged evaluation proven by vortex into a vector
func (ctx *SelfRecursionCtx) defineYs() {
	ranges := []ifaces.ColID{}
	for _, colIDs := range ctx.VortexCtx.CommitmentsByRounds.Inner() {
		ranges = append(ranges, colIDs...)
	}
	ctx.Columns.Ys = verifiercol.NewFromYs(ctx.comp, ctx.VortexCtx.Query, ranges)
}

// Registers the consistency check between Ys and Ualpha
func (ctx *SelfRecursionCtx) consistencyBetweenYsAndUalpha() {

	// We cannot defer the check to the functional dedicated wizard
	// so, we will instead explictly ask the verifier to evaluate
	// the Ys
	ctx.Accessors.CoeffEvalYsAlpha = ifaces.NewAccessor(
		"SELFRECURSION_COEFFEVAL_YS_ALPHA",
		func(run ifaces.Runtime) field.Element {
			ys := ctx.Columns.Ys.GetColAssignment(run)
			alpha := run.GetRandomCoinField(ctx.Coins.Alpha.Name)
			return smartvectors.EvalCoeff(ys, alpha)
		},
		func(api frontend.API, run ifaces.GnarkRuntime) frontend.Variable {
			ys := ctx.Columns.Ys.GetColAssignmentGnark(run)
			alpha := run.GetRandomCoinField(ctx.Coins.Alpha.Name)
			return poly.EvaluateUnivariateGnark(api, ys, alpha)
		},
		ctx.comp.Coins.Round(ctx.Coins.Alpha.Name),
	)

	// Defer the interpolation of Ualpha to a dedicated wizard
	ctx.Accessors.InterpolateUalphaX = functionals.Interpolation(
		ctx.comp,
		ctx.interpolateUAlphaX(),
		accessors.AccessorFromUnivX(ctx.comp, ctx.VortexCtx.Query),
		ctx.Columns.Ualpha,
	)

	// Assumption they should have the same rounds. Not that it's
	// a catastroph if that's the case but that's odd.
	if ctx.Accessors.CoeffEvalYsAlpha.Round != ctx.Accessors.InterpolateUalphaX.Round {
		panic("inconsistency in the round numbers")
	}

	round := ctx.Accessors.CoeffEvalYsAlpha.Round

	// And let the verifier check that they should be both equal
	ctx.comp.InsertVerifier(
		round,
		func(run *wizard.VerifierRuntime) error {
			uAlphaX := ctx.Accessors.InterpolateUalphaX.GetVal(run)
			ysAlpha := ctx.Accessors.CoeffEvalYsAlpha.GetVal(run)
			if uAlphaX != ysAlpha {
				return fmt.Errorf("ConsistencyBetweenYsAndUalpha did not pass")
			}
			return nil
		},
		func(api frontend.API, run *wizard.WizardVerifierCircuit) {
			uAlphaX := ctx.Accessors.InterpolateUalphaX.GetFrontendVariable(api, run)
			ysAlpha := ctx.Accessors.CoeffEvalYsAlpha.GetFrontendVariable(api, run)
			api.AssertIsEqual(uAlphaX, ysAlpha)
		},
	)
}

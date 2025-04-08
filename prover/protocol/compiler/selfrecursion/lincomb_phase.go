package selfrecursion

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/functionals"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
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
	// Includes the precomputed colIds
	if ctx.VortexCtx.IsCommitToPrecomputed() {
		precompColIds := make([]ifaces.ColID, len(ctx.VortexCtx.Items.Precomputeds.PrecomputedColums))
		for i, col := range ctx.VortexCtx.Items.Precomputeds.PrecomputedColums {
			precompColIds[i] = col.GetColID()
		}
		ranges = append(ranges, precompColIds...)
	}
	for _, colIDs := range ctx.VortexCtx.CommitmentsByRounds.Inner() {
		ranges = append(ranges, colIDs...)
	}
	ctx.Columns.Ys = verifiercol.NewFromYs(ctx.comp, ctx.VortexCtx.Query, ranges)
}

type consistencyBetweenYsAndUalphaVerifierAction struct {
	ctx *SelfRecursionCtx
}

func (a *consistencyBetweenYsAndUalphaVerifierAction) Run(run wizard.Runtime) error {
	ys := a.ctx.Columns.Ys.GetColAssignment(run)
	alpha := run.GetRandomCoinField(a.ctx.Coins.Alpha.Name)
	ysAlpha := smartvectors.EvalCoeff(ys, alpha)
	uAlphaX := a.ctx.Accessors.InterpolateUalphaX.GetVal(run)
	if uAlphaX != ysAlpha {
		return fmt.Errorf("ConsistencyBetweenYsAndUalpha did not pass, ysAlphaX=%v uAlphaX=%v", ysAlpha.String(), uAlphaX.String())
	}
	return nil
}

func (a *consistencyBetweenYsAndUalphaVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	ys := a.ctx.Columns.Ys.GetColAssignmentGnark(run)
	alpha := run.GetRandomCoinField(a.ctx.Coins.Alpha.Name)
	uAlphaX := a.ctx.Accessors.InterpolateUalphaX.GetFrontendVariable(api, run)
	ysAlpha := poly.EvaluateUnivariateGnark(api, ys, alpha)
	api.AssertIsEqual(uAlphaX, ysAlpha)
}

// Registers the consistency check between Ys and Ualpha
func (ctx *SelfRecursionCtx) consistencyBetweenYsAndUalpha() {

	// Defer the interpolation of Ualpha to a dedicated wizard
	ctx.Accessors.InterpolateUalphaX = functionals.Interpolation(
		ctx.comp,
		ctx.interpolateUAlphaX(),
		accessors.NewUnivariateX(ctx.VortexCtx.Query, ctx.comp.QueriesParams.Round(ctx.VortexCtx.Query.QueryID)),
		ctx.Columns.Ualpha,
	)

	round := ctx.Accessors.InterpolateUalphaX.Round()

	// And let the verifier check that they should be both equal
	ctx.comp.RegisterVerifierAction(round, &consistencyBetweenYsAndUalphaVerifierAction{
		ctx: ctx,
	})
}

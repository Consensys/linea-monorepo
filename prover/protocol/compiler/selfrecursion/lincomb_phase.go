package selfrecursion

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/functionals"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// RowLinearCombinationPhase implements Check 1 (Reed-Solomon) and Check 2
// (Statement) of the Vortex verifier.
//
// Check 1 — Reed-Solomon check (Schwartz-Zippel):
//
//	UalphaCoeff (T elements) committed by prover. Prover hints
//	UalphaEvals = FFT(UalphaCoeff) (N elements). Verifier checks:
//	    CanonicalEval(UalphaCoeff, β) == LagrangeEval(UalphaEvals, β).
//	After this step UalphaEvals (N) is set and available for the
//	inclusion lookup (Q, UalphaQ) ⊂ (I, UalphaEvals) in ColSelection.
//
// Check 2 — Statement check:
//
//	Compute Ualpha(x) = CanonicalEval(UalphaCoeff, x)   [degree T-1, cheaper].
//	Verify  Ualpha(x) == CanonicalEval(Ys, α)           [evaluations at alpha].
func (ctx *SelfRecursionCtx) RowLinearCombinationPhase() {

	// Check 1: Reed-Solomon check.
	// UalphaCoeff (T coefficients, committed by prover) → FFT hint → UalphaEvals (N evaluations).
	// Schwartz-Zippel: CanonicalEval(UalphaCoeff,β) == LagrangeEval(UalphaEvals,β).
	ctx.Columns.UalphaEvals = reedsolomon.CheckReedSolomon(
		ctx.Comp,
		ctx.VortexCtx.BlowUpFactor,
		ctx.Columns.UalphaCoeff)

	// Create the verifier column ys
	ctx.defineYs()

	// And do the consistency check
	ctx.consistencyBetweenYsAndUalpha()
}

// Gather the alleged evaluation proven by vortex into a vector
func (ctx *SelfRecursionCtx) defineYs() {
	var (
		rangesNonSis = []ifaces.ColID{}
		rangesSis    = []ifaces.ColID{}
	)
	// Includes the precomputed colIds
	if ctx.VortexCtx.IsNonEmptyPrecomputed() {
		precompColIds := make([]ifaces.ColID, len(ctx.VortexCtx.Items.Precomputeds.PrecomputedColums))
		for i, col := range ctx.VortexCtx.Items.Precomputeds.PrecomputedColums {
			precompColIds[i] = col.GetColID()
		}
		// If SIS is applied to precomputed, we need to add the precomputed
		// columns to the rangesSis, otherwise we add them to rangesNonSis
		if ctx.VortexCtx.IsSISAppliedToPrecomputed() {
			rangesSis = append(rangesSis, precompColIds...)
		} else {
			rangesNonSis = append(rangesNonSis, precompColIds...)
		}
	}
	// Collect the SIS round commitments
	for _, colIDs := range ctx.VortexCtx.CommitmentsByRoundsSIS.GetInner() {
		rangesSis = append(rangesSis, colIDs...)
	}
	// Collect the non-SIS round commitments
	for _, colIDs := range ctx.VortexCtx.CommitmentsByRoundsNonSIS.GetInner() {
		rangesNonSis = append(rangesNonSis, colIDs...)
	}
	// append the ranges
	ranges := append(rangesNonSis, rangesSis...)
	ctx.Columns.Ys = verifiercol.NewFromYs(ctx.Comp, ctx.VortexCtx.Query, ranges)
}

type ConsistencyYsUalphaVerifierAction struct {
	Ctx                *SelfRecursionCtx
	InterpolateUalphaX ifaces.Accessor
}

func (a *ConsistencyYsUalphaVerifierAction) Run(run wizard.Runtime) error {
	ys := a.Ctx.Columns.Ys.GetColAssignment(run)
	alpha := run.GetRandomCoinFieldExt(a.Ctx.Coins.Alpha.Name)
	ysAlpha := smartvectors.EvalCoeffExt(ys, alpha)
	uAlphaX := a.InterpolateUalphaX.GetValExt(run)
	if uAlphaX != ysAlpha {
		return fmt.Errorf("ConsistencyBetweenYsAndUalpha did not pass, ysAlphaX=%v uAlphaX=%v", ysAlpha.String(), uAlphaX.String())
	}
	return nil
}

func (a *ConsistencyYsUalphaVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	koalaAPI := koalagnark.NewAPI(api)

	ys := a.Ctx.Columns.Ys.GetColAssignmentGnarkExt(run)
	alpha := run.GetRandomCoinFieldExt(a.Ctx.Coins.Alpha.Name)
	uAlphaX := a.InterpolateUalphaX.GetFrontendVariableExt(api, run)
	ysAlpha := poly.EvaluateUnivariateGnarkExt(api, ys, alpha)
	koalaAPI.AssertIsEqualExt(uAlphaX, ysAlpha)
}

// Registers the consistency check between Ys and Ualpha
func (ctx *SelfRecursionCtx) consistencyBetweenYsAndUalpha() {

	xAccessor := accessors.NewUnivariateX(ctx.VortexCtx.Query, ctx.Comp.QueriesParams.Round(ctx.VortexCtx.Query.QueryID))

	// UalphaCoeff holds T polynomial coefficients.
	// CanonicalEval(UalphaCoeff, x) is a degree-(T-1) Horner evaluation —
	// cheaper than LagrangeEval(UalphaEvals, x) which is degree-(N-1).
	pa, res := functionals.CoeffEvalNoRegisterPA(
		ctx.Comp,
		ctx.interpolateUAlphaX(),
		xAccessor,
		ctx.Columns.UalphaCoeff,
	)
	if pa != nil {
		ctx.Comp.RegisterProverAction(pa.Round, pa)
	}
	ctx.Accessors.InterpolateUalphaX = res

	round := ctx.Accessors.InterpolateUalphaX.Round()

	// And let the verifier check that they should be both equal
	ctx.Comp.RegisterVerifierAction(round, &ConsistencyYsUalphaVerifierAction{
		Ctx:                ctx,
		InterpolateUalphaX: ctx.Accessors.InterpolateUalphaX,
	})
}

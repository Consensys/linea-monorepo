package selfrecursion

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/expr_handle"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/functionals"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Specifies the column opening phase
func (ctx *SelfRecursionCtx) ColumnOpeningPhase() {
	// Registers the limb expanded version of the preimages
	ctx.ColSelection()
	ctx.LinearHashAndMerkle()
	ctx.RootHashGlue()
	ctx.GluePositions()
	// We need this only when there are non zero number
	// of SIS rounds
	if ctx.Columns.ConcatenatedDhQ != nil {
		ctx.RegistersSisPreimageLimbs()
	}
	ctx.CollapsingPhase()
	// The fold phase is only needed if there are non-zero
	// number of SIS rounds
	if ctx.Columns.ConcatenatedDhQ != nil {
		ctx.FoldPhase()
	}
}

// Registers the preimage limbs for the SIS rounds.
//
// Get the preimages (preimage0,preimage1,…,preimaget−1) And range-check
// each of them on the ring-SIS bound
func (ctx *SelfRecursionCtx) RegistersSisPreimageLimbs() {
	wholes := ctx.Columns.WholePreimagesSis
	sisParams := ctx.VortexCtx.SisParams

	limbs := make([]ifaces.Column, len(wholes))
	round := wholes[0].Round()
	limbSize := wholes[0].Size() * sisParams.NumLimbs()

	for i := range limbs {
		limbs[i] = ctx.Comp.InsertCommit(
			round,
			ctx.limbExpandedPreimageName(wholes[i].GetColID()),
			limbSize,
		)
		ctx.Comp.InsertRange(
			round,
			ifaces.QueryIDf("SHORTNESS_%v", limbs[i].GetColID()),
			limbs[i],
			1<<ctx.VortexCtx.SisParams.LogTwoBound,
		)
	}

	ctx.Columns.PreimagesSis = limbs

	ctx.Comp.RegisterProverAction(round, &PreimageLimbsProverAction{
		Ctx:   ctx,
		Limbs: limbs,
	})

}

type PreimageLimbsProverAction struct {
	Ctx   *SelfRecursionCtx
	Limbs []ifaces.Column
}

func (a *PreimageLimbsProverAction) Run(run *wizard.ProverRuntime) {
	parallel.Execute(len(a.Limbs), func(start, end int) {
		for i := start; i < end; i++ {
			whole := a.Ctx.Columns.WholePreimagesSis[i].GetColAssignment(run)
			whole_ := smartvectors.IntoRegVec(whole)
			expanded_ := a.Ctx.SisKey().LimbSplit(whole_)
			expanded := smartvectors.NewRegular(expanded_)
			run.AssignColumn(a.Limbs[i].GetColID(), expanded)
		}
	})
}

type ColSelectionProverAction struct {
	Ctx       *SelfRecursionCtx
	UAlphaQID ifaces.ColID
}

func (a *ColSelectionProverAction) Run(run *wizard.ProverRuntime) {
	q := run.GetRandomCoinIntegerVec(a.Ctx.Coins.Q.Name)
	uAlpha := smartvectors.IntoRegVec(run.GetColumn(a.Ctx.Columns.Ualpha.GetColID()))

	uAlphaQ := make([]field.Element, 0, a.Ctx.Columns.UalphaQ.Size())
	for _, qi := range q {
		uAlphaQ = append(uAlphaQ, uAlpha[qi])
	}

	run.AssignColumn(a.UAlphaQID, smartvectors.NewRegular(uAlphaQ))
}

// Declare the queries justifying the column selection:
//
//   - Build a public column from the selected entries
//     `q=(q0,q1,…,qt−1)`
//
//   - Commits to a column containing the selected entries of
//     the linear combination: `Uα,q`
//
//   - Performs the following lookup constraint:
//     `(q,Uα,q)⊂(I,Uα)`
func (ctx *SelfRecursionCtx) ColSelection() {

	// Build the column q, (auto-assigned)
	ctx.Columns.Q = verifiercol.NewFromIntVecCoin(ctx.Comp, ctx.Coins.Q)

	// Declaration round of the coin Q
	roundQ := ctx.Columns.Q.Round()

	ctx.Columns.UalphaQ = ctx.Comp.InsertCommit(
		roundQ,
		ctx.uAlphaQName(),
		ctx.Coins.Q.Size,
	)

	// And registers the assignment function
	ctx.Comp.RegisterProverAction(roundQ, &ColSelectionProverAction{
		Ctx:       ctx,
		UAlphaQID: ctx.Columns.UalphaQ.GetColID(),
	})

	// Declare an inclusion query to finalize the selection check
	ctx.Comp.InsertInclusion(
		roundQ,
		ctx.selectQInclusion(),
		[]ifaces.Column{
			ctx.Columns.I,
			ctx.Columns.Ualpha,
		},
		[]ifaces.Column{
			ctx.Columns.Q,
			ctx.Columns.UalphaQ,
		},
	)
}

type CollapsingProverAction struct {
	Ctx     *SelfRecursionCtx
	EDualID ifaces.ColID
	SisKey  *ringsis.Key
}

func (a *CollapsingProverAction) Run(run *wizard.ProverRuntime) {
	collapsedPreimage := a.Ctx.Columns.PreimagesSisCollapse.GetColAssignment(run)
	sisKey := a.SisKey

	subDuals := []smartvectors.SmartVector{}
	roundStartAt := 0

	// We add precomputed columns in the calculation only if
	// sis is applied to the precomputed columns
	if a.Ctx.VortexCtx.IsNonEmptyPrecomputed() && a.Ctx.VortexCtx.IsSISAppliedToPrecomputed() {
		numPrecomputeds := len(a.Ctx.VortexCtx.Items.Precomputeds.PrecomputedColums)
		if numPrecomputeds == 0 {
			utils.Panic("The number of precomputeds must be non-zero!")
		}
		preimageSlice := collapsedPreimage.SubVector(
			roundStartAt*sisKey.NumLimbs(),
			(roundStartAt + numPrecomputeds*sisKey.NumLimbs()),
		)
		subDual := sisKey.HashModXnMinus1(smartvectors.IntoRegVec(preimageSlice))
		subDuals = append(subDuals, smartvectors.NewRegular(subDual))
		roundStartAt += numPrecomputeds
	}

	for _, comsInRoundI := range a.Ctx.VortexCtx.CommitmentsByRoundsSIS.GetInner() {
		if len(comsInRoundI) == 0 {
			continue
		}
		preimageSlice := collapsedPreimage.SubVector(
			roundStartAt*sisKey.NumLimbs(),
			(roundStartAt+len(comsInRoundI))*sisKey.NumLimbs(),
		)
		subDual := sisKey.HashModXnMinus1(smartvectors.IntoRegVec(preimageSlice))
		subDuals = append(subDuals, smartvectors.NewRegular(subDual))
		roundStartAt += len(comsInRoundI)
	}

	colPowT := accessors.NewExponent(a.Ctx.Coins.Collapse, a.Ctx.VortexCtx.NbColsToOpen()).GetVal(run)
	eDual := smartvectors.PolyEval(subDuals, colPowT)

	run.AssignColumn(a.EDualID, eDual)

}

type CollapsingVerifierAction struct {
	UAlphaQEval  ifaces.Accessor
	PreImageEval ifaces.Accessor
}

func (a *CollapsingVerifierAction) Run(run wizard.Runtime) error {
	if a.UAlphaQEval.GetVal(run) != a.PreImageEval.GetVal(run) {
		l, r := a.UAlphaQEval.GetVal(run), a.PreImageEval.GetVal(run)
		return fmt.Errorf("consistency between u_alpha and the preimage: mismatch between uAlphaQEval=%v preimages=%v",
			l.String(), r.String())
	}
	return nil
}

func (a *CollapsingVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	api.AssertIsEqual(
		a.UAlphaQEval.GetFrontendVariable(api, run),
		a.PreImageEval.GetFrontendVariable(api, run),
	)
}

/*
Collapsing phase

- Sample the random coin "collapse"
- Collapse all the SIS hashes into a single hash
  - DmergeQCollapse := Fold(DmergeQ, collapse)

- Collapse all the SIS keys into a single sis key
  - Acollapse := \sum_{k} A_{k} (collapse^{t})^k

- Collapse all the preimages
  - PreimageCollapse := \sum_{i} P_{i} (collapse)^i
*/
func (ctx *SelfRecursionCtx) CollapsingPhase() {

	// starting one round after Q is sampled
	round := ctx.Columns.Q.Round() + 1

	// Sampling of r_collapse
	ctx.Coins.Collapse = ctx.Comp.InsertCoin(round, ctx.collapseCoin(), coin.Field)

	// Declare the linear combination of the preimages by collapse coin
	// aka, the collapsed preimage
	// We need this only if there are
	// non-zero number of SIS rounds
	if ctx.Columns.ConcatenatedDhQ != nil {
		ctx.Columns.PreimagesSisCollapse = expr_handle.RandLinCombCol(
			ctx.Comp,
			accessors.NewFromCoin(ctx.Coins.Collapse),
			ctx.Columns.PreimagesSis,
		)
	}

	// Consistency check between the collapsed preimage and UalphaQ
	{
		uAlphaQEval := functionals.CoeffEval(
			ctx.Comp,
			ctx.constencyUalphaQPreimageLeft(),
			ctx.Coins.Collapse,
			ctx.Columns.UalphaQ,
		)

		// The collapsed preimage for the non SIS rounds
		// Computed only if there are non zero number of non SIS rounds
		var (
			// The random linear combination of the collapsed preimage
			// for the non SIS rounds
			preImageNonSisEval ifaces.Accessor
			preImageSisEval    ifaces.Accessor
			// offset denotes the running power of alpha in the BivariateEval for the
			// sis round preimages
			offset       = 0
			preImageEval ifaces.Accessor
		)
		if len(ctx.MIMCMetaData.ToHashSizes) > 0 {
			ctx.Columns.CollapsedPreimagesNonSis = expr_handle.RandLinCombCol(
				ctx.Comp,
				accessors.NewFromCoin(ctx.Coins.Collapse),
				ctx.Columns.WholePreimagesNonSis,
			)

			preImageNonSisEval = functionals.CoeffEval(
				ctx.Comp,
				ctx.constencyUalphaQPreimageRight(),
				ctx.Coins.Alpha,
				ctx.Columns.CollapsedPreimagesNonSis,
			)

			for i := range ctx.MIMCMetaData.ToHashSizes {
				// We add the number of polynomials per non SIS round
				// to the offset
				offset += ctx.MIMCMetaData.ToHashSizes[i]
			}
		}
		/*
			- The preimages are given in the form of several columns of the form:
			  [limb0, limb1, ..., limbL-1, limb0, ... limbL-1, ...], each field element is represented by
			  L limbs (say)
			- Next, we collapse the limbs
			- First, we observe that "\sum_{i=0}^{numb_limbs} limb_i 2^(log_two_bound * i) = jth field element
			  of the preimage" and that this corresponds to evaluating a polynomial whose coefficients are
			  [limb0, ... limbL-1], numb_limbs is the number of limbs required to represent a field element
			- Then making the double sum for all elements in an opened column using alpha as a second
			  evaluation point, we get a bivariate polynomial evaluation
		*/
		// We only compute the preImageSisEval if there are any SIS rounds
		// and the concatenated DhQ is not nil
		if ctx.Columns.ConcatenatedDhQ != nil {
			preImageSisEval = functionals.EvalCoeffBivariate(
				ctx.Comp,
				ctx.constencyUalphaQPreimageRight(),
				ctx.Columns.PreimagesSisCollapse,
				accessors.NewConstant(field.NewElement(1<<ctx.SisKey().LogTwoBound)),
				accessors.NewFromCoin(ctx.Coins.Alpha),
				ctx.VortexCtx.SisParams.NumLimbs(),
				ctx.Columns.WholePreimagesSis[0].Size(),
			)
		}

		// preImageEval := preimageNonSisEval + alpha^offset * preImageSisEval
		if len(ctx.MIMCMetaData.ToHashSizes) > 0 && ctx.Columns.ConcatenatedDhQ != nil {
			preImageEvalSymb := symbolic.Add(
				preImageNonSisEval,
				symbolic.Mul(
					symbolic.Pow(
						accessors.NewFromCoin(ctx.Coins.Alpha),
						offset,
					),
					preImageSisEval,
				),
			)
			preImageEval = accessors.NewFromExpression(preImageEvalSymb, fmt.Sprintf("PREIMAGE_EVAL_%v", ctx.SelfRecursionCnt))
		} else if len(ctx.MIMCMetaData.ToHashSizes) > 0 && ctx.Columns.ConcatenatedDhQ == nil {
			preImageEval = preImageNonSisEval
		} else if len(ctx.MIMCMetaData.ToHashSizes) == 0 && ctx.Columns.ConcatenatedDhQ != nil {
			preImageEval = preImageSisEval
		} else {
			utils.Panic("There are neither SIS nor non SIS round, this should not happen")
		}

		ctx.Comp.RegisterVerifierAction(uAlphaQEval.Round(), &CollapsingVerifierAction{
			UAlphaQEval:  uAlphaQEval,
			PreImageEval: preImageEval,
		})
	}

	// The below code is only executed only if there are non-zero SIS rounds
	if ctx.Columns.ConcatenatedDhQ != nil {
		sisDeg := ctx.VortexCtx.SisParams.OutputSize()
		// Currently, only powers of two SIS degree are allowed
		// (in practice, we restrict ourselves to pure power of two)
		// lattices instances.
		if !utils.IsPowerOfTwo(sisDeg) {
			utils.Panic("Attempting to fold to a non-power of two size : %v", sisDeg)
		}

		// Compute the collapsed hashes
		ctx.Columns.DhQCollapse = functionals.FoldOuter(
			ctx.Comp,
			ctx.Columns.ConcatenatedDhQ,
			accessors.NewFromCoin(ctx.Coins.Collapse),
			ctx.Columns.ConcatenatedDhQ.Size()/sisDeg,
		)

		// sanity-check : the size of DhQCollapse must equal to sisDeg
		if ctx.Columns.DhQCollapse.Size() != sisDeg {
			utils.Panic("the size of DhQ (%v) collapse must equal to the SIS modulus degree (%v)", ctx.Columns.DhQCollapse.Size(), sisDeg)
		}

		//
		// Merging the SIS keys
		//

		// Create an accessor for collapse^t, where t is the number of opened columns
		collapsePowT := accessors.NewExponent(ctx.Coins.Collapse, ctx.VortexCtx.NbColsToOpen())

		// ToDo(arijit): We may not need this any more, after the optional SIS hash feature
		// since some of the Ah and Dh can be nil, we compactify the slice by
		// only retaining the non-nil elements before sending it to the
		// linear combination operator.
		nonNilAh := []ifaces.Column{}
		for _, ah := range ctx.Columns.Ah {
			if ah != nil {
				nonNilAh = append(nonNilAh, ah)
			} else {
				utils.Panic("The Ah column is nil, this should not happen")
			}
		}

		// If there is only a single SIS round, then there is no need for
		// computing an additional linear-combination column. We can simply
		// return the single column. If we were using the random linear
		// combination in this case, then the randomness would be factored out
		// of the linear combination expression and the resulting column would
		// be associated with the round "0" and not as a precomputed column.
		// This would introduces an entire round for just one column and has
		// noticeable effects on the performance of the protocol.
		if len(nonNilAh) > 1 {
			// And computes the linear combination
			ctx.Columns.ACollapsed = expr_handle.RandLinCombCol(
				ctx.Comp,
				collapsePowT,
				nonNilAh,
				ctx.aCollapsedName(),
			)
		}

		if len(nonNilAh) == 1 {
			ctx.Columns.ACollapsed = nonNilAh[0]
		}

		// Declare Edual
		ctx.Columns.Edual = ctx.Comp.InsertCommit(
			round, ctx.eDual(), ctx.VortexCtx.SisParams.OutputSize(),
		)

		// And assign it
		ctx.Comp.RegisterProverAction(round, &CollapsingProverAction{
			Ctx:     ctx,
			EDualID: ctx.Columns.Edual.GetColID(),
			SisKey:  ctx.SisKey(),
		})
	}
}

type FoldPhaseProverAction struct {
	Ctx       *SelfRecursionCtx
	IpQueryID ifaces.QueryID // Changed to ifaces.QueryID explicitly
}

func (a *FoldPhaseProverAction) Run(run *wizard.ProverRuntime) {
	foldedKey := a.Ctx.Columns.ACollapseFold.GetColAssignment(run)
	foldedPreimage := a.Ctx.Columns.PreimageCollapseFold.GetColAssignment(run)
	y := smartvectors.InnerProduct(foldedKey, foldedPreimage)
	run.AssignInnerProduct(a.IpQueryID, y)
}

type FoldPhaseVerifierAction struct {
	Ctx       *SelfRecursionCtx
	IpQueryID ifaces.QueryID
	Degree    int
}

func (a *FoldPhaseVerifierAction) Run(run wizard.Runtime) error {
	edual := a.Ctx.Columns.Edual.GetColAssignment(run)
	dcollapse := a.Ctx.Columns.DhQCollapse.GetColAssignment(run)
	rfold := run.GetRandomCoinField(a.Ctx.Coins.Fold.Name)
	yAlleged := run.GetInnerProductParams(a.IpQueryID).Ys[0]
	yDual := smartvectors.EvalCoeff(edual, rfold)
	yActual := smartvectors.EvalCoeff(dcollapse, rfold)

	var xN, xNminus1, xNplus1 field.Element
	one := field.One()
	xN.Exp(rfold, big.NewInt(int64(a.Degree)))
	xNminus1.Sub(&xN, &one)
	xNplus1.Add(&xN, &one)

	var left, left0, left1, right field.Element
	left0.Mul(&xNplus1, &yDual)
	left1.Mul(&xNminus1, &yActual)
	left.Sub(&left0, &left1)
	right.Double(&yAlleged)

	if left != right {
		return fmt.Errorf("failed the consistency check of the ring-SIS : %v != %v", left.String(), right.String())
	}
	return nil
}

func (a *FoldPhaseVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	edual := a.Ctx.Columns.Edual.GetColAssignmentGnark(run)
	dcollapse := a.Ctx.Columns.DhQCollapse.GetColAssignmentGnark(run)
	rfold := run.GetRandomCoinField(a.Ctx.Coins.Fold.Name)
	yAlleged := run.GetInnerProductParams(a.IpQueryID).Ys[0]
	yDual := poly.EvaluateUnivariateGnark(api, edual, rfold)
	yActual := poly.EvaluateUnivariateGnark(api, dcollapse, rfold)

	one := field.One()
	xN := gnarkutil.Exp(api, rfold, a.Degree)
	xNminus1 := api.Sub(xN, one)
	xNplus1 := api.Add(xN, one)

	left0 := api.Mul(xNplus1, yDual)
	left1 := api.Mul(xNminus1, yActual)
	left := api.Sub(left0, left1)
	right := api.Mul(yAlleged, 2)

	api.AssertIsEqual(left, right)
}

// Registers the final folding phase of the self-recursion
//
//   - Sample the folding random coin r\fold
//
//   - Fold A\merge by rFold to obtain ACollapsed
//
//   - Fold PreimageCollapse by rFold to obtain PreimageCollapseFold
//
//   - Declare and assign the inner-product between PreimageCollapseFold
//     and ACollapsed
//
//   - Perform the final check to evaluate the consistency vs
//     Edual and D\merge,\collapse,q

func (ctx *SelfRecursionCtx) FoldPhase() {

	// The round of declaration should be one more than EDual
	round := ctx.Columns.Edual.Round() + 1

	// Sample rFold
	ctx.Coins.Fold = ctx.Comp.InsertCoin(round, ctx.foldCoinName(), coin.Field)

	// Constructs ACollapsedFold
	ctx.Columns.ACollapseFold = functionals.Fold(
		ctx.Comp, ctx.Columns.ACollapsed,
		accessors.NewFromCoin(ctx.Coins.Fold),
		ctx.VortexCtx.SisParams.OutputSize(),
	)

	// Construct DmergeCollapseFold
	ctx.Columns.PreimageCollapseFold = functionals.Fold(
		ctx.Comp, ctx.Columns.PreimagesSisCollapse,
		accessors.NewFromCoin(ctx.Coins.Fold),
		ctx.VortexCtx.SisParams.OutputSize(),
	)

	// Mark Edual and the DmergeQCollapse fold as proof
	ctx.Comp.Columns.SetStatus(ctx.Columns.DhQCollapse.GetColID(), column.Proof)
	ctx.Comp.Columns.SetStatus(ctx.Columns.Edual.GetColID(), column.Proof)

	// Declare and assign the inner-product
	ctx.Queries.LatticeInnerProd = ctx.Comp.InsertInnerProduct(
		round, ctx.preimagesAndAmergeIP(), ctx.Columns.ACollapseFold,
		[]ifaces.Column{ctx.Columns.PreimageCollapseFold})

	// Assignment part of the inner product
	ctx.Comp.RegisterProverAction(round, &FoldPhaseProverAction{
		Ctx:       ctx,
		IpQueryID: ctx.Queries.LatticeInnerProd.Name(),
	})

	degree := ctx.SisKey().OutputSize()

	// 		If P(X) is of degree 2n

	//	And
	//		- Q(X) = P(X) mod X^n - 1
	//		- R(X) = P(X) mod X^n + 1

	//	Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
	//	Here, we can identify at the point x

	//	yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
	//	We perform this check in the verifier action below.

	ctx.Comp.RegisterVerifierAction(round, &FoldPhaseVerifierAction{
		Ctx:       ctx,
		IpQueryID: ctx.Queries.LatticeInnerProd.Name(),
		Degree:    degree,
	})
}

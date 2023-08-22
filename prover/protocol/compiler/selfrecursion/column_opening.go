package selfrecursion

import (
	"fmt"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/accessors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column/verifiercol"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/expr_handle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/functionals"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

// Specifies the column opening phase
func (ctx *SelfRecursionCtx) ColumnOpeningPhase() {
	// Registers the limb expanded version of the preimages
	ctx.registerMerging()
	ctx.colSelection()
	ctx.registersPreimageLimbs()
	ctx.collapsingPhase()
	ctx.foldPhase()
}

// Registers the preimage limbs
//
// Get the preimages (preimage0,preimage1,…,preimaget−1)And range-check
// each of them on the ring-SIS bound
func (ctx *SelfRecursionCtx) registersPreimageLimbs() {
	wholes := ctx.Columns.WholePreimages
	sisParams := ctx.VortexCtx.SisParams

	// Insert the limb expanded version of the preimages
	limbs := make([]ifaces.Column, len(wholes))
	round := wholes[0].Round()
	limbSize := wholes[0].Size() * sisParams.NumLimbs()

	for i := range limbs {

		// Registers the new column
		limbs[i] = ctx.comp.InsertCommit(
			round,
			ctx.limbExpandedPreimageName(wholes[i].GetColID()),
			limbSize,
		)

		// And also registers a range checks on each
		ctx.comp.InsertRange(
			round,
			ifaces.QueryIDf("SHORTNESS_%v", limbs[i].GetColID()),
			limbs[i],
			1<<ctx.VortexCtx.SisParams.LogTwoBound_,
		)
	}

	// Registers them in the context as well
	ctx.Columns.Preimages = limbs

	// Also assign them immediately
	ctx.comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
		parallel.Execute(len(limbs), func(start, end int) {
			for i := start; i < end; i++ {
				whole := ctx.Columns.WholePreimages[i].GetColAssignment(run)
				whole_ := smartvectors.IntoRegVec(whole)
				expanded_ := ctx.SisKey().LimbSplit(whole_)
				expanded := smartvectors.NewRegular(expanded_)
				run.AssignColumn(limbs[i].GetColID(), expanded)
			}
		})

	})
}

// Registers the merging coin and compute Amerge and Dmerge
//
//   - Sample the Merge coin : rmerge ← FCompute the
//     linear combination of the digests
//     `Dmerge = ∑h rmerge^h Dh`
//
//   - Compute the linear combination of the key shard
//     `Amerge = ∑h rmerge^h Ah`
func (ctx *SelfRecursionCtx) registerMerging() {

	// Definition round for the merge coin. At the same time as Alpha
	round := ctx.comp.Coins.Round(ctx.Coins.Alpha.Name)

	ctx.Coins.Merge = ctx.comp.InsertCoin(round, ctx.mergeCoinName(), coin.Field)

	// Assumption : there are as many Dh as there Ah and there are both
	// nil/non-nil at the same positions
	if len(ctx.Columns.Ah) != len(ctx.Columns.Dh) {

		for i := 0; i < utils.Max(len(ctx.Columns.Ah), len(ctx.Columns.Dh)); i++ {

			dHisNil := true
			aHiSNil := true

			if i < len(ctx.Columns.Dh) {
				dHisNil = ctx.Columns.Dh[i] == nil
			}

			if i < len(ctx.Columns.Ah) {
				aHiSNil = ctx.Columns.Ah[i] == nil
			}

			logrus.Errorf("SELFRECURSION : nilness of Ah %v vs Dh %v\n", aHiSNil, dHisNil)
		}

		utils.Panic(
			"There are %v Ahs but %v Dhs",
			len(ctx.Columns.Ah), len(ctx.Columns.Dh),
		)
	}

	// since some of the Ah and Dh can be nil, we compactify the slice by
	// only retaining the non-nil elements before sending it to the
	// linear combination operator.
	nonNilAh := []ifaces.Column{}
	nonNilDh := []ifaces.Column{}
	for i, ah := range ctx.Columns.Ah {

		if (ah == nil) != (ctx.Columns.Dh[i] == nil) {
			utils.Panic("for round %v, ah and dh should be either both nil or non-nil", i)
		}

		if ah != nil {
			nonNilAh = append(nonNilAh, ah)
			nonNilDh = append(nonNilDh, ctx.Columns.Dh[i])
		}
	}

	// And declare the random linear combinations columns aMerge and dMerge
	// since it is using expr_handle, this is already auto-assigned
	ctx.Columns.Amerge = expr_handle.RandLinCombCol(
		ctx.comp,
		accessors.AccessorFromCoin(ctx.Coins.Merge),
		nonNilAh,
		ctx.aMergeColName(),
	)

	ctx.Columns.Dmerge = expr_handle.RandLinCombCol(
		ctx.comp,
		accessors.AccessorFromCoin(ctx.Coins.Merge),
		nonNilDh,
		ctx.dMergeColName(),
	)
}

// Declare the queries justifying the column selection:
//
//   - Build a public column from the selected entries
//     `q=(q0,q1,…,qt−1)`
//
//   - Commits to a column containing the selected hashes:
//     `D\merge,q`
//
//   - Commits to a column containing the selected entries of
//     the linear combination: `Uα,q`
//
//   - Sample a random coin `r\squash ← F`
//
//   - Computes the foldings:
//     `D\merge,\squash,q = Fold(D\merge,q,r\squash)` and
//     `D\merge,\squash=Fold(D\merge,r\squash)`
//
//   - Performs the following lookup constraint:
//     `(q,D\merge,\squash,q,Uα,q)⊂(I,D\merge,\squash,Uα)`
func (ctx *SelfRecursionCtx) colSelection() {

	// Build the column q, (auto-assigned)
	ctx.Columns.Q = verifiercol.NewFromIntVecCoin(ctx.comp, ctx.Coins.Q)

	// Size of a SIS hash
	sisHashSize := ctx.VortexCtx.SisParams.OutputSize()
	// Declaration round of the coin Q
	roundQ := ctx.Columns.Q.Round()

	// Declare an assign `DmergeQ` and `UalphaQ`
	ctx.Columns.DmergeQ = ctx.comp.InsertCommit(
		roundQ,
		ctx.dMergeQName(),
		ctx.Coins.Q.Size*sisHashSize,
	)

	ctx.Columns.UalphaQ = ctx.comp.InsertCommit(
		roundQ,
		ctx.uAlphaQName(),
		ctx.Coins.Q.Size,
	)

	// And registers the assignment function
	ctx.comp.SubProvers.AppendToInner(
		roundQ,
		func(run *wizard.ProverRuntime) {

			// Load the already assigned columns
			q := run.GetRandomCoinIntegerVec(ctx.Coins.Q.Name)
			uAlpha := smartvectors.IntoRegVec(
				run.GetColumn(ctx.Columns.Ualpha.GetColID()),
			)
			dMerge := smartvectors.IntoRegVec(
				run.GetColumn(ctx.Columns.Dmerge.GetColID()),
			)

			// And select the columns
			uAlphaQ := make([]field.Element, 0, ctx.Columns.UalphaQ.Size())
			dMergeQ := make([]field.Element, 0, ctx.Columns.DmergeQ.Size())

			for _, qi := range q {
				uAlphaQ = append(uAlphaQ, uAlpha[qi])
				dMergeQ = append(dMergeQ, dMerge[qi*sisHashSize:(qi+1)*sisHashSize]...)
			}

			run.AssignColumn(
				ctx.Columns.UalphaQ.GetColID(),
				smartvectors.NewRegular(uAlphaQ),
			)

			run.AssignColumn(
				ctx.Columns.DmergeQ.GetColID(),
				smartvectors.NewRegular(dMergeQ),
			)
		},
	)

	// Now, we need to fold Dmerge and DmergeQ in order to include in the same
	// lookup as the other variables to ensure the column selection was perfor-
	// med truthfully.
	ctx.Coins.Squash = ctx.comp.InsertCoin(roundQ+1, ctx.squashCoin(), coin.Field)

	// Fold Dmerge and DmergeQ over the squash randomness
	ctx.Columns.DmergeQsquash = functionals.Fold(
		ctx.comp,
		ctx.Columns.DmergeQ,
		accessors.AccessorFromCoin(ctx.Coins.Squash),
		sisHashSize,
	)

	ctx.Columns.DmergeSquash = functionals.Fold(
		ctx.comp,
		ctx.Columns.Dmerge,
		accessors.AccessorFromCoin(ctx.Coins.Squash),
		sisHashSize,
	)

	// Declare an inclusion query to finalize the selection check
	ctx.comp.InsertInclusion(
		roundQ+1,
		ctx.selectQInclusion(),
		[]ifaces.Column{
			ctx.Columns.I,
			ctx.Columns.Ualpha,
			ctx.Columns.DmergeSquash,
		},
		[]ifaces.Column{
			ctx.Columns.Q,
			ctx.Columns.UalphaQ,
			ctx.Columns.DmergeQsquash,
		},
	)
}

// Func collapsing phase
//
//   - Sample the Collapse coin : `r\collapse ← F`
//
//   - Compute the collapsed preimage:
//     `preimage_\collapse = ∑i preimagei*r_\collapse^i`
//
//   - Consistency with the row lincheck
//     `EvalCoeff(Uα,q,r\collapse)=?=EvalBivariate(preimage\collapse,2,α)`
//
//   - Compute the collapsed hashes
//     `D\merge,\collapse,q = Fold(D\merge,q,r\collapse)`
//
//   - Evaluate the dual sis hash of the merged key represented by
//     `A\merge` and the collapsed preimage preimage (ignoring, non-shortness
//     of the coefficients) to obtain Edual
func (ctx *SelfRecursionCtx) collapsingPhase() {

	// starting round
	round := ctx.Columns.UalphaQ.Round() + 1

	// Sampling of r_collapse
	ctx.Coins.Collapse = ctx.comp.InsertCoin(round, ctx.collapseCoin(), coin.Field)

	// Declare the linear combination of the preimages by collapse coin
	// aka, the collapsed preimage
	ctx.Columns.PreimagesCollapse = expr_handle.RandLinCombCol(
		ctx.comp,
		accessors.AccessorFromCoin(ctx.Coins.Collapse),
		ctx.Columns.Preimages,
	)

	// Consistency check between the collapsed preimage and UalphaQ
	{
		left := functionals.CoeffEval(
			ctx.comp,
			ctx.constencyUalphaQPreimageLeft(),
			ctx.Coins.Collapse,
			ctx.Columns.UalphaQ,
		)

		right := functionals.EvalCoeffBivariate(
			ctx.comp,
			ctx.constencyUalphaQPreimageRight(),
			ctx.Columns.PreimagesCollapse,
			accessors.AccessorFromConstant(field.NewElement(1<<ctx.SisKey().LogTwoBound)),
			accessors.AccessorFromCoin(ctx.Coins.Alpha),
			ctx.VortexCtx.SisParams.NumLimbs(),
			ctx.Columns.WholePreimages[0].Size(),
		)

		ctx.comp.InsertVerifier(
			left.Round,
			func(run *wizard.VerifierRuntime) error {
				if left.GetVal(run) != right.GetVal(run) {
					l, r := left.GetVal(run), right.GetVal(run)
					return fmt.Errorf("consistency between u_alpha and the preimage: "+
						"mismatch between left and right %v != %v",
						l.String(), r.String(),
					)
				}
				return nil
			},
			func(api frontend.API, run *wizard.WizardVerifierCircuit) {
				api.AssertIsEqual(
					left.GetFrontendVariable(api, run),
					right.GetFrontendVariable(api, run),
				)
			},
		)
	}

	// Compute the collapsed hashes
	ctx.Columns.DmergeQcollapse = functionals.FoldOuter(
		ctx.comp,
		ctx.Columns.DmergeQ,
		accessors.AccessorFromCoin(ctx.Coins.Collapse),
		ctx.Coins.Q.Size,
	)

	// Declare Edual
	ctx.Columns.Edual = ctx.comp.InsertCommit(
		round, ctx.eDual(), ctx.VortexCtx.SisParams.OutputSize(),
	)

	// And assign it
	ctx.comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {

		// Get the collapsed preimage
		collapsedPreimage := ctx.Columns.PreimagesCollapse.GetColAssignment(run)
		sisKey := ctx.SisKey()

		// Returns the list of all the hashes modulo X^n - 1
		subDuals := []smartvectors.SmartVector{}
		roundStartAt := 0
		for _, comsInRoundI := range ctx.VortexCtx.CommitmentsByRounds.Inner() {

			// Check and skip if there is no committed rows
			if len(comsInRoundI) == 0 {
				continue
			}

			// Compute the dual for the chunk I
			preimageSlice := collapsedPreimage.SubVector(
				roundStartAt*sisKey.NumLimbs(),
				(roundStartAt+len(comsInRoundI))*sisKey.NumLimbs(),
			)
			subDual := sisKey.HashModXnMinus1(smartvectors.IntoRegVec(preimageSlice))
			subDuals = append(subDuals, smartvectors.NewRegular(subDual))

			// And update the start cursor
			roundStartAt += len(comsInRoundI)
		}

		mergerCoin := run.GetRandomCoinField(ctx.Coins.Merge.Name)
		eDual := smartvectors.PolyEval(subDuals, mergerCoin)

		run.AssignColumn(ctx.Columns.Edual.GetColID(), eDual)

	})
}

// Registers the final folding phase of the self-recursion
//
//   - Sample the folding random coin r\fold
//
//   - Fold A\merge by rFold to obtain AmergeFold
//
//   - Fold PreimageCollapse by rFold to obtain PreimageCollapseFold
//
//   - Declare and assign the inner-product between PreimageCollapseFold
//     and AmergeFold
//
//   - Perform the final check to evaluate the consistency vs
//     Edual and D\merge,\collapse,q
func (ctx *SelfRecursionCtx) foldPhase() {

	// The round of declaration should be one more than EDual
	round := ctx.Columns.Edual.Round() + 1

	// Sample rFold
	ctx.Coins.Fold = ctx.comp.InsertCoin(round, ctx.foldCoinName(), coin.Field)

	// Constructs AmergeFold
	ctx.Columns.AmergeFold = functionals.Fold(
		ctx.comp, ctx.Columns.Amerge,
		accessors.AccessorFromCoin(ctx.Coins.Fold),
		ctx.VortexCtx.SisParams.OutputSize(),
	)

	// Construct DmergeCollapseFold
	ctx.Columns.PreimageCollapseFold = functionals.Fold(
		ctx.comp, ctx.Columns.PreimagesCollapse,
		accessors.AccessorFromCoin(ctx.Coins.Fold),
		ctx.VortexCtx.SisParams.OutputSize(),
	)

	// Mark Edual and the DmergeQCollapse fold as proof
	ctx.comp.Columns.SetStatus(ctx.Columns.DmergeQcollapse.GetColID(), column.Proof)
	ctx.comp.Columns.SetStatus(ctx.Columns.Edual.GetColID(), column.Proof)

	// Declare and assign the inner-product
	ctx.Queries.LatticeInnerProd = ctx.comp.InsertInnerProduct(
		round, ctx.preimagesAndAmergeIP(), ctx.Columns.AmergeFold,
		[]ifaces.Column{ctx.Columns.PreimageCollapseFold})

	// Assignment part of the inner product
	ctx.comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
		// compute the inner-product
		foldedKey := ctx.Columns.AmergeFold.GetColAssignment(run)                // overshadows the handle
		foldedPreimage := ctx.Columns.PreimageCollapseFold.GetColAssignment(run) // overshadows the handle

		y := smartvectors.InnerProduct(foldedKey, foldedPreimage)
		run.AssignInnerProduct(ctx.preimagesAndAmergeIP(), y)
	})

	degree := ctx.SisKey().Degree

	// And the final check
	// check the folding of the polynomial is correct
	ctx.comp.InsertVerifier(round, func(run *wizard.VerifierRuntime) error {

		// fetch the assignments to edual and dcollapse
		edual := ctx.Columns.Edual.GetColAssignment(run)
		dcollapse := ctx.Columns.DmergeQcollapse.GetColAssignment(run)

		// the folding coin
		rfold := run.GetRandomCoinField(ctx.Coins.Fold.Name)

		// evaluates both edual and dcollapse (seen as polynomial) by
		// coefficients and fetch the result of the inner-product
		yAlleged := run.GetInnerProductParams(ctx.preimagesAndAmergeIP()).Ys[0]
		yDual := smartvectors.EvalCoeff(edual, rfold)
		yActual := smartvectors.EvalCoeff(dcollapse, rfold)

		/*
			If P(X) is of degree 2n

			And
				- Q(X) = P(X) mod X^n - 1
				- R(X) = P(X) mod X^n + 1

			Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
			Here, we can identify at the point x

			yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
		*/
		var xN, xNminus1, xNplus1 field.Element
		one := field.One()
		xN.Exp(rfold, big.NewInt(int64(degree)))
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
	}, func(api frontend.API, run *wizard.WizardVerifierCircuit) {

		// fetch the assignments to edual and dcollapse
		edual := ctx.Columns.Edual.GetColAssignmentGnark(run)
		dcollapse := ctx.Columns.DmergeQcollapse.GetColAssignmentGnark(run)

		// the folding coin
		rfold := run.GetRandomCoinField(ctx.Coins.Fold.Name)

		// evaluates both edual and dcollapse (seen as polynomial) by
		// coefficients and fetch the result of the inner-product
		yAlleged := run.GetInnerProductParams(ctx.preimagesAndAmergeIP()).Ys[0]
		yDual := poly.EvaluateUnivariateGnark(api, edual, rfold)
		yActual := poly.EvaluateUnivariateGnark(api, dcollapse, rfold)

		/*
		   If P(X) is of degree 2n

		   And
		     - Q(X) = P(X) mod X^n - 1
		     - R(X) = P(X) mod X^n + 1

		   Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
		   Here, we can identify at the point x

		   yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
		*/
		one := field.One()
		xN := gnarkutil.Exp(api, rfold, degree)
		xNminus1 := api.Sub(xN, one)
		xNplus1 := api.Add(xN, one)

		left0 := api.Mul(xNplus1, yDual)
		left1 := api.Mul(xNminus1, yActual)
		left := api.Sub(left0, left1)

		right := api.Mul(yAlleged, 2)

		api.AssertIsEqual(left, right)
	})
}

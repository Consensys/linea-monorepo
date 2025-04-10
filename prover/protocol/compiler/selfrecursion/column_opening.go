package selfrecursion

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
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
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/merkle"
	mimcW "github.com/consensys/linea-monorepo/prover/protocol/dedicated/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

// Specifies the column opening phase
func (ctx *SelfRecursionCtx) ColumnOpeningPhase() {
	// Registers the limb expanded version of the preimages
	ctx.colSelection()
	ctx.linearHashAndMerkle()
	ctx.RootHashGlue()
	ctx.GluePositions()
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

	limbs := make([]ifaces.Column, len(wholes))
	round := wholes[0].Round()
	limbSize := wholes[0].Size() * sisParams.NumLimbs()

	for i := range limbs {
		limbs[i] = ctx.comp.InsertCommit(
			round,
			ctx.limbExpandedPreimageName(wholes[i].GetColID()),
			limbSize,
		)
		ctx.comp.InsertRange(
			round,
			ifaces.QueryIDf("SHORTNESS_%v", limbs[i].GetColID()),
			limbs[i],
			1<<ctx.VortexCtx.SisParams.LogTwoBound,
		)
	}

	ctx.Columns.Preimages = limbs

	ctx.comp.RegisterProverAction(round, &preimageLimbsProverAction{
		ctx:   ctx,
		limbs: limbs,
	})

}

type preimageLimbsProverAction struct {
	ctx   *SelfRecursionCtx
	limbs []ifaces.Column
}

func (a *preimageLimbsProverAction) Run(run *wizard.ProverRuntime) {
	parallel.Execute(len(a.limbs), func(start, end int) {
		for i := start; i < end; i++ {
			whole := a.ctx.Columns.WholePreimages[i].GetColAssignment(run)
			whole_ := smartvectors.IntoRegVec(whole)
			expanded_ := a.ctx.SisKey().LimbSplit(whole_)
			expanded := smartvectors.NewRegular(expanded_)
			run.AssignColumn(a.limbs[i].GetColID(), expanded)
			logrus.Infof("Assigned limb column: %v", a.limbs[i].GetColID())
		}
	})
}

type colSelectionProverAction struct {
	ctx       *SelfRecursionCtx
	uAlphaQID ifaces.ColID
}

func (a *colSelectionProverAction) Run(run *wizard.ProverRuntime) {
	q := run.GetRandomCoinIntegerVec(a.ctx.Coins.Q.Name)
	uAlpha := smartvectors.IntoRegVec(run.GetColumn(a.ctx.Columns.Ualpha.GetColID()))

	uAlphaQ := make([]field.Element, 0, a.ctx.Columns.UalphaQ.Size())
	for _, qi := range q {
		uAlphaQ = append(uAlphaQ, uAlpha[qi])
	}

	run.AssignColumn(a.uAlphaQID, smartvectors.NewRegular(uAlphaQ))
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
func (ctx *SelfRecursionCtx) colSelection() {

	// Build the column q, (auto-assigned)
	ctx.Columns.Q = verifiercol.NewFromIntVecCoin(ctx.comp, ctx.Coins.Q)

	// Declaration round of the coin Q
	roundQ := ctx.Columns.Q.Round()

	ctx.Columns.UalphaQ = ctx.comp.InsertCommit(
		roundQ,
		ctx.uAlphaQName(),
		ctx.Coins.Q.Size,
	)

	// And registers the assignment function
	ctx.comp.RegisterProverAction(roundQ, &colSelectionProverAction{
		ctx:       ctx,
		uAlphaQID: ctx.Columns.UalphaQ.GetColID(),
	})

	// Declare an inclusion query to finalize the selection check
	ctx.comp.InsertInclusion(
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

type linearHashMerkleProverAction struct {
	ctx                *SelfRecursionCtx
	concatDhQSize      int
	leavesSize         int
	leavesSizeUnpadded int
}

func (a *linearHashMerkleProverAction) Run(run *wizard.ProverRuntime) {
	openingIndices := run.GetRandomCoinIntegerVec(a.ctx.Coins.Q.Name)
	concatDhQ := make([]field.Element, a.leavesSizeUnpadded*a.ctx.VortexCtx.SisParams.OutputSize())
	linearLeaves := make([]field.Element, a.leavesSizeUnpadded)
	merkleLeaves := make([]field.Element, a.leavesSizeUnpadded)
	merklePositions := make([]field.Element, a.leavesSizeUnpadded)
	merkleRoots := make([]field.Element, a.leavesSizeUnpadded)

	hashSize := a.ctx.VortexCtx.SisParams.OutputSize()
	numOpenedCol := a.ctx.VortexCtx.NbColsToOpen()
	totalNumRounds := a.ctx.comp.NumRounds()
	committedRound := 0

	if a.ctx.VortexCtx.IsCommitToPrecomputed() {
		rootPrecomp := a.ctx.Columns.precompRoot.GetColAssignment(run).Get(0)
		precompColSisHash := a.ctx.VortexCtx.Items.Precomputeds.DhWithMerkle
		for i, selectedCol := range openingIndices {
			srcStart := selectedCol * hashSize
			destStart := i * hashSize
			sisHash := precompColSisHash[srcStart : srcStart+hashSize]
			copy(concatDhQ[destStart:destStart+hashSize], sisHash)
			leaf := mimc.HashVec(sisHash)
			insertAt := i
			linearLeaves[insertAt] = leaf
			merkleLeaves[insertAt] = leaf
			merkleRoots[insertAt] = rootPrecomp
			merklePositions[insertAt].SetInt64(int64(selectedCol))
		}
		committedRound++
		totalNumRounds++
	}

	for round := 0; round <= totalNumRounds; round++ {
		colSisHashName := a.ctx.VortexCtx.SisHashName(round)
		colSisHashSV, found := run.State.TryGet(colSisHashName)
		if !found {
			continue
		}

		rooth := a.ctx.Columns.Rooth[round].GetColAssignment(run).Get(0)
		colSisHash := colSisHashSV.([]field.Element)

		for i, selectedCol := range openingIndices {
			srcStart := selectedCol * hashSize
			destStart := committedRound*numOpenedCol*hashSize + i*hashSize
			sisHash := colSisHash[srcStart : srcStart+hashSize]
			copy(concatDhQ[destStart:destStart+hashSize], sisHash)
			leaf := mimc.HashVec(sisHash)
			insertAt := committedRound*numOpenedCol + i
			linearLeaves[insertAt] = leaf
			merkleLeaves[insertAt] = leaf
			merkleRoots[insertAt] = rooth
			merklePositions[insertAt].SetInt64(int64(selectedCol))
		}

		run.State.TryDel(colSisHashName)
		committedRound++
	}

	numCommittedRound := a.ctx.VortexCtx.NumCommittedRounds()
	if a.ctx.VortexCtx.IsCommitToPrecomputed() {
		numCommittedRound += 1
	}
	if committedRound != numCommittedRound {
		utils.Panic("Committed rounds %v does not match the total number of committed rounds %v", committedRound, numCommittedRound)
	}

	// Assign columns using IDs from ctx.Columns
	run.AssignColumn(a.ctx.Columns.ConcatenatedDhQ.GetColID(), smartvectors.RightZeroPadded(concatDhQ, a.concatDhQSize))
	run.AssignColumn(a.ctx.Columns.MerkleProofsLeaves.GetColID(), smartvectors.RightZeroPadded(merkleLeaves, a.leavesSize))
	run.AssignColumn(a.ctx.Columns.MerkleProofPositions.GetColID(), smartvectors.RightZeroPadded(merklePositions, a.leavesSize))
	run.AssignColumn(a.ctx.Columns.MerkleRoots.GetColID(), smartvectors.RightZeroPadded(merkleRoots, a.leavesSize))
}

func (ctx *SelfRecursionCtx) linearHashAndMerkle() {
	roundQ := ctx.Columns.Q.Round()
	numRound := ctx.VortexCtx.NumCommittedRounds()
	if ctx.VortexCtx.IsCommitToPrecomputed() {
		numRound += 1
	}
	concatDhQSizeUnpadded := ctx.VortexCtx.SisParams.OutputSize() * ctx.VortexCtx.NbColsToOpen() * numRound
	concatDhQSize := utils.NextPowerOfTwo(concatDhQSizeUnpadded)
	leavesSizeUnpadded := ctx.VortexCtx.NbColsToOpen() * numRound
	leavesSize := utils.NextPowerOfTwo(leavesSizeUnpadded)

	ctx.Columns.ConcatenatedDhQ = ctx.comp.InsertCommit(roundQ, ctx.concatenatedDhQ(), concatDhQSize)
	ctx.Columns.MerkleProofsLeaves = ctx.comp.InsertCommit(roundQ, ctx.merkleLeavesName(), leavesSize)
	ctx.Columns.MerkleProofPositions = ctx.comp.InsertCommit(roundQ, ctx.merklePositionssName(), leavesSize)
	ctx.Columns.MerkleRoots = ctx.comp.InsertCommit(roundQ, ctx.merkleRootsName(), leavesSize)

	ctx.comp.RegisterProverAction(roundQ, &linearHashMerkleProverAction{
		ctx:                ctx,
		concatDhQSize:      concatDhQSize,
		leavesSize:         leavesSize,
		leavesSizeUnpadded: leavesSizeUnpadded,
	})

	depth := utils.Log2Ceil(ctx.VortexCtx.NumEncodedCols())
	merkle.MerkleProofCheck(ctx.comp, ctx.merkleProofVerificationName(), depth, leavesSizeUnpadded,
		ctx.Columns.MerkleProofs, ctx.Columns.MerkleRoots, ctx.Columns.MerkleProofsLeaves, ctx.Columns.MerkleProofPositions)
	mimcW.CheckLinearHash(ctx.comp, ctx.linearHashVerificationName(), ctx.Columns.ConcatenatedDhQ,
		ctx.VortexCtx.SisParams.OutputSize(), leavesSizeUnpadded, ctx.Columns.MerkleProofsLeaves)
}

type collapsingProverAction struct {
	ctx     *SelfRecursionCtx
	eDualID ifaces.ColID
	sisKey  *ringsis.Key
}

func (a *collapsingProverAction) Run(run *wizard.ProverRuntime) {
	collapsedPreimage := a.ctx.Columns.PreimagesCollapse.GetColAssignment(run)
	sisKey := a.sisKey

	subDuals := []smartvectors.SmartVector{}
	roundStartAt := 0

	if a.ctx.VortexCtx.IsCommitToPrecomputed() {
		numPrecomputeds := len(a.ctx.VortexCtx.Items.Precomputeds.PrecomputedColums)
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

	for _, comsInRoundI := range a.ctx.VortexCtx.CommitmentsByRounds.Inner() {
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

	colPowT := accessors.NewExponent(a.ctx.Coins.Collapse, a.ctx.VortexCtx.NbColsToOpen()).GetVal(run)
	eDual := smartvectors.PolyEval(subDuals, colPowT)

	run.AssignColumn(a.eDualID, eDual)
}

type collapsingVerifierAction struct {
	uAlphaQEval  ifaces.Accessor
	preImageEval ifaces.Accessor
}

func (a *collapsingVerifierAction) Run(run wizard.Runtime) error {
	if a.uAlphaQEval.GetVal(run) != a.preImageEval.GetVal(run) {
		l, r := a.uAlphaQEval.GetVal(run), a.preImageEval.GetVal(run)
		return fmt.Errorf("consistency between u_alpha and the preimage: mismatch between uAlphaQEval=%v preimages=%v",
			l.String(), r.String())
	}
	return nil
}

func (a *collapsingVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	api.AssertIsEqual(
		a.uAlphaQEval.GetFrontendVariable(api, run),
		a.preImageEval.GetFrontendVariable(api, run),
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
func (ctx *SelfRecursionCtx) collapsingPhase() {

	// starting one round after Q is sampled
	round := ctx.Columns.Q.Round() + 1

	// Sampling of r_collapse
	ctx.Coins.Collapse = ctx.comp.InsertCoin(round, ctx.collapseCoin(), coin.Field)

	// Declare the linear combination of the preimages by collapse coin
	// aka, the collapsed preimage
	ctx.Columns.PreimagesCollapse = expr_handle.RandLinCombCol(
		ctx.comp,
		accessors.NewFromCoin(ctx.Coins.Collapse),
		ctx.Columns.Preimages,
	)

	// Consistency check between the collapsed preimage and UalphaQ
	{
		uAlphaQEval := functionals.CoeffEval(
			ctx.comp,
			ctx.constencyUalphaQPreimageLeft(),
			ctx.Coins.Collapse,
			ctx.Columns.UalphaQ,
		)

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

		preImageEval := functionals.EvalCoeffBivariate(
			ctx.comp,
			ctx.constencyUalphaQPreimageRight(),
			ctx.Columns.PreimagesCollapse,
			accessors.NewConstant(field.NewElement(1<<ctx.SisKey().LogTwoBound)),
			accessors.NewFromCoin(ctx.Coins.Alpha),
			ctx.VortexCtx.SisParams.NumLimbs(),
			ctx.Columns.WholePreimages[0].Size(),
		)

		ctx.comp.RegisterVerifierAction(uAlphaQEval.Round(), &collapsingVerifierAction{
			uAlphaQEval:  uAlphaQEval,
			preImageEval: preImageEval,
		})
	}

	sisDeg := ctx.VortexCtx.SisParams.OutputSize()
	// Currently, only powers of two SIS degree are allowed
	// (in practice, we restrict ourselves to pure power of two)
	// lattices instances.
	if !utils.IsPowerOfTwo(sisDeg) {
		utils.Panic("Attempting to fold to a non-power of two size : %v", sisDeg)
	}

	// Compute the collapsed hashes
	ctx.Columns.DhQCollapse = functionals.FoldOuter(
		ctx.comp,
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

	// since some of the Ah and Dh can be nil, we compactify the slice by
	// only retaining the non-nil elements before sending it to the
	// linear combination operator.
	nonNilAh := []ifaces.Column{}
	for _, ah := range ctx.Columns.Ah {
		if ah != nil {
			nonNilAh = append(nonNilAh, ah)
		}
	}

	// And computes the linear combination
	ctx.Columns.ACollapsed = expr_handle.RandLinCombCol(
		ctx.comp,
		collapsePowT,
		nonNilAh,
		ctx.aCollapsedName(),
	)

	// And edual

	// Declare Edual
	ctx.Columns.Edual = ctx.comp.InsertCommit(
		round, ctx.eDual(), ctx.VortexCtx.SisParams.OutputSize(),
	)

	// And assign it
	ctx.comp.RegisterProverAction(round, &collapsingProverAction{
		ctx:     ctx,
		eDualID: ctx.Columns.Edual.GetColID(),
		sisKey:  ctx.SisKey(),
	})
}

type foldPhaseProverAction struct {
	ctx       *SelfRecursionCtx
	ipQueryID ifaces.QueryID // Changed to ifaces.QueryID explicitly
}

func (a *foldPhaseProverAction) Run(run *wizard.ProverRuntime) {
	foldedKey := a.ctx.Columns.ACollapseFold.GetColAssignment(run)
	foldedPreimage := a.ctx.Columns.PreimageCollapseFold.GetColAssignment(run)
	y := smartvectors.InnerProduct(foldedKey, foldedPreimage)
	run.AssignInnerProduct(a.ipQueryID, y)
}

type foldPhaseVerifierAction struct {
	ctx       *SelfRecursionCtx
	ipQueryID ifaces.QueryID
	degree    int
}

func (a *foldPhaseVerifierAction) Run(run wizard.Runtime) error {
	edual := a.ctx.Columns.Edual.GetColAssignment(run)
	dcollapse := a.ctx.Columns.DhQCollapse.GetColAssignment(run)
	rfold := run.GetRandomCoinField(a.ctx.Coins.Fold.Name)
	yAlleged := run.GetInnerProductParams(a.ipQueryID).Ys[0]
	yDual := smartvectors.EvalCoeff(edual, rfold)
	yActual := smartvectors.EvalCoeff(dcollapse, rfold)

	var xN, xNminus1, xNplus1 field.Element
	one := field.One()
	xN.Exp(rfold, big.NewInt(int64(a.degree)))
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

func (a *foldPhaseVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	edual := a.ctx.Columns.Edual.GetColAssignmentGnark(run)
	dcollapse := a.ctx.Columns.DhQCollapse.GetColAssignmentGnark(run)
	rfold := run.GetRandomCoinField(a.ctx.Coins.Fold.Name)
	yAlleged := run.GetInnerProductParams(a.ipQueryID).Ys[0]
	yDual := poly.EvaluateUnivariateGnark(api, edual, rfold)
	yActual := poly.EvaluateUnivariateGnark(api, dcollapse, rfold)

	one := field.One()
	xN := gnarkutil.Exp(api, rfold, a.degree)
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
func (ctx *SelfRecursionCtx) foldPhase() {

	// The round of declaration should be one more than EDual
	round := ctx.Columns.Edual.Round() + 1

	// Sample rFold
	ctx.Coins.Fold = ctx.comp.InsertCoin(round, ctx.foldCoinName(), coin.Field)

	// Constructs ACollapsedFold
	ctx.Columns.ACollapseFold = functionals.Fold(
		ctx.comp, ctx.Columns.ACollapsed,
		accessors.NewFromCoin(ctx.Coins.Fold),
		ctx.VortexCtx.SisParams.OutputSize(),
	)

	// Construct DmergeCollapseFold
	ctx.Columns.PreimageCollapseFold = functionals.Fold(
		ctx.comp, ctx.Columns.PreimagesCollapse,
		accessors.NewFromCoin(ctx.Coins.Fold),
		ctx.VortexCtx.SisParams.OutputSize(),
	)

	// Mark Edual and the DmergeQCollapse fold as proof
	ctx.comp.Columns.SetStatus(ctx.Columns.DhQCollapse.GetColID(), column.Proof)
	ctx.comp.Columns.SetStatus(ctx.Columns.Edual.GetColID(), column.Proof)

	// Declare and assign the inner-product
	ctx.Queries.LatticeInnerProd = ctx.comp.InsertInnerProduct(
		round, ctx.preimagesAndAmergeIP(), ctx.Columns.ACollapseFold,
		[]ifaces.Column{ctx.Columns.PreimageCollapseFold})

	// Assignment part of the inner product
	ctx.comp.RegisterProverAction(round, &foldPhaseProverAction{
		ctx:       ctx,
		ipQueryID: ctx.Queries.LatticeInnerProd.Name(),
	})

	degree := ctx.SisKey().OutputSize()

	// And the final check
	// check the folding of the polynomial is correct
	// ctx.comp.InsertVerifier(round, func(run wizard.Runtime) error {

	// 	// fetch the assignments to edual and dcollapse
	// 	edual := ctx.Columns.Edual.GetColAssignment(run)
	// 	dcollapse := ctx.Columns.DhQCollapse.GetColAssignment(run)

	// 	// the folding coin
	// 	rfold := run.GetRandomCoinField(ctx.Coins.Fold.Name)

	// 	// evaluates both edual and dcollapse (seen as polynomial) by
	// 	// coefficients and fetch the result of the inner-product
	// 	yAlleged := run.GetInnerProductParams(ctx.preimagesAndAmergeIP()).Ys[0]
	// 	yDual := smartvectors.EvalCoeff(edual, rfold)
	// 	yActual := smartvectors.EvalCoeff(dcollapse, rfold)

	// 	/*
	// 		If P(X) is of degree 2n

	// 		And
	// 			- Q(X) = P(X) mod X^n - 1
	// 			- R(X) = P(X) mod X^n + 1

	// 		Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
	// 		Here, we can identify at the point x

	// 		yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
	// 	*/
	// 	var xN, xNminus1, xNplus1 field.Element
	// 	one := field.One()
	// 	xN.Exp(rfold, big.NewInt(int64(degree)))
	// 	xNminus1.Sub(&xN, &one)
	// 	xNplus1.Add(&xN, &one)

	// 	var left, left0, left1, right field.Element
	// 	left0.Mul(&xNplus1, &yDual)
	// 	left1.Mul(&xNminus1, &yActual)
	// 	left.Sub(&left0, &left1)

	// 	right.Double(&yAlleged)

	// 	if left != right {
	// 		return fmt.Errorf("failed the consistency check of the ring-SIS : %v != %v", left.String(), right.String())
	// 	}

	// 	return nil
	// }, func(api frontend.API, run wizard.GnarkRuntime) {

	// 	// fetch the assignments to edual and dcollapse
	// 	edual := ctx.Columns.Edual.GetColAssignmentGnark(run)
	// 	dcollapse := ctx.Columns.DhQCollapse.GetColAssignmentGnark(run)

	// 	// the folding coin
	// 	rfold := run.GetRandomCoinField(ctx.Coins.Fold.Name)

	// 	// evaluates both edual and dcollapse (seen as polynomial) by
	// 	// coefficients and fetch the result of the inner-product
	// 	yAlleged := run.GetInnerProductParams(ctx.preimagesAndAmergeIP()).Ys[0]
	// 	yDual := poly.EvaluateUnivariateGnark(api, edual, rfold)
	// 	yActual := poly.EvaluateUnivariateGnark(api, dcollapse, rfold)

	// 	/*
	// 	   If P(X) is of degree 2n

	// 	   And
	// 	     - Q(X) = P(X) mod X^n - 1
	// 	     - R(X) = P(X) mod X^n + 1

	// 	   Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
	// 	   Here, we can identify at the point x

	// 	   yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
	// 	*/
	// 	one := field.One()
	// 	xN := gnarkutil.Exp(api, rfold, degree)
	// 	xNminus1 := api.Sub(xN, one)
	// 	xNplus1 := api.Add(xN, one)

	// 	left0 := api.Mul(xNplus1, yDual)
	// 	left1 := api.Mul(xNminus1, yActual)
	// 	left := api.Sub(left0, left1)

	// 	right := api.Mul(yAlleged, 2)

	// 	api.AssertIsEqual(left, right)
	// })

	ctx.comp.RegisterVerifierAction(round, &foldPhaseVerifierAction{
		ctx:       ctx,
		ipQueryID: ctx.Queries.LatticeInnerProd.Name(),
		degree:    degree,
	})
}

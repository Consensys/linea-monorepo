package arithmetics

import (
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"sync"
	"time"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft/fastpoly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

const (
	GLOBAL_REDUCTION                string = "GLOBAL_REDUCTION"
	OFFSET_RANDOMNESS               string = "OFFSET_RANDOMNESS"
	DEGREE_RANDOMNESS               string = "DEGREE_RANDOMNESS"
	QUOTIENT_POLY_TMPL              string = "QUOTIENT_DEG_%v_SHARE_%v"
	EVALUATION_RANDOMESS            string = "EVALUATION_RANDOMNESS"
	UNIVARIATE_EVAL_ALL_HANDLES     string = "UNIV_EVAL_ALL_HANDLES"
	UNIVARIATE_EVAL_QUOTIENT_SHARES string = "UNIV_EVAL_QUOTIENT"
)

// Compute profile
func CompileGlobal(comp *wizard.CompiledIOP) {

	initialNumRound := comp.NumRounds()

	// Initialize two randomness, one to aggregate the constraints
	// at the "offset-level" and then at the degree "level"
	ctx := pocGlobalCtx{}
	ctx.comp = comp
	ctx.selfRecursionCounter = comp.SelfRecursionCount
	ctx.offsetRandomness = comp.InsertCoin(initialNumRound, coin.Name(ctx.derivename(OFFSET_RANDOMNESS)), coin.Field)
	ctx.degreeRandomness = comp.InsertCoin(initialNumRound, coin.Name(ctx.derivename(DEGREE_RANDOMNESS)), coin.Field)

	if !ctx.sortConstraints(comp) {
		logrus.Infof("no global constraints found")
		return
	}
	ctx.aggregateConstraints()

	// Register all the existing commitments
	ctx.organizeTheHandles(comp)

	// The global quotients are (one for each deg), are registered in shares.
	// This is keep the invariant that all
	for _, ratio := range ctx.ratioIndexes {
		for share := 0; share < ratio; share++ {
			comp.InsertCommit(initialNumRound, ifaces.ColID(ctx.derivename(quotientShareName(share, ratio))), ctx.domainSize)
		}
	}

	comp.SubProvers.AppendToInner(initialNumRound, ctx.proverComputeQuotient(comp))

	// Registers the quotient polynomials
	ctx.evaluationRandomness = comp.InsertCoin(initialNumRound+1, coin.Name(ctx.derivename(EVALUATION_RANDOMESS)), coin.Field)
	ctx.allHandlesEval = comp.InsertUnivariate(initialNumRound+1, ifaces.QueryID(ctx.derivename(UNIVARIATE_EVAL_ALL_HANDLES)), ctx.allInvolvedHandles)
	ctx.gatherQuotientHandles(comp)
	ctx.quotientsEval = comp.InsertUnivariate(initialNumRound+1, ifaces.QueryID(ctx.derivename(UNIVARIATE_EVAL_QUOTIENT_SHARES)), ctx.quotientHandles)

	comp.SubProvers.AppendToInner(initialNumRound+1, ctx.proverEvaluationQueries(comp))
	comp.InsertVerifier(initialNumRound+1, ctx.verifierStep, ctx.gnarkVerifierStep)
}

// Global context
type pocGlobalCtx struct {
	// The underlying compiled IOP
	comp *wizard.CompiledIOP
	// Snapshot of the selfrecursion counter, for deduplication
	selfRecursionCounter int

	buckets map[int]map[utils.Range][]query.GlobalConstraint
	// The role of the indexes is for deterministic iterations
	ratioIndexes []int
	rangeIndexes map[int][]utils.Range
	// Aggregate expressions by deg
	aggregatesExpressions    []*symbolic.Expression
	aggregateExpressionBoard []*symbolic.ExpressionBoard
	// Size of the domain expressions
	domainSize int

	handlePerRatio          [][]ifaces.Column
	rootPerRatio            [][]ifaces.Column
	allInvolvedHandles      []ifaces.Column
	allInvolvedRoots        []ifaces.Column
	allInvolvedHandlesIndex map[ifaces.ColID]int

	quotientHandles      []ifaces.Column
	quotientShareHandles [][]ifaces.Column

	// Coins
	offsetRandomness, degreeRandomness, evaluationRandomness coin.Info
	// UnivariateQueries
	allHandlesEval, quotientsEval query.UnivariateEval
}

// Sort the constraints in buckets indexed by the offset profile and the degree
// of the constraint. The function also sanity-checks that all function are on the
// same domain.
// The result is a map of map of bucket list ( deg => (range => list of constraint) )
func (ctx *pocGlobalCtx) sortConstraints(comp *wizard.CompiledIOP) (foundAnyConstraint bool) {

	// Values to assign in the context as an output of the sorting operation
	buckets := map[int]map[utils.Range][]query.GlobalConstraint{}
	ratioIndexes := []int{}
	rangeIndexes := map[int][]utils.Range{}

	// Dispatch the constraint in buckets
	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {

		// Filter only the global constraints
		q, ok := comp.QueriesNoParams.Data(qName).(query.GlobalConstraint)
		if !ok {
			// Not a global constraint
			continue
		}

		// If the domain size is not set, initialize it
		if ctx.domainSize == 0 {
			ctx.domainSize = q.DomainSize
		}

		// Sanity-check
		if ctx.domainSize != q.DomainSize {
			utils.Panic("At this point in the compilation process, we expect all constraints to have the same domain")
		}

		// Mark the constraint as ignored
		comp.QueriesNoParams.MarkAsIgnored(qName)
		// Assigns a bucket to the constraint
		board := q.Expression.Board()

		// By range, account for the fact that we may want to not cancel the constraint
		offsetRange := q.MinMaxOffset()
		if q.NoBoundCancel {
			offsetRange = utils.Range{}
		}

		exprDeg := board.Degree(GetDegree(ctx.domainSize))
		// Account in advance for the cancellator polynomial that we are going to multiply
		// And the fact that we divide the polynomial by X^N - 1.
		// The +1 terms compensates that `exprDeg` returns a degree but we "ratio" is a ratio
		// of numbers of coefficients.
		ratioFloat := float64(exprDeg+offsetRange.Max-offsetRange.Min-ctx.domainSize+1) / float64(ctx.domainSize)
		ratio := int(ratioFloat)
		if float64(ratio) < ratioFloat {
			// Guards against rounding errors
			ratio++
		}
		ratio = utils.Max(ratio, 1)
		ratio = utils.NextPowerOfTwo(ratio)

		// Initialize the outer-maps / slices if the entries are not already allocated
		if _, ok := buckets[ratio]; !ok {
			buckets[ratio] = map[utils.Range][]query.GlobalConstraint{}
			ratioIndexes = append(ratioIndexes, ratio)
			rangeIndexes[ratio] = []utils.Range{}
		}

		// Same for the inner maps
		if _, ok := buckets[ratio][offsetRange]; !ok {
			buckets[ratio][offsetRange] = []query.GlobalConstraint{}
			rangeIndexes[ratio] = append(rangeIndexes[ratio], offsetRange)
		}

		buckets[ratio][offsetRange] = append(buckets[ratio][offsetRange], q)
	}

	// Assign the result
	ctx.buckets = buckets
	ctx.ratioIndexes = ratioIndexes
	ctx.rangeIndexes = rangeIndexes

	return len(ratioIndexes) > 0
}

// Aggregate the constraints expressions into several linear combination
func (ctx *pocGlobalCtx) aggregateConstraints() {

	ctx.aggregatesExpressions = make([]*symbolic.Expression, len(ctx.ratioIndexes))
	ctx.aggregateExpressionBoard = make([]*symbolic.ExpressionBoard, len(ctx.ratioIndexes))

	for i, ratio := range ctx.ratioIndexes {
		sameRangeAggregates := make([]*symbolic.Expression, len(ctx.rangeIndexes[ratio]))

		// Collect the linear combinations of all buckets
		for j, offsets := range ctx.rangeIndexes[ratio] {
			// Collect the expressions of each global constraint in the bucket
			currBucket := ctx.buckets[ratio][offsets]
			bucketExprs := make([]*symbolic.Expression, len(currBucket))

			for i := range currBucket {
				bucketExprs[i] = currBucket[i].Expression
			}

			// And collect the aggregate of the bucket after cancelling on the range
			newExpr := symbolic.NewPolyEval(ctx.offsetRandomness.AsVariable(), bucketExprs)
			sameRangeAggregates[j] = cancelExprOnRange(newExpr, offsets, ctx.domainSize)
		}

		// Then aggregate on a second level the freshly aggregated expressions
		ctx.aggregatesExpressions[i] = symbolic.NewPolyEval(ctx.degreeRandomness.AsVariable(), sameRangeAggregates)
		board := ctx.aggregatesExpressions[i].Board()
		ctx.aggregateExpressionBoard[i] = &board
	}
}

// Cancel an expression on a given range (given a domain)
func cancelExprOnRange(expr *symbolic.Expression, cancelRange utils.Range, domainSize int) *symbolic.Expression {

	res := expr

	// Function that cancels on a specific point
	cancelExprAtPoint := func(expr *symbolic.Expression, i, size int) *symbolic.Expression {
		x := variables.NewXVar()
		omega := fft.GetOmega(size)
		var root field.Element
		root.Exp(omega, big.NewInt(int64(i)))
		return expr.Mul(x.Sub(symbolic.NewConstant(root)))
	}

	if cancelRange.Min < 0 {
		// Cancels the expression on the range [0, -cancelRange.Min)
		for i := 0; i < -cancelRange.Min; i++ {
			res = cancelExprAtPoint(res, i, domainSize)
		}
	}

	if cancelRange.Max > 0 {
		// Cancels the expression on the range (N-cancelRange.Max-1, N-1]
		for i := 0; i < cancelRange.Max; i++ {
			point := domainSize - i - 1 // point at which we want to cancel the constraint
			res = cancelExprAtPoint(res, point, domainSize)
		}
	}

	return res
}

func quotientShareName(share int, ratio int) string {
	return fmt.Sprintf(QUOTIENT_POLY_TMPL, ratio, share)
}

func (ctx *pocGlobalCtx) organizeTheHandles(comp *wizard.CompiledIOP) {

	// A global map of the metadata
	allInvolvedHandlesIndex := map[ifaces.ColID]int{}
	allInvolvedHandles := []ifaces.Column{}
	allInvolvedRoots := []ifaces.Column{}
	allInvolvedRootsSet := collection.NewSet[ifaces.ColID]()

	// We track the handle metadata for each expression to simplify the work
	handlePerRatio := make([][]ifaces.Column, len(ctx.ratioIndexes))
	rootsPerRatio := make([][]ifaces.Column, len(ctx.ratioIndexes))

	// Feed the map with all handle
	for i, expr := range ctx.aggregatesExpressions {
		board := expr.Board()
		uniqueRootsPerRatio := collection.NewSet[ifaces.ColID]()

		for _, metadata := range board.ListVariableMetadata() {
			// Filter in handle metadata only
			handle, ok := metadata.(ifaces.Column)
			if !ok {
				continue
			}

			rootCol := column.RootParents(handle)
			isShiftOrNatural := len(rootCol) == 1 && rootCol[0].Size() == handle.Size()

			// Append the handle (we trust that there are no duplicate of handles
			// within a constraint).
			handlePerRatio[i] = append(handlePerRatio[i], handle)

			if isShiftOrNatural && !uniqueRootsPerRatio.Exists(rootCol[0].GetColID()) {
				rootsPerRatio[i] = append(rootsPerRatio[i], rootCol[0])
			}

			// Get the name of the
			if _, alreadyThere := allInvolvedHandlesIndex[handle.GetColID()]; alreadyThere {
				continue
			}

			allInvolvedHandlesIndex[handle.GetColID()] = len(allInvolvedHandles)
			allInvolvedHandles = append(allInvolvedHandles, handle)

			// If the handle is simply a shift or a natural columns tracks its root
			if isShiftOrNatural && !allInvolvedRootsSet.Exists(rootCol[0].GetColID()) {
				allInvolvedRootsSet.Insert(rootCol[0].GetColID())
				allInvolvedRoots = append(allInvolvedRoots, rootCol[0])
			}
		}
	}

	ctx.handlePerRatio = handlePerRatio
	ctx.rootPerRatio = rootsPerRatio
	ctx.allInvolvedHandles = allInvolvedHandles
	ctx.allInvolvedHandlesIndex = allInvolvedHandlesIndex
	ctx.allInvolvedRoots = allInvolvedRoots
}

func (ctx *pocGlobalCtx) gatherQuotientHandles(comp *wizard.CompiledIOP) {
	ctx.quotientHandles = make([]ifaces.Column, len(ctx.ratioIndexes))
	ctx.quotientShareHandles = make([][]ifaces.Column, len(ctx.rangeIndexes))
	for i, ratio := range ctx.ratioIndexes {
		shareHandles := make([]ifaces.Column, ratio)
		for share := range shareHandles {
			shareHandles[share] = comp.Columns.GetHandle(
				ifaces.ColID(ctx.derivename(quotientShareName(share, ratio))),
			)
		}
		ctx.quotientHandles[i] = column.Interleave(shareHandles...)
		ctx.quotientShareHandles[i] = shareHandles
	}
}

func (ctx *pocGlobalCtx) proverComputeQuotient(comp *wizard.CompiledIOP) wizard.ProverStep {
	return func(run *wizard.ProverRuntime) {

		logrus.Infof("run the prover for the global constraint (quotient computation)")

		// Tracks the time spent on garbage collection
		totalTimeGc := int64(0)

		// Initial step is to compute the FFTs for all committed vectors
		coeffs := map[ifaces.ColID]sv.SmartVector{}
		stopTimer := profiling.LogTimer("Computing the coeffs %v pols of size %v", len(ctx.allInvolvedHandles), ctx.domainSize)

		lock := sync.Mutex{}
		lockRun := sync.Mutex{}

		// Force the GC to run
		tGc := time.Now()
		runtime.GC()
		totalTimeGc += time.Since(tGc).Milliseconds()
		logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc), totalTimeGc)

		// Compute once the FFT of the natural columns
		parallel.ExecuteChunky(len(ctx.allInvolvedRoots), func(start, stop int) {
			for k := start; k < stop; k++ {
				pol := ctx.allInvolvedRoots[k]
				name := pol.GetColID()

				// gets directly a shallow copy in the map of the runtime
				var witness sv.SmartVector
				lockRun.Lock()
				witness, isNatural := run.Columns.TryGet(name)
				lockRun.Unlock()

				// can happen if the column is verifier defined. In that case, no
				// need to protect with a lock. This will not touch run.Columns.
				if !isNatural {
					witness = pol.GetColAssignment(run)
				}
				witness = sv.FFTInverse(witness, fft.DIF, false, 0, 0)

				lock.Lock()
				coeffs[name] = witness
				lock.Unlock()
			}
		})

		parallel.ExecuteChunky(len(ctx.allInvolvedHandles), func(start, stop int) {
			for k := start; k < stop; k++ {
				pol := ctx.allInvolvedHandles[k]

				// short-path, the column is a shifted column of an already
				// present columns rule out interleaved and repeated columns.
				rootCols := column.RootParents(pol)
				if len(rootCols) == 1 && rootCols[0].Size() == pol.Size() {
					// It was already processed by the above loop. Go on with the next entry
					continue
				}

				// normal case for interleaved or repeated columns
				witness := pol.GetColAssignment(run)
				witness = sv.FFTInverse(witness, fft.DIF, false, 0, 0)
				name := pol.GetColID()
				lock.Lock()
				coeffs[name] = witness
				lock.Unlock()
			}
		})

		stopTimer()

		// Force the GC to run
		tGc = time.Now()
		runtime.GC()
		totalTimeGc += time.Since(tGc).Milliseconds()
		logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc).Milliseconds(), totalTimeGc)

		// Take the max quotient degree
		maxRatio := utils.Max(ctx.ratioIndexes...)

		/*
			For the quotient, we precompute the values of (wQ^N - 1)^-1 for w in H, the
			larger domain.

			Those values are D-periodic, thus we only compute a single period.
			(Where D is the ratio of the sizes of the larger and the smaller domain)

			The first value is ignored because it correspond to the case where w^N = 1
			(i.e. w is in the smaller subgroup)
		*/
		annulatorInvVals := fastpoly.EvalXnMinusOneOnACoset(ctx.domainSize, ctx.domainSize*maxRatio)
		annulatorInvVals = field.ParBatchInvert(annulatorInvVals, runtime.GOMAXPROCS(0))

		/*
			Also returns the evaluations of
		*/

		for i := 0; i < maxRatio; i++ {

			computedReeval := map[ifaces.ColID]sv.SmartVector{}
			stopTimer = profiling.LogTimer("Creation of omega")

			/*
				The following computes the quotient polynomial and assigns it

				Omega is a root of unity which generates the domain of evaluation of the
				constraint. Its size coincide with the size of the domain of evaluation.
				For each value of `i`, X will evaluate to gen*omegaQ^numCoset*omega^i.

				Gen is a generator of F^*
			*/
			omega := fft.GetOmega(ctx.domainSize)
			omegaQNumCos := fft.GetOmega(ctx.domainSize * maxRatio)
			omegaQNumCos.Exp(omegaQNumCos, big.NewInt(int64(i)))
			omegaI := field.NewElement(field.MultiplicativeGen)
			omegaI.Mul(&omegaI, &omegaQNumCos)

			// Precomputations of the powers of omega, can be optimized if useful
			omegas := make([]field.Element, ctx.domainSize)
			for i := range omegas {
				omegas[i] = omegaI
				omegaI.Mul(&omegaI, &omega)
			}

			stopTimer()

			for j, ratio := range ctx.ratioIndexes {

				// For instance, if deg = 2 and max deg 8, we enter only if
				// i = 0 or 4 because this correspond to the cosets we are
				// interested in.
				if i%(maxRatio/ratio) != 0 {
					continue
				}

				// With the above example,
				share := i * ratio / maxRatio
				handles := ctx.handlePerRatio[j]
				roots := ctx.rootPerRatio[j]
				board := ctx.aggregateExpressionBoard[j]
				metadatas := board.ListVariableMetadata()

				stopTimer := profiling.LogTimer("ReEvaluate %v pols of size %v on coset %v/%v", len(handles), ctx.domainSize, share, ratio)

				parallel.ExecuteChunky(len(roots), func(start, stop int) {
					for k := start; k < stop; k++ {
						root := roots[k]
						name := root.GetColID()

						lock.Lock()
						_, found := computedReeval[name]
						lock.Unlock()

						if found {
							// it was already computed in a previous iteration of `j`
							continue
						}

						// else it's the first value of j that sees it. so we compute the
						// coset reevaluation.
						reevaledRoot := sv.FFT(coeffs[name], fft.DIT, false, ratio, share)
						lock.Lock()
						computedReeval[name] = reevaledRoot
						lock.Unlock()

					}
				})

				parallel.ExecuteChunky(len(handles), func(start, stop int) {
					for k := start; k < stop; k++ {

						pol := handles[k]
						// short-path, the column is a purely Shifted(Natural) or a Natural
						// (this excludes repeats and/or interleaved columns)
						rootCols := column.RootParents(pol)
						if len(rootCols) == 1 && rootCols[0].Size() == pol.Size() {

							root := rootCols[0]
							name := root.GetColID()

							lock.Lock()
							reevaledRoot, found := computedReeval[name]
							lock.Unlock()

							if !found {
								// it is expected to computed in the above loop
								utils.Panic("did not find the reevaluation of %v", name)
							}

							// Now, we can reuse a soft-rotation of the smart-vector to save memory
							if !pol.IsComposite() {
								// in this case, the right vector was the root so we are done
								continue
							}

							if shifted, isShifted := pol.(column.Shifted); isShifted {
								polName := pol.GetColID()
								res := sv.SoftRotate(reevaledRoot, shifted.Offset)
								lock.Lock()
								computedReeval[polName] = res
								lock.Unlock()
								continue
							}

						}

						name := pol.GetColID()
						lock.Lock()
						_, ok := computedReeval[name]
						lock.Unlock()
						if ok {
							continue
						}

						if _, ok := coeffs[name]; !ok {
							utils.Panic("handle %v not found in the coeffs\n", name)
						}
						res := sv.FFT(coeffs[name], fft.DIT, false, ratio, share)
						lock.Lock()
						computedReeval[name] = res
						lock.Unlock()
					}
				})

				stopTimer()

				// Force the GC to run
				tGc := time.Now()
				runtime.GC()
				totalTimeGc += time.Since(tGc).Milliseconds()
				logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc), totalTimeGc)

				stopTimer = profiling.LogTimer("Batch evaluation of %v pols of size %v (ratio is %v)", len(handles), ctx.domainSize, ratio)

				// Evaluates the constraint expression on the coset
				evalInputs := make([]sv.SmartVector, len(metadatas))
				for k, metadataInterface := range metadatas {
					switch metadata := metadataInterface.(type) {
					case ifaces.Column:
						name := metadata.GetColID()
						evalInputs[k] = computedReeval[name]
					case coin.Info:
						x := run.GetRandomCoinField(metadata.Name)
						evalInputs[k] = sv.NewConstant(x, ctx.domainSize)
					case variables.X:
						evalInputs[k] = metadata.EvalCoset(ctx.domainSize, i, maxRatio, true)
					case variables.PeriodicSample:
						evalInputs[k] = metadata.EvalCoset(ctx.domainSize, i, maxRatio, true)
					case *ifaces.Accessor:
						evalInputs[k] = sv.NewConstant(metadata.GetVal(run), ctx.domainSize)
					default:
						utils.Panic("Not a variable type %v", reflect.TypeOf(metadataInterface))
					}
				}

				// Force the GC to run
				tGc = time.Now()
				runtime.GC()
				totalTimeGc += time.Since(tGc).Milliseconds()
				logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc), totalTimeGc)

				// Note that this will panic if the expression contains "no commitment"
				// This should be caught already by the constructor of the constraint.
				quotientShare := ctx.aggregateExpressionBoard[j].Evaluate(evalInputs)
				quotientShare = sv.ScalarMul(quotientShare, annulatorInvVals[i])
				run.AssignColumn(ctx.quotientShareHandles[j][share].GetColID(), quotientShare)

				stopTimer()

			}

			// Forcefuly clean the memory for the computed reevals
			for colname := range computedReeval {
				computedReeval[colname] = nil
			}

			// Force the GC to run
			tGc := time.Now()
			runtime.GC()
			totalTimeGc += time.Since(tGc).Milliseconds()
			logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc), totalTimeGc)
		}

	}
}

/*
Prover step where he evaluates all polynomials at a random point.
*/
func (ctx *pocGlobalCtx) proverEvaluationQueries(comp *wizard.CompiledIOP) wizard.ProverStep {
	return func(run *wizard.ProverRuntime) {

		stoptimer := profiling.LogTimer("Evaluate the queries for the global constraints")
		r := run.GetRandomCoinField(ctx.evaluationRandomness.Name)

		// Compute the evaluations
		ys := make([]field.Element, len(ctx.allInvolvedHandles))

		parallel.ExecuteChunky(len(ctx.allInvolvedHandles), func(start, stop int) {
			for k := start; k < stop; k++ {
				handle := ctx.allInvolvedHandles[k]
				witness := handle.GetColAssignment(run)
				ys[k] = sv.Interpolate(witness, r) // numcpus = 8 is empirical
			}
		})

		run.AssignUnivariate(ctx.allHandlesEval.QueryID, r, ys...)

		/*
			For the quotient evaluate it on `x = r / g`, where g is the coset
			shift. The generation of the domain is memoized.
		*/
		x := fft.NewDomain(ctx.domainSize * utils.Max(ctx.ratioIndexes...)).FrMultiplicativeGenInv
		x.Mul(&x, &r)

		quotientYs := make([]field.Element, len(ctx.quotientHandles))

		parallel.Execute(len(ctx.quotientHandles), func(start, stop int) {
			for i := start; i < stop; i++ {
				quotient := ctx.quotientHandles[i].GetColAssignment(run)
				quotientYs[i] = sv.Interpolate(quotient, x)

			}
		})

		/*
			as we shifted the evaluation point. No need to do do coset evaluation
			here
		*/
		run.AssignUnivariate(ctx.quotientsEval.QueryID, x, quotientYs...)
		stoptimer()
	}
}

// Verifier step, evaluate the constraint and checks that
func (ctx *pocGlobalCtx) verifierStep(run *wizard.VerifierRuntime) error {

	// Will be assigned to "X", the random point at which we check the constraint.
	r := run.GetRandomCoinField(ctx.evaluationRandomness.Name)

	// Map all the evaluations and checks the evaluations points
	mapYs := make(map[ifaces.ColID]field.Element)

	// Get the parameters
	params := run.GetUnivariateParams(ctx.allHandlesEval.QueryID)
	univQuery := run.GetUnivariateEval(ctx.allHandlesEval.QueryID)

	// Check the evaluation point is consistent with r
	if params.X != r {
		return fmt.Errorf("(verifier of global queries) : Evaluation point of %v is incorrect (%v, expected %v)",
			ctx.allHandlesEval.QueryID, params.X.String(), r.String())
	}

	// Collect the evaluation points
	for j, handle := range univQuery.Pols {
		mapYs[handle.GetColID()] = params.Ys[j]
	}

	// Annulator = X^n - 1, common for all ratios
	one := field.One()
	annulator := r
	annulator.Exp(annulator, big.NewInt(int64(ctx.domainSize)))
	annulator.Sub(&annulator, &one)

	for i, ratio := range ctx.ratioIndexes {

		board := ctx.aggregateExpressionBoard[i]
		metadatas := board.ListVariableMetadata()

		evalInputs := make([]sv.SmartVector, len(metadatas))

		for k, metadataInterface := range metadatas {
			switch metadata := metadataInterface.(type) {
			case ifaces.Column:
				evalInputs[k] = sv.NewConstant(mapYs[metadata.GetColID()], 1)
			case coin.Info:
				evalInputs[k] = sv.NewConstant(run.GetRandomCoinField(metadata.Name), 1)
			case variables.X:
				evalInputs[k] = sv.NewConstant(r, 1)
			case variables.PeriodicSample:
				evalInputs[k] = sv.NewConstant(metadata.EvalAtOutOfDomain(ctx.domainSize, r), 1)
			case *ifaces.Accessor:
				evalInputs[k] = sv.NewConstant(metadata.GetVal(run), 1)
			default:
				utils.Panic("Not a variable type %v in global query (ratio %v)", reflect.TypeOf(metadataInterface), ratio)
			}
		}

		left := board.Evaluate(evalInputs).Get(0)

		// right : r^{n}-1 Q(r)
		qr := run.QueriesParams.MustGet(ctx.quotientsEval.QueryID).(query.UnivariateEvalParams).Ys[i]
		var right field.Element
		right.Mul(&annulator, &qr)

		if left != right {
			return fmt.Errorf("global constraint %v - ratio %v - mismatch at random point - %v != %v", ctx.selfRecursionCounter, ratio, left.String(), right.String())
		}
	}

	return nil
}

// Verifier step, evaluate the constraint and checks that
func (ctx *pocGlobalCtx) gnarkVerifierStep(api frontend.API, c *wizard.WizardVerifierCircuit) {

	// Will be assigned to "X", the random point at which we check the constraint.
	r := c.GetRandomCoinField(ctx.evaluationRandomness.Name)
	annulator := gnarkutil.Exp(api, r, ctx.domainSize)
	annulator = api.Sub(annulator, frontend.Variable(1))

	// Get the parameters
	params := c.GetUnivariateParams(ctx.allHandlesEval.QueryID)
	univQuery := c.GetUnivariateEval(ctx.allHandlesEval.QueryID)
	api.AssertIsEqual(r, params.X) // check the evaluation is consistent with the other stuffs

	// Map all the evaluations and checks the evaluations points
	mapYs := make(map[ifaces.ColID]frontend.Variable)

	// Collect the evaluation points
	for j, handle := range univQuery.Pols {
		mapYs[handle.GetColID()] = params.Ys[j]
	}

	for i, ratio := range ctx.ratioIndexes {

		board := ctx.aggregateExpressionBoard[i]
		metadatas := board.ListVariableMetadata()

		evalInputs := make([]frontend.Variable, len(metadatas))

		for k, metadataInterface := range metadatas {
			switch metadata := metadataInterface.(type) {
			case ifaces.Column:
				evalInputs[k] = mapYs[metadata.GetColID()]
			case coin.Info:
				evalInputs[k] = c.GetRandomCoinField(metadata.Name)
			case variables.X:
				evalInputs[k] = r
			case variables.PeriodicSample:
				evalInputs[k] = metadata.GnarkEvalAtOutOfDomain(api, ctx.domainSize, r)
			case *ifaces.Accessor:
				evalInputs[k] = metadata.GetFrontendVariable(api, c)
			default:
				utils.Panic("Not a variable type %v in global query (ratio %v)", reflect.TypeOf(metadataInterface), ratio)
			}
		}

		left := board.GnarkEval(api, evalInputs)

		// right : r^{n}-1 Q(r)
		ys := c.GetUnivariateParams(ctx.quotientsEval.QueryID).Ys
		qr := ys[i]
		right := api.Mul(annulator, qr)

		api.AssertIsEqual(left, right)
		logrus.Debugf("verifying global constraint : DONE")

	}
}

/*
Returns Size if its a commitment, 1 if it is x and zero else.
*/
func GetDegree(size int) func(iface interface{}) int {
	return func(iface interface{}) int {
		switch v := iface.(type) {
		case ifaces.Column:
			// Univariate polynomials is X. We pad them with zeroes so it is safe
			// to return the domainSize directly.
			if size != v.Size() {
				panic("unconsistent sizes for the commitments")
			}
			// The size gives the number of coefficients , but we return the degree
			// hence the - 1
			return v.Size() - 1
		case coin.Info, *ifaces.Accessor:
			// Coins are treated
			return 0
		case variables.X:
			return 1
		case variables.PeriodicSample:
			return size - size/v.T
		default:
			utils.Panic("Unknown type %v\n", reflect.TypeOf(v))
		}
		panic("unreachable")
	}
}

func (ctx *pocGlobalCtx) derivename(s string, args ...any) string {
	fmts := fmt.Sprintf(s, args...)
	return fmt.Sprintf("%v_%v_%v", GLOBAL_REDUCTION, ctx.comp.SelfRecursionCount, fmts)
}

package arithmetics

import (
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/mempool"
	sv "github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/fft"
	"github.com/consensys/zkevm-monorepo/prover/maths/fft/fastpoly"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/query"
	"github.com/consensys/zkevm-monorepo/prover/protocol/variables"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/collection"
	"github.com/consensys/zkevm-monorepo/prover/utils/gnarkutil"
	"github.com/consensys/zkevm-monorepo/prover/utils/parallel"
	"github.com/consensys/zkevm-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

const (
	GLOBAL_REDUCTION                string = "GLOBAL_REDUCTION"
	OFFSET_RANDOMNESS               string = "OFFSET_RANDOMNESS"
	DEGREE_RANDOMNESS               string = "DEGREE_RANDOMNESS"
	QUOTIENT_POLY_TMPL              string = "QUOTIENT_DEG_%v_SHARE_%v"
	EVALUATION_RANDOMESS            string = "EVALUATION_RANDOMNESS"
	UNIVARIATE_EVAL_ALL_HANDLES     string = "UNIV_EVAL_ALL_HANDLES"
	UNIVARIATE_EVAL_QUOTIENT_SHARES string = "UNIV_EVAL_QUOTIENT_%v_OVER_%v"

	/*
		Explanation for Manual Garbage Collection Thresholds
	*/
	// These two threshold work well for the real-world traces at the moment of writing and a 340GiB memory limit,
	// but this approach can be generalized and further improved.

	// When ctx.domainSize>=524288, proverEvaluationQueries() experiences a heavy workload,
	// consistently hitting the GOMEMLIMIT of 340GiB.
	// This results in numerous auto GCs during CPU-intensive small tasks, significantly degrading performance.
	// In the benchmark input files, GC_DOMAIN_SIZE >= 524288 means only the first call of proverEvaluationQueries().
	// With ctx.domainSize<=262144, manual GC is not necessary as auto GCs triggered by GOMEMLIMIT suffice.
	GC_DOMAIN_SIZE int = 524288

	// Auto GCs are triggered during ReEvaluate and Batch evaluation
	// when len(handles) exceeds approximately 4000, causing performance degradation.
	// This threshold is set to perform manual GCs before ReEvaluate and Batch evaluation
	// only when len(handles) reaches a size substantial enough to trigger auto GC during ReEvaluate and Batch evaluation.
	// Note that the value of GC_HANDLES_SIZE 4000 is derived from experience and analytics on the benchmark input files.
	GC_HANDLES_SIZE int = 4000
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
	ctx.declareUnivOnQuotient()

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
	maxNbNodesPerExpressions int
	// Size of the domain expressions
	domainSize int

	handlePerRatio     [][]ifaces.Column
	rootPerRatio       [][]ifaces.Column
	allInvolvedHandles []ifaces.Column
	allInvolvedRoots   []ifaces.Column

	quotientShareHandles [][]ifaces.Column

	// Coins
	offsetRandomness, degreeRandomness, evaluationRandomness coin.Info
	// UnivariateQueries
	allHandlesEval query.UnivariateEval
	quotientsEval  []query.UnivariateEval
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
	maxNodeCount := 0

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

		// Tracks the largest node count in all aggregated expressions
		maxNodeCount = utils.Max(maxNodeCount, board.CountNodes())
	}

	ctx.maxNbNodesPerExpressions = maxNodeCount
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
	ctx.allInvolvedRoots = allInvolvedRoots
}

func (ctx *pocGlobalCtx) gatherQuotientHandles(comp *wizard.CompiledIOP) {
	ctx.quotientShareHandles = make([][]ifaces.Column, len(ctx.rangeIndexes))
	for i, ratio := range ctx.ratioIndexes {
		shareHandles := make([]ifaces.Column, ratio)
		for share := range shareHandles {
			shareHandles[share] = comp.Columns.GetHandle(
				ifaces.ColID(ctx.derivename(quotientShareName(share, ratio))),
			)
		}
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
		pool := mempool.Create(symbolic.MaxChunkSize).Prewarm(runtime.NumCPU() * ctx.maxNbNodesPerExpressions)

		if ctx.domainSize >= GC_DOMAIN_SIZE {
			// Force the GC to run
			tGc := time.Now()
			runtime.GC()
			totalTimeGc += time.Since(tGc).Milliseconds()
			logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc), totalTimeGc)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
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
			wg.Done()
		}()

		go func() {
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
			wg.Done()
		}()

		wg.Wait()
		stopTimer()

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

			// use sync map
			computedReeval := sync.Map{}

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

				if ctx.domainSize >= GC_DOMAIN_SIZE {
					// Force the GC to run
					tGc := time.Now()
					runtime.GC()
					totalTimeGc += time.Since(tGc).Milliseconds()
					logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc), totalTimeGc)
				}

				stopTimer := profiling.LogTimer("ReEvaluate %v pols of size %v on coset %v/%v", len(handles), ctx.domainSize, share, ratio)

				parallel.ExecuteChunky(len(roots), func(start, stop int) {
					for k := start; k < stop; k++ {
						root := roots[k]
						name := root.GetColID()

						_, found := computedReeval.Load(name)

						if found {
							// it was already computed in a previous iteration of `j`
							continue
						}

						// else it's the first value of j that sees it. so we compute the
						// coset reevaluation.
						reevaledRoot := sv.FFT(coeffs[name], fft.DIT, false, ratio, share)
						computedReeval.Store(name, reevaledRoot)
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

							reevaledRoot, found := computedReeval.Load(name)

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
								res := sv.SoftRotate(reevaledRoot.(sv.SmartVector), shifted.Offset)
								computedReeval.Store(polName, res)
								continue
							}

						}

						name := pol.GetColID()
						_, ok := computedReeval.Load(name)
						if ok {
							continue
						}

						if _, ok := coeffs[name]; !ok {
							utils.Panic("handle %v not found in the coeffs\n", name)
						}
						res := sv.FFT(coeffs[name], fft.DIT, false, ratio, share)
						computedReeval.Store(name, res)
					}
				})

				stopTimer()

				if len(handles) >= GC_HANDLES_SIZE {
					// Force the GC to run
					tGc := time.Now()
					runtime.GC()
					totalTimeGc += time.Since(tGc).Milliseconds()
					logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc), totalTimeGc)
				}

				stopTimer = profiling.LogTimer("Batch evaluation of %v pols of size %v (ratio is %v)", len(handles), ctx.domainSize, ratio)

				// Evaluates the constraint expression on the coset
				evalInputs := make([]sv.SmartVector, len(metadatas))
				for k, metadataInterface := range metadatas {
					switch metadata := metadataInterface.(type) {
					case ifaces.Column:
						//name := metadata.GetColID()
						//evalInputs[k] = computedReeval[name]
						value, _ := computedReeval.Load(metadata.GetColID())
						evalInputs[k] = value.(sv.SmartVector)
					case coin.Info:
						evalInputs[k] = sv.NewConstant(run.GetRandomCoinField(metadata.Name), ctx.domainSize)
					case variables.X:
						evalInputs[k] = metadata.EvalCoset(ctx.domainSize, i, maxRatio, true)
					case variables.PeriodicSample:
						evalInputs[k] = metadata.EvalCoset(ctx.domainSize, i, maxRatio, true)
					case ifaces.Accessor:
						evalInputs[k] = sv.NewConstant(metadata.GetVal(run), ctx.domainSize)
					default:
						utils.Panic("Not a variable type %v", reflect.TypeOf(metadataInterface))
					}
				}

				if len(handles) >= GC_HANDLES_SIZE {
					// Force the GC to run
					tGc := time.Now()
					runtime.GC()
					totalTimeGc += time.Since(tGc).Milliseconds()
					logrus.Infof("global constraints : spent %v ms in gc, total time %v ms", time.Since(tGc), totalTimeGc)
				}

				// Note that this will panic if the expression contains "no commitment"
				// This should be caught already by the constructor of the constraint.
				quotientShare := ctx.aggregateExpressionBoard[j].Evaluate(evalInputs, pool)
				quotientShare = sv.ScalarMul(quotientShare, annulatorInvVals[i])
				run.AssignColumn(ctx.quotientShareHandles[j][share].GetColID(), quotientShare)

				stopTimer()

			}

			// Forcefuly clean the memory for the computed reevals
			computedReeval.Range(func(k, v interface{}) bool {
				computedReeval.Delete(k)
				return true
			})
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

		witnesses := make([]sv.SmartVector, len(ctx.allInvolvedHandles))
		// Compute the evaluations
		parallel.Execute(len(ctx.allInvolvedHandles), func(start, stop int) {
			for i := start; i < stop; i++ {
				handle := ctx.allInvolvedHandles[i]
				witness := handle.GetColAssignment(run)
				witnesses[i] = witness

			}
		})

		ys := sv.BatchInterpolate(witnesses, r)
		run.AssignUnivariate(ctx.allHandlesEval.QueryID, r, ys...)

		/*
			For the quotient evaluate it on `x = r / g`, where g is the coset
			shift. The generation of the domain is memoized.
		*/

		var (
			maxRatio          = utils.Max(ctx.ratioIndexes...)
			mulGenInv         = fft.NewDomain(maxRatio * ctx.domainSize).FrMultiplicativeGenInv
			rootInv           = fft.GetOmega(maxRatio * ctx.domainSize)
			quotientEvalPoint field.Element
			wg                = &sync.WaitGroup{}
		)

		rootInv.Inverse(&rootInv)
		quotientEvalPoint.Mul(&mulGenInv, &r)

		for i := range ctx.quotientsEval {
			wg.Add(1)
			go func(i int, evalPoint field.Element) {
				var (
					q  = ctx.quotientsEval[i]
					ys = make([]field.Element, len(q.Pols))
				)

				parallel.Execute(len(q.Pols), func(start, stop int) {
					for i := start; i < stop; i++ {
						c := q.Pols[i].GetColAssignment(run)
						ys[i] = sv.Interpolate(c, evalPoint)
					}
				})

				run.AssignUnivariate(q.Name(), evalPoint, ys...)
				wg.Done()
			}(i, quotientEvalPoint)
			quotientEvalPoint.Mul(&quotientEvalPoint, &rootInv)
		}

		wg.Wait()

		/*
			as we shifted the evaluation point. No need to do do coset evaluation
			here
		*/
		stoptimer()
	}
}

// Verifier step, evaluate the constraint and checks that
func (ctx *pocGlobalCtx) verifierStep(run *wizard.VerifierRuntime) error {

	var (
		// Will be assigned to "X", the random point at which we check the constraint.
		r = run.GetRandomCoinField(ctx.evaluationRandomness.Name)
		// Map all the evaluations and checks the evaluations points
		mapYs = make(map[ifaces.ColID]field.Element)
		// Get the parameters
		params           = run.GetUnivariateParams(ctx.allHandlesEval.QueryID)
		univQuery        = run.GetUnivariateEval(ctx.allHandlesEval.QueryID)
		quotientYs, errQ = ctx.recombineQuotientSharesEvaluation(run, r)
	)

	if errQ != nil {
		return fmt.Errorf("invalid evaluation point for the quotients: %v", errQ.Error())
	}

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
			case ifaces.Accessor:
				evalInputs[k] = sv.NewConstant(metadata.GetVal(run), 1)
			default:
				utils.Panic("Not a variable type %v in global query (ratio %v)", reflect.TypeOf(metadataInterface), ratio)
			}
		}

		left := board.Evaluate(evalInputs).Get(0)

		// right : r^{n}-1 Q(r)
		qr := quotientYs[i]
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
	quotientYs := ctx.recombineQuotientSharesEvaluationGnark(api, c, r)

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
			case ifaces.Accessor:
				evalInputs[k] = metadata.GetFrontendVariable(api, c)
			default:
				utils.Panic("Not a variable type %v in global query (ratio %v)", reflect.TypeOf(metadataInterface), ratio)
			}
		}

		left := board.GnarkEval(api, evalInputs)

		// right : r^{n}-1 Q(r)
		qr := quotientYs[i]
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
		case coin.Info, ifaces.Accessor:
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

// declareUnivOnQuotient declares the univariate queries over all the quotient
// shares, making sure that the shares needing to be evaluated over the same
// point are in the same query. This is a req from the upcoming naturalization
// compiler.
func (ctx *pocGlobalCtx) declareUnivOnQuotient() {

	var (
		round       = ctx.quotientShareHandles[0][0].Round()
		maxRatio    = utils.Max(ctx.ratioIndexes...)
		queriesPols = make([][]ifaces.Column, maxRatio)
	)

	for i, ratio := range ctx.ratioIndexes {
		var (
			jumpBy = maxRatio / ratio
		)
		for j := range ctx.quotientShareHandles[i] {
			queriesPols[j*jumpBy] = append(queriesPols[j*jumpBy], ctx.quotientShareHandles[i][j])
		}
	}

	ctx.quotientsEval = make([]query.UnivariateEval, maxRatio)

	for i := range queriesPols {
		ctx.quotientsEval[i] = ctx.comp.InsertUnivariate(
			round+1,
			ifaces.QueryID(ctx.derivename(UNIVARIATE_EVAL_QUOTIENT_SHARES, i, maxRatio)),
			queriesPols[i],
		)
	}
}

// recombineQuotientSharesEvaluation returns the evaluations of the quotients
// on point r
func (ctx pocGlobalCtx) recombineQuotientSharesEvaluation(run *wizard.VerifierRuntime, r field.Element) ([]field.Element, error) {

	var (
		// res stores the list of the recombined quotient evaluations for each
		// combination.
		recombinedYs = make([]field.Element, len(ctx.ratioIndexes))
		// ys stores the values of the quotient shares ordered by ratio
		qYs      = make([][]field.Element, utils.Max(ctx.ratioIndexes...))
		maxRatio = utils.Max(ctx.ratioIndexes...)
		// shiftedR = r / g where g is the generator of the multiplicative group
		shiftedR field.Element
		// mulGen is the generator of the multiplicative group
		mulGenInv = fft.NewDomain(maxRatio * ctx.domainSize).FrMultiplicativeGenInv
		// omegaN is a root of unity generating the domain of size `domainSize
		// * maxRatio`
		omegaN = fft.GetOmega(ctx.domainSize * maxRatio)
	)

	shiftedR.Mul(&r, &mulGenInv)

	for i, q := range ctx.quotientsEval {
		params := run.GetUnivariateParams(q.Name())
		qYs[i] = params.Ys

		// Check that the provided value for x is the right one
		providedX := params.X
		var expectedX field.Element
		expectedX.Inverse(&omegaN)
		expectedX.Exp(expectedX, big.NewInt(int64(i)))
		expectedX.Mul(&expectedX, &shiftedR)
		if providedX != expectedX {
			return nil, fmt.Errorf("bad X value")
		}
	}

	for i, ratio := range ctx.ratioIndexes {
		var (
			jumpBy = maxRatio / ratio
			ys     = make([]field.Element, ratio)
		)

		for j := range ctx.quotientShareHandles[i] {
			ys[j] = qYs[j*jumpBy][0]
			qYs[j*jumpBy] = qYs[j*jumpBy][1:]
		}

		var (
			m          = ctx.domainSize
			n          = ctx.domainSize * ratio
			omegaRatio = fft.GetOmega(ratio)
			rPowM      field.Element
			// outerFactor stores m/n*(r^n - 1)
			outerFactor   = shiftedR
			one           = field.One()
			omegaRatioInv field.Element
			res           field.Element
			ratioInvField = field.NewElement(uint64(ratio))
		)

		rPowM.Exp(shiftedR, big.NewInt(int64(m)))
		ratioInvField.Inverse(&ratioInvField)
		omegaRatioInv.Inverse(&omegaRatio)

		for k := range ys {

			// tmp stores ys[k] / ((r^m / omegaRatio^k) - 1)
			var tmp field.Element
			tmp.Exp(omegaRatioInv, big.NewInt(int64(k)))
			tmp.Mul(&tmp, &rPowM)
			tmp.Sub(&tmp, &one)
			tmp.Div(&ys[k], &tmp)

			res.Add(&res, &tmp)
		}

		outerFactor.Exp(shiftedR, big.NewInt(int64(n)))
		outerFactor.Sub(&outerFactor, &one)
		outerFactor.Mul(&outerFactor, &ratioInvField)
		res.Mul(&res, &outerFactor)
		recombinedYs[i] = res
	}

	return recombinedYs, nil
}

// recombineQuotientSharesEvaluation returns the evaluations of the quotients
// on point r
func (ctx pocGlobalCtx) recombineQuotientSharesEvaluationGnark(api frontend.API, run *wizard.WizardVerifierCircuit, r frontend.Variable) []frontend.Variable {

	var (
		// res stores the list of the recombined quotient evaluations for each
		// combination.
		recombinedYs = make([]frontend.Variable, len(ctx.ratioIndexes))
		// ys stores the values of the quotient shares ordered by ratio
		qYs      = make([][]frontend.Variable, utils.Max(ctx.ratioIndexes...))
		maxRatio = utils.Max(ctx.ratioIndexes...)
		// shiftedR = r / g where g is the generator of the multiplicative group
		shiftedR frontend.Variable
		// mulGen is the generator of the multiplicative group
		mulGenInv = fft.NewDomain(maxRatio * ctx.domainSize).FrMultiplicativeGenInv
		// omegaN is a root of unity generating the domain of size `domainSize
		// * maxRatio`
		omegaN = fft.GetOmega(ctx.domainSize * maxRatio)
	)

	shiftedR = api.Mul(r, mulGenInv)

	for i, q := range ctx.quotientsEval {
		params := run.GetUnivariateParams(q.Name())
		qYs[i] = params.Ys

		// Check that the provided value for x is the right one
		providedX := params.X
		var expectedX frontend.Variable
		expectedX = api.Inverse(omegaN)
		expectedX = gnarkutil.Exp(api, expectedX, i)
		expectedX = api.Mul(expectedX, shiftedR)
		api.AssertIsEqual(providedX, expectedX)
	}

	for i, ratio := range ctx.ratioIndexes {
		var (
			jumpBy = maxRatio / ratio
			ys     = make([]frontend.Variable, ratio)
		)

		for j := range ctx.quotientShareHandles[i] {
			ys[j] = qYs[j*jumpBy][0]
			qYs[j*jumpBy] = qYs[j*jumpBy][1:]
		}

		var (
			m          = ctx.domainSize
			n          = ctx.domainSize * ratio
			omegaRatio = fft.GetOmega(ratio)
			// outerFactor stores m/n*(r^n - 1)
			one           = field.One()
			omegaRatioInv field.Element
			res           = frontend.Variable(0)
			ratioInvField = field.NewElement(uint64(ratio))
		)

		rPowM := gnarkutil.Exp(api, shiftedR, m)
		ratioInvField.Inverse(&ratioInvField)
		omegaRatioInv.Inverse(&omegaRatio)

		for k := range ys {

			// tmp stores ys[k] / ((r^m / omegaRatio^k) - 1)
			var omegaInvPowK field.Element
			omegaInvPowK.Exp(omegaRatioInv, big.NewInt(int64(k)))
			tmp := api.Mul(omegaInvPowK, rPowM)
			tmp = api.Sub(tmp, one)
			tmp = api.Div(ys[k], tmp)

			res = api.Add(res, tmp)
		}

		outerFactor := gnarkutil.Exp(api, shiftedR, n)
		outerFactor = api.Sub(outerFactor, one)
		outerFactor = api.Mul(outerFactor, ratioInvField)
		res = api.Mul(res, outerFactor)
		recombinedYs[i] = res
	}

	return recombinedYs

}

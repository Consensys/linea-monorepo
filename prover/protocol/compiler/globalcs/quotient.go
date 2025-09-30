package globalcs

import (
	"math/big"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"golang.org/x/sync/singleflight"

	"github.com/sirupsen/logrus"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	ppool "github.com/consensys/linea-monorepo/prover/utils/parallel/pool"
)

const (
	/*
		Explanation for Manual Garbage Collection Thresholds
	*/
	// These two thresholds work well for the real-world traces at the moment of writing and a 340GiB memory limit,
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

// QuotientCtx collects all the internal fields needed to compute the quotient
type QuotientCtx[T zk.Element] struct {

	// DomainSize is the domain over which the global constraints are computed
	DomainSize int

	// Ratio lists the ratio found in the global constraints
	//
	// See [mergingCtx.Ratios]
	Ratios []int

	// ColumnsForRatio[k] stores all the columns involved in the aggregate
	// expressions for ratio Ratios[k]
	ColumnsForRatio [][]ifaces.Column[T]

	// RootsPerRatio[k] stores all the root columns involved in the aggregate
	// expressions for ration Ratios[k]. By root column we mean the underlying
	// column that are actually committed to. For instance, if Shift(A, 1) is
	// in ColumnsForRatio[k], we will have A in RootPerRatio[k]
	RootsForRatio [][]ifaces.Column[T]

	// AllInvolvedColumns stores the union of the ColumnForRatio[k] for all k
	AllInvolvedColumns []ifaces.Column[T]

	// AllInvolvedRoots stores the union of the RootsForRatio[k] for all k
	AllInvolvedRoots []ifaces.Column[T]

	// AggregateExpressions[k] stores the aggregate expression for Ratios[k]
	AggregateExpressions []*symbolic.Expression[T]

	// AggregateExpressionsBoard[k] stores the topological sorting of
	// AggregateExpressions[k]
	AggregateExpressionsBoard []symbolic.ExpressionBoard[T]

	// QuotientShares[k] stores for each k, the list of the Ratios[k] shares
	// of the quotient for the AggregateExpression[k]
	QuotientShares [][]ifaces.Column[T]

	// MaxNbExprNode stores the largest number of node AggregateExpressionBoard[*]
	// has. This is used to dimension the memory pool during the prover time.
	MaxNbExprNode int
}

// createQuotientCtx constructs a [quotientCtx] from a list of ratios and aggregated
// expressions. The function organizes the handles but does not declare anything
// in the current wizard.CompiledIOP.
func createQuotientCtx[T zk.Element](comp *wizard.CompiledIOP, ratios []int, aggregateExpressions []*symbolic.Expression[T]) QuotientCtx[T] {

	var (
		allInvolvedHandlesIndex = map[ifaces.ColID]int{}
		allInvolvedRootsSet     = collection.NewSet[ifaces.ColID]()
		_, _, domainSize        = wizardutils.AsExpr(aggregateExpressions[0])
		ctx                     = QuotientCtx[T]{
			DomainSize:                domainSize,
			Ratios:                    ratios,
			AggregateExpressions:      aggregateExpressions,
			AggregateExpressionsBoard: make([]symbolic.ExpressionBoard[T], len(ratios)),
			AllInvolvedColumns:        []ifaces.Column[T]{},
			AllInvolvedRoots:          []ifaces.Column[T]{},
			ColumnsForRatio:           make([][]ifaces.Column[T], len(ratios)),
			RootsForRatio:             make([][]ifaces.Column[T], len(ratios)),
			QuotientShares:            generateQuotientShares[T](comp, ratios, domainSize),
		}
	)

	for k, expr := range ctx.AggregateExpressions {

		var (
			board               = expr.Board()
			uniqueRootsForRatio = collection.NewSet[ifaces.ColID]()
		)

		ctx.AggregateExpressionsBoard[k] = board
		ctx.MaxNbExprNode = max(ctx.MaxNbExprNode, board.CountNodes())

		// This loop scans the metadata looking for columns with the goal of
		// populating the collections composing quotientCtx.
		for _, metadata := range board.ListVariableMetadata() {

			// Scan in column metadata only
			col, ok := metadata.(ifaces.Column[T])
			if !ok {
				continue
			}

			var (
				rootCol = column.RootParents(col)
			)

			// Append the handle (we trust that there are no duplicate of handles
			// within a constraint). This works because the symbolic package has
			// automatic simplification routines that ensure that an expression
			// does not refer to duplicates of the same variable.
			ctx.ColumnsForRatio[k] = append(ctx.ColumnsForRatio[k], col)

			if !uniqueRootsForRatio.Exists(rootCol.GetColID()) {
				ctx.RootsForRatio[k] = append(ctx.RootsForRatio[k], rootCol)
			}

			// Get the name of the
			if _, alreadyThere := allInvolvedHandlesIndex[col.GetColID()]; alreadyThere {
				continue
			}

			allInvolvedHandlesIndex[col.GetColID()] = len(ctx.AllInvolvedColumns)
			ctx.AllInvolvedColumns = append(ctx.AllInvolvedColumns, col)

			// If the handle is simply a shift or a natural columns tracks its root
			if !allInvolvedRootsSet.Exists(rootCol.GetColID()) {
				allInvolvedRootsSet.Insert(rootCol.GetColID())
				ctx.AllInvolvedRoots = append(ctx.AllInvolvedRoots, rootCol)
			}
		}
	}

	return ctx
}

// generateQuotientShares declares and returns the quotient share columns
func generateQuotientShares[T zk.Element](comp *wizard.CompiledIOP, ratios []int, domainSize int) [][]ifaces.Column[T] {

	var (
		quotientShares = make([][]ifaces.Column[T], len(ratios))
		currRound      = comp.NumRounds() - 1
	)

	for i, ratio := range ratios {
		quotientShares[i] = make([]ifaces.Column[T], ratio)
		for k := range quotientShares[i] {
			quotientShares[i][k] = comp.InsertCommit(
				currRound,
				ifaces.ColID(deriveName(comp, QUOTIENT_POLY_TMPL, ratio, k)),
				domainSize,
			)
		}
	}

	return quotientShares
}

// Run implements the [wizard.ProverAction] interface and embeds the logic to
// compute the quotient shares.
func (ctx *QuotientCtx[T]) Run(run *wizard.ProverRuntime) {

	// Initial step is to compute the FFTs for all committed vectors
	coeffs := sync.Map{} // (ifaces.ColID <=> sv.SmartVector)

	// Compute once the FFT of the natural columns

	domain0 := fft.NewDomain(uint64(ctx.DomainSize), fft.WithCache())
	parallel.Execute(len(ctx.AllInvolvedRoots), func(start, stop int) {

		for k := start; k < stop; k++ {
			pol := ctx.AllInvolvedRoots[k]
			name := pol.GetColID()

			// gets directly a shallow copy in the map of the runtime
			var witness sv.SmartVector
			witness, isAssigned := run.Columns.TryGet(name)

			// can happen if the column is verifier defined. In that case, no
			// need to protect with a lock. This will not touch run.Columns.
			if !isAssigned {
				witness = pol.GetColAssignment(run)
			}

			if smartvectors.IsBase(witness) {
				res := make([]field.Element, witness.Len())
				witness.WriteInSlice(res)
				domain0.FFTInverse(res, fft.DIF, fft.WithNbTasks(1))
				coeffs.Store(name, smartvectors.NewRegular(res))
				continue
			}

			res := make([]fext.Element, witness.Len())
			witness.WriteInSliceExt(res)
			domain0.FFTInverseExt(res, fft.DIF, fft.WithNbTasks(1))
			coeffs.Store(name, smartvectors.NewRegularExt(res))
		}

	})

	// Take the max quotient degree
	maxRatio := utils.Max(ctx.Ratios...)

	/*
		For the quotient, we precompute the values of (wQ^N - 1)^-1 for w in H, the
		larger domain.

		Those values are D-periodic, thus we only compute a single period.
		(Where D is the ratio of the sizes of the larger and the smaller domain)

		The first value is ignored because it correspond to the case where w^N = 1
		(i.e. w is in the smaller subgroup)
	*/
	annulatorInvVals := fastpolyext.EvalXnMinusOneOnACoset(ctx.DomainSize, ctx.DomainSize*maxRatio)
	annulatorInvVals = fext.ParBatchInvert(annulatorInvVals, runtime.GOMAXPROCS(0))

	// this space we allocate, always to avoid relying on a pool with syncs, and fragmented memory
	// or too many allocs.
	scratch := make([]fext.Element, len(ctx.AllInvolvedRoots)*ctx.DomainSize)

	for i := 0; i < maxRatio; i++ {
		// we use Atomic add on this to determine the next free slot in our scratch space
		var scratchOffset int64

		// use sync map to store the coset evaluated polynomials
		var group singleflight.Group
		computedReeval := sync.Map{} // (ifaces.ColID <=> sv.SmartVector)

		for j, ratio := range ctx.Ratios {

			// For instance, if deg = 2 and max deg 8, we enter only if
			// i = 0 or 4 because this corresponds to the cosets we are
			// interested in.
			if i%(maxRatio/ratio) != 0 {
				continue
			}

			// With the above example, if we are in the ratio = 2 and maxRatio = 8
			// and i = 1 (it can only be 0 <= i < ratio).
			var (
				share   = i * ratio / maxRatio
				handles = ctx.ColumnsForRatio[j]
				// roots     = ctx.RootsForRatio[j]
				board     = ctx.AggregateExpressionsBoard[j]
				metadatas = board.ListVariableMetadata()
			)

			shift := computeShift(uint64(ctx.DomainSize), ratio, share)
			domain := fft.NewDomain(uint64(ctx.DomainSize), fft.WithShift(shift), fft.WithCache())

			computeRoot := func(name ifaces.ColID) (any, error) {
				_v, _ := coeffs.Load(name)
				v := _v.(sv.SmartVector)
				n := atomic.AddInt64(&scratchOffset, 1)
				start := (n - 1) * int64(ctx.DomainSize)
				end := n * int64(ctx.DomainSize)
				var reevaledRoot []fext.Element
				if end > int64(len(scratch)) {
					reevaledRoot = make([]fext.Element, ctx.DomainSize)
				} else {
					reevaledRoot = scratch[start:end]
				}

				// TODO @gbotrel coeffs are mostly base vectors;
				// but the code in eval somewhere can't mix well ext and base,
				// we could save additional memory here.
				v.WriteInSliceExt(reevaledRoot)
				domain.FFTExt(reevaledRoot, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))

				res := smartvectors.NewRegularExt(reevaledRoot)
				computedReeval.Store(name, res)

				return res, nil
			}

			ppool.ExecutePoolChunky(len(handles), func(k int) {

				pol := handles[k]
				// short-path, the column is a purely Shifted(Natural) or a Natural
				// (this excludes repeats and/or interleaved columns)
				root := column.RootParents(pol)
				rootName := root.GetColID()

				reevaledRoot, _, _ := group.Do(string(rootName), func() (interface{}, error) { return computeRoot(rootName) })

				// Now, we can reuse a soft-rotation of the smart-vector to save memory
				if !pol.IsComposite() {
					// in this case, the right vector was the root so we are done
					return
				}

				if shifted, isShifted := pol.(column.Shifted[T]); isShifted {
					polName := pol.GetColID()
					res := sv.SoftRotateExt(reevaledRoot.(sv.SmartVector), shifted.Offset)
					computedReeval.Store(polName, res)
					return
				}

				polName := pol.GetColID()
				group.Do(string(polName), func() (interface{}, error) {
					v, _ := coeffs.Load(polName)

					n := uint64(v.(sv.SmartVector).Len())

					res := make([]field.Element, n)
					v.(sv.SmartVector).WriteInSlice(res)
					domain.FFT(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))

					// res := sv.FFT(v.(sv.SmartVector), fft.DIT, false, ratio, share, fft.WithNbTasks(2))
					computedReeval.Store(polName, smartvectors.NewRegular(res))
					return nil, nil
				})
			})

			// Evaluates the constraint expression on the coset
			evalInputs := make([]sv.SmartVector, len(metadatas))

			for k, metadataInterface := range metadatas {

				switch metadata := metadataInterface.(type) {
				case ifaces.Column[T]:
					value, ok := computedReeval.Load(metadata.GetColID())
					if !ok {
						utils.Panic("did not find the reevaluation of %v", metadata.GetColID())
					}
					evalInputs[k] = value.(sv.SmartVector)

				case coin.Info:
					if metadata.IsBase() {
						evalInputs[k] = sv.NewConstant(run.GetRandomCoinField(metadata.Name), ctx.DomainSize)
					} else {
						evalInputs[k] = sv.NewConstantExt(run.GetRandomCoinFieldExt(metadata.Name), ctx.DomainSize)
					}
				case variables.X[T]:
					evalInputs[k] = metadata.EvalCoset(ctx.DomainSize, i, maxRatio, true)
				case variables.PeriodicSample[T]:
					evalInputs[k] = metadata.EvalCoset(ctx.DomainSize, i, maxRatio, true)
				case ifaces.Accessor[T]:
					if metadata.IsBase() {
						evalInputs[k] = sv.NewConstant(metadata.GetVal(run), ctx.DomainSize)
					} else {
						evalInputs[k] = sv.NewConstantExt(metadata.GetValExt(run), ctx.DomainSize)
					}
				default:
					utils.Panic("Not a variable type %v", reflect.TypeOf(metadataInterface))
				}
			}

			// Note that this will panic if the expression contains "no commitment"
			// This should be caught already by the constructor of the constraint.
			quotientShare := ctx.AggregateExpressionsBoard[j].Evaluate(evalInputs)
			if re, ok := quotientShare.(*sv.RegularExt); ok {
				vq := extensions.Vector(*re)
				vq.ScalarMul(vq, &annulatorInvVals[i])
			} else {
				quotientShare = sv.ScalarMulExt(quotientShare, annulatorInvVals[i])
			}

			run.AssignColumn(ctx.QuotientShares[j][share].GetColID(), quotientShare)

		}

		// Forcefully clean the memory for the computed reevals
		computedReeval = sync.Map{}
		group = singleflight.Group{}
	}

	logrus.Infof("[global-constraint] msg=\"computed the quotient\"")

}

func computeShift(n uint64, cosetRatio int, cosetID int) field.Element {
	var shift field.Element
	cardinality := ecc.NextPowerOfTwo(uint64(n))
	frMulGen := fft.GeneratorFullMultiplicativeGroup()
	omega, _ := fft.Generator(cardinality * uint64(cosetRatio))
	omega.Exp(omega, big.NewInt(int64(cosetID)))
	shift.Mul(&frMulGen, &omega)
	return shift
}

package globalcs

import (
	"math/big"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"

	"github.com/sirupsen/logrus"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/fft/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/arena"
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
)

// QuotientCtx collects all the internal fields needed to compute the quotient
type QuotientCtx struct {

	// DomainSize is the domain over which the global constraints are computed
	DomainSize int

	// Ratio lists the ratio found in the global constraints
	//
	// See [mergingCtx.Ratios]
	Ratios []int

	// RootsPerRatio[k] stores all the root columns involved in the aggregate
	// expressions for ration Ratios[k]. By root column we mean the underlying
	// column that are actually committed to. For instance, if Shift(A, 1) is
	// in ColumnsForRatio[k], we will have A in RootPerRatio[k]
	RootsForRatio [][]ifaces.Column

	// AllInvolvedColumns stores the union of the ColumnForRatio[k] for all k
	AllInvolvedColumns []ifaces.Column

	// AggregateExpressions[k] stores the aggregate expression for Ratios[k]
	AggregateExpressions []*symbolic.Expression

	// AggregateExpressionsBoard[k] stores the topological sorting of
	// AggregateExpressions[k]
	AggregateExpressionsBoard []symbolic.ExpressionBoard

	// QuotientShares[k] stores for each k, the list of the Ratios[k] shares
	// of the quotient for the AggregateExpression[k]
	QuotientShares [][]ifaces.Column

	// MaxNbExprNode stores the largest number of node AggregateExpressionBoard[*]
	// has. This is used to dimension the memory pool during the prover time.
	MaxNbExprNode int
}

// createQuotientCtx constructs a [quotientCtx] from a list of ratios and aggregated
// expressions. The function organizes the handles but does not declare anything
// in the current wizard.CompiledIOP.
func createQuotientCtx(comp *wizard.CompiledIOP, ratios []int, aggregateExpressions []*symbolic.Expression) QuotientCtx {

	var (
		allInvolvedColumns = map[ifaces.ColID]struct{}{}
		_, _, domainSize   = wizardutils.AsExpr(aggregateExpressions[0])
		ctx                = QuotientCtx{
			DomainSize:                domainSize,
			Ratios:                    ratios,
			AggregateExpressions:      aggregateExpressions,
			AggregateExpressionsBoard: make([]symbolic.ExpressionBoard, len(ratios)),
			AllInvolvedColumns:        []ifaces.Column{},
			RootsForRatio:             make([][]ifaces.Column, len(ratios)),
			QuotientShares:            generateQuotientShares(comp, ratios, domainSize),
		}
	)

	for k, expr := range ctx.AggregateExpressions {

		var (
			board               = expr.Board()
			uniqueRootsForRatio = map[ifaces.ColID]struct{}{}
		)

		ctx.AggregateExpressionsBoard[k] = board
		ctx.MaxNbExprNode = max(ctx.MaxNbExprNode, board.CountNodes())

		// This loop scans the metadata looking for columns with the goal of
		// populating the collections composing quotientCtx.
		for _, metadata := range board.ListVariableMetadata() {

			// Scan in column metadata only
			col, ok := metadata.(ifaces.Column)
			if !ok {
				continue
			}

			// we keep track of all involved columns
			colName := col.GetColID()
			if _, ok := allInvolvedColumns[colName]; !ok {
				allInvolvedColumns[colName] = struct{}{}
				ctx.AllInvolvedColumns = append(ctx.AllInvolvedColumns, col)
			}

			// first time we see this root for this ratio
			rootCol := column.RootParents(col)
			rootName := rootCol.GetColID()
			if _, ok := uniqueRootsForRatio[rootName]; !ok {
				uniqueRootsForRatio[rootName] = struct{}{}
				ctx.RootsForRatio[k] = append(ctx.RootsForRatio[k], rootCol)
			}
		}
	}

	return ctx
}

// generateQuotientShares declares and returns the quotient share columns
func generateQuotientShares(comp *wizard.CompiledIOP, ratios []int, domainSize int) [][]ifaces.Column {

	var (
		quotientShares = make([][]ifaces.Column, len(ratios))
		currRound      = comp.NumRounds() - 1
	)

	for i, ratio := range ratios {
		quotientShares[i] = make([]ifaces.Column, ratio)
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
func (ctx *QuotientCtx) Run(run *wizard.ProverRuntime) {

	start := time.Now()
	// Initial step is to compute the FFTs for all committed vectors
	pool := mempool.CreateFromSyncPool(symbolic.MaxChunkSize).Prewarm(runtime.GOMAXPROCS(0) * ctx.MaxNbExprNode)

	if ctx.DomainSize >= GC_DOMAIN_SIZE {
		// Force the GC to run
		runtime.GC()
	}

	// Take the max quotient degree
	maxRatio := utils.Max(ctx.Ratios...)

	// for all involved roots, count the one we need some memory for
	// loops mirror the computing loop below.
	maxNbAllocs := 0
	for i := 0; i < maxRatio; i++ {
		nbAllocs := int64(0)
		for j, ratio := range ctx.Ratios {
			if i%(maxRatio/ratio) != 0 {
				continue
			}

			roots := ctx.RootsForRatio[j]
			for k := range roots {
				root := roots[k]
				name := root.GetColID()
				var witness sv.SmartVector
				witness, isNatural := run.Columns.TryGet(name)
				if !isNatural {
					witness = root.GetColAssignment(run)
				}

				switch witness.(type) {
				case *sv.Constant, *sv.PaddedCircularWindow:
					continue
				}
				nbAllocs++
			}
		}
		maxNbAllocs = max(maxNbAllocs, int(nbAllocs))
	}
	domain := fft.NewDomain(uint64(ctx.DomainSize), fft.WithCache())

	// we need up to maxNbAllocs allocations of size ctx.DomainSize at a given time
	vArena := arena.NewVectorArena[field.Element](int(maxNbAllocs) * ctx.DomainSize)

	/*
		For the quotient, we precompute the values of (wQ^N - 1)^-1 for w in H, the
		larger domain.

		Those values are D-periodic, thus we only compute a single period.
		(Where D is the ratio of the sizes of the larger and the smaller domain)

		The first value is ignored because it correspond to the case where w^N = 1
		(i.e. w is in the smaller subgroup)
	*/
	annulatorInvVals := fastpoly.EvalXnMinusOneOnACoset(ctx.DomainSize, ctx.DomainSize*maxRatio)
	annulatorInvVals = field.ParBatchInvert(annulatorInvVals, runtime.GOMAXPROCS(0))

	for i := 0; i < maxRatio; i++ {

		// use sync map to store the coset evaluated polynomials
		computedReeval := sync.Map{}
		vArena.Reset(0)

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
				share     = i * ratio / maxRatio
				roots     = ctx.RootsForRatio[j]
				board     = ctx.AggregateExpressionsBoard[j]
				metadatas = board.ListVariableMetadata()
			)

			shift := computeShift(uint64(ctx.DomainSize), ratio, share)
			domainCoset := fft.NewDomain(uint64(ctx.DomainSize), fft.WithCache(), fft.WithShift(shift))

			ppool.ExecutePoolChunky(len(roots), func(k int) {

				root := roots[k]
				name := root.GetColID()
				if _, found := computedReeval.Load(name); found {
					// it was already computed in a previous iteration of `j`
					return
				}

				// get the column value
				v, isNatural := run.Columns.TryGet(name)
				if !isNatural {
					v = root.GetColAssignment(run)
				}

				// now we re-evaluate on the coset
				reevaledRoot := reevalOnCoset(v, vArena, domain, domainCoset)
				computedReeval.Store(name, reevaledRoot)
			})

			// Evaluates the constraint expression on the coset
			evalInputs := make([]sv.SmartVector, len(metadatas))

			ppool.ExecutePoolChunky(len(metadatas), func(k int) {
				metadataInterface := metadatas[k]
				switch metadata := metadataInterface.(type) {
				case ifaces.Column:
					root := column.RootParents(metadata)
					rootName := root.GetColID()

					reevaledRoot, found := computedReeval.Load(rootName)
					if !found {
						// it is expected to computed in the above loop
						utils.Panic("did not find the reevaluation of %v", rootName)
					}

					if !metadata.IsComposite() {
						// natural and verifier columns.
						evalInputs[k] = reevaledRoot.(sv.SmartVector)
						return
					}

					if shifted, isShifted := metadata.(column.Shifted); isShifted {
						reevaledRoot = sv.SoftRotate(reevaledRoot.(sv.SmartVector), shifted.Offset)
						evalInputs[k] = reevaledRoot.(sv.SmartVector)
						return
					}
					utils.Panic("unexpected composite column type %T", metadata)
				case coin.Info:
					evalInputs[k] = sv.NewConstant(run.GetRandomCoinField(metadata.Name), ctx.DomainSize)
				case variables.X:
					evalInputs[k] = metadata.EvalCoset(ctx.DomainSize, i, maxRatio, true)
				case variables.PeriodicSample:
					evalInputs[k] = metadata.EvalCoset(ctx.DomainSize, i, maxRatio, true)
				case ifaces.Accessor:
					evalInputs[k] = sv.NewConstant(metadata.GetVal(run), ctx.DomainSize)
				default:
					utils.Panic("Not a variable type %v", reflect.TypeOf(metadataInterface))
				}
			})

			// Note that this will panic if the expression contains "no commitment"
			// This should be caught already by the constructor of the constraint.
			quotientShare := board.Evaluate(evalInputs, pool)
			quotientShare = sv.ScalarMul(quotientShare, annulatorInvVals[i])
			run.AssignColumn(ctx.QuotientShares[j][share].GetColID(), quotientShare)

		}

	}

	vArena = nil

	if ctx.DomainSize >= GC_DOMAIN_SIZE {
		// Force the GC to run
		runtime.GC()
	}

	logrus.Infof("[global-constraint] msg=\"computed the quotient\" took %v", time.Since(start))

}

// reevalOnCoset takes a vector v in evaluation form on the base domain
// and returns the vector evaluated on the coset defined by (cosetRatio, cosetID)
func reevalOnCoset(v sv.SmartVector, vArena *arena.VectorArena, domain, domainCoset *fft.Domain) sv.SmartVector {
	skipInverse := false
	switch x := v.(type) {
	case *sv.Constant:
		return x

	case *sv.PaddedCircularWindow:

		interval := x.Interval()
		if interval.IntervalLen == 1 && interval.Start() == 0 && x.PaddingVal_.IsZero() {
			// It's a multiple of the first Lagrange polynomial c * (1 + x + x^2 + x^3 + ...)
			// The ifft is (c) = (c/N, c/N, c/N, ...)
			constTerm := field.NewElement(uint64(x.Len()))
			constTerm.Inverse(&constTerm)
			constTerm.Mul(&constTerm, &x.Window_[0])
			v = sv.NewConstant(constTerm, x.Len())
			skipInverse = true
		}
	}

	res := arena.Get[field.Element](vArena, v.Len())
	v.WriteInSlice(res)

	if !skipInverse {
		domain.FFTInverse(res, fft.DIF, fft.WithNbTasks(2))
	}

	domainCoset.FFT(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(2))

	return sv.NewRegular(res)
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

package globalcs

import (
	"math/big"
	"reflect"
	"runtime"
	"sync"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"

	"github.com/sirupsen/logrus"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpoly"
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
	"github.com/consensys/linea-monorepo/prover/utils/arena"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// QuotientCtx collects all the internal fields needed to compute the quotient
type QuotientCtx struct {

	// DomainSize is the domain over which the global constraints are computed
	DomainSize int

	// Ratio lists the ratio found in the global constraints
	//
	// See [mergingCtx.Ratios]
	Ratios []int

	// ShiftedColumnsForRatio[k] stores all the columns involved in the aggregate
	// expressions for ratio Ratios[k]
	ShiftedColumnsForRatio [][]ifaces.Column

	// RootsPerRatio[k] stores all the root columns involved in the aggregate
	// expressions for ration Ratios[k]. By root column we mean the underlying
	// column that are actually committed to. For instance, if Shift(A, 1) is
	// in ColumnsForRatio[k], we will have A in RootPerRatio[k]
	RootsForRatio [][]ifaces.Column

	// AllInvolvedColumns stores the union of the ColumnForRatio[k] for all k
	AllInvolvedColumns []ifaces.Column

	// AllInvolvedRoots stores the union of the RootsForRatio[k] for all k
	AllInvolvedRoots []ifaces.Column

	// AggregateExpressions[k] stores the aggregate expression for Ratios[k]
	AggregateExpressions []*symbolic.Expression

	// AggregateExpressionsBoard[k] stores the topological sorting of
	// AggregateExpressions[k]
	AggregateExpressionsBoard []symbolic.ExpressionBoard

	// QuotientShares[k] stores for each k, the list of the Ratios[k] shares
	// of the quotient for the AggregateExpression[k]
	QuotientShares [][]ifaces.Column
}

// createQuotientCtx constructs a [quotientCtx] from a list of ratios and aggregated
// expressions. The function organizes the handles but does not declare anything
// in the current wizard.CompiledIOP.
func createQuotientCtx(comp *wizard.CompiledIOP, ratios []int, aggregateExpressions []*symbolic.Expression) QuotientCtx {

	var (
		allInvolvedHandlesIndex = map[ifaces.ColID]struct{}{}
		allInvolvedRootsSet     = collection.NewSet[ifaces.ColID]()
		_, _, domainSize        = wizardutils.AsExpr(aggregateExpressions[0])
		ctx                     = QuotientCtx{
			DomainSize:                domainSize,
			Ratios:                    ratios,
			AggregateExpressions:      aggregateExpressions,
			AggregateExpressionsBoard: make([]symbolic.ExpressionBoard, len(ratios)),
			AllInvolvedColumns:        []ifaces.Column{},
			AllInvolvedRoots:          []ifaces.Column{},
			ShiftedColumnsForRatio:    make([][]ifaces.Column, len(ratios)),
			RootsForRatio:             make([][]ifaces.Column, len(ratios)),
			QuotientShares:            generateQuotientShares(comp, ratios, domainSize),
		}
	)

	for k, expr := range ctx.AggregateExpressions {

		var (
			board               = expr.Board()
			uniqueRootsForRatio = collection.NewSet[ifaces.ColID]()
		)

		ctx.AggregateExpressionsBoard[k] = board

		// This loop scans the metadata looking for columns with the goal of
		// populating the collections composing quotientCtx.
		for _, metadata := range board.ListVariableMetadata() {

			// Scan in column metadata only
			col, ok := metadata.(ifaces.Column)
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
			if _, isShifted := col.(column.Shifted); isShifted {
				// TODO @gbotrel confirm we only get shifted and natural columns.
				ctx.ShiftedColumnsForRatio[k] = append(ctx.ShiftedColumnsForRatio[k], col)
			}

			if !uniqueRootsForRatio.Exists(rootCol.GetColID()) {
				ctx.RootsForRatio[k] = append(ctx.RootsForRatio[k], rootCol)
				uniqueRootsForRatio.Insert(rootCol.GetColID())
			}

			// Get the name of the
			if _, alreadyThere := allInvolvedHandlesIndex[col.GetColID()]; alreadyThere {
				continue
			}

			allInvolvedHandlesIndex[col.GetColID()] = struct{}{}
			ctx.AllInvolvedColumns = append(ctx.AllInvolvedColumns, col)

			// If the handle is simply a shift or a natural columns tracks its root
			if !allInvolvedRootsSet.Exists(rootCol.GetColID()) {
				allInvolvedRootsSet.Insert(rootCol.GetColID())
				ctx.AllInvolvedRoots = append(ctx.AllInvolvedRoots, rootCol)
			}
		}
	}

	// TODO @gbotrel this context preparation should compute the exact memory needed for the arenas
	// in Run and prepare them here.

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
				false,
			)
		}
	}

	return quotientShares
}

// compute the quotient shares.
func (ctx *QuotientCtx) Run(run *wizard.ProverRuntime) {

	// Initial step is to compute the FFTs for all committed vectors
	coeffs := sync.Map{} // (ifaces.ColID <=> sv.SmartVector)

	// Compute once the FFT of the natural columns

	domain0 := fft.NewDomain(uint64(ctx.DomainSize), fft.WithCache())

	arenaBase := arena.NewVectorArena[field.Element](ctx.DomainSize * len(ctx.AllInvolvedRoots))

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
				res := arena.Get[field.Element](arenaBase, ctx.DomainSize)
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
	var onceAnnulatorExt, onceAnnulatorBase sync.Once
	var annulatorInvValsExt []fext.Element
	var annulatorInvVals []field.Element

	arenaExt := arena.NewVectorArena[fext.Element](ctx.DomainSize * len(ctx.AllInvolvedRoots))

	for i := 0; i < maxRatio; i++ {
		// reset the scratch offset for this round
		arenaExt.Reset(0)

		// use sync map to store the coset evaluated polynomials
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
				share     = i * ratio / maxRatio
				handles   = ctx.ShiftedColumnsForRatio[j]
				roots     = ctx.RootsForRatio[j]
				board     = ctx.AggregateExpressionsBoard[j]
				metadatas = board.ListVariableMetadata()
			)

			shift := computeShift(uint64(ctx.DomainSize), ratio, share)
			domain := fft.NewDomain(uint64(ctx.DomainSize), fft.WithShift(shift), fft.WithCache())

			parallel.Execute(len(roots), func(start, stop int) {
				for k := start; k < stop; k++ {
					root := roots[k]
					rootName := root.GetColID()

					// load the coeff
					_v, _ := coeffs.Load(rootName)
					vCoeffs := _v.(sv.SmartVector)

					// most of the coeffs are regular, so we optimize for that case
					// and do a fft on the base field.
					if vr, ok := vCoeffs.(*sv.Regular); ok {
						reevaledRoot := arena.Get[field.Element](arenaExt, ctx.DomainSize)
						copy(reevaledRoot, *vr)
						domain.FFT(reevaledRoot, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))

						res := smartvectors.NewRegular(reevaledRoot)
						computedReeval.Store(rootName, res)
						continue
					}

					reevaledRoot := arena.Get[fext.Element](arenaExt, ctx.DomainSize)
					vCoeffs.WriteInSliceExt(reevaledRoot)
					domain.FFTExt(reevaledRoot, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))

					res := smartvectors.NewRegularExt(reevaledRoot)
					computedReeval.Store(rootName, res)
				}
			})

			for _, pol := range handles {
				root := column.RootParents(pol)
				rootName := root.GetColID()

				reevaledRoot, _ := computedReeval.Load(rootName)

				if shifted, isShifted := pol.(column.Shifted); isShifted {
					polName := pol.GetColID()
					var res sv.SmartVector
					switch ssv := reevaledRoot.(sv.SmartVector).(type) {
					case *sv.Regular:
						res = sv.SoftRotate(ssv, shifted.Offset)
					case *sv.RegularExt:
						res = sv.SoftRotateExt(ssv, shifted.Offset)
					}
					computedReeval.Store(polName, res)
					continue
				}
				panic("never called")
				// TODO @gbotrel confirm we only get shifted and natural columns.
			}

			// Evaluates the constraint expression on the coset
			evalInputs := make([]sv.SmartVector, len(metadatas))

			for k, metadataInterface := range metadatas {

				switch metadata := metadataInterface.(type) {
				case ifaces.Column:
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
				case variables.X:
					evalInputs[k] = metadata.EvalCoset(ctx.DomainSize, i, maxRatio, true)
				case variables.PeriodicSample:
					evalInputs[k] = metadata.EvalCoset(ctx.DomainSize, i, maxRatio, true)
				case ifaces.Accessor:
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
			quotientShare := board.Evaluate(evalInputs)
			switch quotientShare := quotientShare.(type) {
			case *sv.Regular:
				onceAnnulatorBase.Do(func() {
					annulatorInvVals = fastpoly.EvalXnMinusOneOnACoset(ctx.DomainSize, ctx.DomainSize*maxRatio)
					annulatorInvVals = field.ParBatchInvert(annulatorInvVals, runtime.GOMAXPROCS(0))
				})

				vq := field.Vector(*quotientShare)
				vq.ScalarMul(vq, &annulatorInvVals[i])
			case *sv.RegularExt:
				onceAnnulatorExt.Do(func() {
					annulatorInvValsExt = fastpolyext.EvalXnMinusOneOnACoset(ctx.DomainSize, ctx.DomainSize*maxRatio)
					annulatorInvValsExt = fext.ParBatchInvert(annulatorInvValsExt, runtime.GOMAXPROCS(0))
				})
				vq := extensions.Vector(*quotientShare)
				vq.ScalarMul(vq, &annulatorInvValsExt[i])
			default:
				// quotientShare = sv.ScalarMulExt(quotientShare, annulatorInvVals[i])
				utils.Panic("unexpected type %T", quotientShare)
			}

			run.AssignColumn(ctx.QuotientShares[j][share].GetColID(), quotientShare)
		}

		// Forcefully clean the memory for the computed reevals
		computedReeval = sync.Map{}
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

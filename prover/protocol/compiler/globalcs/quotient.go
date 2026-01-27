package globalcs

import (
	"math/big"
	"runtime"
	"sort"
	"sync"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"

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
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
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

	// AggregateExpressions[k] stores the aggregate expression for Ratios[k]
	AggregateExpressions []*symbolic.Expression

	// QuotientShares[k] stores for each k, the list of the Ratios[k] shares
	// of the quotient for the AggregateExpression[k]
	QuotientShares [][]ifaces.Column

	// ConstraintsByRatio[r] stores the list of indices k such that Ratios[k] == r
	ConstraintsByRatio map[int][]int
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
			ConstraintsByRatio:        make(map[int][]int),
		}
	)

	for k, ratio := range ratios {
		ctx.ConstraintsByRatio[ratio] = append(ctx.ConstraintsByRatio[ratio], k)
	}

	for k, expr := range ctx.AggregateExpressions {

		var (
			board               = expr.Board()
			uniqueRootsForRatio = collection.NewSet[ifaces.ColID]()
		)

		ctx.AggregateExpressionsBoard[k] = board

		// This loop scans the metadata looking for columns with the goal of
		// populating the collections composing quotientCtx.
		for _, metadata := range metadatas {

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
	stopTimer := profiling.LogTimer("computed the quotient (domain size %d)", ctx.DomainSize)
	defer stopTimer()

	domain0 := fft.NewDomain(uint64(ctx.DomainSize), fft.WithCache())

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

		for ratio, constraintsIndices := range ctx.ConstraintsByRatio {

			// For instance, if deg = 2 and max deg 8, we enter only if
			// i = 0 or 4 because this corresponds to the cosets we are
			// interested in.
			if i%(maxRatio/ratio) != 0 {
				continue
			}

			// reset the scratch offset for this round
			arenaExt.Reset(0)

			var (
				share = i * ratio / maxRatio
				shift = computeShift(uint64(ctx.DomainSize), ratio, share)
			)

			// 1. Identify all unique roots needed for this ratio group
			uniqueRootsMap := make(map[ifaces.ColID]int)
			var uniqueRoots []ifaces.Column

			for _, j := range constraintsIndices {
				for _, root := range ctx.RootsForRatio[j] {
					id := root.GetColID()
					if _, ok := uniqueRootsMap[id]; !ok {
						uniqueRootsMap[id] = len(uniqueRoots)
						uniqueRoots = append(uniqueRoots, root)
					}
				}
			}

			rootResults := make([]sv.SmartVector, len(uniqueRoots))

			domain := fft.NewDomain(uint64(ctx.DomainSize), fft.WithShift(shift), fft.WithCache())

			parallel.Execute(len(uniqueRoots), func(start, stop int) {
				for k := start; k < stop; k++ {
					root := uniqueRoots[k]
					rootName := root.GetColID()

					var witness sv.SmartVector
					witness, isAssigned := run.Columns.TryGet(rootName)

					// can happen if the column is verifier defined. In that case, no
					// need to protect with a lock. This will not touch run.Columns.
					if !isAssigned {
						witness = root.GetColAssignment(run)
					}

					// TODO @gbotrel handle the case where the witness is a constant and skip the ffts
					if smartvectors.IsBase(witness) {
						res := arena.Get[field.Element](arenaExt, ctx.DomainSize)
						witness.WriteInSlice(res)
						domain0.FFTInverse(res, fft.DIF, fft.WithNbTasks(2))
						domain.FFT(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(2))
						rootResults[k] = smartvectors.NewRegular(res)
					} else {
						res := arena.Get[fext.Element](arenaExt, ctx.DomainSize)
						witness.WriteInSliceExt(res)
						domain0.FFTInverseExt(res, fft.DIF, fft.WithNbTasks(2))
						domain.FFTExt(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(2))
						rootResults[k] = smartvectors.NewRegularExt(res)
					}

				}
			})

			computedReeval := make(map[ifaces.ColID]sv.SmartVector, len(uniqueRoots))
			for k, root := range uniqueRoots {
				computedReeval[root.GetColID()] = rootResults[k]
			}

			for _, j := range constraintsIndices {
				var (
					handles   = ctx.ShiftedColumnsForRatio[j]
					board     = ctx.AggregateExpressionsBoard[j]
					metadatas = board.ListVariableMetadata()
				)

				for _, pol := range handles {
					polName := pol.GetColID()
					if _, ok := computedReeval[polName]; ok {
						continue
					}

					root := column.RootParents(pol)
					rootName := root.GetColID()

					reevaledRoot := computedReeval[rootName]

					if shifted, isShifted := pol.(column.Shifted); isShifted {
						var res sv.SmartVector
						switch ssv := reevaledRoot.(type) {
						case *sv.Regular:
							res = sv.SoftRotate(ssv, shifted.Offset)
						case *sv.RegularExt:
							res = sv.SoftRotateExt(ssv, shifted.Offset)
						}
						computedReeval[polName] = res
						continue
					}
					panic("never called")
				}

				// Evaluates the constraint expression on the coset
				evalInputs := make([]sv.SmartVector, len(metadatas))

				for k, metadataInterface := range metadatas {

					switch metadata := metadataInterface.(type) {
					case ifaces.Column:
						value, ok := computedReeval[metadata.GetColID()]
						if !ok {
							utils.Panic("did not find the reevaluation of %v", metadata.GetColID())
						}
						evalInputs[k] = value

					case coin.Info:
						if metadata.IsBase() {
							utils.Panic("unsupported, coins are always over field extensions")

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
		}
	}
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

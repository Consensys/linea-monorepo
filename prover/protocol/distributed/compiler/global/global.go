package global

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/constants"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	edc "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/execution_data_collector"
)

type segmentID int

type DistributionInputs struct {
	ModuleComp  *wizard.CompiledIOP
	InitialComp *wizard.CompiledIOP
	// module Discoverer used to detect the relevant part of the query to the module
	Disc distributed.ModuleDiscoverer
	// Name of the module
	ModuleName distributed.ModuleName
	// number of segments for the module
	NumSegments int
	// the ID of the segment in the module
	SegID int
}

func DistributeGlobal(in DistributionInputs) {

	var (
		bInputs = boundaryInputs{
			moduleComp:  in.ModuleComp,
			numSegments: in.NumSegments,

			provider: boundaries{
				boundaryCol:          in.ModuleComp.Columns.GetHandle("PROVIDER"),
				lastPosOnBoundaryCol: 0,
				boundaryOpenings:     collection.NewMapping[query.LocalOpening, int](),
			},
			receiver: boundaries{
				boundaryCol:          in.ModuleComp.Columns.GetHandle("RECEIVER"),
				boundaryOpenings:     collection.NewMapping[query.LocalOpening, int](),
				lastPosOnBoundaryCol: 0,
			},
		}

		provider = bInputs.provider.boundaryCol
		receiver = bInputs.receiver.boundaryCol
	)

	for _, qName := range in.InitialComp.QueriesNoParams.AllUnignoredKeys() {

		q, ok := in.InitialComp.QueriesNoParams.Data(qName).(query.GlobalConstraint)
		if !ok {
			continue
		}

		if in.Disc.ExpressionIsInModule(q.Expression, in.ModuleName) {

			// apply global constraint over the segment.
			in.ModuleComp.InsertGlobal(0,
				q.ID,
				AdjustExpressionForGlobal(in.ModuleComp, q.Expression, in.NumSegments),
			)

			// collect the boundaries for provider and receiver
			BoundariesForProvider(&bInputs, q)
			BoundariesForReceiver(&bInputs, q)

		}

		// @Azam the risk is that some global constraints may be skipped here.
		// we can prevent this by tagging the query as ignored from the initialComp,
		// and at the end make sure that no query has remained in initial CompiledIOP.
	}

	// get the hash of the provider and the receiver
	var (
		colOnes            = verifiercol.NewConstantCol(field.One(), provider.Size())
		mimcHasherProvider = edc.NewMIMCHasher(in.ModuleComp, provider, colOnes, "MIMC_HASHER_PROVIDER")
		mimcHasherReceiver = edc.NewMIMCHasher(in.ModuleComp, receiver, colOnes, "MIMC_HASHER_RECEIVER")
	)

	mimcHasherProvider.DefineHasher(in.ModuleComp, "DISTRIBUTED_GLOBAL_QUERY_MIMC_HASHER_PROVIDER")
	mimcHasherReceiver.DefineHasher(in.ModuleComp, "DISTRIBUTED_GLOBAL_QUERY_MIMC_HASHER_RECEIVER")

	var (
		openingHashProvider = in.ModuleComp.InsertLocalOpening(0, "ACCESSOR_FROM_HASH_PROVIDER", mimcHasherProvider.HashFinal)
		openingHashReceiver = in.ModuleComp.InsertLocalOpening(0, "ACCESSOR_FROM_HASH_RECEIVER", mimcHasherReceiver.HashFinal)
	)

	// declare the hash of the provider/receiver as the public inputs.
	in.ModuleComp.PublicInputs = append(in.ModuleComp.PublicInputs,
		wizard.PublicInput{
			Name: constants.GlobalProviderPublicInput,
			Acc:  accessors.NewLocalOpeningAccessor(openingHashProvider, 0),
		})

	in.ModuleComp.PublicInputs = append(in.ModuleComp.PublicInputs,
		wizard.PublicInput{
			Name: constants.GlobalReceiverPublicInput,
			Acc:  accessors.NewLocalOpeningAccessor(openingHashReceiver, 0),
		})

	in.ModuleComp.RegisterProverAction(0, &proverActionForBoundaries{
		provider: boundaryAssignments{
			boundaries:  bInputs.provider,
			hashOpening: openingHashProvider,
			mimcHash:    *mimcHasherProvider,
		},

		receiver: boundaryAssignments{
			boundaries:  bInputs.receiver,
			hashOpening: openingHashReceiver,
			mimcHash:    *mimcHasherReceiver,
		},
	})

}

type boundaryInputs struct {
	moduleComp  *wizard.CompiledIOP
	provider    boundaries
	receiver    boundaries
	numSegments int
	segID       int
}

type boundaries struct {
	boundaryCol          ifaces.Column
	boundaryOpenings     collection.Mapping[query.LocalOpening, int]
	lastPosOnBoundaryCol int
}

func AdjustExpressionForGlobal(
	comp *wizard.CompiledIOP,
	expr *symbolic.Expression,
	numSegments int,
) *symbolic.Expression {

	var (
		board          = expr.Board()
		metadatas      = board.ListVariableMetadata()
		translationMap = collection.NewMapping[string, *symbolic.Expression]()
		colTranslation ifaces.Column
		size           = column.ExprIsOnSameLengthHandles(&board)
	)

	for _, metadata := range metadatas {

		// For each slot, get the expression obtained by replacing the commitment
		// by the appropriated column.

		switch m := metadata.(type) {
		case ifaces.Column:

			switch col := m.(type) {
			case column.Natural:
				colTranslation = comp.Columns.GetHandle(m.GetColID())

			case verifiercol.VerifierCol:
				// panic happens specially for the case of FromAccessors
				panic("unsupported for now, unless module discoverer can capture such columns")

			case column.Shifted:
				colTranslation = column.Shift(comp.Columns.GetHandle(col.Parent.GetColID()), col.Offset)

			}

			translationMap.InsertNew(m.String(), ifaces.ColumnAsVariable(colTranslation))
		case variables.X:
			utils.Panic("unsupported, the value of `x` in the unsplit query and the split would be different")
		case variables.PeriodicSample:
			// Check that the period is not larger than the domain size. If
			// the period is smaller this is a no-op because the period does
			// not change.
			segSize := size / numSegments

			if m.T > segSize {

				panic("unsupported")
			}
			translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))
		default:
			// Repass the same variable (for coins or other types of single-valued variable)
			translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))
		}

	}
	return expr.Replay(translationMap)
}

func BoundariesForProvider(in *boundaryInputs, q query.GlobalConstraint) {

	var (
		board          = q.Board()
		offsetRange    = q.MinMaxOffset()
		maxShift       = offsetRange.Max
		colsInExpr     = distributed.ListColumnsFromExpr(q.Expression, false)
		colsOnProvider = onBoundaries(colsInExpr, maxShift, &in.provider)
		numBoundaries  = offsetRange.Max - offsetRange.Min
		size           = column.ExprIsOnSameLengthHandles(&board)
		segSize        = size / in.numSegments
	)
	for _, col := range colsInExpr {
		for i := 0; i < numBoundaries; i++ {
			if colsOnProvider.Exists(col.GetColID()) {

				pos := colsOnProvider.MustGet(col.GetColID())

				if i < maxShift-column.StackOffsets(col) {
					// take from provider, since the size of the provider is different from size of the expression
					// take it via accessor.
					var (
						index            = pos[0] + i
						name             = ifaces.QueryIDf("%v_%v_%v", q.ID, "FROM_PROVIDER_AT", index)
						loProvider       = in.moduleComp.InsertLocalOpening(0, name, column.Shift(in.provider.boundaryCol, index))
						accessorProvider = accessors.NewLocalOpeningAccessor(loProvider, 0)
						indexOnCol       = segSize - (maxShift - column.StackOffsets(col) - i)
						nameExpr         = ifaces.QueryIDf("%v_%v_%v", "CONSISTENCY_AGAINST_PROVIDER", col.GetColID(), i)
						colInModule      ifaces.Column
					)

					// replace col with its replacement in the module.
					if shifted, ok := col.(column.Shifted); ok {
						colInModule = in.moduleComp.Columns.GetHandle(shifted.Parent.GetColID())
					} else {
						colInModule = in.moduleComp.Columns.GetHandle(col.GetColID())
					}

					// add the localOpening to the map
					in.provider.boundaryOpenings.InsertNew(loProvider, index)
					// impose that loProvider = loCol
					in.moduleComp.InsertLocal(0, nameExpr,
						symbolic.Sub(accessorProvider, column.Shift(colInModule, indexOnCol)),
					)

				}
			}
		}
	}

}

func BoundariesForReceiver(in *boundaryInputs, q query.GlobalConstraint) {

	var (
		offsetRange    = q.MinMaxOffset()
		maxShift       = offsetRange.Max
		colsInExpr     = distributed.ListColumnsFromExpr(q.Expression, false)
		colsOnReceiver = onBoundaries(colsInExpr, maxShift, &in.receiver)
		numBoundaries  = offsetRange.Max - offsetRange.Min
		comp           = in.moduleComp
		colInModule    ifaces.Column
	)

	for i := 0; i < numBoundaries; i++ {

		translationMap := collection.NewMapping[string, *symbolic.Expression]()

		for _, col := range colsInExpr {

			// replace col with its replacement in the module.
			if shifted, ok := col.(column.Shifted); ok {
				colInModule = in.moduleComp.Columns.GetHandle(shifted.Parent.GetColID())
			} else {
				colInModule = in.moduleComp.Columns.GetHandle(col.GetColID())
			}

			if colsOnReceiver.Exists(col.GetColID()) {
				pos := colsOnReceiver.MustGet(col.GetColID())

				if i < maxShift-column.StackOffsets(col) {
					// take from receiver, since the size of the receiver is different from size of the expression
					// take it via accessor.
					var (
						index    = pos[0] + i
						name     = ifaces.QueryIDf("%v_%v_%v", q.ID, "FROM_RECEIVER_AT", index)
						lo       = comp.InsertLocalOpening(0, name, column.Shift(in.receiver.boundaryCol, index))
						accessor = accessors.NewLocalOpeningAccessor(lo, 0)
					)
					// add the localOpening to the map
					in.receiver.boundaryOpenings.InsertNew(lo, index)
					// in.receiverOpenings = append(in.receiverOpenings, lo)
					// translate the column
					translationMap.InsertNew(string(col.GetColID()), accessor.AsVariable())
				} else {
					// take the rest from the column
					tookFromReceiver := maxShift - column.StackOffsets(col)
					translationMap.InsertNew(string(col.GetColID()), ifaces.ColumnAsVariable(column.Shift(colInModule, i-tookFromReceiver)))
				}

			} else {
				translationMap.InsertNew(string(col.GetColID()), ifaces.ColumnAsVariable((column.Shift(colInModule, i))))
			}

		}

		// If this is the first segment check for NoBoundCancel.
		// q.NoBoundCancel is false by default, which in this case we should not check the boundaries.
		if in.segID != 0 || q.NoBoundCancel {
			expr := q.Expression.Replay(translationMap)
			name := ifaces.QueryIDf("%v_%v_%v", "CONSISTENCY_AGAINST_RECEIVER", q.ID, i)
			comp.InsertLocal(0, name, expr)
		}

	}

}

// it indicates the column list having the provider cells (i.e.,
// some cells of the columns are needed to be provided to the next segment)
func onBoundaries(colsInExpr []ifaces.Column, maxShift int, b *boundaries) collection.Mapping[ifaces.ColID, [2]int] {

	var (
		ctr            = b.lastPosOnBoundaryCol
		colsOnReceiver = collection.NewMapping[ifaces.ColID, [2]int]()
	)
	for _, col := range colsInExpr {
		// number of boundaries from the column (that falls on the receiver) is
		// maxShift - column.StackOffsets(col)
		newCtr := ctr + maxShift - column.StackOffsets(col)

		// it does not have any cell on the receiver.
		if newCtr == ctr {
			continue
		}

		colsOnReceiver.InsertNew(col.GetColID(), [2]int{ctr, newCtr})
		ctr = newCtr

	}

	b.lastPosOnBoundaryCol = ctr
	return colsOnReceiver

}

// it generates natural verifier columns, from a given verifier column
func createVerifierColForModule(col ifaces.Column, numSegments int) ifaces.Column {

	if vcol, ok := col.(verifiercol.VerifierCol); ok {

		switch v := vcol.(type) {
		case verifiercol.ConstCol:
			return verifiercol.NewConstantCol(v.F, v.Size()/numSegments)
		default:
			panic("unsupported")
		}
	}
	return nil
}

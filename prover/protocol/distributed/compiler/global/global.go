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
			moduleComp:       in.ModuleComp,
			numSegments:      in.NumSegments,
			provider:         in.ModuleComp.Columns.GetHandle("PROVIDER"),
			receiver:         in.ModuleComp.Columns.GetHandle("RECEIVER"),
			providerOpenings: []query.LocalOpening{},
			receiverOpenings: []query.LocalOpening{},
		}
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
		colOnes             = verifiercol.NewConstantCol(field.One(), bInputs.provider.Size())
		mimcHasherProvider  = edc.NewMIMCHasher(in.ModuleComp, bInputs.provider, colOnes, "MIMC_HASHER_PROVIDER")
		mimicHasherReceiver = edc.NewMIMCHasher(in.ModuleComp, bInputs.receiver, colOnes, "MIMC_HASHER_RECEIVER")
	)

	mimcHasherProvider.DefineHasher(in.ModuleComp, "DISTRIBUTED_GLOBAL_QUERY_MIMC_HASHER_PROVIDER")
	mimcHasherProvider.DefineHasher(in.ModuleComp, "DISTRIBUTED_GLOBAL_QUERY_MIMC_HASHER_RECEIVER")

	var (
		openingHashProvider = in.ModuleComp.InsertLocalOpening(0, "ACCESSOR_FROM_HASH_PROVIDER", mimcHasherProvider.HashFinal)
		openingHashReceiver = in.ModuleComp.InsertLocalOpening(0, "ACCESSOR_FROM_HASH_RECEIVER", mimicHasherReceiver.HashFinal)
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
		provider:         bInputs.provider,
		receiver:         bInputs.receiver,
		providerOpenings: bInputs.providerOpenings,
		receiverOpenings: bInputs.receiverOpenings,

		mimicHasherProvider: *mimcHasherProvider,
		mimicHasherReceiver: *mimicHasherReceiver,
		hashOpeningProvider: openingHashProvider,
		hashOpeningReceiver: openingHashReceiver,
	})

}

type boundaryInputs struct {
	moduleComp                         *wizard.CompiledIOP
	numSegments                        int
	provider                           ifaces.Column
	receiver                           ifaces.Column
	providerOpenings, receiverOpenings []query.LocalOpening
	segID                              int
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

				panic("unsupported, since this depends on the segment ID, unless the module discoverer can detect such cases")
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
		provider       = in.provider
		maxShift       = offsetRange.Max
		colsInExpr     = distributed.ListColumnsFromExpr(q.Expression, false)
		colsOnProvider = onBoundaries(colsInExpr, maxShift)
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
						name             = ifaces.QueryIDf("%v_%v", "FROM_PROVIDER_AT", index)
						loProvider       = in.moduleComp.InsertLocalOpening(0, name, column.Shift(provider, index))
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

					// add the localOpening to the list
					in.providerOpenings = append(in.providerOpenings, loProvider)
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
		receiver       = in.receiver
		maxShift       = offsetRange.Max
		colsInExpr     = distributed.ListColumnsFromExpr(q.Expression, false)
		colsOnReceiver = onBoundaries(colsInExpr, maxShift)
		numBoundaries  = offsetRange.Max - offsetRange.Min
		comp           = in.moduleComp
		colInModule    ifaces.Column
		// list of local openings by the boundary index
		allLists = make([][]query.LocalOpening, numBoundaries)
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
						name     = ifaces.QueryIDf("%v_%v", "FROM_RECEIVER_AT", index)
						lo       = comp.InsertLocalOpening(0, name, column.Shift(receiver, index))
						accessor = accessors.NewLocalOpeningAccessor(lo, 0)
					)
					// add the localOpening to the list
					allLists[i] = append(allLists[i], lo)
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

	// order receiverOpenings column by column
	for i := 0; i < numBoundaries; i++ {
		for _, list := range allLists {
			if len(list) > i {
				in.receiverOpenings = append(in.receiverOpenings, list[i])
			}
		}
	}

}

// it indicates the column list having the provider cells (i.e.,
// some cells of the columns are needed to be provided to the next segment)
func onBoundaries(colsInExpr []ifaces.Column, maxShift int) collection.Mapping[ifaces.ColID, [2]int] {

	var (
		ctr            = 0
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

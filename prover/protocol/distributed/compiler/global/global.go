package global

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

type segmentID int

type provider struct {
	// Provider stores the boundaries that the segment should
	// provide to the next segment (for all the global constraints).
	provider ifaces.Column
	// last index in the provider that is checked by the verifier (for local constraints over boundaries).
	lastIndex int
}

type receiver struct {
	// Receiver stores the boundaries that the segment should
	// receive from its previous segment (for all the global constraints).
	receiver ifaces.Column
	// last index in the provider that is checked by the verifier (for local constraints over boundaries).
	lastIndex int
}

type DistributionInputs struct {
	ModuleComp  *wizard.CompiledIOP
	InitialComp *wizard.CompiledIOP
	// module Discoverer used to detect the relevant part of the query to the module
	Disc distributed.ModuleDiscoverer
	// Name of the module
	ModuleName distributed.ModuleName
	// number of segments for the module
	NumSegments int
}

func DistributeGlobal(in DistributionInputs) {

	var (
		bInputs = boundaryInputs{
			comp:               in.ModuleComp,
			numSegments:        in.NumSegments,
			provider:           provider{provider: in.ModuleComp.Columns.GetHandle("PROVIDER")},
			receiver:           receiver{receiver: in.ModuleComp.Columns.GetHandle("RECEIVER")},
			providerOpenings:   []query.LocalOpening{},
			receiverOpenings:   []query.LocalOpening{},
			colAgainstProvider: []query.LocalOpening{},
			colAgainstReceiver: []query.LocalOpening{},
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
			var (
				board    = q.Board()
				metadata = board.ListVariableMetadata()
			)

			for _, m := range metadata {
				switch t := m.(type) {
				case ifaces.Column:
					collectBoundariesForProvider(&bInputs, q, t)
					collectBoundariesForReceiver(&bInputs, q, t)

				}

			}
		}

		// the risk is that some global constraints may be skipped here.
		// we can prevent this by tagging the query as ignored from the initialComp,
		// and at the end make sure that no query has remained in initial CompiledIOP.
	}

	bInputs.comp.RegisterProverAction(0, &paramAssignments{
		provider:           bInputs.provider.provider,
		receiver:           bInputs.receiver.receiver,
		providerOpenings:   bInputs.providerOpenings,
		receiverOpenings:   bInputs.receiverOpenings,
		colAgainstProvider: bInputs.colAgainstProvider,
		colAgainstReceiver: bInputs.colAgainstReceiver})

	bInputs.comp.RegisterVerifierAction(0, &verifierChecks{
		providerOpenings:   bInputs.providerOpenings,
		receiverOpenings:   bInputs.receiverOpenings,
		colAgainstProvider: bInputs.colAgainstProvider,
		colAgainstReceiver: bInputs.colAgainstReceiver,
	})

}

type boundaryInputs struct {
	comp                                   *wizard.CompiledIOP
	numSegments                            int
	provider                               provider
	receiver                               receiver
	isFirst, isLast                        bool
	providerOpenings, receiverOpenings     []query.LocalOpening
	colAgainstProvider, colAgainstReceiver []query.LocalOpening
}

func checkBoundaries(in boundaryInputs, q query.GlobalConstraint) {

}

func AdjustExpressionForGlobal(comp *wizard.CompiledIOP,
	expr *symbolic.Expression, numSegments int,
) *symbolic.Expression {

	var (
		board          = expr.Board()
		metadatas      = board.ListVariableMetadata()
		translationMap = collection.NewMapping[string, *symbolic.Expression]()
		colTranslation ifaces.Column
		// column is split fairly among segments.
		// size    = column.ExprIsOnSameLengthHandles(&board)
		// segSize = size / numSegments
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
				// Create the split in live
				//colTranslation = col.Split(comp, segID*segSize, (segID+1)*segSize)

			// Shift the subparent, if the offset is larger than the subparent
			// we repercute it on the num
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
			/* translated := symbolic.NewVariable(metadata)

			if m.T > segSize {

				// Here, there are two possibilities. (1) The current slot is
				// on a portion of the Periodic sample where everything is
				// zero or (2) the current slot matches a portion of the
				// periodic sampling containing a 1. To determine which is
				// the current situation, we need to find out where the slot
				// is located compared to the period.
				var (
					slotStartAt = (segID * segSize) % m.T
					slotStopAt  = slotStartAt + segSize
				)

				if m.Offset >= slotStartAt && m.Offset < slotStopAt {
					translated = variables.NewPeriodicSample(segSize, m.Offset%segSize)
				} else {
					translated = symbolic.NewConstant(0)
				}
			}

			// And we can just pass it over because the period does not change
			translationMap.InsertNew(m.String(), translated)
			*/
		default:
			// Repass the same variable (for coins or other types of single-valued variable)
			translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))
		}

	}
	return expr.Replay(translationMap)
}

func GetMaxShift(expr *symbolic.Expression) int {
	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
		maxshift = 0
	)
	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:
			if shifted, ok := t.(column.Shifted); ok {
				maxshift = max(maxshift, shifted.Offset)
			}
		}

	}
	return maxshift
}

func GetMinShift(expr *symbolic.Expression) int {
	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
		minshift = 0
	)
	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:
			if shifted, ok := t.(column.Shifted); ok {
				minshift = min(minshift, shifted.Offset)
			}
		}

	}
	return minshift
}

func collectBoundariesForProvider(in *boundaryInputs, q query.GlobalConstraint, t ifaces.Column) {
	var (
		numSegments    = in.numSegments
		segmentSize    = t.Size() / numSegments
		col            ifaces.Column
		numBoundariesP int
		maxShift       = GetMaxShift(q.Expression)
	)

	if !utils.IsPowerOfTwo(segmentSize) {
		panic("the segmentSize is not power of two")
	}

	if shifted, ok := t.(column.Shifted); ok {
		// number of boundaries from the current column
		numBoundariesP = maxShift - shifted.Offset
		col = in.comp.Columns.GetHandle(shifted.Parent.GetColID())

	} else {
		numBoundariesP = maxShift
		col = in.comp.Columns.GetHandle(t.GetColID())
	}

	ctrP := in.provider.lastIndex
	for i := segmentSize - numBoundariesP; i < segmentSize; i++ {

		var (
			indexP  = (i - (segmentSize - numBoundariesP)) + in.provider.lastIndex
			nameP   = fmt.Sprintf("%v_%v_%v", q.ID, "PROVIDER_OPENING", indexP)
			nameCol = fmt.Sprintf("%v_%v_%v_%v", q.ID, col.GetColID(), "AGAINST_PROVIDER", i)
		)

		in.providerOpenings = append(in.providerOpenings,

			in.comp.InsertLocalOpening(0, ifaces.QueryID(nameP),
				column.Shift(in.provider.provider, indexP),
			),
		)

		in.colAgainstProvider = append(in.colAgainstProvider,

			in.comp.InsertLocalOpening(0, ifaces.QueryID(nameCol),
				column.Shift(col, i),
			),
		)

		ctrP++
	}
	in.provider.lastIndex += numBoundariesP
}

func collectBoundariesForReceiver(in *boundaryInputs, q query.GlobalConstraint, t ifaces.Column) {

	var (
		col            ifaces.Column
		numBoundariesR int
		minShift       = GetMinShift(q.Expression)
	)

	if shifted, ok := t.(column.Shifted); ok {
		// number of boundaries from the current column
		numBoundariesR = shifted.Offset - minShift
		col = in.comp.Columns.GetHandle(shifted.Parent.GetColID())

	} else {
		numBoundariesR = -minShift
		col = in.comp.Columns.GetHandle(t.GetColID())
	}

	ctrR := in.receiver.lastIndex

	for i := 0; i < numBoundariesR; i++ {

		var (
			nameR   = fmt.Sprintf("%v_%v_%v", q.ID, "RECEIVER_OPENING", i+in.receiver.lastIndex)
			nameCol = fmt.Sprintf("%v_%v_%v_%v", q.ID, col.GetColID(), "AGAINST_RECEIVER", i)
		)

		in.receiverOpenings = append(in.receiverOpenings,

			in.comp.InsertLocalOpening(0, ifaces.QueryID(nameR),
				column.Shift(in.receiver.receiver, i+in.receiver.lastIndex),
			),
		)

		in.colAgainstReceiver = append(in.colAgainstReceiver,

			in.comp.InsertLocalOpening(0, ifaces.QueryID(nameCol),
				column.Shift(col, i),
			),
		)

		ctrR++
	}
	in.receiver.lastIndex += numBoundariesR
}

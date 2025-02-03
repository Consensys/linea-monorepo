package global

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
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

	for _, qName := range in.InitialComp.QueriesNoParams.AllUnignoredKeys() {

		q, ok := in.InitialComp.QueriesNoParams.Data(qName).(query.GlobalConstraint)
		if !ok {
			continue
		}

		if in.Disc.ExpressionIsInModule(q.Expression, in.ModuleName) {
			distributeAmongSegments(in.ModuleComp, q, in.NumSegments)
		}

		// the risk is that some global constraints may be skipped here.
		// we can prevent this by tagging the query as ignored from the initialComp,
		// and at the end make sure that no query has remained in initial CompiledIOP.
	}

}

// it distribute the columns of the expression among segments of the same module.
func distributeAmongSegments(comp *wizard.CompiledIOP, q query.GlobalConstraint, numSegments int) {

	// apply global constraint over each segment.
	for segID := 0; segID < numSegments; segID++ {
		comp.InsertGlobal(0,
			ifaces.QueryIDf("%v_DISTRIBUTED_GLOBALQ_SLOT_%v", q.ID, segID),
			AdjustExpressionForGlobal(comp, q.Expression, segID, numSegments),
		)

		// It checks the boundaries against provider and receiver columns of the segment.
		// checkBoundaries(comp, q, segID, numSegments)
	}

}

type boundaryInputs struct {
	comp            *wizard.CompiledIOP
	q               query.GlobalConstraint
	numSegments     int
	provider        provider
	receiver        receiver
	isFirst, isLast bool
}

func checkBouandaries(in boundaryInputs) {

	var (
		numSegments   = in.numSegments
		board         = in.q.Board()
		metadata      = board.ListVariableMetadata()
		maxshift      = GetMaxShift(in.q.Expression)
		numBoundaries int
		provider      = in.provider
		receiver      = in.receiver
	)

	// TBD: sanity checks
	// if this is the last  segment provider should be empty
	// if it is the first segment the receiver should be empty.

	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:

			var (
				segmentSize = t.Size() / numSegments
				lastRow     = segmentSize - 1
				// access the counterpart of the column in the module.
				col = in.comp.Columns.GetHandle(t.GetColID())
			)

			if !utils.IsPowerOfTwo(segmentSize) {
				panic("the segmentSize is not power of two")
			}

			if shifted, ok := t.(column.Shifted); ok {
				// number of boundaries from the current column
				numBoundaries = maxshift - shifted.Offset

			} else {
				numBoundaries = maxshift
			}

			if !in.isLast {
				ctrP := provider.lastIndex
				for i := lastRow - numBoundaries; i <= lastRow; i++ {
					name := fmt.Sprintf("%v_%v_%v", in.q.ID, t.GetColID(), i)
					in.comp.InsertLocal(0, ifaces.QueryID(name),
						symbolic.Sub(
							column.Shift(provider.provider, provider.lastIndex),
							column.Shift(col, lastRow-i-1),
						),
					)
					ctrP++
				}
				provider.lastIndex += numBoundaries
			}

			if !in.isFirst {
				ctrR := receiver.lastIndex
				for i := 0; i < numBoundaries; i++ {
					name := fmt.Sprintf("%v_%v_%v", in.q.ID, t.GetColID(), i)
					in.comp.InsertLocal(0, ifaces.QueryID(name),
						symbolic.Sub(
							column.Shift(receiver.receiver, receiver.lastIndex),
							column.Shift(col, i),
						),
					)
					ctrR++
				}
				receiver.lastIndex += numBoundaries
			}

			// reset the number of Boundaries for the next column
			numBoundaries = 0
		}

	}
}

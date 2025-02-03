package global

import (
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

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

func checkBoundaries(comp *wizard.CompiledIOP, q query.GlobalConstraint, segmentID, numSegments int) {
	// get the provider and the receiver Columns of the segment.
	// Provider stores the boundaries that the segment should provide to the next segment (for all the global constraints).
	// Receiver stores the boundaries that the segment should receive from its previous segment (for all the global constraints).
	provider := comp.Columns.GetHandle("PROVIDER")
	receiver := comp.Columns.GetHandle("RECEIVER")
	checkConsistency(comp, q, provider, segmentID, numSegments, "PROVIDER")
	checkConsistency(comp, q, receiver, segmentID, numSegments, "RECEIVER")
}
func checkConsistency(
	comp *wizard.CompiledIOP,
	q query.GlobalConstraint,
	col ifaces.Column,
	segmentID, numSEgments int,
	hint string) {

	panic("unimplemented")
}

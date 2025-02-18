package distributedprojection

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

func CompileDistributedProjection(comp *wizard.CompiledIOP) {

	for _, qName := range comp.QueriesParams.AllUnignoredKeys() {
		// Filter out non distributed projection queries
		distributedprojection, ok := comp.QueriesParams.Data(qName).(query.DistributedProjection)
		if !ok {
			continue
		}

		// This ensures that the distributed projection query is not used again in the
		// compilation process. We know that the query was not already ignored at the beginning
		// because we are iterating over the unignored keys.
		comp.QueriesParams.MarkAsIgnored(qName)
		round := comp.QueriesParams.Round(qName)
		compile(comp, round, distributedprojection)
	}
}

func compile(comp *wizard.CompiledIOP, round int, distributedprojection query.DistributedProjection) {
	var (
		pa = &distribuedProjectionProverAction{
			Name:                    distributedprojection.ID,
			FilterA:                 make([]*symbolic.Expression, len(distributedprojection.Inp)),
			FilterB:                 make([]*symbolic.Expression, len(distributedprojection.Inp)),
			ColumnA:                 make([]*symbolic.Expression, len(distributedprojection.Inp)),
			ColumnB:                 make([]*symbolic.Expression, len(distributedprojection.Inp)),
			HornerA:                 make([]ifaces.Column, len(distributedprojection.Inp)),
			HornerB:                 make([]ifaces.Column, len(distributedprojection.Inp)),
			HornerA0:                make([]query.LocalOpening, len(distributedprojection.Inp)),
			HornerB0:                make([]query.LocalOpening, len(distributedprojection.Inp)),
			EvalCoins:               make([]coin.Info, len(distributedprojection.Inp)),
			IsA:                     make([]bool, len(distributedprojection.Inp)),
			IsB:                     make([]bool, len(distributedprojection.Inp)),
			CumNumOnesPrevSegmentsA: make([]big.Int, len(distributedprojection.Inp)),
			CumNumOnesPrevSegmentsB: make([]big.Int, len(distributedprojection.Inp)),
			CumNumOnesCurrSegmentA:  make([]field.Element, len(distributedprojection.Inp)),
			CumNumOnesCurrSegmentB:  make([]field.Element, len(distributedprojection.Inp)),
		}
	)
	pa.Push(comp, distributedprojection)
	pa.RegisterQueries(comp, round, distributedprojection)
	comp.RegisterProverAction(round, pa)
	comp.RegisterVerifierAction(round, &distributedProjectionVerifierAction{
		Name:                    pa.Name,
		HornerA0:                pa.HornerA0,
		HornerB0:                pa.HornerB0,
		IsA:                     pa.IsA,
		IsB:                     pa.IsB,
		EvalCoins:               pa.EvalCoins,
		CumNumOnesPrevSegmentsA: pa.CumNumOnesPrevSegmentsA,
		CumNumOnesPrevSegmentsB: pa.CumNumOnesPrevSegmentsB,
		NumOnesCurrSegmentA:     pa.CumNumOnesCurrSegmentA,
		NumOnesCurrSegmentB:     pa.CumNumOnesCurrSegmentB,
		FilterA: pa.FilterA,
		FilterB: pa.FilterB,
	})

}

package logderiv

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// compile [query.LogDerivativeSum] query
func CompileLogDerivSum(comp *wizard.CompiledIOP) {

	// Collect all the lookup queries into "lookups"
	for _, qName := range comp.QueriesParams.AllUnignoredKeys() {

		// Filter out non lookup queries
		logDeriv, ok := comp.QueriesParams.Data(qName).(query.LogDerivativeSum)
		if !ok {
			continue
		}

		// This ensures that the LogDerivativeSum query is not used again in the
		// compilation process. We know that the query was already ignored at
		// the beginning because we are iterating over the unignored keys.
		comp.QueriesParams.MarkAsIgnored(qName)
		zEntries := logDeriv.Inputs
		va := lookup.FinalEvaluationCheck{}
		for _, entry := range zEntries {
			zC := &lookup.ZCtx{
				Round:            entry.Round,
				Size:             entry.Size,
				SigmaNumerator:   entry.Numerator,
				SigmaDenominator: entry.Denominator,
			}

			// z-packing compile
			zC.Compile(comp)
			// prover step; Z assignments
			zAssignmentTask := lookup.ZAssignmentTask(*zC)
			comp.SubProvers.AppendToInner(zC.Round, func(run *wizard.ProverRuntime) {
				zAssignmentTask.Run(run)
			})
			va.ZOpenings = append(va.ZOpenings, zC.ZOpenings...)
		}
		// verifer step
		lastRound := comp.NumRounds() - 1
		comp.RegisterVerifierAction(lastRound, &va)
	}

}

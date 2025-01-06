package logderiv

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
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
		va := FinalEvaluationCheck{}
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
		va.LogDeriveSumID = qName
		// verifer step
		lastRound := comp.NumRounds() - 1
		comp.RegisterVerifierAction(lastRound, &va)
	}

}

type FinalEvaluationCheck struct {
	// the name of a lookupTable in the pack, this can help for debugging.
	Name string
	// ZOpenings lists all the openings of all the zCtx
	ZOpenings []query.LocalOpening
	// query ID
	LogDeriveSumID ifaces.QueryID
}

// Run implements the [wizard.VerifierAction]
func (f *FinalEvaluationCheck) Run(run *wizard.VerifierRuntime) error {

	// zSum stores the sum of the ending values of the zs as queried
	// in the protocol via the local opening queries.
	zSum := field.Zero()
	for k := range f.ZOpenings {
		temp := run.GetLocalPointEvalParams(f.ZOpenings[k].ID).Y
		zSum.Add(&zSum, &temp)
	}

	claimedSum := run.GetLogDerivSumParams(f.LogDeriveSumID).Sum
	if zSum != claimedSum {
		return fmt.Errorf("log-derivate lookup, the final evaluation check failed for %v,", f.Name)
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction]
func (f *FinalEvaluationCheck) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {

	claimedSum := run.GetLogDerivSumParams(f.LogDeriveSumID)
	// SigmaSKSum stores the sum of the ending values of the SigmaSs as queried
	// in the protocol via the
	zSum := frontend.Variable(field.Zero())
	for k := range f.ZOpenings {
		temp := run.GetLocalPointEvalParams(f.ZOpenings[k].ID).Y
		zSum = api.Add(zSum, temp)
	}

	api.AssertIsEqual(zSum, claimedSum)
}

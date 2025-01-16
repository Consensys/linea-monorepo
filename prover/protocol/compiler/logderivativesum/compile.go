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

	// Collect all the logDerivativeSum queries
	for _, qName := range comp.QueriesParams.AllUnignoredKeys() {

		// Filter out non other types of queries
		logDeriv, ok := comp.QueriesParams.Data(qName).(query.LogDerivativeSum)
		if !ok {
			continue
		}

		// This ensures that the LogDerivativeSum query is not used again in the
		// compilation process. We know that the query was already ignored at
		// the beginning because we are iterating over the unignored keys.
		comp.QueriesParams.MarkAsIgnored(qName)

		var (
			zEntries = logDeriv.Inputs
			va       = FinalEvaluationCheck{
				LogDerivSumID: qName,
			}
			lastRound = logDeriv.Round
		)

		for _, entry := range zEntries {

			// get the Numerator and Denominator from the input and prepare their compilation.
			zC := &lookup.ZCtx{
				Round:            lastRound,
				Size:             entry.Size,
				SigmaNumerator:   entry.Numerator,
				SigmaDenominator: entry.Denominator,
			}

			// z-packing compile; it imposes the correct accumulation over Numerator and Denominator.
			zC.Compile(comp)

			// prover step; Z assignments
			zAssignmentTask := lookup.ZAssignmentTask(*zC)
			comp.SubProvers.AppendToInner(zC.Round, func(run *wizard.ProverRuntime) {
				zAssignmentTask.Run(run)
			})

			// collect all the zOpening for all the z columns
			va.ZOpenings = append(va.ZOpenings, zC.ZOpenings...)
		}

		// verifer step
		comp.RegisterVerifierAction(lastRound, &va)
	}

}

type FinalEvaluationCheck struct {
	// ZOpenings lists all the openings of all the zCtx
	ZOpenings []query.LocalOpening
	// query ID
	LogDerivSumID ifaces.QueryID
}

// Run implements the [wizard.VerifierAction]
func (f *FinalEvaluationCheck) Run(run wizard.Runtime) error {

	// zSum stores the sum of the ending values of the zs as queried
	// in the protocol via the local opening queries.
	zSum := field.Zero()
	for k := range f.ZOpenings {
		temp := run.GetLocalPointEvalParams(f.ZOpenings[k].ID).Y
		zSum.Add(&zSum, &temp)
	}

	claimedSum := run.GetLogDerivSumParams(f.LogDerivSumID).Sum
	if zSum != claimedSum {
		return fmt.Errorf("log-derivate-sum; the final evaluation check failed for %v\n"+
			"given %v but calculated %v,",
			f.LogDerivSumID, claimedSum.String(), zSum.String())
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction]
func (f *FinalEvaluationCheck) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	claimedSum := run.GetLogDerivSumParams(f.LogDerivSumID).Sum
	// SigmaSKSum stores the sum of the ending values of the SigmaSs as queried
	// in the protocol via the
	zSum := frontend.Variable(field.Zero())
	for k := range f.ZOpenings {
		temp := run.GetLocalPointEvalParams(f.ZOpenings[k].ID).Y
		zSum = api.Add(zSum, temp)
	}

	api.AssertIsEqual(zSum, claimedSum)
}

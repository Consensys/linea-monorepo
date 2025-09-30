package logderivativesum

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"slices"

	"github.com/consensys/gnark/frontend"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// compile [query.LogDerivativeSum] query
func CompileLogDerivativeSum[T zk.Element](comp *wizard.CompiledIOP[T][T]) {

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
			va       = FinalEvaluationCheck[T]{
				LogDerivSumID: qName,
			}
			lastRound   = logDeriv.Round
			proverTasks = make([]ProverTaskAtRound[T], lastRound+1)
		)

		sizes := utils.MapFunc(zEntries.Parts, func(part query.LogDerivativeSumPart[T]) int {
			return part.Size
		})

		slices.SortFunc(sizes, func(a, b int) int { return a - b }) // sorts the slice of integers
		sizes = slices.Compact(sizes)
		sizes = slices.Clip(sizes)

		for _, size := range sizes {

			if !utils.IsPowerOfTwo(size) {
				utils.Panic("the size of a log-derivative cannot be a non-power of two: %v", size)
			}

			// get the Numerator and Denominator from the input and prepare their compilation.
			zC := ZCtx[T]{
				Round: lastRound,
				Size:  size,
			}

			// This accumulates all the parts whose size is equal to size in zC
			for _, part := range zEntries.Parts {
				if part.Size == size {
					zC.SigmaNumerator = append(zC.SigmaNumerator, part.Num)
					zC.SigmaDenominator = append(zC.SigmaDenominator, part.Den)
				}
			}

			// z-packing compile; it imposes the correct accumulation over Numerator and Denominator.
			zC.Compile(comp)

			// prover step; Z assignments
			zAssignmentTask := ZCtx[T](zC)
			proverTasks[zC.Round].pushZAssignment(zAssignmentTask)

			// collect all the zOpening for all the z columns
			va.ZOpenings = append(va.ZOpenings, zC.ZOpenings...)
		}

		for round, task := range proverTasks {
			if task.numTasks() > 0 {
				comp.RegisterProverAction(round, task)
			}
		}

		// verifer step
		comp.RegisterVerifierAction(lastRound, &va)
	}

}

type FinalEvaluationCheck[T zk.Element] struct {
	// ZOpenings lists all the openings of all the zCtx
	ZOpenings []query.LocalOpening[T]
	// query ID
	LogDerivSumID ifaces.QueryID
	// skip verifier action
	skipped bool `serde:"omit"`
}

// Run implements the [wizard.VerifierAction]
func (f *FinalEvaluationCheck[T]) Run(run wizard.Runtime) error {

	tmps := make([]fext.GenericFieldElem, 0)

	// zSum stores the sum of the ending values of the zs as queried
	// in the protocol via the local opening queries.
	zSum := fext.GenericFieldZero()

	for k := range f.ZOpenings {
		zOpening := run.GetLocalPointEvalParams(f.ZOpenings[k].ID)
		temp := zOpening.ToGenericGroupElement()
		tmps = append(tmps, temp)
		zSum.Add(&temp)
	}

	logDerivSumParam := run.GetLogDerivSumParams(f.LogDerivSumID)
	claimedSum := logDerivSumParam.Sum

	if !zSum.IsEqual(&claimedSum) {
		return fmt.Errorf("log-derivate-sum; the final evaluation check failed for %v\n"+
			"given %v but calculated %v\n"+
			"partial-sums=(len %v) %v",
			f.LogDerivSumID, claimedSum.String(), zSum.String(), len(tmps), vectorext.PrettifyGeneric(tmps),
		)
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction]
func (f *FinalEvaluationCheck[T]) RunGnark(api frontend.API, run wizard.GnarkRuntime[T]) {

	claimedSum := run.GetLogDerivSumParams(f.LogDerivSumID).Sum
	// SigmaSKSum stores the sum of the ending values of the SigmaSs as queried
	// in the protocol via the
	zSum := T(field.Zero())
	for k := range f.ZOpenings {
		// not using panic so that the compiler does not ask us to comment out
		// the remainder of the loop
		utils.Panic("this shoud default to extension fields")
		temp := run.GetLocalPointEvalParams(f.ZOpenings[k].ID).ExtY
		zSum = api.Add(zSum, temp)
	}

	api.AssertIsEqual(zSum, claimedSum)
}

func (f *FinalEvaluationCheck[T]) Skip() {
	f.skipped = true
}

func (f *FinalEvaluationCheck[T]) IsSkipped() bool {
	return f.skipped
}

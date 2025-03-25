package permutation

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// proverTaskAtRound implements the [wizard.ProverAction] interface and is
// responsible for assigning the Z polynomials of all the queries for which the
// Z polynomial needs to be assigned in the current round
type proverTaskAtRound []*ZCtx

// Run implements the [wizard.ProverAction interface]. The tasks will spawn
// a goroutine for each tasks and wait for all of them to finish. The approach
// for parallelization can be justified if the number of go-routines stays low
// (e.g. less than 1000s).
func (p proverTaskAtRound) Run(run *wizard.ProverRuntime) {

	wg := &sync.WaitGroup{}
	wg.Add(len(p))

	for i := range p {
		// the passing of the index `i` is there to ensure that the go-routine
		// is running over a local copy of `i` which is not incremented every
		// time the loop goes to the next iteration.
		go func(i int) {
			p[i].run(run)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

// run assigns all the Zs in parallel and set the parameters for their
// corresponding last values openings.
func (z *ZCtx) run(run *wizard.ProverRuntime) {

	for i := range z.Zs {
		var (
			numerator   []field.Element
			denominator []field.Element
		)

		if packingArity*i < len(z.NumeratorFactors) {
			numerator = column.EvalExprColumn(run, z.NumeratorFactorsBoarded[i]).IntoRegVecSaveAlloc()
		} else {
			numerator = vector.Repeat(field.One(), z.Size)
		}

		if packingArity*i < len(z.DenominatorFactors) {
			denominator = column.EvalExprColumn(run, z.DenominatorFactorsBoarded[i]).IntoRegVecSaveAlloc()
		} else {
			denominator = vector.Repeat(field.One(), z.Size)
		}

		denominator = field.BatchInvert(denominator)

		for i := range denominator {
			numerator[i].Mul(&numerator[i], &denominator[i])
			if i > 0 {
				numerator[i].Mul(&numerator[i], &numerator[i-1])
			}
		}

		run.AssignColumn(z.Zs[i].GetColID(), smartvectors.NewRegular(numerator))
		run.AssignLocalPoint(z.ZOpenings[i].Name(), numerator[len(numerator)-1])
	}

}

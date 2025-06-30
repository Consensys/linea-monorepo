package permutation

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// ProverTaskAtRound implements the [wizard.ProverAction] interface and is
// responsible for assigning the Z polynomials of all the queries for which the
// Z polynomial needs to be assigned in the current round
type ProverTaskAtRound []*ZCtx

// Run implements the [wizard.ProverAction interface]. The tasks will spawn
// a goroutine for each tasks and wait for all of them to finish. The approach
// for parallelization can be justified if the number of go-routines stays low
// (e.g. less than 1000s).
func (p ProverTaskAtRound) Run(run *wizard.ProverRuntime) {

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

// run assigns all the Zs in parallel and set the parameters for their
// corresponding last values openings.
func (z *ZCtx) run(run *wizard.ProverRuntime) {
	for i := range z.Zs {
		var (
			numerator   smartvectors.SmartVector
			denominator smartvectors.SmartVector
		)

		if packingArity*i < len(z.NumeratorFactors) {
			numerator = column.EvalExprColumn(run, z.NumeratorFactorsBoarded[i])
		} else {
			numerator = smartvectors.NewConstant(field.One(), z.Size)
		}

		if packingArity*i < len(z.DenominatorFactors) {
			denominator = column.EvalExprColumn(run, z.DenominatorFactorsBoarded[i])
		} else {
			denominator = smartvectors.NewConstant(field.One(), z.Size)
		}

		if smartvectors.IsBase(numerator) && smartvectors.IsBase(denominator) {
			// If both numerator and denominator are base
			denominatorSlice, _ := denominator.IntoRegVecSaveAllocBase()
			denominatorSlice = field.BatchInvert(denominatorSlice)

			numeratorSlice, _ := numerator.IntoRegVecSaveAllocBase()

			for i := range denominatorSlice {
				numeratorSlice[i].Mul(&numeratorSlice[i], &denominatorSlice[i])
				if i > 0 {
					numeratorSlice[i].Mul(&numeratorSlice[i], &numeratorSlice[i-1])
				}
			}

			run.AssignColumn(z.Zs[i].GetColID(), smartvectors.NewRegular(numeratorSlice))
			run.AssignLocalPoint(z.ZOpenings[i].Name(), numeratorSlice[len(numeratorSlice)-1])
		} else {
			// at least one of the numerator or denominator is over field extensions
			denominatorSlice := denominator.IntoRegVecSaveAllocExt()
			denominatorSlice = fext.BatchInvert(denominatorSlice)

			numeratorSlice := numerator.IntoRegVecSaveAllocExt()

			for i := range denominatorSlice {
				numeratorSlice[i].Mul(&numeratorSlice[i], &denominatorSlice[i])
				if i > 0 {
					numeratorSlice[i].Mul(&numeratorSlice[i], &numeratorSlice[i-1])
				}
			}

			run.AssignColumn(z.Zs[i].GetColID(), smartvectors.NewRegularExt(numeratorSlice))
			run.AssignLocalPointExt(z.ZOpenings[i].Name(), numeratorSlice[len(numeratorSlice)-1])
		}

	}
}

// AssignPermutationGranddProduct assigns the grand product query
type AssignPermutationGrandProduct struct {
	Query *query.GrandProduct
	// IsPartial indicates that the permuation queries contains public
	// terms to evaluate explictly by the verifier. In that case, the
	// result of the query is not one and must be computed explicitly.
	IsPartial bool
}

func (a AssignPermutationGrandProduct) Run(run *wizard.ProverRuntime) {
	y := fext.One()
	if a.IsPartial {
		y = a.Query.Compute(run)
	}
	run.AssignGrandProduct(a.Query.ID, y)
}

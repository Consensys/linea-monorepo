package inclusion

import (
	"runtime/debug"
	"sync"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	lookUp "github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// proverTaskAtRound implements the [wizard.ProverAction] interface. It gathers
// all the operations related to all compiled tables altogether that have to be
// done at a particular round.
//
// Namely, if applied to the round N. The action will be responsible for
// assigning the M column for tables compiled on round N and the SigmaS/SigmaT
// and their respective LocalOpening for the tables compiled at round N-1.
//
// All these actions are performed in parallel.
type proverTaskAtRound struct {

	// ZAssignmentTasks lists all the tasks consisting of assigning the
	// columns SigmaS and SigmaT for the given round.
	ZAssignmentTasks []zAssignmentTask
}

// Run implements the [wizard.ProverAction interface]. The tasks will spawn
// a goroutine for each tasks and wait for all of them to finish. The approach
// for parallelization can be justified if the number of go-routines stays low
// (e.g. less than 1000s).
func (p proverTaskAtRound) Run(run *wizard.ProverRuntime) {

	wg := &sync.WaitGroup{}
	wg.Add(p.numTasks())

	var (
		panicTrace []byte
		panicMsg   any
		panicOnce  = &sync.Once{}
	)

	for i := range p.ZAssignmentTasks {
		// the passing of the index `i` is there to ensure that the go-routine
		// is running over a local copy of `i` which is not incremented every
		// time the loop goes to the next iteration.
		go func(i int) {

			// In case the subtask panics, we recover so that we can repanic in
			// the main goroutine. Simplifying the process of tracing back the
			// error and allowing to test the panics.
			defer func() {
				if r := recover(); r != nil {
					panicOnce.Do(func() {
						panicMsg = r
						panicTrace = debug.Stack()
					})
				}

				wg.Done()
			}()

			p.ZAssignmentTasks[i].run(run)
		}(i)
	}

	wg.Wait()

	if len(panicTrace) > 0 {
		utils.Panic("Had a panic: %v\nStack: %v\n", panicMsg, string(panicTrace))
	}
}

// pushZAssignment appends an [sigmaAssignmentTask] to the list of tasks
func (p *proverTaskAtRound) pushZAssignment(s zAssignmentTask) {
	p.ZAssignmentTasks = append(p.ZAssignmentTasks, s)
}

// numTasks returns the total number of tasks that are scheduled in the
// [proverTaskAtRound].
func (p *proverTaskAtRound) numTasks() int {
	return len(p.ZAssignmentTasks)
}

// zAssignmentTask represents a prover task of assignming the columns
// SigmaS and SigmaT for a specific lookup table.
// sigmaAssignment
type zAssignmentTask lookUp.ZCtx

func (z zAssignmentTask) run(run *wizard.ProverRuntime) {
	parallel.Execute(len(z.ZDenominatorBoarded), func(start, stop int) {
		for frag := start; frag < stop; frag++ {

			var (
				numeratorMetadata = z.ZNumeratorBoarded[frag].ListVariableMetadata()
				denominator       = wizardutils.EvalExprColumn(run, z.ZDenominatorBoarded[frag]).IntoRegVecSaveAlloc()
				numerator         []field.Element
				packedZ           = field.BatchInvert(denominator)
			)

			if len(numeratorMetadata) == 0 {
				numerator = vector.Repeat(field.One(), z.Size)
			}

			if len(numeratorMetadata) > 0 {
				numerator = wizardutils.EvalExprColumn(run, z.ZNumeratorBoarded[frag]).IntoRegVecSaveAlloc()
			}

			for k := range packedZ {
				packedZ[k].Mul(&numerator[k], &packedZ[k])
				if k > 0 {
					packedZ[k].Add(&packedZ[k], &packedZ[k-1])
				}
			}

			run.AssignColumn(z.Zs[frag].GetColID(), sv.NewRegular(packedZ))
			run.AssignLocalPoint(z.ZOpenings[frag].ID, packedZ[len(packedZ)-1])
		}
	})
}

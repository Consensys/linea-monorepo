package innerproduct

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// ProverTask implements the [wizard.ProverAction] interface and as such
// implements the prover work of the compilation step. It works by calling
// in parallel the prover tasks of the sub-compilation steps.
type ProverTask[T zk.Element] struct {
	Pt []*ContextForSize[T]
}

// Run implements the [wizard.ProverAction] interface.
func (p ProverTask[T]) Run(run *wizard.ProverRuntime) {

	wg := &sync.WaitGroup{}
	wg.Add(len(p.Pt))

	for i := range p.Pt {
		// Passing the loop index ensures each go routine is storing the value
		// of i in a different variable so that there is no race condition over
		// i.
		go func(i int) {
			p.Pt[i].run(run)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

// run partially implements the prover runtime associated with the current
// partial compilation context. Its role is to assign Summation and its opening.
func (ctx *ContextForSize[T]) run(run *wizard.ProverRuntime) {

	var (
		size      = ctx.Summation.Size()
		collapsed = column.EvalExprColumn(run, ctx.CollapsedBoard).IntoRegVecSaveAllocExt()
		summation = make([]fext.Element, size)
	)

	summation[0] = collapsed[0]
	for i := 0; i+1 < size; i++ {
		summation[i+1].Add(&summation[i], &collapsed[i+1])
	}

	run.AssignColumn(ctx.Summation.GetColID(), smartvectors.NewRegularExt(summation))
	run.AssignLocalPointExt(ctx.SummationOpening.ID, summation[len(summation)-1])
}

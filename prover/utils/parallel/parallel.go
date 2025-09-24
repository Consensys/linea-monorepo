package parallel

import (
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// Execute process in parallel the work function
func Execute(nbIterations int, work func(int, int), maxCpus ...int) {
	nbTasks := runtime.GOMAXPROCS(0)

	if len(maxCpus) == 1 {
		nbTasks = maxCpus[0]
		if nbTasks < 1 {
			nbTasks = 1
		}
	}

	nbGoRoutines := runtime.NumGoroutine()
	nbCpus := runtime.GOMAXPROCS(0) * 3
	for nbTasks+nbGoRoutines > nbCpus && nbTasks > 1 {
		nbTasks--
	}
	nbTasks = max(1, nbTasks)

	if nbTasks == 1 {
		work(0, nbIterations)
		return
	}

	nbIterationsPerCpus := nbIterations / nbTasks

	// more CPUs than tasks: a CPU will work on exactly one iteration
	if nbIterationsPerCpus < 1 {
		nbIterationsPerCpus = 1
		nbTasks = nbIterations
	}

	var wg sync.WaitGroup

	extraTasks := nbIterations - (nbTasks * nbIterationsPerCpus)
	extraTasksOffset := 0

	var (
		panicTrace []byte
		panicMsg   any
		panicOnce  = &sync.Once{}
	)

	// stack := debug.Stack()

	// filters := []string{
	// 	"state-management/smt/tree.go",
	// 	"AssignSplitColumnProverAction",
	// }
	// skip := false
	// for _, f := range filters {
	// 	if strings.Contains(string(stack), f) {
	// 		skip = true
	// 		break
	// 	}
	// }

	// stack := string(debug.Stack())
	for i := 0; i < nbTasks; i++ {
		wg.Add(1)
		_start := i*nbIterationsPerCpus + extraTasksOffset
		_end := _start + nbIterationsPerCpus
		if extraTasks > 0 {
			_end++
			extraTasks--
			extraTasksOffset++
		}

		go func() {
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

			// start := time.Now()
			work(_start, _end)
			// took := time.Since(start)

			// if !skip && took < time.Millisecond {
			// 	// this is a very fast task
			// 	fmt.Printf("parallel task [%d,%d] took %s\n%s\n", _start, _end, took, stack)
			// }

		}()
	}

	wg.Wait()

	if len(panicTrace) > 0 {
		utils.Panic("Had a panic: %v\nStack: %v\n", panicMsg, string(panicTrace))
	}
}

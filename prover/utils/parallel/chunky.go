package parallel

import (
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

// Execute in parallel. This implementation is optimized for the case where
// the workload is a sequence of unbalanced and long computation time jobs.
func ExecuteChunky(nbIterations int, work func(start, stop int), numcpus ...int) {

	numcpu := runtime.GOMAXPROCS(0)
	if len(numcpus) > 0 && numcpus[0] > 0 {
		numcpu = numcpus[0]
	}

	// Then just call the trivial function
	if nbIterations < numcpu {
		logrus.Debugf("Loss in parallelization time numpcpu = %v and nbIterator = %v", numcpu, nbIterations)
		Execute(nbIterations, work, numcpu)
		return
	}

	// The jobs are sent one by one to the workers
	jobChan := make(chan int, nbIterations)
	for i := 0; i < nbIterations; i++ {
		jobChan <- i
	}

	// The wait group ensures that all the children goroutine have terminated
	// before we close the
	wg := sync.WaitGroup{}
	wg.Add(nbIterations)

	// Each goroutine consumes the jobChan to
	for p := 0; p < numcpu; p++ {
		go func() {
			for i := range jobChan {
				work(i, i+1)
				wg.Done()
			}
		}()
	}

	wg.Wait()
	close(jobChan)
}

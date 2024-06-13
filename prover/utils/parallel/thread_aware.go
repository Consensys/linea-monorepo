package parallel

import (
	"runtime"
	"sync"
)

type ThreadInit func(threadID int)
type Worker func(taskID, threadID int)

func ExecuteThreadAware(nbIterations int, init ThreadInit, worker Worker, numcpus ...int) {

	numcpu := runtime.GOMAXPROCS(0)
	if len(numcpus) > 0 && numcpus[0] > 0 {
		numcpu = numcpus[0]
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
		threadID := p
		go func() {
			init(threadID)
			for taskID := range jobChan {
				worker(taskID, threadID)
				wg.Done()
			}
		}()
	}

	wg.Wait()
	close(jobChan)

}

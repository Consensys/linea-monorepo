package parallel

import (
	"runtime"
	"sync"
)

var queue chan func() = make(chan func())
var available chan struct{} = make(chan struct{}, runtime.GOMAXPROCS(0))
var once sync.Once

func ExecutePool(task func()) {
	once.Do(run)

	ch := make(chan struct{}, 1)
	queue <- func() {
		task()
		close(ch)
	}

	<-ch
}

func ExecutePoolChunky(nbIterations int, work func(k int)) {
	wg := sync.WaitGroup{}
	wg.Add(nbIterations)

	for i := 0; i < nbIterations; i++ {
		k := i
		queue <- func() {
			work(k)
			wg.Done()
		}
	}

	wg.Wait()
}

func run() {
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		available <- struct{}{}
	}

	go scheduler()
}

func scheduler() {
	for {
		<-available
		task := <-queue
		go func() {
			task()
			available <- struct{}{}
		}()
	}
}

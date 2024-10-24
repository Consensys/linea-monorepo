package parallel

import (
	"runtime"
	"sync"
)

var queue chan func() = make(chan func())
var available chan struct{} = make(chan struct{})
var once sync.Once

func ExecutePool(task func()) chan struct{} {
	once.Do(run)

	ch := make(chan struct{}, 1)
	queue <- func() {
		task()
		close(ch)
	}
	return ch
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

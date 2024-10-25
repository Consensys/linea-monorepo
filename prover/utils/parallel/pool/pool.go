package pool

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
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

func ExecutePoolChunkyWithCache(nbIterations int, lagerPool mempool.MemPool, work func(k int, localPool *mempool.SliceArena)) {
	once.Do(run)

	wg := sync.WaitGroup{}
	wg.Add(nbIterations)

	pool := make(chan *mempool.SliceArena, runtime.GOMAXPROCS(0))
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		pool <- mempool.WrapsWithMemCache(lagerPool)
	}

	for i := 0; i < nbIterations; i++ {
		k := i
		localPool := <-pool

		queue <- func() {
			work(k, localPool)
			pool <- localPool
			wg.Done()
		}
	}

	wg.Wait()

	for localPool := range pool {
		localPool.TearDown()
	}

	close(pool)
}

func ExecutePoolChunky(nbIterations int, work func(k int)) {
	once.Do(run)

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

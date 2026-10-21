package predecoding

import (
	"runtime"
	"sync"
	"testing"
)

func TestCountValidWords(t *testing.T) {
	workers := runtime.NumCPU()
	const total = uint64(1) << 32
	chunk := total / uint64(workers)
	invalid := make([]uint64, workers)
	noop := make([]uint64, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := uint64(w) * chunk; i < uint64(w+1)*chunk; i++ {
				switch classifyInstruction(uint32(i)) {
				case computeInvalid:
					invalid[w]++
				case computeNoOp:
					noop[w]++
				}
			}
		}(w)
	}
	wg.Wait()
	var inv, nop uint64
	for w := range invalid { inv += invalid[w]; nop += noop[w] }
	t.Logf("total=%d invalid=%d noop=%d valid_nonnoop=%d", total, inv, nop, total-inv-nop)
}

package predecoding

import (
	"hash/fnv"
	"runtime"
	"sync"
	"testing"
	"time"
)

// Decodes every 32-bit word once and fingerprints the output.
func TestExhaustiveDecodeTiming(t *testing.T) {
	workers := runtime.NumCPU()
	const total = uint64(1) << 32
	chunk := total / uint64(workers)
	sums := make([]uint64, workers)
	start := time.Now()
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			h := fnv.New64a()
			var bw bitWriter
			for i := uint64(w) * chunk; i < uint64(w+1)*chunk; i++ {
				bw.buf = bw.buf[:0]
				bw = bitWriter{buf: bw.buf}
				decodeInstruction(&bw, uint32(i))
				h.Write(bw.buf)
			}
			sums[w] = h.Sum64()
		}(w)
	}
	wg.Wait()
	t.Logf("workers=%d words=%d elapsed=%s fp0=%x", workers, total, time.Since(start), sums[0])
}

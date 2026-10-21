package predecoding

import (
	"encoding/binary"
	"runtime"
	"sort"
	"sync"
	"testing"

	"golang.org/x/arch/riscv64/riscv64asm"
)

// Groups words Lineth accepts but riscv-opcodes rejects, by (opcode, funct3, compute_op).
func TestOverAcceptBreakdown(t *testing.T) {
	workers := runtime.NumCPU()
	const total = uint64(1) << 32
	type k struct{ op, f3, c uint32 }
	parts := make([]map[k]uint64, workers)
	smp := make([]map[k]uint32, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			m := map[k]uint64{}
			s := map[k]uint32{}
			var b [4]byte
			for i := total * uint64(w) / uint64(workers); i < total*uint64(w+1)/uint64(workers); i++ {
				x := uint32(i)
				if x&3 != 3 || x&0x7f == opcodeCUSTOM1 {
					continue
				}
				c := classifyInstruction(x)
				if c == computeInvalid {
					continue
				}
				binary.LittleEndian.PutUint32(b[:], x)
				if _, err := riscv64asm.Decode(b[:]); err == nil {
					continue
				}
				key := k{x & 0x7f, (x >> 12) & 7, c}
				m[key]++
				if _, ok := s[key]; !ok {
					s[key] = x
				}
			}
			parts[w], smp[w] = m, s
		}(w)
	}
	wg.Wait()
	all := map[k]uint64{}
	sm := map[k]uint32{}
	for w := range parts {
		for key, v := range parts[w] {
			all[key] += v
			if _, ok := sm[key]; !ok {
				sm[key] = smp[w][key]
			}
		}
	}
	keys := make([]k, 0, len(all))
	for key := range all {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return all[keys[i]] > all[keys[j]] })
	for _, key := range keys {
		t.Logf("opcode=%07b funct3=%03b compute_op=%3d count=%d sample=%#08x", key.op, key.f3, key.c, all[key], sm[key])
	}
}

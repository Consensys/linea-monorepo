package predecoding

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"testing"

	"golang.org/x/arch/riscv64/riscv64asm"
)

var allowed = map[string]bool{}

func init() {
	for _, s := range []string{"LUI", "AUIPC", "JAL", "JALR", "BEQ", "BNE", "BLT", "BGE", "BLTU", "BGEU",
		"LB", "LH", "LW", "LD", "LBU", "LHU", "LWU", "SB", "SH", "SW", "SD",
		"ADDI", "SLTI", "SLTIU", "XORI", "ORI", "ANDI", "SLLI", "SRLI", "SRAI",
		"ADD", "SUB", "SLL", "SLT", "SLTU", "XOR", "SRL", "SRA", "OR", "AND",
		"FENCE", "ECALL", "EBREAK",
		"ADDIW", "SLLIW", "SRLIW", "SRAIW", "ADDW", "SUBW", "SLLW", "SRLW", "SRAW",
		"MUL", "MULH", "MULHSU", "MULHU", "DIV", "DIVU", "REM", "REMU", "MULW", "DIVW", "DIVUW", "REMW", "REMUW"} {
		allowed[s] = true
	}
}

type res struct {
	mismatch map[string]uint64
	samples  map[string]uint32
	opMap    map[uint32]map[string]bool
}

// Compares Lineth's classifyInstruction against x/arch riscv64asm (generated from riscv-opcodes) on all 2^32 words.
func TestDifferentialAgainstRiscvOpcodes(t *testing.T) {
	workers := runtime.NumCPU()
	const total = uint64(1) << 32
	chunk := total / uint64(workers)
	results := make([]res, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			r := res{map[string]uint64{}, map[string]uint32{}, map[uint32]map[string]bool{}}
			var b [4]byte
			for i := uint64(w) * chunk; i < uint64(w+1)*chunk; i++ {
				x := uint32(i)
				l := classifyInstruction(x)
				lv := l != computeInvalid
				key := ""
				if x&3 != 3 {
					if lv {
						key = "lineth-valid-on-16bit-encoding"
					}
				} else if x&0x7f == opcodeCUSTOM1 {
					continue
				} else {
					binary.LittleEndian.PutUint32(b[:], x)
					inst, err := riscv64asm.Decode(b[:])
					name := ""
					if err == nil {
						name = inst.Op.String()
					}
					rv := err == nil && allowed[name]
					switch {
					case lv && !rv:
						key = "lineth-valid/ref-not-RV64IM:" + name
					case !lv && rv:
						key = "lineth-invalid/ref-RV64IM:" + name
					case lv && rv && l != computeNoOp:
						m := r.opMap[l]
						if m == nil {
							m = map[string]bool{}
							r.opMap[l] = m
						}
						m[name] = true
					}
				}
				if key != "" {
					r.mismatch[key]++
					if _, ok := r.samples[key]; !ok {
						r.samples[key] = x
					}
				}
			}
			results[w] = r
		}(w)
	}
	wg.Wait()
	mis := map[string]uint64{}
	smp := map[string]uint32{}
	ops := map[uint32]map[string]bool{}
	for _, r := range results {
		for k, v := range r.mismatch {
			mis[k] += v
			if _, ok := smp[k]; !ok {
				smp[k] = r.samples[k]
			}
		}
		for k, m := range r.opMap {
			if ops[k] == nil {
				ops[k] = map[string]bool{}
			}
			for n := range m {
				ops[k][n] = true
			}
		}
	}
	keys := make([]string, 0, len(mis))
	for k := range mis {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("MISMATCH %-55s count=%d sample=%#08x", k, mis[k], smp[k])
	}
	multi := 0
	rev := map[string][]uint32{}
	for op, names := range ops {
		if len(names) > 1 {
			multi++
			t.Logf("compute_op %d maps to several ref ops: %v", op, names)
		}
		for n := range names {
			rev[n] = append(rev[n], op)
		}
	}
	for n, l := range rev {
		if len(l) > 1 {
			t.Logf("ref op %s maps to several compute_ops: %v", n, l)
		}
	}
	t.Log(fmt.Sprintf("summary: mismatch-categories=%d compute_ops-seen=%d ambiguous=%d", len(mis), len(ops), multi))
}

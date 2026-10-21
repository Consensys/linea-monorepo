package predecoding

import (
	"encoding/binary"
	"runtime"
	"sort"
	"sync"
	"testing"

	"golang.org/x/arch/riscv64/riscv64asm"
)

// allowed is the set of reference-decoder mnemonics (as printed by
// riscv64asm's Op.String()) that count as "valid RV64IM" for this test.
// The reference decoder understands far more than RV64IM (A, F, D, C, Zicsr,
// ...). Lineth only supports RV64IM, so anything outside this set is treated
// as invalid on the reference side.
var allowed = map[string]bool{}

// init fills `allowed` from a hardcoded list of RV64IM mnemonics, grouped as:
//   - RV32I base: U-type (LUI, AUIPC), jumps (JAL, JALR), branches (BEQ..BGEU),
//     loads (LB..LHU), stores (SB..SW), ALU immediates (ADDI..SRAI),
//     ALU register-register (ADD..AND), and system (FENCE, ECALL, EBREAK)
//   - RV64I additions: LWU/LD/SD and the 32-bit "W" ops (ADDIW, SLLIW, ..., SRAW)
//   - M extension: MUL/MULH/MULHSU/MULHU, DIV/DIVU, REM/REMU and their W variants
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

// res is one worker's partial result; results are merged after all workers
// finish, so no locking is needed while scanning.
type res struct {
	// mismatch counts words per mismatch category (see the key names below).
	mismatch map[string]uint64
	// samples keeps the first word seen per mismatch category, for reproduction.
	samples map[string]uint32
	// opMap records, for words both sides accept, which reference mnemonics
	// were seen for each Lineth compute_op.
	opMap map[uint32]map[string]bool
}

// Compares Lineth's classifyInstruction against x/arch riscv64asm (generated from riscv-opcodes) on all 2^32 words.
// It only logs; it never fails the test, so read the -v output for mismatches.
func TestDiffRefRVOP(t *testing.T) {
	// Split the 2^32 input space into one contiguous chunk per CPU.
	workers := runtime.NumCPU()
	const total = uint64(1) << 32
	results := make([]res, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			r := res{map[string]uint64{}, map[string]uint32{}, map[uint32]map[string]bool{}}
			var b [4]byte
			for i := total * uint64(w) / uint64(workers); i < total*uint64(w+1)/uint64(workers); i++ {
				x := uint32(i)
				// Lineth's verdict: anything other than computeInvalid is "accepted".
				l := classifyInstruction(x)
				lv := l != computeInvalid
				// key names the mismatch category for this word; empty means no mismatch.
				key := ""
				if x&3 != 3 {
					// Low two bits != 0b11 means a 16-bit compressed (RVC) encoding.
					// Lineth has no C extension, so it must reject all of these.
					if lv {
						key = "lineth-valid-on-16bit-encoding"
					}
				} else if x&0x7f == opcodeCUSTOM1 {
					// Lineth-specific custom opcode; the reference decoder knows nothing
					// about it, so there is nothing to compare against.
					continue
				} else {
					// Decode with the reference (riscv64asm expects little-endian bytes).
					binary.LittleEndian.PutUint32(b[:], x)
					inst, err := riscv64asm.Decode(b[:])
					name := ""
					if err == nil {
						name = inst.Op.String()
					}
					// Reference verdict: decodes successfully AND is an RV64IM mnemonic.
					rv := err == nil && allowed[name]
					switch {
					case lv && !rv:
						// Lineth accepts a word the spec says is not RV64IM (soundness hole).
						key = "lineth-valid/ref-not-RV64IM:" + name
					case !lv && rv:
						// Lineth rejects a real RV64IM instruction (completeness gap).
						key = "lineth-invalid/ref-RV64IM:" + name
					case lv && rv && l != computeNoOp:
						// Both accept: record which reference mnemonic landed in which
						// compute_op, to check the mapping is consistent below.
						// computeNoOp is excluded because it intentionally groups many ops.
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

	// Merge per-worker results.
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
	// Report mismatch categories in sorted order, with count and one sample word.
	keys := make([]string, 0, len(mis))
	for k := range mis {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("MISMATCH %-55s count=%d sample=%#08x", k, mis[k], smp[k])
	}
	// Mapping consistency, forward: one compute_op covering several reference ops.
	// Also builds the reverse index (reference op -> compute_ops) for the check below.
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
	// Mapping consistency, reverse: one reference op split across several compute_ops.
	for n, l := range rev {
		if len(l) > 1 {
			t.Logf("ref op %s maps to several compute_ops: %v", n, l)
		}
	}
	t.Logf("summary: mismatch-categories=%d compute_ops-seen=%d ambiguous=%d", len(mis), len(ops), multi)
}

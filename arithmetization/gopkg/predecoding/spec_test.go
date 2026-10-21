package predecoding

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"testing"

	"golang.org/x/arch/riscv64/riscv64asm"
)

// Checks every 32-bit word's full record (compute_op, imm, rs1, rs2, rd)
// against an independent spec: compute_op numbering from constants.zkc (what
// the interpreter executes), instruction identity and operands from
// riscv64asm (generated from riscv-opcodes), and the documented x0 folding.

type class int

const (
	cR class = iota
	cI
	cSH
	cLOAD
	cJALR
	cSTORE
	cBR
	cJAL
	cU
	cSYS
	cFENCE
)

type specEntry struct {
	name string // constants.zkc name; JAL/JALR use base name, WB variant derived
	cl   class
	fold bool // rd==x0 folds to NO_OP
}

var spec = map[string]specEntry{
	"LB": {"READ8_SGN_WB", cLOAD, true}, "LH": {"READ16_SGN_WB", cLOAD, true}, "LW": {"READ32_SGN_WB", cLOAD, true},
	"LD": {"READ64_WB", cLOAD, true}, "LBU": {"READ8_ZEXT_WB", cLOAD, true}, "LHU": {"READ16_ZEXT_WB", cLOAD, true},
	"LWU": {"READ32_ZEXT_WB", cLOAD, true},
	"ADDI": {"OP_ADDI_WB", cI, true}, "SLTI": {"OP_SLTI_WB", cI, true}, "SLTIU": {"OP_SLTIU_WB", cI, true},
	"XORI": {"OP_XORI_WB", cI, true}, "ORI": {"OP_ORI_WB", cI, true}, "ANDI": {"OP_ANDI_WB", cI, true},
	"SLLI": {"OP_SLLI_WB", cSH, true}, "SRLI": {"OP_SRLI_WB", cSH, true}, "SRAI": {"OP_SRAI_WB", cSH, true},
	"ADDIW": {"OP_ADDIW_WB", cI, true}, "SLLIW": {"OP_SLLIW_WB", cSH, true}, "SRLIW": {"OP_SRLIW_WB", cSH, true},
	"SRAIW": {"OP_SRAIW_WB", cSH, true},
	"JALR": {"ITYPE_JALR", cJALR, false}, "ECALL": {"ITYPE_ECALL", cSYS, false}, "EBREAK": {"ITYPE_EBREAK", cSYS, false},
	"ADD": {"RTYPE_ADD_WB", cR, true}, "SUB": {"RTYPE_SUB_WB", cR, true}, "SLL": {"RTYPE_SLL_WB", cR, true},
	"SLT": {"RTYPE_SLT_WB", cR, true}, "SLTU": {"RTYPE_SLTU_WB", cR, true}, "XOR": {"RTYPE_XOR_WB", cR, true},
	"SRL": {"RTYPE_SRL_WB", cR, true}, "SRA": {"RTYPE_SRA_WB", cR, true}, "OR": {"RTYPE_OR_WB", cR, true},
	"AND": {"RTYPE_AND_WB", cR, true}, "MUL": {"RTYPE_MUL_WB", cR, true}, "MULH": {"RTYPE_MULH_WB", cR, true},
	"MULHSU": {"RTYPE_MULHSU_WB", cR, true}, "MULHU": {"RTYPE_MULHU_WB", cR, true}, "DIV": {"RTYPE_DIV_WB", cR, true},
	"DIVU": {"RTYPE_DIVU_WB", cR, true}, "REM": {"RTYPE_REM_WB", cR, true}, "REMU": {"RTYPE_REMU_WB", cR, true},
	"ADDW": {"RTYPE_ADDW_WB", cR, true}, "SUBW": {"RTYPE_SUBW_WB", cR, true}, "SLLW": {"RTYPE_SLLW_WB", cR, true},
	"SRLW": {"RTYPE_SRLW_WB", cR, true}, "SRAW": {"RTYPE_SRAW_WB", cR, true}, "MULW": {"RTYPE_MULW_WB", cR, true},
	"DIVW": {"RTYPE_DIVW_WB", cR, true}, "DIVUW": {"RTYPE_DIVUW_WB", cR, true}, "REMW": {"RTYPE_REMW_WB", cR, true},
	"REMUW": {"RTYPE_REMUW_WB", cR, true},
	"SB": {"STYPE_STORE8", cSTORE, false}, "SH": {"STYPE_STORE16", cSTORE, false}, "SW": {"STYPE_STORE32", cSTORE, false},
	"SD": {"STYPE_STORE64", cSTORE, false},
	"BEQ": {"BRANCH_BEQ", cBR, false}, "BNE": {"BRANCH_BNE", cBR, false}, "BLT": {"BRANCH_BLT", cBR, false},
	"BGE": {"BRANCH_BGE", cBR, false}, "BLTU": {"BRANCH_BLTU", cBR, false}, "BGEU": {"BRANCH_BGEU", cBR, false},
	"JAL": {"JTYPE_JAL", cJAL, false},
	"LUI": {"UTYPE_LUI_WB", cU, true}, "AUIPC": {"UTYPE_AUIPC_WB", cU, true},
	"FENCE": {"NO_OP", cFENCE, false},
}

func loadZkcConstants(t *testing.T) map[string]uint64 {
	b, err := os.ReadFile("../../src/main/common/constants.zkc")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?m)^const (\w+):ComputeOp = (\d+)`)
	m := map[string]uint64{}
	for _, g := range re.FindAllStringSubmatch(string(b), -1) {
		v, _ := strconv.ParseUint(g[2], 10, 64)
		m[g[1]] = v
	}
	return m
}

type rec struct{ op, imm, rs1, rs2, rd uint64 }

func readRecord(buf []byte) rec {
	pos := 0
	get := func(w int) uint64 {
		var v uint64
		for i := 0; i < w; i++ {
			bit := (buf[pos/8] >> (7 - uint(pos%8))) & 1
			v = v<<1 | uint64(bit)
			pos++
		}
		return v
	}
	return rec{get(8), get(64), get(5), get(5), get(5)}
}

func regNum(a riscv64asm.Arg) uint64 { return uint64(a.(riscv64asm.Reg) - riscv64asm.X0) }

func expected(inst riscv64asm.Inst, x uint32, c map[string]uint64) (rec, bool) {
	e, ok := spec[inst.Op.String()]
	if !ok {
		return rec{}, false
	}
	rd := uint64((x >> 7) & 31)
	var r rec
	switch e.cl {
	case cFENCE:
		return rec{op: 0}, true
	case cJAL:
		r.op = c["JTYPE_JAL"]
		if rd != 0 {
			r.op = c["JTYPE_JAL_WB"]
		}
	case cJALR:
		r.op = c["ITYPE_JALR"]
		if rd != 0 {
			r.op = c["ITYPE_JALR_WB"]
		}
	default:
		r.op = c[e.name]
		if e.fold && rd == 0 {
			return rec{op: 0}, true
		}
	}
	a := inst.Args
	switch e.cl {
	case cR:
		r.rd, r.rs1, r.rs2 = regNum(a[0]), regNum(a[1]), regNum(a[2])
	case cI:
		r.rd, r.rs1, r.imm = regNum(a[0]), regNum(a[1]), uint64(int64(a[2].(riscv64asm.Simm).Imm))
	case cSH:
		r.rd, r.rs1, r.imm = regNum(a[0]), regNum(a[1]), uint64(a[2].(riscv64asm.Uimm).Imm)
	case cLOAD, cJALR:
		o := a[1].(riscv64asm.RegOffset)
		r.rd, r.rs1, r.imm = regNum(a[0]), uint64(o.OfsReg-riscv64asm.X0), uint64(int64(o.Ofs.Imm))
	case cSTORE:
		o := a[1].(riscv64asm.RegOffset)
		r.rs2, r.rs1, r.imm = regNum(a[0]), uint64(o.OfsReg-riscv64asm.X0), uint64(int64(o.Ofs.Imm))
	case cBR:
		r.rs1, r.rs2, r.imm = regNum(a[0]), regNum(a[1]), uint64(int64(a[2].(riscv64asm.Simm).Imm))
	case cJAL:
		r.rd, r.imm = regNum(a[0]), uint64(int64(a[1].(riscv64asm.Simm).Imm))
	case cU:
		r.rd, r.imm = regNum(a[0]), uint64(int64(int32(a[1].(riscv64asm.Uimm).Imm<<12)))
	case cSYS:
		r.imm = uint64((x >> 20) & 0xfff)
	}
	return r, true
}

// Lineth-specific custom-1 spec (decode.go doc + constants.zkc): funct7 == 0,
// funct3 0/1/2 -> KECCAK/POSEIDON2/WRITE_OUTPUT with R-type operands, no x0 fold.
func expectedCustom1(x uint32, c map[string]uint64) rec {
	f3, f7 := (x>>12)&7, x>>25
	names := []string{"RTYPE_KECCAK", "RTYPE_POSEIDON2", "RTYPE_WRITE_OUTPUT"}
	if f7 != 0 || f3 > 2 {
		return rec{op: 255, rs1: uint64((x >> 15) & 31), rs2: uint64((x >> 20) & 31), rd: uint64((x >> 7) & 31)}
	}
	return rec{op: c[names[f3]], rs1: uint64((x >> 15) & 31), rs2: uint64((x >> 20) & 31), rd: uint64((x >> 7) & 31)}
}

func TestRecordsAgainstSpec(t *testing.T) {
	c := loadZkcConstants(t)
	if os.Getenv("MUTATE") == "1" { // negative control: swap ADD/SUB expectations
		a, b := spec["ADD"], spec["SUB"]
		a.name, b.name = b.name, a.name
		spec["ADD"], spec["SUB"] = a, b
	}
	t.Logf("constants.zkc: %d ComputeOp constants", len(c))
	workers := runtime.NumCPU()
	const total = uint64(1) << 32
	type res struct {
		checked, opMis, operandMis, custom uint64
		cats                               map[string]uint64
		samples                            map[string]string
	}
	results := make([]res, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			r := res{cats: map[string]uint64{}, samples: map[string]string{}}
			var b [4]byte
			var bw bitWriter
			lo := total * uint64(w) / uint64(workers)
			hi := total * uint64(w+1) / uint64(workers)
			note := func(k, s string) {
				r.cats[k]++
				if _, ok := r.samples[k]; !ok {
					r.samples[k] = s
				}
			}
			for i := lo; i < hi; i++ {
				x := uint32(i)
				if x&3 != 3 {
					continue
				}
				bw = bitWriter{buf: bw.buf[:0]}
				decodeInstruction(&bw, x)
				got := readRecord(bw.buf)
				if x&0x7f == opcodeCUSTOM1 {
					r.custom++
					want := expectedCustom1(x, c)
					if got.op != want.op {
						note("custom1-op", fmt.Sprintf("%#08x got=%d want=%d", x, got.op, want.op))
					} else if want.op != 255 && (got.imm != 0 || got.rs1 != want.rs1 || got.rs2 != want.rs2 || got.rd != want.rd) {
						note("custom1-operands", fmt.Sprintf("%#08x got=%+v want=%+v", x, got, want))
					}
					continue
				}
				binary.LittleEndian.PutUint32(b[:], x)
				inst, err := riscv64asm.Decode(b[:])
				if err != nil {
					continue
				}
				want, ok := expected(inst, x, c)
				if !ok {
					continue
				}
				r.checked++
				if got.op != want.op {
					r.opMis++
					note("op:"+inst.Op.String(), fmt.Sprintf("%#08x %v got=%d want=%d", x, inst, got.op, want.op))
					continue
				}
				if want.op == 0 {
					continue
				}
				if got != want {
					r.operandMis++
					note("operands:"+inst.Op.String(), fmt.Sprintf("%#08x %v got=%+v want=%+v", x, inst, got, want))
				}
			}
			results[w] = r
		}(w)
	}
	wg.Wait()
	var checked, opMis, operandMis, custom uint64
	cats := map[string]uint64{}
	samples := map[string]string{}
	for _, r := range results {
		checked += r.checked
		opMis += r.opMis
		operandMis += r.operandMis
		custom += r.custom
		for k, v := range r.cats {
			cats[k] += v
			if _, ok := samples[k]; !ok {
				samples[k] = r.samples[k]
			}
		}
	}
	keys := make([]string, 0, len(cats))
	for k := range cats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("MISMATCH %-22s count=%d e.g. %s", k, cats[k], samples[k])
	}
	t.Logf("summary: RV64IM words checked=%d op-mismatch=%d operand-mismatch=%d custom1-words=%d", checked, opMis, operandMis, custom)
	if len(cats) > 0 {
		t.Errorf("%d mismatch categories", len(cats))
	}
}

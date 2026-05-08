// GPU symbolic expression compiler.
//
// Compiles a DAG of arithmetic operations over E4 (KoalaBear degree-4 extension)
// into bytecode for parallel GPU evaluation.  One GPU thread per vector element,
// zero warp divergence — every thread executes the identical instruction stream.
//
//  ┌─────────────────────────┐         ┌──────────────────────────┐
//  │  NodeOp[] (topo-sorted) │         │  kern_symbolic_eval      │
//  │                         │         │                          │
//  │  liveness analysis      │  H2D    │  thread i:               │
//  │  register allocation    │ ──────▶ │    E4 slots[S]           │
//  │  bytecode emission      │         │    for pc in program:    │
//  │                         │         │      execute(i)          │
//  │  → GPUProgram           │         │    out[i] = slots[R]     │
//  └─────────────────────────┘         └──────────────────────────┘
//
// The NodeOp representation is decoupled from linea-monorepo's symbolic package.
// A thin adapter in the monorepo converts ExpressionBoard.Nodes[] → []NodeOp.
//
// Bytecode format (uint32 words):
//
//   OP_CONST   (0): [0, dst, const_idx]                        3 words
//   OP_INPUT   (1): [1, dst, input_id]                         3 words
//   OP_MUL     (2): [2, dst, n, s₀, e₀, ..., sₙ, eₙ]        3 + 2n words
//   OP_LINCOMB (3): [3, dst, n, s₀, c₀, ..., sₙ, cₙ]        3 + 2n words
//   OP_POLYEVAL(4): [4, dst, n, s₀, s₁, ..., sₙ]             3 + n words
package symbolic

// Opcodes — match CUDA kernel's switch cases exactly.
const (
	OpConst    = 0
	OpInput    = 1
	OpProduct  = 2
	OpLinComb  = 3
	OpPolyEval = 4
)

// NodeOp describes one node in a topologically-sorted expression DAG.
//
//	Kind=OpConst:    leaf, ConstVal holds the E4 value (Montgomery form)
//	Kind=OpInput:    leaf, references the next input variable
//	Kind=OpLinComb:  Σ Coeffs[k] · Children[k],   Coeffs = small integers
//	Kind=OpProduct:  Π Children[k]^Coeffs[k],      Coeffs = exponents ≥ 0
//	Kind=OpPolyEval: Horner(Children[0]=x, Children[1..]=coefficients)
type NodeOp struct {
	Kind     int
	Children []int      // indices into nodes array (child < self)
	Coeffs   []int      // LinComb: coefficients, Product: exponents
	ConstVal [4]uint32  // E4 constant, [b0.a0, b0.a1, b1.a0, b1.a1]
}

// GPUProgram holds compiled bytecode ready for GPU evaluation.
type GPUProgram struct {
	Bytecode   []uint32 // packed GPU instructions
	Constants  []uint32 // E4 constants (4 uint32 each)
	NumSlots   int
	ResultSlot int
	NumInputs  int
}

// CompileGPU compiles a topologically-sorted DAG into GPU bytecode.
//
// Algorithm (identical to CPU compiler in symbolic/compiler.go):
//  1. Liveness analysis — determine last use of each node
//  2. Register allocation — assign slots, reuse dead slots
//  3. Instruction emission — emit uint32 bytecode per node
func CompileGPU(nodes []NodeOp) *GPUProgram {
	n := len(nodes)
	if n == 0 {
		return &GPUProgram{}
	}

	// ── 1. Liveness: lastUse[i] = latest node index that reads node i ────
	lastUse := make([]int, n)
	for i := range lastUse {
		lastUse[i] = -1
	}
	lastUse[n-1] = n // output node is implicitly live
	for i, node := range nodes {
		for _, c := range node.Children {
			if i > lastUse[c] {
				lastUse[c] = i
			}
		}
	}

	// ── 2. Register allocation + 3. Instruction emission ─────────────────
	pgm := &GPUProgram{
		Bytecode:  make([]uint32, 0, n*4),
		Constants: make([]uint32, 0),
	}

	slots := make([]int, n)
	freeSlots := make([]int, 0, 32)
	nextSlot := 0

	alloc := func() int {
		if k := len(freeSlots); k > 0 {
			s := freeSlots[k-1]
			freeSlots = freeSlots[:k-1]
			return s
		}
		s := nextSlot
		nextSlot++
		return s
	}

	inputCursor := 0

	for i, node := range nodes {
		dst := alloc()
		slots[i] = dst

		switch node.Kind {
		case OpConst:
			ci := len(pgm.Constants) / 4
			pgm.Constants = append(pgm.Constants,
				node.ConstVal[0], node.ConstVal[1],
				node.ConstVal[2], node.ConstVal[3])
			pgm.Bytecode = append(pgm.Bytecode, OpConst, uint32(dst), uint32(ci))

		case OpInput:
			pgm.Bytecode = append(pgm.Bytecode, OpInput, uint32(dst), uint32(inputCursor))
			inputCursor++

		case OpLinComb:
			nc := len(node.Children)
			pgm.Bytecode = append(pgm.Bytecode, OpLinComb, uint32(dst), uint32(nc))
			for k, c := range node.Children {
				pgm.Bytecode = append(pgm.Bytecode,
					uint32(slots[c]), uint32(int32(node.Coeffs[k])))
			}

		case OpProduct:
			nc := len(node.Children)
			pgm.Bytecode = append(pgm.Bytecode, OpProduct, uint32(dst), uint32(nc))
			for k, c := range node.Children {
				pgm.Bytecode = append(pgm.Bytecode,
					uint32(slots[c]), uint32(node.Coeffs[k]))
			}

		case OpPolyEval:
			nc := len(node.Children)
			pgm.Bytecode = append(pgm.Bytecode, OpPolyEval, uint32(dst), uint32(nc))
			for _, c := range node.Children {
				pgm.Bytecode = append(pgm.Bytecode, uint32(slots[c]))
			}

		default:
			panic("vortex: CompileGPU: unknown node kind")
		}

		// Free dead children
		freed := make(map[int]bool)
		for _, c := range node.Children {
			if lastUse[c] == i && !freed[slots[c]] {
				freeSlots = append(freeSlots, slots[c])
				freed[slots[c]] = true
			}
		}
	}

	pgm.NumSlots = nextSlot
	pgm.ResultSlot = slots[n-1]
	pgm.NumInputs = inputCursor
	return pgm
}

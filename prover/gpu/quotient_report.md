# GPU Quotient Computation — Architecture Analysis & Plan

## Executive Summary

The quotient computation (`protocol/compiler/globalcs/quotient.go`) consists of two
distinct phases with very different GPU characteristics:

1. **Input preparation** — IFFT + FFT per root column per coset (embarrassingly parallel NTTs)
2. **Symbolic evaluation** — DAG-driven element-wise computation over prepared vectors

Your proposed idea — "homogenize everything to E4 on GPU, stream inputs with
per-input conversion instructions, then run a branchless bytecode VM" — is
**architecturally sound** and aligns well with what's already scaffolded in
`gpu/symbolic/`. The key insight is correct: the GPU doesn't benefit from the
CPU's SmartVector polymorphism; it benefits from uniformity.

However, the plan needs refinement in several areas. This report covers what
works, what doesn't, and proposes the optimal architecture.

---

## 1. Current CPU Quotient Pipeline

```
For each coset i ∈ [0, maxRatio):
  For each ratio group:
    ┌─ Phase 1: Input Preparation ──────────────────────────────────────┐
    │                                                                    │
    │  For each unique root column (parallel across roots):              │
    │    witness = run.Columns.Get(rootName)                             │
    │    if base:                                                        │
    │      IFFT(witness, DIF)  → coefficients                           │
    │      FFT(coeffs, DIT, OnCoset) → evaluations on coset             │
    │    else (extension):                                               │
    │      IFFTExt(witness, DIF) → coefficients                         │
    │      FFTExt(coeffs, DIT, OnCoset) → evaluations on coset          │
    │                                                                    │
    │  For each shifted column:                                          │
    │    SoftRotate(root_eval, offset) → rotated view (free, pointer)    │
    │                                                                    │
    │  For coins, accessors, X, PeriodicSample:                          │
    │    Materialize as Constant/ConstantExt/Regular SmartVectors        │
    └────────────────────────────────────────────────────────────────────┘
    ┌─ Phase 2: Expression Evaluation ──────────────────────────────────┐
    │                                                                    │
    │  For each constraint j in this ratio group:                        │
    │    evalInputs = [SmartVector for each board variable]              │
    │    quotientShare = board.Evaluate(evalInputs)                      │
    │      → Chunked parallel: 32-element chunks, vmBase or vmExt       │
    │      → Bytecode VM: opConst, opInput, opMul, opLinComb, opPolyEval│
    │                                                                    │
    │    quotientShare *= annulatorInv[i]  (vanishing poly division)     │
    │    run.AssignColumn(quotientShare)                                 │
    └────────────────────────────────────────────────────────────────────┘
```

### Cost breakdown (typical: DomainSize=2^20, ~200 roots, maxRatio=4)

| Phase | Work | CPU Estimate |
|-------|------|-------------|
| IFFT + FFT per root per coset | 4 × 200 × 2 FFTs of size 2^20 | ~60-80% of time |
| Expression evaluation | 4 × K boards × 2^20 elements | ~15-25% |
| Annulator, assignments, misc | Scalar mul + copies | ~5% |

**Key observation**: FFTs dominate. The symbolic evaluation is secondary but
still significant for large constraint counts.

---

## 2. Analysis of the Proposed "Homogeneous E4 GPU" Approach

### 2.1 What you proposed

> Allocate the space needed on device for the full evaluation, stream one by
> one the input with an "instruction" on how to convert it (fft, ifft, eval
> coset etc) in a staging area and copy it (on device) to a vector of field
> extension elements. That way, the actual compute is super simple and
> straightforward, branchless.

### 2.2 What's right about this

**The core principle is correct.** The GPU symbolic VM already works this way
(in `gpu/symbolic/` and `kern_symbolic_eval` in `kb.cu`):

- Every thread executes identical bytecode → zero warp divergence
- All slots are E4 (16 bytes) → uniform memory access pattern
- Input loading handles the heterogeneity via `SymInputDesc` tags:
  - tag=0 (KB): base field → embed as `(val, 0, 0, 0)` in E4
  - tag=1 (CONST_E4): broadcast constant to all threads
  - tag=2 (ROT_KB): rotated base field `d_ptr[(i+offset)%n]`
  - tag=3 (E4_VEC): native E4 vector

This is exactly the "instruction on how to convert" you described. The
heterogeneity is confined to input loading, and the compute is branchless.

### 2.3 What needs refinement

**Problem 1: FFTs shouldn't be "streamed as input conversion instructions"**

The proposed idea of streaming each input "with an instruction on how to convert
it (fft, ifft, eval coset etc)" conflates two very different operations:

- **Input conversion** (embed base→E4, broadcast constant, apply rotation):
  these are O(1) per element, perfectly suited for per-thread inline handling
  during the symbolic VM's `OP_INPUT` instruction.

- **FFT/IFFT**: these are O(n log n) operations with complex data-dependent
  access patterns (butterfly networks). They cannot be folded into the
  per-element symbolic VM. They require dedicated NTT kernels with shared memory
  tiling, multi-stage butterfly decomposition, and carefully tuned memory access
  patterns.

**Recommendation**: Keep FFTs as a separate phase. The GPU pipeline should be:

```
Phase 1: Batch NTT (all roots for this coset, in parallel)
Phase 2: Symbolic VM evaluation (bytecode, all inputs device-resident)
```

**Problem 2: The IFFT→FFT sequence should be fused/cached**

Currently, for each coset iteration, every root column does:
1. IFFT(witness) → coefficients
2. FFT(coefficients, coset_shift) → evaluations on coset

The IFFT produces the same coefficient vector regardless of coset. So:
- Cache the coefficient form after the first coset iteration
- For subsequent cosets, only do the coset FFT (saves 50% of NTTs)

On GPU, this is straightforward: keep `d_coefficients[root]` resident.

**Problem 3: Memory model for the symbolic VM**

The current CUDA kernel allocates `E4 slots[SYM_MAX_SLOTS]` (2048 slots) per
thread in local memory. Each slot is 16 bytes → 32 KB per thread.

With a GPU running ~100K concurrent threads: 32 KB × 100K = 3.2 GB just for
slots. This is fine for small-to-medium programs but becomes a concern if:
- `NumSlots` is large (many live intermediate values)
- `n` (domain size) is very large, requiring many concurrent threads

For most quotient expressions (10-100 nodes, 5-20 live slots after register
allocation), this works well. The liveness-based register allocator in
`CompileGPU()` already minimizes slot count.

**Problem 4: Output goes back to host (currently)**

`kb_sym_eval` writes results to `h_out` (host buffer) via D2H. But the quotient
shares are immediately assigned as columns and will later be committed via
Vortex. If the Vortex commit pipeline also runs on GPU, the result should stay
on device.

**Recommendation**: Add a `kb_sym_eval_device` variant that writes to a
device buffer, avoiding the D2H→H2D round-trip.

---

## 3. The Real Architecture Proposal

### 3.1 Two-phase GPU pipeline for quotient computation

```
┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 0: COMPILE-TIME PREPROCESSING (once, at compilation)         │
│                                                                     │
│  For each constraint j:                                             │
│    1. Convert ExpressionBoard.Nodes[] → NodeOp[]                    │
│       (adapter: map symbolic.Node operators to NodeOp kinds)        │
│    2. CompileGPU(nodeOps) → GPUProgram (bytecode + constants)       │
│    3. CompileSymGPU(dev, gpuProgram) → GPUSymProgram (device handle)│
│    4. Build input mapping:                                          │
│       variable[k] → { type: column|coin|X|periodic|accessor,       │
│                        rootColID, shiftOffset, ... }                │
│                                                                     │
│  Precompute:                                                        │
│    - NTT domain twiddle factors (cached in gpu.Device)              │
│    - Annulator inverse values (maxRatio elements)                   │
│    - Coset shift table for all coset IDs                            │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 1: BATCH NTT (per coset iteration)                           │
│                                                                     │
│  For coset i:                                                       │
│    If i == 0 (first coset):                                         │
│      For each unique root column:                                   │
│        H2D: witness → d_witness[root]                               │
│        Batch INTT: d_witness → d_coeffs[root]  (cache these!)      │
│        Batch NTT(coset_shift): d_coeffs → d_eval[root]             │
│    Else:                                                            │
│      For each unique root column:                                   │
│        copy d_coeffs[root] → d_eval[root]  (reuse cached coeffs)   │
│        Batch NTT(coset_shift): d_eval → d_eval[root]               │
│                                                                     │
│  For variables.X:                                                   │
│    Precompute d_x_coset[i] = [shift·ω^0, shift·ω^1, ..., shift·ω^(n-1)] │
│    (geometric sequence, one kernel launch)                          │
│                                                                     │
│  For PeriodicSample:                                                │
│    Precompute d_periodic[i] on device (small computation)           │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 2: SYMBOLIC VM (per constraint, per coset)                   │
│                                                                     │
│  For each constraint j in this ratio group:                         │
│    Build SymInput[] descriptors:                                    │
│      Column → SymInputKB(d_eval[root]) or SymInputRotKB(d_eval, offset) │
│      Coin → SymInputConstE4(val)                                   │
│      X → SymInputKB(d_x_coset) or SymInputE4Vec(d_x_coset)        │
│      PeriodicSample → SymInputKB(d_periodic)                       │
│      Accessor → SymInputConstE4(val)                                │
│                                                                     │
│    Launch kern_symbolic_eval:                                       │
│      → n threads, each executes bytecode, writes E4 result          │
│      → Output: d_quotient_share[j] (device-resident)               │
│                                                                     │
│    Scalar multiply by annulator inverse:                            │
│      d_quotient_share[j] *= annulatorInv[i]                        │
│      (one kernel: element-wise E4 × E4 scalar)                     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 3: OUTPUT                                                     │
│                                                                     │
│  Option A (current path): D2H each quotient share → host SmartVector│
│    → Assigned via run.AssignColumn()                                │
│    → Later committed via Vortex (potentially re-uploaded to GPU)    │
│                                                                     │
│  Option B (optimal): Keep on device                                 │
│    → Quotient shares stay as d_quotient_share[j]                    │
│    → Vortex commit reads them directly from GPU memory              │
│    → Zero D2H for quotient shares (enormous win for large domains)  │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Memory budget analysis

For DomainSize = 2^20 = 1M elements:

| Buffer | Size | Notes |
|--------|------|-------|
| d_witness[root] (200 roots) | 200 × 1M × 4B = 800 MB | H2D, then overwritten |
| d_coeffs[root] (cached) | 200 × 1M × 4B = 800 MB | Kept across cosets |
| d_eval[root] (per coset) | 200 × 1M × 4B = 800 MB | Rewritten each coset |
| d_x_coset | 1M × 4B = 4 MB | Recomputed per coset |
| d_quotient_share (per constraint) | 1M × 16B = 16 MB | E4 output per constraint |
| Symbolic VM slots (in registers/local) | ~32 KB per thread | Auto-managed |
| **Total for Phase 1+2** | **~2.4 GB** | Fits easily in 90+ GB |

For DomainSize = 2^22 = 4M elements:

| Buffer | Size | Notes |
|--------|------|-------|
| d_coeffs[root] (200 roots) | 200 × 4M × 4B = 3.2 GB | Cached |
| d_eval[root] | 200 × 4M × 4B = 3.2 GB | Per coset |
| d_quotient_shares (all) | K × 4M × 16B | Depends on K |
| **Total** | **~6.4 GB + outputs** | Fine for 90 GB |

For DomainSize = 2^24 = 16M elements (future scaling):

| Buffer | Size | Notes |
|--------|------|-------|
| d_coeffs[root] (200 roots) | 200 × 16M × 4B = 12.8 GB | |
| d_eval[root] | 200 × 16M × 4B = 12.8 GB | |
| **Total** | **~25.6 GB + outputs** | Still fits |

**Verdict**: Your observation that "our GPU has 90GB+ RAM so we should be fine
for the current benchmarks and tests" is correct. Even at DomainSize = 2^24 with
200 roots, we use ~26 GB for coefficients + evaluations. With aggressive root
count (500+), we might approach 50-60 GB, still within budget.

The real memory pressure comes not from the quotient alone but from the Vortex
commit pipeline running concurrently (encoded matrices, Merkle trees, etc.).
Coordinating GPU memory across quotient + Vortex is a scheduling problem, not a
fundamental limitation.

### 3.3 What about extension field columns?

Most witness columns are base field (4 bytes per element). Extension field
columns arise from:
- Random coin challenges (always constant-broadcast, negligible memory)
- Some computed intermediate columns

The existing code already handles this with `IsBase()` branching:
```go
if smartvectors.IsBase(witness) {
    // base FFT path (4 bytes/element)
} else {
    // extension FFT path (16 bytes/element)
}
```

On GPU, the NTT kernels already support both base and E4. The symbolic VM
operates entirely in E4. So the approach is:
- NTT base columns using base-field batch NTT (4x more efficient)
- NTT extension columns using E4 batch NTT
- Symbolic VM doesn't care — it reads via `SymInputDesc` which embeds base→E4

No issues here.

---

## 4. Critical Optimization: Caching Coefficient Form

This is the single highest-ROI optimization for the quotient, independent of GPU.

### Current flow (wasteful)
```
For coset 0: IFFT(witness) → coeffs → FFT(coeffs, coset_0) → eval_0
For coset 1: IFFT(witness) → coeffs → FFT(coeffs, coset_1) → eval_1   ← REDUNDANT IFFT!
For coset 2: IFFT(witness) → coeffs → FFT(coeffs, coset_2) → eval_2   ← REDUNDANT IFFT!
For coset 3: IFFT(witness) → coeffs → FFT(coeffs, coset_3) → eval_3   ← REDUNDANT IFFT!
```

### Optimized flow (cache coefficients)
```
IFFT(witness) → coeffs  [ONCE]
For coset 0: copy(coeffs) → FFT(copy, coset_0) → eval_0
For coset 1: copy(coeffs) → FFT(copy, coset_1) → eval_1
For coset 2: copy(coeffs) → FFT(copy, coset_2) → eval_2
For coset 3: copy(coeffs) → FFT(copy, coset_3) → eval_3
```

**Savings**: For 200 roots × maxRatio=4: eliminates 600 IFFTs. On GPU, each
batch IFFT of 200 polynomials of size 2^20 takes ~100-200ms, so this saves
~300-600ms. On CPU, the savings are proportionally larger (maybe 20-30s).

On GPU this is natural: `d_coeffs[root]` stays resident, and each coset
iteration does `cudaMemcpy(d_eval, d_coeffs, D2D)` + `batch_ntt(d_eval, coset_twiddles)`.

### Even better: fused IFFT→coset-FFT as twiddle scaling

Mathematically, evaluating a polynomial `f` on coset `g·H` (where `H` = domain
of roots of unity) is equivalent to:

```
f(g·ωⁱ) = IFFT of f → coefficients cₖ → multiply cₖ by gᵏ → FFT → evaluations
```

The "multiply `cₖ` by `gᵏ`" step is just an element-wise scaling. So:

```
Once:     IFFT(witness) → d_coeffs
Per coset: d_eval = d_coeffs ⊙ coset_scale_table[coset_id]
           FFT(d_eval) → evaluations on coset
```

Where `coset_scale_table[coset_id][k] = g^k · (ω_big^coset_id)^k`. This is a
precomputable geometric sequence.

The fused NTT tile kernel (`kern_batch_ntt_fused` in `kb.cu`) already supports
coset scaling as part of the butterfly decomposition. This means we can
potentially eliminate the separate D2D copy and scaling step entirely.

---

## 5. Should We Revamp the Symbolic Package?

### 5.1 No fundamental revamp needed

The existing symbolic package architecture is well-suited for GPU acceleration:

- **Topological sort + DAG deduplication**: essential for correct evaluation order,
  already done
- **Liveness analysis + register allocation**: directly maps to GPU slot allocation
  (already in `CompileGPU`)
- **Bytecode VM**: same opcodes work on both CPU and GPU (already verified)
- **Expression tree → Board → Bytecode pipeline**: clean separation of concerns

### 5.2 Adapter layer needed

What's missing is a clean adapter from `ExpressionBoard.Nodes[]` (which uses
Go interface types like `symbolic.LinComb`, `symbolic.Product`, etc.) to
`[]NodeOp` (the GPU-portable representation).

This adapter should:

```go
func BoardToNodeOps(board *symbolic.ExpressionBoard) []NodeOp {
    ops := make([]NodeOp, len(board.Nodes))
    for i, node := range board.Nodes {
        switch op := node.Operator.(type) {
        case symbolic.Constant:
            val := op.Val.GetExt()
            ops[i] = NodeOp{
                Kind: OpConst,
                ConstVal: [4]uint32{
                    uint32(val.B0.A0[0]), uint32(val.B0.A1[0]),
                    uint32(val.B1.A0[0]), uint32(val.B1.A1[0]),
                },
            }
        case symbolic.Variable:
            ops[i] = NodeOp{Kind: OpInput}
        case symbolic.LinComb:
            ops[i] = NodeOp{Kind: OpLinComb, Children: childIDs(node), Coeffs: op.Coeffs}
        case symbolic.Product:
            ops[i] = NodeOp{Kind: OpProduct, Children: childIDs(node), Coeffs: op.Exponents}
        case symbolic.PolyEval:
            ops[i] = NodeOp{Kind: OpPolyEval, Children: childIDs(node)}
        }
    }
    return ops
}
```

This is thin glue code, not a revamp.

### 5.3 What about the CPU path?

The CPU symbolic evaluator should remain unchanged. It's well-optimized for its
use case:
- Chunked parallel evaluation (32-element chunks for AVX-512)
- Base/extension field specialization
- SmartVector polymorphism avoids materializing constants

The GPU path is an alternative execution backend, not a replacement.

---

## 6. Memory Layout Alternatives Considered

### 6.1 "Flat materialized inputs" approach (your suggestion)

Allocate `NumInputs × DomainSize × 16 bytes` (all inputs as E4 vectors), then
run the symbolic VM reading from this flat buffer.

**Pros**: Simple, cache-friendly for input reads
**Cons**: Wastes memory. Base field columns occupy 4x more space than needed.
Constants broadcast to full vectors (DomainSize × 16B for a single scalar).

**Verdict**: The `SymInputDesc` tagged-input approach is strictly better. It
avoids materializing constants and keeps base fields compact. The per-thread
branching on `desc->tag` (4 cases) is perfectly predicted by the GPU — all
threads in a warp read the same tag, so there's zero divergence.

### 6.2 "Intermediate results as flat buffer" approach

Instead of per-thread `E4 slots[]` in local memory, use a global buffer:
`d_slots[slot_id × n + thread_id]`.

**Pros**: No per-thread stack pressure, can handle arbitrarily many slots
**Cons**: Much worse memory access pattern. Each bytecode instruction would
require global memory reads/writes instead of register/L1 access. For typical
programs (5-20 live slots), the per-thread approach with L1 caching dominates.

**Verdict**: Keep per-thread slots for programs with `NumSlots < ~64`. For
extremely large programs (unlikely in practice), fall back to global buffer.
The current `SYM_MAX_SLOTS = 2048` limit is generous.

### 6.3 "Warp-cooperative evaluation" approach

Instead of one thread per element, use one warp (32 threads) per 32 elements,
with shared memory for intermediate slots.

**Pros**: Better memory coalescing for slot reads/writes
**Cons**: More complex kernel, limited shared memory per block
**Verdict**: Over-engineered for the current use case. The per-thread approach
with L1 caching already achieves good throughput. Revisit only if profiling
shows slot access as a bottleneck.

---

## 7. Scalability Concerns

### 7.1 When does this approach break?

The "all inputs on GPU" approach works when:
```
NumRoots × DomainSize × 4B × 2 (coeffs + eval) < GPU_MEMORY × 0.6
```

With 90 GB GPU and 60% budget:
- 54 GB available
- DomainSize = 2^20, 200 roots: 200 × 1M × 8B = 1.6 GB ✓
- DomainSize = 2^22, 200 roots: 200 × 4M × 8B = 6.4 GB ✓
- DomainSize = 2^24, 500 roots: 500 × 16M × 8B = 64 GB ✗ (exceeds budget)

### 7.2 Scaling strategies when it doesn't fit

**Strategy A: Tile across roots**
Process roots in batches. Cache coefficients for batch_0, evaluate all cosets,
free, then process batch_1. This is the natural extension.

**Strategy B: Tile across domain**
Split the domain into tiles. Each tile does its own NTTs and symbolic evaluation.
This is more complex because FFTs have global data dependencies — you'd need
to use the Bluestein algorithm or process complete NTTs off-device.

**Strategy C: Multi-GPU**
Partition roots across GPUs. Each GPU handles a subset of roots and constraints.
Constraints that share roots need coordination, but this is manageable.

**Recommendation**: Start with Strategy A. It requires minimal code changes and
handles the 2^24 × 500 roots case cleanly. Multi-GPU is a later concern.

---

## 8. Comparison: GPU Quotient vs. Alternatives

### 8.1 Could we restructure the quotient computation itself?

The quotient computation is fundamentally:
```
Q(x) = Σ_j α^j · C_j(x) / (x^n - 1)
```
evaluated on `maxRatio` cosets. There's not much room to change the mathematics.

One possible restructuring: instead of evaluating per-constraint boards
independently, merge all constraints for a ratio group into a single large
board. This would:
- Increase DAG node sharing (common subexpressions across constraints)
- Reduce kernel launch overhead (one launch instead of K launches)
- Potentially increase register pressure (larger program)

**Assessment**: The merging already happens via `PolyEval(mergingCoin, constraints)`.
Each `AggregateExpression[k]` is already the Schwartz-Zippel combination of all
constraints for that ratio. So there's typically one board per ratio group, not
one per constraint. The "K launches per coset" concern is actually "one launch
per ratio group per coset", which is a small number.

### 8.2 Could we skip the quotient entirely?

No. The quotient polynomial is a fundamental part of the PLONK protocol. Without
it, the verifier cannot check constraint satisfaction.

### 8.3 Could we compute the quotient in coefficient form directly?

Instead of evaluating on cosets and then doing IFFT, compute the quotient
polynomial coefficients directly via polynomial division in coefficient space.

This is theoretically possible but:
- Polynomial multiplication in coefficient space is O(n²) without FFT
- The symbolic expressions involve products, requiring convolution
- The coset-evaluation approach already gives us O(n log n) per column

Not worth pursuing.

---

## 9. Implementation Roadmap

### Phase 1: Infrastructure (no user-visible change)

1. **Add `BoardToNodeOps` adapter** in `gpu/symbolic/`
   - Convert `ExpressionBoard` → `[]NodeOp`
   - Map variable metadata → `SymInputDesc` descriptors
   - Handle the input ordering (variables listed by board position)

2. **Add device-output variant of `kb_sym_eval`**
   - `kb_sym_eval_device(...)` → writes to device buffer
   - Avoids D2H for quotient shares that will be re-committed

3. **Add batch NTT wrapper in `gpu/`**
   - Wrap existing NTT kernels for multi-polynomial batch processing
   - Support both base and E4 fields
   - Support coset scaling (fused with NTT butterfly if possible)

### Phase 2: GPU quotient prototype

4. **Implement `QuotientCtx.RunGPU(run *wizard.ProverRuntime)`**

   ```go
   func (ctx *QuotientCtx) RunGPU(run *wizard.ProverRuntime) {
       // Phase 0: Compile boards to GPU programs (once, cached)
       programs := ctx.compileGPUPrograms(dev)

       // Phase 1: Upload witnesses, compute coefficient form
       coeffs := ctx.uploadAndIFFT(dev, run)

       // Phase 2: For each coset
       for i := 0; i < maxRatio; i++ {
           evals := ctx.cosetFFT(dev, coeffs, i)
           xCoset := ctx.computeXCoset(dev, i)

           for _, j := range constraintsForCoset(i) {
               inputs := ctx.buildSymInputs(evals, xCoset, run, j)
               result := EvalSymGPU(dev, programs[j], inputs, ctx.DomainSize)
               // multiply by annulator inverse
               // assign or keep on device
           }
       }
   }
   ```

5. **Integrate coefficient caching**
   - After first coset: coefficients are already on GPU
   - Subsequent cosets: D2D copy + coset NTT only

### Phase 3: Optimization

6. **Fuse NTT + coset scaling** using existing fused tile kernel
7. **Profile and tune** SYM_MAX_SLOTS, thread block size, memory layout
8. **Add device-resident quotient shares** path for Vortex integration
9. **Benchmark vs CPU** — expect 20-50x speedup on FFT phase, 5-10x on symbolic phase

### Phase 4: Production hardening

10. **Cross-backend equivalence tests** (CPU vs GPU quotient shares must match exactly)
11. **Memory pressure handling** — graceful fallback to tiled processing
12. **Multi-GPU support** — partition roots across devices

---

## 10. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Register pressure from large boards | Low | Medium | SYM_MAX_SLOTS=2048 is generous; liveness minimizes live slots |
| Memory overflow for very large domains | Low (current) | High | Tiled processing (Strategy A from §7.2) |
| Numerical mismatch CPU vs GPU | Low | Critical | Bit-exact field arithmetic (Montgomery form); extensive cross-tests |
| Kernel launch overhead | Medium | Low | One launch per ratio group per coset; can batch if needed |
| PeriodicSample evaluation on GPU | Medium | Low | Small computation; precompute on CPU and upload, or add small kernel |
| D2H bottleneck for quotient shares | Medium | Medium | Device-resident path (Phase 3, item 8) |

---

## 11. Conclusion

Your intuition is correct: the GPU quotient pipeline should **homogenize
inputs to E4 and run a branchless bytecode VM**. The infrastructure for this
already exists in `gpu/symbolic/`. The main work is:

1. Wrapping the FFT phase as batch NTT on GPU (separate from symbolic VM)
2. Adding the `BoardToNodeOps` adapter
3. Caching coefficient form across cosets (50% FFT savings)
4. Wiring it together in `QuotientCtx.RunGPU()`

The approach is sound, scalable to current GPU memory (90 GB), and has a clear
path to multi-GPU when needed. The symbolic package doesn't need a revamp — it
needs a thin GPU compilation adapter.

**Expected end-to-end speedup for quotient computation: 10-30x** depending on
domain size and constraint count, with FFT batch NTT providing the bulk of
the improvement and the symbolic VM providing a secondary 5-10x on the
evaluation phase.

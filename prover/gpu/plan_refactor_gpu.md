# GPU PlonK Refactor Plan

## Context

Two GPU-accelerated PlonK provers coexist in `prover/gpu/`:

| Package | Lines | Files | Curves | Performance |
|---|---|---|---|---|
| `gpu/plonk` | 9,526 | 20 | BLS12-377 only | baseline (fast) |
| `gpu/plonk2` | 15,941 | 56 | BN254, BLS12-377, BW6-761 | ~4× slower than plonk |

`gpu/plonk2` was designed as a validation-first multi-curve foundation. It is 4× slower than `gpu/plonk` on BLS12-377 and the codebase is harder to read, reason about, and extend. This plan eliminates the gap by rewriting `gpu/plonk2` using the same orchestration philosophy as `gpu/plonk`, applied to all three curves, via code generation.

**Working mode.** This plan is executed on the `prover/gpu2` branch. No PR is opened against `main` while phases 1–7 are in flight; phase boundaries are marked by checkpoint commits. Phase 8 (deprecation of `gpu/plonk`) is the only step that interacts with `main`, and only after explicit authorization.

---

## Root Cause Analysis: Why plonk2 Is 4× Slower

Understanding the causes is mandatory before writing any code. Do not skip this section.

### 1. No CUDA stream pipelining (estimated impact: −20–30%)

Every kernel call in `gpu/plonk2` is synchronous. `gpu/plonk` accepts an optional `gpu.StreamID` on every `FrVector`, `FFTDomain`, and MSM operation, allowing H2D transfers, NTT computations, and D2H copies to overlap across CUDA streams. This is not a minor optimization — for a prover that alternates between large NTT passes and MSM phases, pipelining is table-stakes.

### 2. MSM coordinate system and inner-loop formula (estimated impact: −10–20%)

`gpu/plonk` uses Twisted Edwards coordinates for the BLS12-377 SRS. The earlier draft of this section claimed TE was 96 bytes/point vs 144 bytes/point for affine — that comparison was wrong. For BLS12-377, Fp is 377 bits = 6 × 64-bit limbs = 48 bytes, and **both** affine SW and compact TE are an `(X, Y)` pair, so both are 96 bytes per point. The 144-byte figure that appeared in this section was the Jacobian work-buffer layout used inside the kernel, not the SRS storage layout. Storage cost is not the differentiator.

The real, measurable advantages of the legacy TE path are:

- ~10–20% fewer field multiplications inside the bucket accumulation kernel (TE mixed-add ≈ 9M vs SW mixed-add ≈ 11M)
- Cleaner handling of the curve identity (no infinity flag branch) inside the inner loop
- A pre-converted on-disk SRS asset format that avoids an SW→TE conversion at startup

These are real but localized wins. They are not the dominant cause of the 4× gap. See **Decision 4** for why we accept giving them up in `plonk2`.

### 3. Raw `[]uint64` with runtime curve dispatch (estimated impact: −10–15%)

`gpu/plonk2` represents all field elements as `[]uint64` slices and dispatches operations through the `curveProofOps` interface. This means:
- Every field operation has a runtime bounds check on the slice
- No escape analysis: temporaries heap-allocate instead of stack-allocating
- Interface dispatch adds indirection on every hot-path call
- The compiler cannot see the limb count statically, blocking vectorization

`gpu/plonk` uses typed `fr.Element` (`[4]uint64` array) throughout. The compiler knows sizes at compile time.

### 4. No work buffer lifecycle management (estimated impact: −10–15%)

`gpu/plonk` has `PinWorkBuffers()/ReleaseWorkBuffers()` on the MSM context: it pre-allocates all sort buffers once and reuses them across multiple MSM calls within a proof. It also has `OffloadPoints()/ReloadPoints()` to reclaim 6+ GiB of VRAM during the quotient phase (when MSM is not running). `gpu/plonk2` allocates and frees for each call.

### 5. Scatter of orchestration logic (qualitative impact: maintainability)

`gpu/plonk2` splits the prover across `prove.go` (461 lines), `generic_prove_cuda.go` (1,450 lines), three `generic_finalize_*.go` files (~1,550 lines combined), `generic_prepare_cuda.go` (519 lines), `generic_quotient_cuda.go` (601 lines), `memory_plan.go`, `msm_plan.go`, `ntt_plan.go`, and more. The orchestration logic is not in one place, making phase ordering, buffer ownership, and stream dependency hard to audit. `gpu/plonk/prove.go` is 2,727 lines but it is *one file* with a clear layered architecture (`GPUProvingKey → gpuInstance → gpuProver`).

---

## Design Decisions

### Decision 1: Kill the raw `[]uint64` API on the Go side

**No `interface{}`, no `any`, no raw `[]uint64` field buffers in any public or internal API.** Every per-curve package uses typed arrays:

- BN254, BLS12-377: `[4]uint64` for `fr.Element`
- BW6-761: `[6]uint64` for `fr.Element`

This matches gnark-crypto's representation and lets the compiler enforce correctness, eliminate bounds checks, and stack-allocate temporaries.

### Decision 2: Code generation over hand-written generics

Follow the gnark-crypto `internal/generator` + bavard pattern exactly. Three typed, specialized packages are generated from templates. The templates are derived from `gpu/plonk` structure. A small amount of duplicated generated code is acceptable and preferable to a shared abstraction that forces runtime dispatch.

### Decision 3: Replicate gpu/plonk orchestration as the template source of truth

`gpu/plonk/prove.go` is the reference. The `GPUProvingKey → gpuInstance → gpuProver` layering, the stream dependency graph, the buffer ownership model, the `PinWorkBuffers` / `OffloadPoints` lifecycle — all of these are proven correct and fast. Templates must preserve this structure.

### Decision 4: Affine short-Weierstrass for all three curves; defer Twisted Edwards

Every per-curve package generates an affine SW-based MSM. The legacy Twisted Edwards path stays inside `gpu/plonk` only and is used as a single-device performance reference.

Rationale:

- The on-disk TE SRS asset format diverges from the rest of the production stack (KZG serialization, ceremony output, recursive verifier inputs). Maintaining two SRS asset pipelines is real operational debt — separate conversion tooling, separate validation, separate disk footprint.
- If `plonk2/bls12377` (affine) lands within ~15–20% of legacy `gpu/plonk` (TE) on a single device, the simpler asset and code path wins. We can revisit TE later if measurement clearly justifies the complexity.
- Removing the `{{if .UseTE}}` conditional collapses three meaningfully different generated packages into one shape, and removes the only branch where `bls12377/` would diverge from `bn254/` and `bw6761/`.
- BN254 and BW6-761 had to be affine anyway. Choosing affine across the board collapses three coordinate systems into one.

If the affine BLS12-377 prover falls outside the 20% acceptance band after Phase 5, **do not** silently re-introduce TE. Open a separate proposal that quantifies the gap and the asset/operational cost, and decide explicitly. Multi-GPU acceleration (Decision 6) is the more attractive lever to close any remaining gap.

### Decision 5: C++ templates for CUDA kernels, not copy-paste

Kernel algorithms (NTT butterfly, Pippenger accumulation, Z factor computation, gate accumulation) are parameterized by `template<int NLimbs>` with explicit instantiations for NLimbs=4 and NLimbs=6. The MSM kernel is affine short-Weierstrass for all three curves (Decision 4), so there is no per-curve coordinate-system header. Per-curve `.cu` files exist only to instantiate the templates with the correct field modulus and limb count.

### Decision 6: One Prover at a time; multiple GPUs collaborate inside it

Production hardware constraint: host RAM (witness buffers, SRS staging, in-flight commitments, pinned transfer regions) only fits **one** in-flight proof at a time, regardless of how many GPUs are present. Therefore additional GPUs are not used to run more provers in parallel — they are used to make a single proof faster.

The `Prover` struct owns a `DeviceGroup` (slice of `gpu.Device`, length 1, 2, or 4). Inside `Prove`, the orchestrator splits and pipelines work across the group:

- **MSM split** — the bucket method partitions naturally on the scalar/point dimension. For an MSM over `N` points, dispatch the first `N/k` points to GPU 0, the next `N/k` to GPU 1, etc. Each device runs its own Pippenger pass and produces a partial G1 result; the host adds the partials.
- **Independent-polynomial concurrency** — `iFFT(L)`, `iFFT(R)`, `iFFT(O)` are independent, as are the three commitment MSMs and the three quotient-piece MSMs. Schedule each onto its own device.
- **Phase pipelining** — while GPU 0 finishes the quotient gate-accumulation kernel, GPU 1 can start an MSM as soon as its inputs are ready. The dependency graph in `gpu/plonk/prove.go` is the source of truth for what can overlap.
- **No P2P required for v1** — partial results are tiny (a few G1 points). Round-trip them through pinned host memory. Direct GPU-to-GPU transfer is an optimization to revisit only if profiling identifies host bandwidth as the bottleneck.

See the **Multi-GPU Strategy** section for the concrete decision rules an agent must apply when modifying the orchestrator.

### Decision 7: CPU fallback stays in the root dispatcher, not in per-curve packages

The per-curve packages (`bn254/`, `bls12377/`, `bw6761/`) are pure GPU implementations — they return an error when CUDA is unavailable. The root `gpu/plonk2/prover.go` wraps them and handles the CPU fallback. This keeps the generated code clean.

---

## Multi-GPU Strategy

A `Prover` is parameterized by a `DeviceGroup` (`[]gpu.Device`). When the group has size 1, the orchestrator behaves identically to a single-GPU prover. When size > 1, the rules below apply. Multi-GPU is **intra-prover only** — there is never more than one `Prover` running on a host at once.

### Mental model

The PlonK prover is a DAG of phases, not a linear pipeline. With one GPU, dependencies are serialized through CUDA streams on a single device. With multiple GPUs, two more axes open up:

1. **Spatial split** of a single large operation across devices (only safe for operations whose math is partition-friendly: MSM, element-wise FrVector ops, batched independent FFTs of separate polynomials).
2. **Temporal pipelining** of independent phases onto different devices (FFT of one polynomial on G0 while MSM of another runs on G1).

Both axes should be exploited, but never at the cost of correctness or readable orchestration. When in doubt, pick the simpler distribution — usually that is "assign each independent op to a different GPU" rather than "split one op across GPUs."

### When to split a single operation across GPUs

Split if **all three** are true:

- The operation cost dominates a phase boundary (i.e., end-to-end time is sensitive to it). MSM over the full SRS qualifies. Small element-wise ops do not.
- The math admits a clean partition with O(small) host-side combine. MSM partitions on the points axis (each device holds a disjoint slice of the SRS). FrVector ops partition on indices.
- The per-device chunk is still large enough to amortize launch and sync overhead. Below ~2²² scalars, an MSM split typically loses to a single-GPU MSM because of the host-side combine and synchronization cost. The split threshold is a tunable, not a hard-coded constant.

**Do NOT split a single NTT across devices in v1.** A single 1D NTT requires data exchange between devices at every stage where the butterfly distance crosses the partition. The 2D / six-step decomposition that makes this tractable adds substantial complexity and is out of scope.

### When to assign independent operations to different GPUs

This is the cheaper, safer form of multi-GPU acceleration. Assign whenever:

- Two operations have no data dependency in the current phase.
- Their outputs are needed in the same later phase, and dispatching them sequentially would extend the critical path.

Concrete examples in the PlonK prover:

| Phase | Independent work | 2-GPU sketch | 4-GPU sketch |
|---|---|---|---|
| LRO upload + iFFT | `iFFT(L)`, `iFFT(R)`, `iFFT(O)` | L→G0, R→G1, O→G0 (after L) | L→G0, R→G1, O→G2, blinding→G3 |
| LRO commit | `MSM(L)`, `MSM(R)`, `MSM(O)` | each MSM split 2-way across G0,G1 | one MSM per GPU on G0..G2 |
| Quotient | gate-accum, perm-accum, H decomposition | gate→G0, perm→G1 | gate→G0, perm→G1, decomp→G2 |
| H commit | `MSM(H0)`, `MSM(H1)`, `MSM(H2)` | each split 2-way | one per GPU on G0..G2 |
| Linearization | KZG opening, evaluations | KZG→G0, evals→G1 | KZG→G0, evals→G1, batch open→G2 |

These mappings are illustrative — the orchestrator chooses them at runtime based on the dependency graph. Hard-coded `if nDevices == 2` switches in the templates are forbidden; they break the moment a curve has different phase costs or a host has 3 GPUs.

### Data placement rules

1. **SRS slice per device.** During `gpuInstance` construction, the SRS is partitioned across the device group: each device pins and uploads only the slice of points it owns for the MSM split. Total VRAM use across the group is the same as the single-GPU SRS — distributing the SRS is the point.
2. **Per-proof scratch is per-device.** Each device has its own iFFT scratch, twiddle table, and MSM work buffers. The `PinWorkBuffers / OffloadPoints` lifecycle from `gpu/plonk` runs independently per device.
3. **Witness polynomials live where they are first consumed.** If `iFFT(L)` runs on G0, the canonical L stays on G0. When the later `MSM(L)` is split across G0 and G1, only the scalar slice owned by G1 is shipped via pinned host memory. Avoid GPU-to-GPU copies in v1.
4. **Commitments and partial results round-trip through pinned host memory.** A G1 partial sum is a few hundred bytes. Don't over-engineer the transfer.

### Synchronization rules

- Each device has its own set of CUDA streams. There is no shared global stream.
- Cross-device dependencies are expressed via host-side `errgroup`/channel handoff. The orchestrator never calls `cudaStreamWaitEvent` across devices in v1.
- Every cross-device sync is a critical-path event. Count them. If a phase has more than two cross-device syncs, redesign the phase before adding more.
- `defer dev.Synchronize()` at the end of `Prove` is **not** a sync strategy. Synchronize the device(s) you actually wrote to, in the order their results are consumed.

### Concurrency primitives in the orchestrator

The orchestrator is plain Go: `errgroup.Group` for fan-out, channels for handoff of GPU-resident handles, explicit `sync.WaitGroup` for barriers. No goroutine pools, no work-stealing schedulers, no actor frameworks. Phase ordering is static and known at compile time; dynamic scheduling buys nothing here and adds debugging cost.

```go
// Sketch — see prove.go.tmpl for the full pattern.
g, ctx := errgroup.WithContext(ctx)
var lH, rH, oH gpuFrHandle  // GPU-resident handles, each tagged with the device that owns them
g.Go(func() error { return iFFTOn(devs[0],            &lH, witness.L) })
g.Go(func() error { return iFFTOn(devs[1%len(devs)],  &rH, witness.R) })
g.Go(func() error { return iFFTOn(devs[2%len(devs)],  &oH, witness.O) })
if err := g.Wait(); err != nil { return err }

// Now LRO are blinded-canonical on their owning devices.
// Commit phase splits each MSM across all devs.
cL, cR, cO, err := commitLRO(devs, lH, rH, oH)
```

### Decision rules an agent must follow when changing the orchestrator

1. **Read `gpu/plonk/prove.go` first.** That single file is the canonical phase ordering and stream dependency graph.
2. **Identify the phase you're modifying.** Name it (e.g., "computeQuotient"). Phase boundaries are hard-won — do not edit them casually.
3. **List dependencies in and out of the phase.** Write them down in the checkpoint commit message. If a new dependency edge crosses a device boundary, justify the cost.
4. **Default to assigning independent ops to different devices.** Only resort to splitting a single op across devices when measurement shows the op is on the critical path and partition-friendly.
5. **Measure cross-device sync overhead before adding more concurrency.** A `nsys` or `nvprof` trace showing the new schedule is required when a change touches cross-device coordination.
6. **Never assume a fixed number of GPUs.** All scheduling code must work for `len(devs) ∈ {1, 2, 4}`. The single-GPU case must hit the same fast path the legacy `gpu/plonk` prover does — no extra synchronization, no extra allocations.
7. **Reject schedules that cannot be expressed in the existing template structure.** If you find yourself writing a new orchestration framework, stop. The point of the rewrite is to keep `prove.go` a readable, layered file.

### What multi-GPU is NOT for

- It is not for running multiple independent proofs in parallel on the same host. Host memory does not allow this; the calling layer above `Prover` is expected to serialize proofs.
- It is not a substitute for kernel-level optimization. If a kernel is slow, fix the kernel; do not paper over it by sharding the work.
- It is not a correctness tool. Two-device runs must produce byte-identical proofs to one-device runs, full stop. This is a non-negotiable test gate.

---

## Target File Structure

```
gpu/
├── device.go               ← unchanged (shared Device, StreamID, Error types)
├── gpu.go                  ← unchanged
├── plonk/                  ← FROZEN (kept as reference until bls12377/ is validated)
│   └── ...                 ← do not modify
├── plonk2/
│   ├── doc.go              ← package documentation + //go:generate
│   ├── prover.go           ← multi-curve dispatcher, CPU fallback, options
│   ├── options.go          ← prover options (slimmed: remove legacyBLS, tracePath detail)
│   ├── stub.go             ← non-CUDA build stub (panics with clear message)
│   ├── bn254/              ← GENERATED — do not edit by hand
│   │   ├── fr.go           (typed FrVector, [4]uint64 arrays, SoA on GPU)
│   │   ├── fr_test.go      (gopter property-based tests)
│   │   ├── fft.go          (stream-aware FFTDomain, typed twiddles)
│   │   ├── fft_test.go
│   │   ├── msm.go          (affine SRS, pinned memory, chunking, work buffer lifecycle)
│   │   ├── msm_test.go
│   │   ├── prove.go        (orchestration mirroring gpu/plonk, typed)
│   │   └── plonk_test.go
│   ├── bls12377/           ← GENERATED — do not edit by hand
│   │   ├── fr.go
│   │   ├── fr_test.go
│   │   ├── fft.go
│   │   ├── fft_test.go
│   │   ├── msm.go          (affine SW SRS, pinned memory, multi-GPU split)
│   │   ├── msm_test.go
│   │   ├── prove.go        (DeviceGroup-aware orchestration)
│   │   └── plonk_test.go
│   ├── bw6761/             ← GENERATED — do not edit by hand
│   │   ├── fr.go           ([6]uint64 typed FrVector)
│   │   ├── fr_test.go
│   │   ├── fft.go
│   │   ├── fft_test.go
│   │   ├── msm.go
│   │   ├── msm_test.go
│   │   ├── prove.go
│   │   └── plonk_test.go
│   └── internal/
│       └── bench_icicle/   ← moved from plonk2 root (icicle_*.go, bench_* files)
│           └── ...
└── internal/
    └── generator/          ← NEW code generator
        ├── main.go
        ├── config/
        │   ├── curve.go
        │   ├── bn254.go
        │   ├── bls12377.go
        │   └── bw6761.go
        ├── plonk/
        │   ├── generate.go
        │   └── template/
        │       ├── fr.go.tmpl
        │       ├── fr_test.go.tmpl
        │       ├── fft.go.tmpl
        │       ├── fft_test.go.tmpl
        │       ├── msm.go.tmpl
        │       ├── msm_test.go.tmpl
        │       ├── prove.go.tmpl
        │       ├── plonk_test.go.tmpl
        │       └── templates.go    (//go:embed *)
        └── common/
            └── generator.go        (bavard wrapper, identical to gnark-crypto)
```

**Files deleted from `gpu/plonk2/` root:**

| File | Reason |
|---|---|
| `generic_prove_cuda.go` | Replaced by generated `bls12377/prove.go` etc. |
| `generic_finalize_bls12377_cuda.go` | Replaced by generated prove.go template |
| `generic_finalize_bn254_cuda.go` | Replaced by generated prove.go template |
| `generic_finalize_bw6761_cuda.go` | Replaced by generated prove.go template |
| `generic_prepare_cuda.go` | Folded into generated prove.go |
| `generic_prepare_stub.go` | Folded into stub.go |
| `generic_prove_stub.go` | Folded into stub.go |
| `generic_quotient_cuda.go` | Folded into generated prove.go |
| `generic_scratch_cuda.go` | Eliminated; scratch is typed per-curve |
| `curve.go` | Eliminated; typed packages don't need curve ID enum |
| `curve_proof_ops_cuda.go` | Eliminated; no more interface dispatch |
| `fr.go` | Replaced by generated per-curve fr.go |
| `fft.go` | Replaced by generated per-curve fft.go |
| `msm.go` | Replaced by generated per-curve msm.go |
| `g1.go` | Replaced by generated per-curve msm.go |
| `commit.go` | Folded into generated msm.go |
| `quotient.go` | Folded into generated prove.go |
| `memory_plan.go` | Absorbed into generated prove.go preamble |
| `msm_plan.go` | Absorbed into generated msm.go |
| `msm_window.go` | Absorbed into generated msm.go |
| `ntt_plan.go` | Absorbed into generated fft.go |
| `device.go` | Use `gpu.Device` from parent package directly |
| `prove_gpu_cuda.go` | Thin; fold into prover.go |
| `prove_gpu_stub.go` | Fold into stub.go |
| `icicle_roots_cuda.go` | Move to `internal/bench_icicle/` |
| `icicle_compare_cuda_test.go` | Move to `internal/bench_icicle/` |
| All `bench_*_test.go` in root | Move to `internal/bench_icicle/` |
| `srs_assets_test.go` | Move to per-curve `*_test.go` |

---

## Code Generator Design

### Curve Config

```go
// gpu/internal/generator/config/curve.go
type Curve struct {
    Name         string   // "bn254", "bls12377", "bw6761"
    Package      string   // "bn254", "bls12377", "bw6761"
    FrLimbs      int      // 4 or 6
    FpLimbs      int      // 4, 6, or 12
    ScalarBits   int      // 254, 253, 377
    FrModulus    string   // decimal string
    FpModulus    string   // decimal string
    // Gnark-crypto import paths
    GnarkCryptoFr  string // "github.com/consensys/gnark-crypto/ecc/bn254/fr"
    GnarkCryptoFFT string // "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
    GnarkCryptoKZG string // "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
    GnarkCurve     string // "github.com/consensys/gnark-crypto/ecc/bn254"
    GnarkCS        string // "github.com/consensys/gnark/constraint/bn254"
    GnarkPlonk     string // "github.com/consensys/gnark/backend/plonk/bn254"
}
```

### Template Variables (used inside .tmpl files)

```
{{.Name}}           → "bn254"
{{.Package}}        → "bn254"
{{.FrLimbs}}        → 4
{{.FpLimbs}}        → 4
{{.ScalarBits}}     → 254
{{.GnarkCryptoFr}}  → import path
```

### Generator Entry Point

```go
// gpu/internal/generator/main.go
func main() {
    curves := []config.Curve{config.BN254, config.BLS12377, config.BW6761}
    for _, curve := range curves {
        outputDir := filepath.Join("../../plonk2", curve.Package)
        if err := plonk.Generate(curve, outputDir, gen); err != nil {
            log.Fatal(err)
        }
    }
    // Run gofmt + goimports on output dirs
}
```

### Template Style (follow gnark-crypto conventions exactly)

Each `.tmpl` file begins with:
```
// Code generated by gpu/internal/generator DO NOT EDIT
```

Tests use gopter:
```go
// fr_test.go.tmpl
func TestFrVectorAdd(t *testing.T) {
    parameters := gopter.DefaultTestParameters()
    properties := gopter.NewProperties(parameters)
    properties.Property("GPU add matches CPU add", prop.ForAll(
        func(a, b {{.Name}}fr.Element) bool { ... },
        GenFr(), GenFr(),
    ))
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func GenFr() gopter.Gen {
    return func(_ *gopter.GenParameters) *gopter.GenResult {
        var e {{.Name}}fr.Element
        e.MustSetRandom()
        return gopter.NewGenResult(e, gopter.NoShrinker)
    }
}
```

---

## CUDA Refactoring Strategy

### Kernel Organization

```
cuda/
├── include/
│   ├── gnark_gpu.h         ← C ABI (add curve-parameterized NLimbs variants)
│   ├── field_ops.cuh       ← template<int NLimbs> field arithmetic
│   ├── ntt.cuh             ← template<int NLimbs> NTT butterfly
│   ├── pippenger.cuh       ← template<int NLimbs> bucket method (affine SW)
│   └── plonk_kernels.cuh   ← template<int NLimbs> gate accum, Z factors, etc.
└── src/
    ├── field_bn254.cu      ← explicit instantiation: field_ops<4> for BN254 modulus
    ├── field_bls12377.cu   ← explicit instantiation: field_ops<4> for BLS12-377 modulus
    ├── field_bw6761.cu     ← explicit instantiation: field_ops<6> for BW6-761 modulus
    ├── ntt_bn254.cu
    ├── ntt_bls12377.cu
    ├── ntt_bw6761.cu
    ├── msm_bn254.cu
    ├── msm_bls12377.cu
    ├── msm_bw6761.cu
    └── plonk_kernels.cu    ← instantiated for all three curves
```

### CUDA Documentation Standard

Every kernel must have a comment block that:
1. Names the algorithm (e.g., "Cooley-Tukey radix-2 DIT NTT, iterative, in-place")
2. States the input/output preconditions (Montgomery form? SoA layout? Which stream?)
3. Notes any non-obvious invariant (e.g., "twiddle factors must be in bit-reversed order")
4. States the thread/block layout and why it was chosen

Example:
```cpp
// ntt_butterfly_kernel: single stage of a Cooley-Tukey DIT NTT.
//
// Each thread processes one butterfly pair at distance `half_m`.
// Input must be in natural order; output is in bit-reversed order after
// log2(n) passes. Twiddle factors are in Montgomery form, pre-computed.
//
// Thread layout: 1D grid, BLOCK_SIZE=256 threads per block, ceil(n/2) threads total.
// Shared memory not used; access pattern is stride-2 which is coalesced at this granularity.
template<int NLimbs>
__global__ void ntt_butterfly_kernel(
    uint64_t* __restrict__ data,
    const uint64_t* __restrict__ twiddles,
    uint32_t half_m,
    uint32_t n
);
```

### Stream Management Pattern (replicate from gpu/plonk)

Each `prove.go` template instantiates named streams at the top of the proof:
```go
s0, _ := dev.NewStream()  // H2D witness upload
s1, _ := dev.NewStream()  // NTT passes
s2, _ := dev.NewStream()  // MSM
s3, _ := dev.NewStream()  // D2H for Fiat-Shamir challenges
defer dev.DestroyStream(s0, s1, s2, s3)

// Phase ordering with explicit sync points documented inline
```

Every kernel wrapper in the per-curve package accepts `stream ...gpu.StreamID` as a variadic last argument. When no stream is provided, the default stream is used (safe for tests, suboptimal for production).

### Multi-Device Support

`Prover` owns a `DeviceGroup` (`[]gpu.Device`, length 1, 2, or 4), fixed at construction time. No global state, no peer-to-peer setup, no second `Prover` instance running on the same host. The orchestrator inside `Prove` cooperates across the group as described in the **Multi-GPU Strategy** section above.

```go
// Single GPU (most common deployment)
p, _ := bn254.NewProver(gpu.DeviceGroup{gpu.Device(0)}, ccs, pk)

// Two GPUs cooperating on one proof
p, _ := bn254.NewProver(gpu.DeviceGroup{gpu.Device(0), gpu.Device(1)}, ccs, pk)
proof, err := p.Prove(witness)
```

The single-device path must remain a fast path with no extra synchronization or allocation versus the legacy `gpu/plonk` shape. `len(devs) == 1` should branch out of any errgroup/channel scaffolding and call the local kernel directly.

---

## Migration Phases

Each phase must pass its acceptance criteria before the next begins. Do not batch phases.

---

### Phase 0 — Baseline Measurement (prerequisite, no code changes)

**Goal:** Establish performance baselines that later phases must meet or beat.

**Actions:**
1. Build `gpu/plonk` with `go build -tags cuda ./gpu/plonk/...`
2. Run `gpu/plonk` benchmarks and record results in `gpu/plonk2/internal/bench_icicle/baseline_plonk.txt`
3. Run `gpu/plonk2` benchmarks and record in `baseline_plonk2.txt`
4. Record: end-to-end prove time, FFT time, MSM time, peak VRAM usage
5. Commit baseline files — they must not be overwritten until Phase 6

**Benchmark command (gpu/plonk):**
```bash
go test -tags cuda -v -bench=BenchmarkGPUProve -benchtime=3x \
    -run=^$ ./gpu/plonk/ 2>&1 | tee gpu/plonk2/internal/bench_icicle/baseline_plonk.txt
go test -tags cuda -v -bench=BenchmarkFFT -bench=BenchmarkMSM -benchtime=5x \
    -run=^$ ./gpu/plonk/ 2>&1 | tee -a gpu/plonk2/internal/bench_icicle/baseline_plonk.txt
```

**Acceptance criteria:** Baselines recorded, committed, reviewed by a human.

---

### Phase 1 — Generator Infrastructure

**Goal:** Build the code generator skeleton, configuration, and bavard wrapper. No generated code yet.

**Actions:**
1. Create `gpu/internal/generator/common/generator.go` — exact copy of gnark-crypto's `internal/generator/common/generator.go` with package path updated
2. Create `gpu/internal/generator/config/curve.go` with the `Curve` struct
3. Create `gpu/internal/generator/config/bn254.go`, `bls12377.go`, `bw6761.go`
4. Create `gpu/internal/generator/plonk/generate.go` — calls `bavard.BatchGenerator` for each template entry
5. Create `gpu/internal/generator/main.go` — loops over curves, calls plonk.Generate, runs gofmt+goimports
6. Create `gpu/internal/generator/plonk/template/templates.go` with `//go:embed *`
7. Write a single smoke-test template `fr_stub.go.tmpl` that produces a valid `doc.go` in each output directory

**Acceptance criteria:**
```bash
cd gpu/internal/generator && go run . 2>&1   # must exit 0
go build ./gpu/plonk2/bn254/...              # must compile (doc.go only)
go build ./gpu/plonk2/bls12377/...
go build ./gpu/plonk2/bw6761/...
```

---

### Phase 2 — FrVector Templates (fr.go, fr_test.go)

**Goal:** Generate typed, stream-aware FrVector implementations for all three curves. Validate correctness against CPU reference.

**Template inputs from gpu/plonk/fr.go:**
- `FrVector` struct (SoA layout, limb arrays as typed fields, not slices)
- `NewFrVector`, `Free`, finalizer
- `CopyFromHost`, `CopyToHost`, `CopyFromDevice` (all accept variadic `gpu.StreamID`)
- `Add`, `Sub`, `Mul`, `AddMul`, `AddScalarMul`, `ScalarMul`
- `ScaleByPowers`, `BatchInvert`, `SetZero`
- Parameterize: `{{.FrLimbs}}` for array sizing, `{{.GnarkCryptoFr}}` for imports

**Test template (`fr_test.go.tmpl`) must cover:**
- `Add(a, b) == b + a` (commutativity, gopter)
- `Mul(a, Inv(a)) == 1` (when a ≠ 0, gopter)
- `BatchInvert` result matches scalar loop of gnark-crypto `fr.Element.Inverse`
- `ScaleByPowers(omega)` matches `[omega^0, omega^1, ..., omega^{n-1}]` computed on CPU
- Round-trip: `CopyFromHost → CopyToHost` is identity
- All tests compile and pass with `-tags cuda` absent (stub returns ErrNoCUDA)

**Acceptance criteria:**
```bash
# Without CUDA (stub path):
go test ./gpu/plonk2/bn254/... ./gpu/plonk2/bls12377/... ./gpu/plonk2/bw6761/...

# With CUDA:
go test -tags cuda -v -count=3 ./gpu/plonk2/bn254/...
go test -tags cuda -v -count=3 ./gpu/plonk2/bls12377/...
go test -tags cuda -v -count=3 ./gpu/plonk2/bw6761/...
```
All must pass with zero failures.

**Performance gate:**
```bash
go test -tags cuda -bench=BenchmarkFrVectorAdd -bench=BenchmarkFrVectorBatchInvert \
    -benchtime=5x ./gpu/plonk2/bls12377/
```
Must match or beat `gpu/plonk` fr benchmarks from Phase 0 baseline within 5%.

---

### Phase 3 — FFT Templates (fft.go, fft_test.go)

**Goal:** Generate typed, stream-aware FFTDomain implementations.

**Template inputs from gpu/plonk/fft.go:**
- `GPUFFTDomain` struct (holds GPU-resident twiddle factors as typed array pointer, not `[]uint64`)
- `NewFFTDomain(dev gpu.Device, size int) (*GPUFFTDomain, error)`
- `FFT`, `FFTInverse`, `BitReverse`, `CosetFFT`, `CosetFFTInverse`
- `Butterfly4Inverse` (for decomposed iFFT(4n) in quotient computation)
- All operations accept variadic `gpu.StreamID`
- `Free` + finalizer

**Test template (`fft_test.go.tmpl`) must cover:**
- `FFT(FFTInverse(v)) == v` for random vectors (gopter, sizes 2^10 to 2^24)
- `CosetFFT(CosetFFTInverse(v)) == v`
- FFT output matches gnark-crypto CPU `fft.Domain.FFT` exactly (byte-for-byte after `CopyToHost`)
- `BitReverse` matches gnark-crypto `fft.BitReverse`
- Tests compile without CUDA (stub)

**Acceptance criteria:**
```bash
go test ./gpu/plonk2/bn254/... ./gpu/plonk2/bls12377/... ./gpu/plonk2/bw6761/...
go test -tags cuda -v -count=3 ./gpu/plonk2/bn254/...
go test -tags cuda -v -count=3 ./gpu/plonk2/bls12377/...
go test -tags cuda -v -count=3 ./gpu/plonk2/bw6761/...
```

**Performance gate:**
```bash
go test -tags cuda -bench=BenchmarkFFT -bench=BenchmarkFFTInverse \
    -bench=BenchmarkCosetFFT -benchtime=5x ./gpu/plonk2/bls12377/
```
Must be within 5% of `gpu/plonk` FFT benchmarks from Phase 0 baseline.

---

### Phase 4 — MSM Templates (msm.go, msm_test.go)

**Goal:** Generate typed MSM implementations with the full optimization suite from gpu/plonk.

**Template inputs from gpu/plonk/msm.go (affine SW path only — TE remains in `gpu/plonk`):**

All three curves share the same MSM template. Differences are limb counts and import paths only.

- `G1Affine` typed pinned host memory: `[2 * FpLimbs]uint64` per point
- `G1MSM` context with `PinWorkBuffers()`, `ReleaseWorkBuffers()`, `OffloadPoints()`, `ReloadPoints()` — per device when the prover holds a `DeviceGroup`
- `MultiExp(scalars ...[]fr.Element) ([]G1Jac, error)` — variadic multi-MSM, single device
- `MultiExpSplit(devs []gpu.Device, scalars ...[]fr.Element) ([]G1Jac, error)` — partitions points across `devs`, runs partial Pippenger per device, sums on the host. The `len(devs) == 1` path must fall through to `MultiExp` with no overhead.
- `msmChunkThreshold = 1<<27` — point-count threshold per device above which sort buffers are halved by chunking
- All operations accept variadic `gpu.StreamID`

Curve-specific only by `{{.FrLimbs}}`, `{{.FpLimbs}}`, `{{.GnarkCurve}}`. No `{{if .UseTE}}` branches in any template.

**Test template (`msm_test.go.tmpl`) must cover:**
- `MultiExp(scalars)` result matches gnark-crypto `G1Affine.MultiExp` (small N ≤ 1000, gopter)
- `MultiExp` is consistent across multiple calls with same inputs
- Chunked MSM (N > msmChunkThreshold) matches unchunked result (test with mock threshold)
- `OffloadPoints / ReloadPoints` round-trip: result unchanged
- `MultiExpSplit(devs, ...)` matches `MultiExp(...)` byte-for-byte for `len(devs) ∈ {1, 2, 4}` (skip the >1 cases when only one GPU is visible)
- Tests compile without CUDA

**Acceptance criteria:**
```bash
go test ./gpu/plonk2/bn254/... ./gpu/plonk2/bls12377/... ./gpu/plonk2/bw6761/...
go test -tags cuda -v -count=3 -timeout=20m ./gpu/plonk2/bn254/...
go test -tags cuda -v -count=3 -timeout=20m ./gpu/plonk2/bls12377/...
go test -tags cuda -v -count=3 -timeout=20m ./gpu/plonk2/bw6761/...
```

**Performance gate:**
```bash
go test -tags cuda -bench=BenchmarkMSM -benchtime=3x ./gpu/plonk2/bls12377/
```
Single-device MSM must be within **20%** of `gpu/plonk` MSM from Phase 0 baseline. The wider band (vs the 5% used for FFT and FrVector) reflects giving up the TE inner-loop advantage; affine SW genuinely is slower per point on a single device. If the gap exceeds 20%, profile the bucket-accumulation kernel before adjusting the budget.

If the host has 2+ GPUs, also run:
```bash
go test -tags cuda -bench=BenchmarkMSMSplit -benchtime=3x ./gpu/plonk2/bls12377/
```
A 2-GPU split MSM should be ≥ 1.6× the single-GPU MSM. Below 1.4× points to host combine overhead, PCIe contention, or an unbalanced point partition.

For BN254 and BW6-761 there is no legacy baseline. Record these as new baselines.

---

### Phase 5a — Prover Templates, Single Device (prove.go, plonk_test.go)

**Goal:** Generate the full PlonK prover per curve, single-device only, with stream-pipelined orchestration matching gpu/plonk's `GPUProvingKey → gpuInstance → gpuProver` architecture. Multi-GPU support is layered on in Phase 5b without disturbing this phase's fast path.

**Template inputs from gpu/plonk/prove.go:**

Layers (all three must appear in the template):
1. `GPUProvingKey` — thin wrapper (gnark VerifyingKey + lazy gpuInstance)
2. `gpuInstance` — GPU-persistent resources: MSM context, FFT domain, permutation table, selector polys. Created once, reused across proofs. Holds a `DeviceGroup` even in 5a (length 1).
3. `gpuProver` — per-proof mutable state, implements prove phases as methods

Phases in order (preserve this ordering exactly — it encodes stream dependencies):
1. `preSolve` — run gnark constraint solver to get LRO in Lagrange form
2. `uploadLRO` — H2D copy of L, R, O on stream s0; start NTT on s1 immediately
3. `computeBlindedCanonical` — iFFT LRO, apply blinding polynomials
4. `commitLRO` — MSM for L, R, O commitments on stream s2 while Z is being prepared
5. `computeZ` — PlonkZComputeFactors + iFFT + blinding
6. `commitZ` — MSM for Z commitment
7. `computeQuotient` — gate accum + perm boundary on GPU, decompose into H0/H1/H2
8. `commitH` — MSM for H0, H1, H2
9. `evalLROZH` — polynomial evaluations at zeta (Horner on GPU)
10. `computeLinearizedPoly` — curve-specific linearized poly (template-specialized per curve)
11. `openingProof` — KZG batch opening (curve-specific)

Template variables:
- `{{.GnarkPlonk}}`, `{{.GnarkCS}}`, etc. for import paths
- `{{.FrLimbs}}`, `{{.FpLimbs}}` for typed array sizing inside prover state

There are **no** `{{if .UseTE}}` blocks (Decision 4). The `OffloadPoints / ReloadPoints` lifecycle is preserved unconditionally because it pays for itself on affine too: it frees ~6 GiB during the quotient phase regardless of coordinate system.

**Test template (`plonk_test.go.tmpl`) must cover:**
- Prove → Verify round-trip for a small (≤ 100 constraint) circuit (no CUDA tag required for compilation)
- Prove → Verify for a medium circuit (≥ 2^18 constraints, CUDA required)
- Proof is identical to a gnark CPU proof for the same witness and randomness seed (byte-for-byte)
- Test build tag: `//go:build cuda` only for tests requiring a GPU; others compile unconditionally

**Acceptance criteria:**
```bash
# Compilation without CUDA:
go test ./gpu/plonk2/bn254/... ./gpu/plonk2/bls12377/... ./gpu/plonk2/bw6761/...

# Full single-device test suite with CUDA:
go test -tags cuda -v -timeout=30m ./gpu/plonk2/bn254/...
go test -tags cuda -v -timeout=30m ./gpu/plonk2/bls12377/...
go test -tags cuda -v -timeout=30m ./gpu/plonk2/bw6761/...
```
Zero failures. GPU proof byte-identical to CPU proof.

**Performance gate (primary single-device success criterion):**
```bash
go test -tags cuda -bench=BenchmarkGPUProve -benchtime=3x ./gpu/plonk2/bls12377/
```
Result must be within **20%** of `gpu/plonk` baseline from Phase 0 on a single device. The 20% band (vs the 10% used in earlier drafts) reflects giving up the TE inner-loop advantage per Decision 4. Recovering further is the job of Phase 5b's multi-GPU work, not of re-introducing TE.

If single-device is outside 20%, do not proceed to Phase 5b. Profile with `nsys`. Common causes:
- Missing stream pipelining (uploads not overlapped with compute)
- Misaligned pinned-memory transfers
- SRS chunking threshold mistuned for the host's PCIe bandwidth
- Default-stream fallback inside a kernel wrapper that should accept a stream

---

### Phase 5b — Multi-GPU Orchestration

**Goal:** Extend the per-curve `prove.go` to make use of all devices in the `DeviceGroup`, following the **Multi-GPU Strategy** section's decision rules. The single-device path from Phase 5a must continue to hit its existing benchmark — 5b changes nothing for `len(devs) == 1`.

**Actions:**
1. Audit `prove.go.tmpl` for phases listed in the Multi-GPU Strategy table. For each phase, decide whether to apply (a) independent-op assignment, (b) single-op split, or (c) leave single-device because it is not on the critical path.
2. Wrap independent-op fan-out in `errgroup.Group`. Each goroutine pins its goroutine to a specific device with `runtime.LockOSThread` only if measurement shows it matters; otherwise let the kernel wrapper manage device context.
3. Implement per-device `MSM` contexts inside `gpuInstance`. Each device owns its slice of the SRS, its own work buffers, and its own `OffloadPoints / ReloadPoints` lifecycle. The slice partition is fixed at construction.
4. Implement host-side combine for split MSMs (sum of partial G1 Jacobian points). This is a few-cycle scalar operation; keep it inline in `prove.go`.
5. Add `BenchmarkGPUProveMultiGPU` and `BenchmarkMSMSplit` to the per-curve bench template.

**Test additions:**
- For every existing prove test, parameterize over `len(devs) ∈ {1, 2}` (and `{1, 2, 4}` when 4 GPUs are visible). All variants must produce byte-identical proofs.
- Failure to schedule (e.g., `len(devs) == 3`) must return a clear error, not silently degrade.
- Stress test: 100 sequential proves on a 2-GPU group, verify no leak (VRAM, pinned host memory, file descriptors) by comparing `nvidia-smi` snapshots before and after.

**Acceptance criteria:**
```bash
# All 5a tests still pass with len(devs) == 1:
go test -tags cuda -v -timeout=30m ./gpu/plonk2/...

# Multi-GPU tests (skipped automatically if only one GPU is visible):
go test -tags cuda -v -timeout=45m -run=MultiGPU ./gpu/plonk2/...
```

**Performance gate (multi-GPU):**
```bash
go test -tags cuda -bench=BenchmarkGPUProveMultiGPU -benchtime=3x ./gpu/plonk2/bls12377/
```

| Configuration | Required speedup vs single-GPU on same host |
|---|---|
| 2 GPUs | ≥ 1.5× |
| 4 GPUs | ≥ 2.5× |

Anything below those bars indicates the orchestrator is leaving devices idle or oversynchronizing. Re-read the Multi-GPU Strategy decision rules and profile with `nsys` showing per-device timelines before changing anything.

If after Phase 5b the multi-GPU BLS12-377 prover beats the legacy `gpu/plonk` (single-GPU TE) prover by any margin on the same host, the rewrite has met its goal. Decision 4 is vindicated.

---

### Phase 6 — Root Package Cleanup

**Goal:** Delete dead code, rewire the root `prover.go` dispatcher, move benchmarks.

**Actions:**
1. Delete all files listed in the "Files deleted from gpu/plonk2/ root" table above
2. Rewrite `gpu/plonk2/prover.go` as a thin multi-curve dispatcher:
   - `NewProver(dev, ccs, pk, opts...) (*Prover, error)` — inspects curve ID, instantiates typed per-curve prover
   - `Prove(witness, opts...) (gnarkplonk.Proof, error)` — delegates, applies CPU fallback from options
3. Move `icicle_*.go` and all `bench_*_test.go` from root to `internal/bench_icicle/`
4. Verify `gpu/plonk2/options.go` only exports options that root `prover.go` needs; delete `WithLegacyBLS12377Backend` and `WithTrace` (trace belongs in per-curve prove.go if needed)
5. Write `gpu/plonk2/stub.go` (build tag `!cuda`) that returns `ErrNoCUDA` from all exported functions with a clear message

**Acceptance criteria:**
```bash
# Non-CUDA build must compile cleanly:
go build ./gpu/plonk2/...

# CUDA build must compile cleanly:
go build -tags cuda ./gpu/plonk2/...

# All per-curve tests still pass:
go test -tags cuda -v -timeout=30m ./gpu/plonk2/...

# No reference to deleted files in any import:
grep -r "generic_prove\|generic_finalize\|curve_proof_ops\|generic_quotient\|generic_prepare" \
    gpu/plonk2/ && echo "FAIL: stale references" || echo "OK"
```

---

### Phase 7 — CUDA Documentation Pass

**Goal:** Every kernel has a compliant documentation comment. No functional changes.

**Actions:**
For each kernel in `cuda/src/`:
1. Read the kernel body
2. Add the required comment block: algorithm name, preconditions, thread layout, non-obvious invariants
3. Add a one-line comment on every non-trivial formula line (e.g., next to the Montgomery reduction step, the SW mixed-add formula, the Pippenger bucket accumulation)
4. If a kernel has a magic constant (e.g., modulus limbs, window size), name it and explain the choice

**Acceptance criteria:**
- Every `__global__` function has a documentation comment of ≥ 3 lines
- Build still passes: `cd cuda && cmake --build build 2>&1 | grep -c error` must output `0`
- No test regressions

---

### Phase 8 — Deprecation of gpu/plonk (deferred)

**Do not execute this phase until bls12377/ has been in production for at least one release cycle.**

**Actions:**
1. Add `// Deprecated: use gpu/plonk2/bls12377 instead.` to `gpu/plonk` package doc
2. Update all call sites in the monorepo to import `gpu/plonk2` (root dispatcher)
3. Delete `gpu/plonk` in a separate PR after all call sites are migrated

---

## Testing Protocol

### Running Tests Without a GPU

All test files must compile and the non-CUDA subset must pass without a GPU. Use build tags:

```go
//go:build cuda   // gates the entire file behind CUDA requirement
```

vs

```go
// no build tag — always compiled; uses stub which returns ErrNoCUDA
func TestFrVectorAddNoCUDA(t *testing.T) {
    _, err := bn254.NewFrVector(gpu.NoDevice, 16)
    require.ErrorIs(t, err, gpu.ErrNoCUDA)
}
```

CI pipeline must run both:
```bash
# Step 1 — always runs (no GPU needed):
go test ./gpu/plonk2/...

# Step 2 — runs only on GPU-enabled runners:
go test -tags cuda ./gpu/plonk2/...
```

### Property-Based Test Requirements

Every mathematical primitive must have a gopter property test:

| Primitive | Property to test |
|---|---|
| `FrVector.Add` | commutativity: GPU(a+b) == GPU(b+a) |
| `FrVector.Mul` | associativity: GPU((a*b)*c) == GPU(a*(b*c)) |
| `FrVector.BatchInvert` | GPU(BatchInvert(v))[i] * v[i] == 1 ∀ nonzero v[i] |
| `FrVector.ScaleByPowers` | GPU result matches CPU loop ∀ omega, n |
| `FFTDomain.FFT` | GPU(FFT(v)) == gnark-crypto-CPU(FFT(v)) |
| `FFTDomain.FFTInverse` | GPU(FFTInverse(GPU(FFT(v)))) == v |
| `G1MSM.MultiExp` | GPU(MultiExp(pts, scalars)) == gnark-crypto-CPU(MultiExp) |
| Prove | GPU(Prove(w)) is accepted by gnark-crypto Verify |

Minimum `gopter.DefaultTestParameters().MinSuccessfulTests` = 50 for CUDA tests (GPU tests are slow; do not set to 1000).

### Regression Tests

After Phase 5a, add a fixed-vector regression test for each curve:

```go
// plonk_test.go — deterministic test circuit with fixed seed
func TestProveRegression_{{.Name}}(t *testing.T) {
    // Fixed circuit, fixed witness, fixed randomness (use deterministic hash source)
    // Expected proof bytes hard-coded as hex constant in test file
    // Fails if proof changes — forces explicit update of expected bytes
}
```

This catches silent behavioral regressions when kernels are modified.

---

## Benchmarking Protocol

### What to Measure

For each curve and each phase gate, measure these four metrics:

| Metric | Command |
|---|---|
| End-to-end prove time | `BenchmarkGPUProve` |
| FFT throughput (size 2^24) | `BenchmarkFFT_2_24`, `BenchmarkCosetFFT_2_24` |
| MSM throughput (2^26 points) | `BenchmarkMSM_2_26` |
| Peak VRAM usage | `nvidia-smi dmon -s m` during benchmark |

### Benchmark File Structure

Each per-curve package gets a `bench_test.go` generated from a template:

```go
// bench_test.go.tmpl
//go:build cuda

package {{.Package}}_test

import (
    "testing"
    // ... imports ...
)

func BenchmarkFFT_2_24(b *testing.B) {
    dev := requireGPU(b)
    domain, _ := NewFFTDomain(dev, 1<<24)
    v := randomFrVector(dev, 1<<24)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        domain.FFT(v)
        dev.Synchronize()
    }
    b.ReportAllocs()
}

func BenchmarkMSM_2_26(b *testing.B) { ... }
func BenchmarkGPUProve(b *testing.B) { ... }
```

### Pass/Fail Criteria

| Phase | Metric | Threshold |
|---|---|---|
| Phase 2 | FrVector ops | Within 5% of gpu/plonk baseline |
| Phase 3 | FFT, CosetFFT | Within 5% of gpu/plonk baseline |
| Phase 4 | MSM, single device (BLS12-377) | Within 20% of gpu/plonk baseline (affine vs TE — see Decision 4) |
| Phase 4 | MSM, 2-GPU split (BLS12-377) | ≥ 1.6× the single-device MSM (when 2 GPUs available) |
| Phase 5a | End-to-end prove, single device (BLS12-377) | Within 20% of gpu/plonk baseline |
| Phase 5a | Peak VRAM single device (BLS12-377, 2^27 pts) | ≤ gpu/plonk baseline + 500 MiB |
| Phase 5a | End-to-end prove, single device (BN254) | Recorded as new baseline |
| Phase 5a | End-to-end prove, single device (BW6-761) | Recorded as new baseline |
| Phase 5b | End-to-end prove, 2-GPU (BLS12-377) | ≥ 1.5× single-device prove on same host |
| Phase 5b | End-to-end prove, 4-GPU (BLS12-377) | ≥ 2.5× single-device prove on same host |
| Phase 5b | Peak VRAM per device, 2-GPU (BLS12-377) | ≤ single-device VRAM × 0.6 (SRS partitioned) |

If a phase benchmark fails its gate, stop. Do not proceed. Profile before continuing:

```bash
# CUDA profiling:
nsys profile --trace=cuda,nvtx go test -tags cuda -bench=BenchmarkGPUProve \
    -benchtime=1x ./gpu/plonk2/bls12377/

# Memory profiling:
nvidia-smi dmon -s m -d 1 &
go test -tags cuda -bench=BenchmarkGPUProve -benchtime=1x ./gpu/plonk2/bls12377/
kill %1
```

### Comparative Benchmark (bls12377 only)

After Phase 5a (single-device) and again after Phase 5b (multi-GPU), run a side-by-side comparison and commit the output:

```bash
# Single-device, after Phase 5a:
go test -tags cuda -bench=. -benchtime=5x ./gpu/plonk/                  > bench_plonk_legacy.txt
go test -tags cuda -bench=. -benchtime=5x ./gpu/plonk2/bls12377/        > bench_plonk2_bls12377_single.txt
benchstat bench_plonk_legacy.txt bench_plonk2_bls12377_single.txt

# Multi-GPU, after Phase 5b (2-GPU host required):
GPU_DEVICES=0,1 go test -tags cuda -bench=BenchmarkGPUProveMultiGPU \
    -benchtime=5x ./gpu/plonk2/bls12377/                                > bench_plonk2_bls12377_multi.txt
benchstat bench_plonk_legacy.txt bench_plonk2_bls12377_multi.txt
```

After Phase 5a, the `benchstat` output may show up to 20% regression vs legacy on single-device prove (deliberate, per Decision 4). After Phase 5b, the multi-GPU run must beat or match legacy single-device prove on the same hardware. If it does not, do not move to Phase 6.

---

## What NOT to Do

These are explicit prohibitions. If you are about to do any of these, stop and re-read the plan.

1. **Do not create a shared `gpu/base/` package** with curve-indexed functions. That is just plonk2's current approach with a different name. The point is typed specializations.

2. **Do not use `[]uint64` in any public or internal API** inside the generated per-curve packages. Raw uint64 slices have no semantic meaning. Use `fr.Element`, `G1Affine`, etc.

3. **Do not use `interface{}` or `any` to represent field elements or curve points** inside the generated packages.

4. **Do not remove stream support from any operation.** If a function currently takes no stream in gpu/plonk, check whether it should. When in doubt, add `stream ...gpu.StreamID`.

5. **Do not fold the CPU fallback into the per-curve packages.** CPU fallback is a deployment policy, not a cryptographic concern. Keep it in the root `prover.go`.

6. **Do not delete `gpu/plonk` until Phase 8 is explicitly authorized.** It is the reference implementation and performance baseline.

7. **Do not skip phase acceptance criteria to "move faster."** A benchmark regression discovered in Phase 6 is far more expensive to debug than one caught in Phase 3.

8. **Do not add features not in this plan** (statistical ZK, new commitment schemes, recursive proofs). This refactor is a rewrite for correctness, performance, and maintainability. Scope is frozen.

9. **Do not modify any CUDA kernel without running the full benchmark suite.** Kernel changes have non-obvious register pressure and occupancy effects.

10. **Do not commit generated files without the `// Code generated by gpu/internal/generator DO NOT EDIT` header.** Without this header, reviewers will attempt to edit generated code by hand.

11. **Do not re-introduce a Twisted Edwards path inside the generated packages** without an explicit, measured proposal that revisits Decision 4. The legacy TE implementation in `gpu/plonk` stays as a reference; production runs through affine SW.

12. **Do not run multiple `Prover` instances on the same host.** Multi-GPU is intra-prover via `DeviceGroup`. Host memory does not allow two in-flight proofs and the orchestrator above the `Prover` is expected to serialize.

13. **Do not split a single NTT across devices in v1.** A 1D NTT requires data exchange at every stage where butterfly distance crosses the partition boundary. The 2D / six-step decomposition that makes this tractable is out of scope for this rewrite.

14. **Do not hard-code `nDevices` (1, 2, 4) anywhere in the orchestrator or templates.** The schedule must be expressed in terms of `len(devs)` and work for any size in `{1, 2, 4}`.

15. **Do not open a PR against `main` while phases 1–7 are in flight.** Work lives on `prover/gpu2`; phase boundaries are checkpoint commits. Phase 8 is the only phase that touches `main`, and only when explicitly authorized.

---

## Agent Instructions

When an agent is assigned a phase from this plan, it must:

1. **Read this document in full before writing any code.** In particular, the Multi-GPU Strategy and "What NOT to Do" sections.
2. **Read `gpu/plonk/prove.go`, `gpu/plonk/msm.go`, `gpu/plonk/fft.go`, `gpu/plonk/fr.go`** — these are the reference implementations to templatize.
3. **Read `gnark-crypto/internal/generator/main.go` and at least one `.go.tmpl` file** to internalize the bavard pattern before writing templates.
4. **Run acceptance criteria commands verbatim** — do not paraphrase them or substitute equivalent commands.
5. **Report benchmark numbers in the phase checkpoint commit message**, not just "tests pass." Include the `benchstat` output where applicable.
6. **Do not advance to the next phase until that phase's benchmark gates pass.** Phases are sequenced for a reason.
7. **Do not push to `main` or open a PR.** Work is on `prover/gpu2`; phase boundaries are checkpoint commits on this branch.

For Phase 5a specifically: the agent must produce a `benchstat` comparison against the Phase 0 baseline and include it in the checkpoint commit message. If the comparison shows a regression > 20% on BLS12-377 single-device prove, identify the cause (missing stream pipelining, wrong chunking threshold, default-stream fallback inside a kernel wrapper) and fix it before moving to Phase 5b.

For Phase 5b specifically: the agent must produce per-device `nsys` timeline screenshots (or a textual summary if profiling tools are unavailable) showing that no device is idle during MSM-heavy phases. The 2-GPU and 4-GPU benchmarks must each be in a separate `benchstat` comparison against the Phase 5a single-device baseline.

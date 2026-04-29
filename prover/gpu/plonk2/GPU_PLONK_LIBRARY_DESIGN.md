# GPU PlonK Library Design and Roadmap

This document describes a target architecture for a lean CUDA library with Go
bindings that can replace gnark's CPU PlonK prover for BN254, BLS12-377, and
BW6-761.

It is based on reading the current branch code, not only the Markdown notes:

- `gpu/plonk` is a working BLS12-377 prover path with a specialized
  Twisted-Edwards MSM, GPU NTTs, quotient kernels, persistent proving-key state,
  pinned host buffers, and explicit stream/event scheduling.
- `gpu/plonk2` is a curve-generic primitive layer. It has curve-indexed Fr
  vectors, NTT domains, quotient kernels, raw KZG commitments, resident
  short-Weierstrass MSM handles, and setup-commitment tests across all three
  target curves.
- `gpu/plonk2` is not yet a drop-in full prover. Full GPU proving benchmarks in
  the current tree still call the older `gpu/plonk` BLS12-377 prover.
- The generic `plonk2` MSM has improved beyond the earliest correctness
  baseline, but it still has production gaps: no batched shared-base API,
  host-side Montgomery correction, limited memory planning enforcement, no
  chunked execution policy, no large-bucket scheduling, and no full prover
  integration.

## Goal

The goal is a small, auditable GPU backend that can be used as:

```go
proof, err := gpuplonk.Prove(dev, ccs, pk, witness, opts...)
```

with the same proof bytes and verifier compatibility as:

```go
proof, err := plonk.Prove(ccs, pk, witness, opts...)
```

The GPU implementation should accelerate the dominant PlonK work while leaving
the PlonK protocol, transcript, proof shape, and gnark public APIs intact.

## Non-Goals

- Do not fork the PlonK protocol unless gnark makes a protocol hook impossible.
- Do not expose CUDA tuning knobs in the user-facing API until defaults are
  correct and stable.
- Do not make BLS12-377 Twisted-Edwards layout the generic architecture. It can
  remain an optional backend after the affine path is production-grade.
- Do not optimize before each primitive is proven equivalent to gnark-crypto on
  all target curves.
- Do not create broad utility packages for one-off prover needs.

## Current Branch Assessment

### `gpu/plonk`

`gpu/plonk` is production-shaped but curve-specific. The important patterns are:

- `GPUProvingKey` owns the verifying key, SRS input, and lazy `gpuInstance`.
- `gpuInstance` owns persistent device resources: MSM handle, FFT domain,
  permutation table, and selector/permutation polynomials in GPU memory.
- `gpuProver` owns per-proof mutable state and implements the prove phases.
- Large host buffers are preallocated once per instance to reduce GC pressure.
- Hot fixed polynomials are uploaded once and reused by device-to-device copies.
- MSM work buffers are pinned for commitment waves, then released before
  quotient phases that need VRAM.
- Quotient computation uses multiple working vectors and a two-stream pipeline
  to overlap selector copies with FFT work.
- Memory pressure is handled pragmatically with best-effort GPU-resident paths
  and host fallback paths.

The weaknesses are equally important:

- It is hard-wired to BLS12-377 gnark types.
- It uses a specialized BLS12-377 Twisted-Edwards MSM surface.
- Prover orchestration and low-level kernels are entangled in one package.
- Several wrappers panic on misuse, which is tolerable internally but not ideal
  at a drop-in library boundary.

### `gpu/plonk2`

`gpu/plonk2` has the right generic primitive boundary:

- `Curve` and `CurveInfo` describe scalar/base field widths.
- `FrVector` stores device scalars in SoA layout and accepts gnark-crypto raw
  AoS Montgomery host buffers.
- `FFTDomain` owns curve-specific twiddles and exposes forward, inverse, bit
  reversal, and coset transforms.
- Quotient helpers cover the PlonK kernels needed by the old prover flow:
  gate accumulation, permutation boundary, Z factors, prefix product,
  blinded-coset reduction, and inverse butterfly.
- `G1MSM` keeps a short-Weierstrass affine SRS resident and exposes `CommitRaw`.
- Tests compare raw commitments, KZG commitments, KZG opening quotients, and
  setup selector commitments against gnark for BN254, BLS12-377, and BW6-761.

Observed gaps:

- There is no curve-generic full prover path.
- `CommitRaw` returns to Go for Montgomery-layout correction.
- MSM is still one commitment at a time. PlonK needs commitment waves.
- MSM scratch planning exists in Go but is not yet a hard runtime contract.
- Resident MSM buffers can be pinned, but there is no chunking/offload policy
  comparable to `gpu/plonk`.
- The current CUDA cleanup path in `gpu/cuda/src/plonk2/msm.cu` frees
  `d_buckets` twice in one error cleanup block. This is a concrete hygiene
  issue and should be fixed before performance work continues.
- NTTs are generic but simpler than the optimized BLS12-377 NTT in `gpu/plonk`;
  order and residency are not yet represented as an internal plan.
- The non-CUDA build currently fails on this macOS host because
  `gpu/threadlocal.go` calls `unix.Gettid`, which is not available on Darwin.

## Design Principles

1. Keep the public surface small.

   Users should see device selection, proving-key preparation, proof
   generation, and explicit cleanup. They should not see buckets, windows,
   CUB workspaces, streams, or host pinning unless they opt into diagnostics.

2. Keep the C ABI flat.

   The Go/C boundary should pass opaque handles, curve IDs, raw pointers, sizes,
   and explicit stream IDs. Avoid ABI structs that mirror C++ template internals.

3. Keep curve differences data-driven.

   Curve-specific facts belong in compact descriptors and C++ template
   instantiations. Go orchestration should not be copied per curve.

4. Keep protocol logic in Go.

   CUDA kernels should implement data-parallel arithmetic, transforms, MSM, and
   fused polynomial kernels. Transcript binding, proof assembly, and gnark type
   compatibility should stay in Go.

5. Prefer resident data to repeated transfers.

   SRS points, twiddles, permutation tables, selectors, and stable polynomials
   should live on device for the lifetime of a prepared proving key.

6. Make memory planning explicit.

   Every major allocation must be classified as persistent, per-proof, per-wave,
   or per-kernel. BW6-761 must drive the memory model because it has the widest
   fields and largest point/bucket footprint.

7. Make async execution explicit.

   Every operation should either be documented as enqueue-only or as
   synchronizing. Hidden `cudaStreamSynchronize` calls should be removed from
   hot paths unless they are part of a documented API contract.

8. Optimize only behind correctness tests.

   Each kernel transformation must preserve CPU/GPU equality tests before
   benchmark numbers matter.

9. Less code is better, but not less structure.

   The target is fewer concepts, fewer public entrypoints, and clear ownership.
   Comments and design documents should explain invariants and dataflow, not
   restate obvious code.

## Package Shape

Recommended package split:

```text
gpu/
  device.go                 Shared CUDA device, streams, events, diagnostics
  memory.go                 Pinned host and device allocation helpers
  plonk2/
    doc.go                  Public package overview
    prove.go                Drop-in generic Prove entrypoint
    proving_key.go          Prepared proving-key lifecycle
    instance.go             Persistent per-circuit GPU state
    prover.go               Per-proof orchestration
    curve.go                Curve descriptors
    raw.go                  Safe raw-layout adapters
    fr.go                   Device scalar vectors
    fft.go                  NTT domains and plans
    msm.go                  Resident MSM handle and commitment API
    msm_plan.go             Memory planner and run planner
    quotient.go             PlonK quotient kernels
    opening.go              KZG opening helpers
    tests/...               Curve-generic correctness tests
  cuda/
    include/gnark_gpu.h     Stable C ABI
    src/plonk2/             CUDA implementation
```

The split should be done gradually. Do not move files just for aesthetics while
the backend is still changing. The first goal is to make ownership and dataflow
clear in code, then move files if it reduces cognitive load.

## Public Go API

The top-level proving API should be small:

```go
type Prover struct {
    // owns prepared GPU state and a CPU fallback policy
}

func NewProver(dev *gpu.Device, ccs constraint.ConstraintSystem, pk plonk.ProvingKey, opts ...Option) (*Prover, error)
func (p *Prover) Prove(fullWitness witness.Witness, opts ...backend.ProverOption) (plonk.Proof, error)
func (p *Prover) Close() error

func Prove(
    dev *gpu.Device,
    ccs constraint.ConstraintSystem,
    pk plonk.ProvingKey,
    fullWitness witness.Witness,
    opts ...backend.ProverOption,
) (plonk.Proof, error)
```

The drop-in function may prepare and close internally. High-throughput callers
should use `NewProver` so SRS points, twiddles, selectors, and buffers persist
across proofs.

Options should initially be limited:

- `WithCPUFallback(bool)` for operational rollout.
- `WithMemoryLimit(bytes uint64)` for hosts that need a hard VRAM budget.
- `WithPinnedHostLimit(bytes uint64)` for system RAM pressure.
- `WithTrace(path string)` for phase timing and allocation diagnostics.

Benchmark-only tuning should stay internal or behind build tags. Public window
sizes and bucket knobs are premature.

## C ABI Contract

The C ABI should expose opaque handles and explicit lifecycle:

```c
gnark_gpu_error_t gnark_gpu_plonk2_context_prepare(...);
gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_alloc(...);
gnark_gpu_error_t gnark_gpu_plonk2_ntt_domain_create(...);
gnark_gpu_error_t gnark_gpu_plonk2_msm_create(...);
gnark_gpu_error_t gnark_gpu_plonk2_msm_run(...);
void gnark_gpu_plonk2_msm_destroy(...);
```

Rules:

- Every allocation has one owner and one destroy function.
- Every handle records curve ID, device ID, capacity, and stream ownership.
- Every run function validates count, curve, device, and capacity.
- No run function allocates large temporary buffers if a prepared handle exists.
- Stream-aware variants should not synchronize unless the name says `sync`.
- Error reporting should include enough context to diagnose OOM versus invalid
  shape versus CUDA launch failure.

The current single header is acceptable for now, but it should be organized into
sections that mirror lifecycle, memory, Fr, NTT, MSM, quotient, and diagnostics.

## Data Layout

Host boundaries should use gnark-crypto layouts:

- Fr elements: raw AoS Montgomery words.
- G1 points: short-Weierstrass affine raw words, X then Y.
- KZG commitments and proof points: gnark curve types on the Go side.

Device layouts should be chosen for throughput:

- Fr vectors: SoA by limb for coalesced element-wise field kernels.
- Affine SRS: initially AoS raw points for minimal code; later evaluate SoA or
  precomputed forms with benchmarks.
- Buckets: curve-specialized projective/Jacobian arrays.
- Assignment keys/values: compact 32-bit keys and packed point index/sign where
  point counts permit it.

No code should rely on Go struct memory layout unless it is isolated in one raw
adapter file with tests.

## Prover Dataflow

The target prover should keep the proven `gpu/plonk` phase structure:

```text
prepare proving key
  parse SparseR1CS trace
  prepare SRS MSM
  prepare FFT domain
  upload permutation table
  convert stable setup polynomials when needed
  upload stable polynomials
  allocate reusable host/device buffers

prove
  solve witness
  complete Qk with public inputs and BSB22 commitments
  iFFT L/R/O/Qk
  blind and commit L/R/O
  derive gamma/beta
  build Z and commit Z
  compute quotient and commit H1/H2/H3
  derive zeta
  evaluate/open polynomials
  build linearized polynomial and opening proof
  assemble gnark-compatible proof
```

The generic implementation should first reproduce the BLS12-377 proof produced
by `gpu/plonk`, then enable BN254 and BW6-761.

## Prepared Proving Key

`PreparedProvingKey` should own persistent state:

- curve descriptor;
- gnark verifying key;
- domain size and log size;
- resident canonical and Lagrange SRS handles as needed;
- FFT domain handles for `n` and any decomposed quotient domains;
- device permutation table;
- resident fixed polynomials;
- reusable MSM work buffers;
- reusable quotient buffers where memory permits;
- pinned host buffers within a configured limit.

Preparation should be explicit and measurable. The user should be able to time
setup separately from proof generation.

## Memory Model

Every allocation should be visible in a memory plan:

```text
persistent:
  SRS points
  twiddle tables
  selector/permutation polynomials
  permutation table

per-proof:
  witness vectors
  L/R/O/Z blinded polynomials
  Qk patched polynomial
  quotient working vectors
  opening working vectors

per-commitment-wave:
  scalar staging
  assignment keys/values
  CUB sort temp
  bucket offsets/ends
  buckets
  window partials/results

host pinned:
  SRS staging when not loaded from disk directly
  hot polynomial buffers
  scalar upload buffers
```

Memory planner requirements:

- Return an exact upper bound for each allocation class for a given curve,
  domain size, point count, window policy, batch size, and chunking policy.
- Ask CUDA for free/total VRAM during preparation and before large phases.
- Select window bits and chunk size from the plan, not from scattered constants.
- For BW6-761, prefer chunking over optimistic full-resident execution.
- Release MSM wave buffers before quotient if the quotient needs the memory.
- Avoid page-locking tens of GiB of host RAM without an explicit configured cap.

## Streams and Scheduling

The default stream model should be simple:

- stream 0: arithmetic/FFT/quotient compute;
- stream 1: host-device and device-device transfers;
- stream 2: MSM pipeline;
- optional stream 3: large-bucket or batch-overlap work after the MSM planner
  proves it helps.

Rules:

- Go code records dependencies through explicit events.
- CUDA run functions do not create streams.
- CUDA run functions accept a stream or use the handle's default stream.
- Synchronization happens at phase boundaries, before host reads, and before
  proof assembly.
- Traces record enqueue time, GPU event elapsed time, bytes moved, and peak
  memory estimate.

Avoid using concurrency to hide unclear ownership. If two goroutines can touch
one CUDA handle, the handle must document whether it is single-flight,
internally locked, or caller-serialized.

## NTT Design

Current `plonk2` NTTs are useful but should gain an internal plan:

```go
type NTTPlan struct {
    Curve Curve
    Size int
    Direction Direction
    InputOrder Order
    OutputOrder Order
    InputResidency Residency
    OutputResidency Residency
    Batch int
}
```

The public API can stay small while the prover uses the plan internally.

Important work:

- Track natural versus bit-reversed order to avoid standalone bit reversals.
- Add batched NTTs for commitment waves and quotient cosets.
- Keep coset scaling fused with transforms where profitable.
- Reuse twiddle tables and local power tables.
- Compare generic NTT kernels against the optimized BLS12-377 `gpu/plonk` NTT.

Acceptance:

- Roundtrip equality for all curves and supported sizes.
- Coset FFT equality against gnark-crypto.
- Benchmarks for forward, inverse, coset, and batched transforms.

## MSM Design

MSM is the critical path. The target backend should keep one affine
short-Weierstrass input contract for all curves and optimize the private
pipeline.

Baseline pipeline:

```text
scalars -> signed windows -> assignment keys/values
        -> radix sort
        -> bucket metadata
        -> bucket accumulation
        -> per-window reduction
        -> final window combination
        -> device-side normalization/correction
```

Required design points:

- Resident SRS handle with explicit capacity.
- Shared-base batched commitments for PlonK waves.
- Preallocated CUB sort temp and work buffers.
- Window and chunk policy selected by `MSMRunPlan`.
- Device-resident output option for prover phases that can consume the result.
- Device-side Montgomery correction.
- Large-bucket segmentation so one skewed bucket does not serialize a block.
- Parallel per-window reduction; no one-thread window running sums.
- Optional base precomputation only after memory planning proves it is worth it.

`MSMRunPlan` should include:

```go
type MSMRunPlan struct {
    Curve Curve
    Points int
    ScalarBits int
    WindowBits int
    Windows int
    BatchSize int
    ChunkPoints int
    SharedBases bool
    PrecomputeFactor int
    LargeBucketFactor int
    Bytes MSMMemoryPlan
}
```

The first production path should not expose this publicly. It should be logged
in traces and benchmark output.

Acceptance:

- CPU/GPU equality for random scalars, zero scalars, one-hot scalars, repeated
  points, infinity points where valid, and adversarial bucket-skew scalars.
- KZG commitment equality for all target curves.
- Setup-commitment equality for all selector/permutation commitments.
- Benchmarks for 16, 1K, 16K, 64K, 1M, and large memory-permitting sizes.
- No large `cudaMalloc` or `cudaFree` inside repeated `CommitRaw` when work
  buffers are pinned.

## Quotient and Opening Design

The quotient path should be ported after the generic NTT and MSM surfaces are
stable enough.

Rules:

- Keep the mathematical flow from gnark and `gpu/plonk`.
- Keep fixed polynomials resident.
- Use device-to-device copies for selector reuse.
- Preserve host fallback paths when VRAM is insufficient.
- Make every CPU/GPU boundary explicit.

The following kernels are already present in `plonk2` and should be validated
as orchestration is wired:

- blinded-coset reduction;
- L1 denominator computation;
- batch inversion or replacement scan;
- permutation boundary accumulation;
- gate accumulation;
- Z factor computation;
- Z prefix product;
- inverse butterfly for decomposed quotient recovery.

Opening work should avoid building large folded polynomials on the Go heap when
the data is already resident. The first version can keep the current CPU/GPU
split from `gpu/plonk`; later milestones can move folding and Horner quotient
work to GPU where benchmarks justify it.

## Error Handling and Cleanup

Library boundaries should return errors. Internal invariant violations may
panic only in private helpers where callers cannot recover.

Required cleanup discipline:

- Every Go handle has `Close` or `Free` and a finalizer only as a leak guard.
- Explicit `Close` is used in production paths.
- C++ cleanup labels must be idempotent and free each pointer at most once.
- Tests should include allocation-failure or early-error paths where practical.
- CUDA errors should include the phase name and curve.

## Testing Strategy

Non-CUDA tests:

- curve descriptor validation;
- memory plan arithmetic;
- raw layout conversion tests;
- API shape and fallback behavior;
- gnark CPU reference E2E tests for all target curves.

CUDA correctness tests:

- Fr add/sub/mul/scalar/addmul/batch invert for all curves;
- NTT forward/inverse/coset for all curves and sizes;
- G1 affine operations and MSM for all curves;
- KZG commit/open quotient equality for all curves;
- PlonK setup commitment equality for all curves;
- full proof generation and verification for all curves;
- negative witness tests where CPU proving fails and GPU proving must fail.

Performance tests:

- cold prepare time;
- warm proof time;
- phase timings;
- H2D/D2H bytes;
- peak planned and observed VRAM;
- pinned host bytes;
- CPU baseline versus GPU baseline;
- old `gpu/plonk` BLS12-377 versus new generic BLS12-377.

CI should include non-CUDA compile/tests by default. CUDA tests can remain on
GPU runners, but their benchmark commands and hardware metadata must be logged.

## Milestones

Each milestone should be independently reviewable and should leave tests green.

### Milestone 0: Hygiene and Compile Baseline

- Fix the double-free cleanup in `gpu/cuda/src/plonk2/msm.cu`.
- Fix or guard `gpu/threadlocal.go` so non-CUDA macOS builds compile.
- Run `go test ./gpu/plonk2 ./gpu/plonk` without CUDA tags.
- Run existing CUDA tests on a GPU host.
- Add a short status section to the worklog with exact commands and hardware.

Exit criterion: no known cleanup bug, non-CUDA package tests compile, and the
current CUDA suite still passes on a GPU machine.

### Milestone 1: API Contract and Prepared Prover Skeleton

- Add `Prover`, `NewProver`, `Prove`, and `Close` skeletons.
- Detect BN254, BLS12-377, and BW6-761 gnark proving key types.
- Return a clear unsupported-curve error for other curves.
- Add CPU fallback option and tests.
- Do not accelerate anything yet.

Exit criterion: the GPU package can be called as a drop-in wrapper and returns
the same proof as CPU through fallback.

### Milestone 2: Memory Planner as Runtime Contract

- Extend `PlanMSMMemory` to cover all real buffers used by CUDA.
- Add NTT, quotient, and prepared-key memory estimates.
- Query device memory and choose window/chunk policy during preparation.
- Log the selected plan in tests and benchmarks.
- Refuse plans that exceed the configured memory limit.

Exit criterion: every large allocation in the prover has a corresponding plan
entry and no large phase uses undocumented scratch memory.

### Milestone 3: MSM Correctness Hardening

- Move Montgomery correction into CUDA.
- Add structured-scalar and bucket-skew MSM tests.
- Add infinity and malformed-point tests if supported by the input contract.
- Make repeated `CommitRaw` with pinned buffers allocation-free in steady state.
- Fix any raw-layout assumptions into one adapter layer.

Exit criterion: MSM and KZG equality tests cover all edge cases on all curves.

### Milestone 4: MSM Throughput Plan

- Add internal `MSMRunPlan`.
- Implement chunked execution for memory-limited runs.
- Add bucket metadata with bucket sizes and non-empty bucket lists.
- Add large-bucket segmentation.
- Keep the old accumulation path available behind an internal switch until the
  new path is correct.

Exit criterion: the new path matches CPU and is faster or neutral on the
setup-commitment benchmark.

### Milestone 5: Batched Shared-Base Commitments

- Add private batched commitment API on `G1MSM`.
- Reuse scalar staging and sort buffers across a commitment wave.
- Return gnark-compatible commitments.
- Apply to PlonK setup commitments first.

Exit criterion: setup selector/permutation commitments for all curves use one
resident MSM handle and a batched or wave-aware execution path.

### Milestone 6: NTT Plan and Batched Transforms

- Add internal order/residency tracking.
- Avoid unnecessary bit reversals in the prover flow.
- Add batched transforms where PlonK phases naturally evaluate many polynomials.
- Compare generic BLS12-377 NTT to the old optimized NTT.

Exit criterion: NTT work is never repeated or reordered accidentally, and
benchmarks show the cost of each remaining bit reversal.

### Milestone 7: Generic BLS12-377 Full Prover

- Port `gpu/plonk` orchestration to the generic `plonk2` primitives.
- Keep the proof type and transcript exactly compatible with gnark.
- Start with BLS12-377 to compare against the existing GPU prover.
- Preserve CPU fallback behind an option.

Exit criterion: BLS12-377 GPU proofs verify and match gnark's public behavior;
benchmarks compare CPU, old GPU, and new generic GPU.

### Milestone 8: BN254 and BW6-761 Full Prover

- Enable the same generic prover flow for BN254.
- Enable BW6-761 with conservative chunking and memory limits.
- Add curve-specific benchmark cases for setup, prove, quotient, and opening.

Exit criterion: all three curves produce verifying proofs and have documented
phase timings.

### Milestone 9: Production Rollout Controls

- Add environment or option based rollout in downstream prover call sites.
- Emit traces for phase timing and memory decisions.
- Add CPU fallback on OOM and unsupported shapes where acceptable.
- Document operational requirements: CUDA version, GPU memory, host pinned
  memory, build tags, and expected benchmark commands.

Exit criterion: callers can enable the GPU prover safely without removing the
CPU path.

### Milestone 10: Optional Specialization

- Re-evaluate BLS12-377 Twisted-Edwards fast path as an opt-in backend.
- Evaluate SRS precomputation factors per curve.
- Evaluate multi-GPU proof-level scheduling.
- Move additional opening/folding work to GPU only if measured.

Exit criterion: every specialization has a benchmark-backed reason to exist.

## Small Task Backlog

The following tasks are deliberately small and mostly independent:

- Add a CUDA cleanup test or review checklist for single-free ownership.
- Replace raw `unsafe.Slice` point copies with named raw adapter helpers.
- Add `CurveInfo.PointWords()` and `CurveInfo.ScalarWords()` helpers.
- Add tests for `defaultMSMWindowBits` at BN254, BLS12-377, and BW6-761
  boundary sizes.
- Add `MSMRunPlan.String()` for benchmark logs.
- Add `MemoryPlan.TotalPersistent()` and `TotalScratch()` helpers.
- Add a benchmark that calls `CommitRaw` 100 times after `PinWorkBuffers`.
- Add a benchmark that alternates short and full scalar counts on one MSM.
- Add structured scalar tests that force one huge bucket.
- Add a test that releases and re-pins MSM buffers repeatedly.
- Add a device-side output path for MSM.
- Add device-side Montgomery correction.
- Add chunked MSM for `count > plan.ChunkPoints`.
- Add non-empty bucket count metadata.
- Add bucket-size histogram diagnostics under trace mode.
- Add large-bucket segmentation for one curve, then enable for all curves.
- Add private batched `CommitRawN`.
- Port setup-commitment benchmark to `CommitRawN`.
- Add NTT order enum and internal assertions.
- Add batched NTT API without using it.
- Use NTT order tracking in one quotient subphase.
- Add `PreparedProvingKey` skeleton with CPU fallback.
- Move `gpu/plonk` phase comments into reusable `plonk2` docs.
- Port only BLS12-377 L/R/O iFFT and commitment into `plonk2` prover.
- Port BLS12-377 Z construction.
- Port BLS12-377 quotient.
- Port BLS12-377 opening.
- Add BN254 compile-time type switch.
- Add BW6-761 compile-time type switch.
- Add full-prover E2E for each curve with a tiny circuit.
- Add invalid-witness full-prover tests for each curve.
- Add phase trace JSONL for prepare and prove.
- Add a benchmark report generator that consumes trace JSONL.

## Acceptance Bar

The library is not production-ready until all of the following are true:

- `gpuplonk.Prove` is a drop-in replacement for gnark PlonK on BN254,
  BLS12-377, and BW6-761.
- Full GPU proofs verify with gnark verifiers on all three curves.
- Invalid witnesses fail.
- CPU fallback remains available.
- All large allocations are planned and traceable.
- Repeated warm proofs do not allocate large GPU buffers in hot loops.
- The generic BLS12-377 path is within a defensible range of the old specialized
  `gpu/plonk` path, or the remaining gap is measured and explicitly accepted.
- BW6-761 has a conservative memory policy that does not assume a 98 GiB GPU.
- CUDA and Go ownership rules are documented at the call sites that matter.
- Benchmarks include exact commands, hardware, CUDA version, and commit SHA.

## Immediate Recommendation

Do not start by wiring the full prover. First fix hygiene, make the memory plan
authoritative, and harden MSM correctness/performance. The full prover will then
be a controlled orchestration port rather than a debugging session across
protocol logic, CUDA memory ownership, and curve arithmetic all at once.

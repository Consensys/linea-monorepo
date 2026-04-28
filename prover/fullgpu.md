# Full-GPU Prover Architecture

This document is the target design for making the on-the-fly Limitless prover
worth running on a `g7e.24xlarge`: 96 vCPUs, roughly 1 TiB RAM, and four
NVIDIA RTX PRO 6000 Blackwell GPUs with about 96 GiB VRAM each.

The CPU reference on `r8a.24xlarge` is about 10-15 minutes for the inner proof
including on-the-fly compilation, plus about 5 minutes for the outer proof. The
GPU target is an end-to-end proof below 5 minutes. Reaching that requires a
pipeline, not isolated kernel swaps: the GPUs must own the large polynomial
work, the CPU must keep compiling, loading, assigning, and reducing while GPUs
are busy, and host/device transfers must be deliberate.

## Current Pipeline

`cmd/prover prove` with `serialization=false` enters
`backend/execution/limitless.ProveOnTheFly`:

1. Build the Limitless zkEVM and distributed module graph in memory.
2. Compile GL/LPP/conglomeration segment circuits while trace parsing and the
   bootstrapper run.
3. Segment the bootstrapper runtime into GL and LPP witnesses.
4. Prove GL segments, then derive shared randomness.
5. Prove LPP segments.
6. Hierarchically conglomerate GL/LPP segment proofs in parallel with segment
   production.
7. Run the final execution circuit proof through gnark PlonK.

The heavy inner segment path is:

1. `compiled.ProveSegmentKoala`
2. `wizard.RunProverUntilRound`
3. `globalcs.QuotientCtx.Run`
4. Vortex `ColumnAssignmentProverAction`
5. Vortex `LinearCombinationComputationProverAction`
6. Vortex `OpenSelectedColumnsProverAction`
7. recursion/conglomeration proving over the segment proof

The outer path is `execCirc.MakeProof -> circuits.ProveCheck -> plonk.Prove`.
The final BLS conglomeration and the outer proof are the natural targets for
the BLS12-377 GPU PlonK layer.

## Principles

- One active segment prover per GPU by default. On this host that means
  `LIMITLESS_GPU_COUNT=4` and `LIMITLESS_SUBPROVER_JOBS=4`.
- Keep each segment pinned to one GPU. No hot cross-GPU movement; this machine
  has PCIe, not NVLink.
- Lock the goroutine to its OS thread before calling CUDA. CUDA current-device
  state is per-thread, so every GPU worker must call `cudaSetDevice`.
- Treat Vortex encoded matrices as device-resident data. Full D2H of encoded
  matrices destroys the GPU win and fills the Go heap.
- Keep memory lifetimes explicit. Cached pipelines are useful only within a
  short phase; unbounded caches become OOM bugs.
- Prefer recomputation over residency when residency would exceed the VRAM
  budget. A 250 ms recommit is cheaper than a 96 GiB OOM or a corrupted proof.
- Leave CPU cores busy with compilation, trace expansion, witness IO,
  precomputed setup loading, and conglomeration while GPUs run quotient and
  Vortex work.

## Target Scheduling

The runtime should use these knobs on `g7e.24xlarge`:

```bash
make bin/prover-cuda
# equivalent: GO_BUILD_TAGS=debug,cuda make bin/prover

LIMITLESS_GPU_COUNT=4 \
LIMITLESS_SUBPROVER_JOBS=4 \
LIMITLESS_MERGE_JOBS=1 \
LIMITLESS_GPU_QUOTIENT=1 \
LIMITLESS_GPU_VORTEX=1 \
LIMITLESS_GPU_PROFILE=1 \
GOGC=500 \
GOMEMLIMIT=900GiB \
./bin/prover prove --config ... --in ... --out ...
```

`LIMITLESS_MERGE_JOBS=1` is conservative while the Koala Vortex path uses
per-device cached pipelines. More merge workers are only safe when they are
scheduled onto idle GPUs or when the merge path is CPU-only and cannot race the
same GPU pipeline.

## Incremental Vortex-Only Pipeline

The first stable proving target should keep quotient computation on CPU and
move only SIS-applied Vortex rounds to GPU:

```bash
LIMITLESS_GPU_PIPELINE=vortex-only \
LIMITLESS_GPU_COUNT=4 \
LIMITLESS_SUBPROVER_JOBS=2 \
LIMITLESS_MERGE_JOBS=1 \
LIMITLESS_GPU_PROFILE=1 \
LIMITLESS_GPU_VORTEX_RECOMMIT=1 \
GOGC=500 \
GOMEMLIMIT=900GiB \
./bin/prover prove --config ... --in ... --out ...
```

This deliberately leaves the quotient FFT, quotient coset FFTs, and quotient
symbolic evaluation on CPU. The machine has enough CPU to overlap that work
with GPU Vortex, and it gives a smaller correctness surface than enabling the
current full GPU quotient path.

The quotient code logs an FFT-placement estimate in this mode. Current
mainnet-limitless measurements consistently select CPU: moving only quotient
FFTs to GPU would require D2H of every re-evaluated root so the CPU expression
boards can still run, adding about 10-65 GiB of transfer on large boards while
the observed CPU FFT portions are usually milliseconds to a few seconds. The
bottleneck is quotient expression evaluation, not the FFT alone. GPU quotient
should therefore return only when the expression evaluation and annulator
scaling stay on GPU too.

## Device Affinity

Every GL/LPP goroutine must:

1. choose `deviceID = scheduleSlot % LIMITLESS_GPU_COUNT`;
2. call `runtime.LockOSThread`;
3. create or reuse `gpu.GetDeviceN(deviceID)`;
4. call `dev.Bind()` (`cudaSetDevice`);
5. register the device in `gpu.CurrentDevice`;
6. clear the thread-local device and unlock the OS thread on exit.

GPU dispatch code must use `gpu.CurrentDevice()`, not `gpu.GetDevice()`, so
quotient and Vortex stay on the segment's assigned GPU.

## Vortex Design

Vortex is the largest inner-proof win and the highest OOM risk.

The correct GPU shape is:

1. Upload raw committed rows to pinned host memory.
2. H2D only raw rows.
3. RS encode on GPU.
4. SIS hash on GPU.
5. Poseidon2 leaves and Merkle tree on GPU.
6. D2H only small outputs: Merkle root/tree, SIS hashes when self-recursion
   needs them, UAlpha, and opened columns.

The wrong shape is returning `EncodedMatrix` to Go. That forces full D2H,
multi-GiB Go heap allocations, SmartVector reconstruction, and GC pressure.

### Residency Modes

The protocol needs all committed rounds later for UAlpha and openings, but a
single cached CUDA pipeline is overwritten on every commit. A round handle must
therefore use one of these modes:

- `host`: CPU fallback, stores the host `EncodedMatrix`.
- `snapshot`: after commit, copy `d_encoded_col` to a per-round GPU buffer by
  D2D. Fast for UAlpha/open, but consumes `nRows * sizeCodeword * 4` bytes per
  SIS round.
- `recommit`: keep only the Merkle tree/root and raw column IDs. Re-run the GPU
  commit at UAlpha time and again at opening time. Slower, but bounded VRAM.

The default should be adaptive:

- if estimated per-segment SIS snapshots fit under
  `LIMITLESS_GPU_VORTEX_SNAPSHOT_BUDGET_GIB` (default 48 GiB), use snapshots;
- otherwise use recommit mode;
- `LIMITLESS_GPU_VORTEX_RECOMMIT=1` forces the low-memory mode;
- `LIMITLESS_GPU_VORTEX_SNAPSHOT=0` disables snapshots.

This is the clean compromise for production: small segments get the fastest
path; huge segments stop OOMing.

One important exception remains: self-recursed Vortex contexts must currently
use snapshots. `recursion.ExtractWitness` can materialize encoded matrices
from prover state before the open phase, so a pure recommit handle is not
sufficient there yet. Low-memory recommit is therefore enabled only for
non-self-recursed contexts until witness extraction learns a recompute path.

The shared CUDA pipeline must stay locked across commit and any immediate read
of `d_encoded_col` or the pinned Merkle/SIS buffers. Locking only the commit
call is insufficient: another same-device worker can otherwise overwrite the
pipeline before snapshot, UAlpha, or selected-column extraction completes.

### Vortex Cache Lifetime

`GPUVortex` pipelines cache multi-GiB buffers:

- raw row H2D workspace;
- encoded matrix;
- SIS output;
- Poseidon leaves;
- Merkle tree;
- pinned input/tree/SIS buffers.

Pipelines are keyed by `(deviceID, nCols, maxNRows, rate)`. The cache must be
evicted per device after `OpenSelectedColumns`, when the segment has no further
use for that level's pipeline. Evicting from quotient code is unsafe because
another goroutine sharing the device may still be inside a Vortex call.

Eviction must also wait for the per-pipeline mutex before freeing CUDA buffers.
The map entry can be removed first, but `kb_vortex_pipeline_free` must not run
while `CommitDirectAndThen` is still reading pinned SIS hashes, the Merkle tree,
or `d_encoded_col`.

## Quotient Design

`globalcs.QuotientCtx.Run` should offload when
`LIMITLESS_GPU_QUOTIENT=1`:

1. compile symbolic boards into GPU bytecode;
2. pack base roots into a reused pinned buffer;
3. H2D once per ratio group;
4. batch IFFT roots on GPU;
5. batch coset FFT per coset on GPU;
6. run symbolic evaluation on GPU;
7. assign quotient shares back to Wizard.

Current bottlenecks to keep removing:

- host-side extension-field SoA marshaling;
- host-built `variables.X` and `PeriodicSample` vectors;
- D2H of full E4 symbolic results before annulator scaling;
- serial H2D then IFFT where streams can overlap.

The next clean version should cache device-side deterministic auxiliary vectors
per `(deviceID, domainSize, maxRatio, coset, variable)` and should move
annulator scaling into the GPU eval output path.

## PlonK Design

The BLS12-377 GPU PlonK work should be integrated in two places:

1. final BLS conglomeration proof;
2. outer execution proof.

The existing `gpu/plonk` layer already has CUDA MSM/FFT infrastructure and
work-buffer pinning. The desired integration is an env-gated replacement for
`plonk.Prove` inside `circuits.ProveCheck`, using a GPU proving key converted
once from gnark's proving key and cached per setup digest.

Outer proof target: reduce the current 4-5 minute CPU outer proof to about
1-2 minutes. If the GPU PlonK implementation cannot yet prove the exact gnark
setup, it must remain behind `LIMITLESS_GPU_OUTER=1` and verify every proof
against the existing verifying key.

## Memory Budget

Per GPU budget on this host:

- 96 GiB physical VRAM;
- keep at least 8-12 GiB free for CUDA allocator fragmentation and transient
  kernels;
- keep Vortex pipeline + snapshots below about 80 GiB;
- use recompute mode when snapshots would exceed the budget.

Host RAM budget:

- `GOMEMLIMIT=900GiB`;
- avoid full encoded matrix D2H;
- keep compiled segment objects released as soon as their last witness using
  that module is done;
- run GC between GL and LPP, and after LPP before the long conglomeration tail;
- keep scratch data on NVMe when using serialized witnesses.

Pinned host memory must be capped and released at phase boundaries. It is not
normal Go heap and can starve the OS if treated as a cache with no limit.

## Expected Critical Path

The desired overlap is:

```text
CPU: compile segments + setup load + trace/bootstrapper
GPU:                         idle or prewarm

CPU: feed GL witnesses, verify, serialize proofs, conglomerate ready pairs
GPU: GL quotient + Vortex, one segment per GPU

CPU: derive shared randomness, release GL, feed LPP, conglomerate
GPU: LPP quotient + Vortex, one segment per GPU

CPU: final response assembly, proof verification, residual reductions
GPU: final BLS conglomeration proof, then outer execution proof
```

The main performance target is not peak utilization inside a single kernel. It
is keeping all four GPUs busy with segment work while CPU-side compilation and
conglomeration proceed without causing heap or VRAM spikes.

## Current Findings

- The Makefile's `bin/prover` target currently builds with `debug` only. A GPU
  binary must include the `cuda` tag, otherwise all GPU code is compiled out.
- Multi-GPU Vortex requires explicit `cudaSetDevice` on the locked OS thread.
  Without it, work intended for devices 1-3 silently lands on device 0.
- The Vortex drop-in API that returns host `EncodedMatrix` is too expensive for
  production sizes. Device-resident handles are mandatory.
- Per-round Vortex handles cannot point at the shared cached pipeline unless
  no later commit can overwrite it. They must snapshot or recommit.
- Evicting the Vortex cache from quotient code is unsafe under concurrent
  workers. Evict at the end of the owning segment's Vortex open step.
- The quotient GPU path is correct on focused tests but still has host-side
  overhead and large-board fallback. Keep it off for the vortex-only baseline;
  re-enable it only when expression evaluation and annulator scaling stay on
  GPU, not for GPU-only FFTs.
- The first four-GPU end-to-end attempt failed after 8m23s and peaked at about
  963 GiB RSS. It did not produce a proof. The GPU trace showed 18 quotient
  phases totalling 362.8s and 30 Vortex SIS commits totalling 68.5s.
- The vortex-only end-to-end attempt with CPU quotient and GPU Vortex failed
  after 9m17s in `CommitState.ExtractSISHashes`, not in quotient code. The
  crash exposed a Vortex cache lifetime bug: eviction could free a cached
  pipeline while another goroutine was still inside `CommitDirectAndThen`.
  Focused CUDA tests now cover this race.
- In vortex-only mode, the quotient FFT-placement estimator selects CPU for
  the current hybrid. GPU-only FFTs would add large H2D/D2H traffic and compete
  with Vortex VRAM while leaving the dominant quotient expression evaluation on
  CPU.
- The worst quotient calls were not GPU-kernel-bound. Large expression boards
  exceeded the current symbolic-eval slot limit and fell back to CPU, producing
  single calls of 169.9s and 97.9s. This must be fixed with chunked or
  rescheduled GPU symbolic evaluation; raising the slot limit would put huge
  per-thread local arrays in CUDA and is the wrong architecture.
- `LIMITLESS_GPU_VORTEX_RECOMMIT=1` still produced 30 Vortex snapshots because
  self-recursed contexts require encoded matrices for witness extraction. The
  next Vortex fix must make recursion witness extraction work from recomputed
  or sliced encoded data so self-recursed contexts can use low-memory mode.

## Validation Gates

Each change must pass in this order:

1. rebuild CUDA artifacts when `.cu`, `.cuh`, or headers change:
   `cmake --build gpu/cuda/build --target gnark_gpu -j`;
2. focused CUDA tests:
   `go test -tags cuda ./gpu/vortex ./gpu/quotient`;
3. vortex-only protocol tests:
   `LIMITLESS_GPU_PIPELINE=vortex-only LIMITLESS_GPU_COUNT=4 go test -tags cuda ./protocol/compiler/vortex ./protocol/compiler/globalcs`;
4. representative benchmarks:
   `go test -tags cuda -bench 'BenchmarkCommitGPUResident|BenchmarkGPUQuotient' ./gpu/vortex ./protocol/compiler/globalcs`;
5. end-to-end proof with `LIMITLESS_GPU_PROFILE=1`;
6. inspect GPU JSONL, wall clock, host RSS, and per-GPU VRAM peaks.

Proof validity is non-negotiable: any GPU proof path must keep the existing
Wizard and PlonK verification checks enabled.

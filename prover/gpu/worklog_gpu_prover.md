# GPU Limitless Prover — Worklog

Started: 2026-04-27. Owner: gautam.botrel.

Hardware (g7e.24xlarge, current):
- 4× NVIDIA RTX PRO 6000 Blackwell Server Edition, 97,887 MiB VRAM each,
  driver 590.48.01, CUDA 13.1.
- 96 vCPUs, 999 GiB host RAM.
- No swap is currently configured. The old g7e.12xlarge notes below are
  historical and should not drive new scheduling decisions.

Reference run (CPU-only, r8a.24xlarge, ~700 GiB RAM): ~10–15 min inner proof
plus ~4–5 min outer proof = 15–20 min total. Goal: at least 4× on the inner
proof using all four GPUs, plus a meaningful cut on the outer.

Current goal: use all four GPUs and the full 1 TiB host RAM budget to get the
end-to-end proof below 5 minutes. CPU-only reference runs may fit in memory on
this instance, but they are not the target operating mode.

## Method

Don't guess. Measure first, optimize the actual bottleneck, validate against existing reference results, document each step (success and failure). Every step lands behind an env-var or build tag; no silent default changes.

---

## Inventory of the prover

```
limitless.Prove (backend/execution/limitless/prove.go)
├── RunBootstrapper                     CPU-bound; segments witness, dumps to /tmp/witnesses
├── GL phase  (numConcurrentSubProverJobs=4, ~N_GL segments)
│   └── RunGL → compiledGL.ProveSegmentKoala
│       └── proveSegment (protocol/distributed/segment_compilation.go:394)
│           ├── wizard.RunProverUntilRound (inner Koala-Vortex IOP)
│           │   ├── globalcs.QuotientCtx.Run    ← FFT + symbolic eval        (CPU hot)
│           │   ├── vortex.ColumnAssignment     ← RS-encode + Merkle (+SIS)  (CPU hot)
│           │   ├── vortex.LinearComb           ← α-LC over committed rows   (CPU hot)
│           │   └── vortex.OpenSelectedColumns  ← column extract + Merkle p. (CPU)
│           └── wizard.RunProverUntilRound on recursionCompKoala (smaller, same shape)
├── LPP phase (same shape, ~N_LPP segments, runs after GL)
├── Conglomeration  (hierarchical merge; final merge uses BLS12-377)
│   └── cong.ProveSegmentKoala / ProveSegmentBLS  (PlonkInWizard inside)
└── execCirc.MakeProof                          ← gnark plonk over BLS12-377
```

Compiled module sizes on disk (mmap-backed): GL 7–15 GiB, LPP ~7 GiB each. Eighteen module types. With concurrency 4, working set inside the prover for live GL/LPP runs is roughly `4 × (compiled-mod + per-segment heap)`.

### Mapped CPU hot paths and the GPU prototype that would replace them

| CPU site | File:line | Existing GPU replacement | Status |
|---|---|---|---|
| Quotient: global iFFT, per-coset FFT, expression eval | `protocol/compiler/globalcs/quotient.go:199` (`QuotientCtx.Run`) | `gpu/quotient.RunGPU` (+ `gpu/symbolic`) | API ready, **not wired**. CPU fallback already inside `RunGPU` for boards >8192 slots. |
| Vortex per-round commit (RS encode + Poseidon2 Merkle, optional SIS) | `protocol/compiler/vortex/prover.go:72` (`ColumnAssignmentProverAction.Run`), calls `vortex_koalabear.CommitMerkleWith[out]SIS` | `gpu/vortex.GPUVortex.Commit` returning `*CommitState` | API ready, **not wired**. |
| Vortex linear combination | `protocol/compiler/vortex/prover.go:158` (`LinearCombinationComputationProverAction.Run`) | `(*CommitState).LinComb(α)` | API ready, **not wired**. |
| Vortex column open + Merkle proofs | `protocol/compiler/vortex/prover.go:264` (`OpenSelectedColumnsProverAction.Run`) | `(*CommitState).ExtractColumns(cols)` + Merkle proofs (still CPU; Merkle proofs are small) | Partial — column extract on GPU; Merkle proof gen still CPU. |
| Final BLS conglomeration merge (gnark plonk inside PlonkInWizard) | `RecursedSegmentCompilation.ProveSegmentBLS` → recursion stage `RunProverUntilRound(forBLS=true)` | `gpu/plonk` (BLS12-377) | API present, **not wired**. |
| Outer execution proof | `circuits/execution/circuit.go:138` (`circuits.ProveCheck` → gnark `plonk.Prove`) | `gpu/plonk` | Same. |

### Observations about the GPU prototype layer

- Singleton device: `gpu.GetDevice()` is `sync.Once` and always returns device 0 (`gpu/singleton.go:21`). The Go wrapper has no multi-device API. The C layer in `gpu/cuda/src/plonk/api.cu` already calls `cudaSetDevice(...)` per entrypoint, so multi-context is feasible — the gap is on the Go side.
- `gpu/quotient.RunGPU` is signature-compatible with what `QuotientCtx.Run` already has: `(domainSize, ratios, boards, rootsForRatio, shiftedForRatio, quotientShares, constraintsByRatio)`. It compiles boards via `BoardToNodeOps` and falls back to CPU per-board when slots > 8192.
- `gpu/vortex` already designed around a *device-resident* `CommitState`: it caches `GPUVortex` instances keyed by `(nCols, maxNRows, rate)` to avoid re-allocation between calls (`gpu/vortex/commit_merkle.go:38`). `EvictPipelineCache` is the explicit release. This is exactly the lifecycle the per-segment prover needs.
- `gpu/plonk/WORKLOG.md` shows the BLS plonk prover is mature (1.28×–1.37× MSM speedup on RTX PRO 6000 already landed; full PlonK over n=2²⁷ runs in ~78s). The 2²⁷-domain footprint is ~96 GiB pinned host + ~30+ GiB VRAM transient. With 96 GiB VRAM per GPU, this fits.

### Memory + IO realities on this box

- 499 GiB host RAM, 192 GiB total VRAM. The CPU baseline (700–800 GiB) does not fit. Three levers:
  1. Move encoded matrices out of Go heap onto GPU (largest single saving — Vortex commitments at every round).
  2. Use mmap for compiled circuits and witnesses (already done).
  3. Cap `numConcurrentSubProverJobs` to match GPU concurrency (= 2), which reduces simultaneous in-flight heap by 2×.
- `/tmp/witnesses` lives on `/` (130 GB free). Witnesses can run several tens of GB. **Mount nvme1n1 (3.5 TB) and redirect `witnessDir`.**
- The 2 GPUs are pure compute peers — no NVLink between them. Cross-GPU traffic = host RAM. Stick to per-segment GPU pinning, no cross-device hot loops.

---

## The plan

Each phase ends with a numeric measurement and a correctness validation. Land each phase on its own commit; revert if the validation regresses.

### Phase 0 — environment + baselining

**0.1** Mount nvme1n1 at `/scratch` (ext4, mounted read-write). Repoint `witnessDir` to `/scratch/witnesses` and dump intermediate artifacts (perf JSONL, profile traces) under `/scratch/runs/`. Document the mount step in this worklog (one-time per machine setup).

**0.2** Add per-phase GPU instrumentation to the JSONL `perfLogger` (`backend/execution/limitless/perf_log.go`):
  - Enrich existing `phase`/`job` events with `vram_used_mb` and `vram_peak_mb` per device.
  - Emit `gpu_event` records inside `proveSegment` with `{phase, device_id, ms}` for: quotient, vortex_commit, vortex_lincomb, vortex_open, recursion_quotient, recursion_vortex.
  - Behind `LIMITLESS_GPU_PROFILE=1`. Default off.

**0.3** Capture a *partial* CPU baseline. We can't run end-to-end CPU on this box. Instead:
  - Run a single GL segment for one small module (TINY-STUFFS) all-CPU. Record per-action timings.
  - Run a single LPP segment for one small module all-CPU.
  - Persist as `gpu/baselines/cpu_single_segment.json`. This is what we'll compare individual GPU swaps against.

**0.4** Capture an end-to-end timing reference *with the existing 1-GPU prototype unwired*. The current binary will run CPU-only and likely OOM. Record the OOM point (which phase, which heap) so we know which Phase-1 swap matters most for memory.

**Exit criterion:** JSONL emits per-phase GPU timings (still 0 ms because nothing is wired); single-segment CPU timings recorded; OOM phase identified.

### Phase 1 — single-GPU correctness

Goal: end-to-end proof on this box, GPU-accelerated, on **one** GPU only. Validation: proof verifies (this is enforced by `wizard.VerifyUntilRound` inside `proveSegment`).

**1.1** Add `gpu.GetDeviceN(id int) *Device` and a small device pool (slice of `*Device`, populated lazily via `New(WithDeviceID(id))`). Keep `GetDevice()` as `GetDeviceN(0)`. No behaviour change for existing callers. (`gpu/singleton.go`, ~30 LoC.)

**1.2** Wire `gpu/quotient.RunGPU` into `globalcs.QuotientCtx.Run`. Behind env var `LIMITLESS_GPU_QUOTIENT=1`. Steps:
  - At entry: if env on and `gpu.GetDevice() != nil`, call `quotient.RunGPU(...)` with the same arguments the CPU path computes from `ctx`.
  - On error: log + fall back to CPU path.
  - Validation: a single segment proof verifies, both with and without the env var, and `VerifyUntilRound` succeeds.
  - Microbench: time the `quotient` JSONL phase before/after on TINY-STUFFS, ARITH-OPS, BLS-PAIRING (3 representative module sizes).

**1.3** Wire `gpu/vortex` into `protocol/compiler/vortex/prover.go`. Behind env var `LIMITLESS_GPU_VORTEX=1`. Three actions to swap:
  - **ColumnAssignment**: replace `vortex_koalabear.CommitMerkleWithSIS/WithoutSIS` with `GPUVortex.Commit(rows)` returning `(*CommitState, root)`. Store the `*CommitState` in `run.State` under the same key (`VortexProverStateName(round)`). The downstream actions read it back.
  - **LinearCombination**: replace the CPU `vortex.LinearCombination` with `(*CommitState).LinComb(α)` accumulating across all committed states for the round set. Where `committedSV` mixes GPU `*CommitState` and host `vortex_koalabear.EncodedMatrix` (e.g. precomputeds), use the existing hybrid path documented in `MEMORY.md`: GPU lincomb on the GPU-resident states + host lincomb on the rest, then `α^offset · gpu_part + host_part`.
  - **OpenSelectedColumns**: pull selected columns via `(*CommitState).ExtractColumns(cols)` (small D2H), then run the existing CPU Merkle-proof generator over the host trees.
  - Free `*CommitState`s right after `OpenSelectedColumns` and call `gpuvortex.EvictPipelineCache()` between recursion levels (per `MEMORY.md` notes).
  - Validation: proof verifies with the env var on; `VerifyUntilRound` exact match.
  - Microbench: per-round commit timings before/after.

**1.4** Run the *full* limitless prover end-to-end on this machine with **both** env vars on, GPU 0 only, `numConcurrentSubProverJobs=2`. Expect:
  - Inner proof time drops materially.
  - Heap peak drops because encoded matrices are no longer on the Go heap.
  - Possibly still slow (1 GPU) — that's fine; correctness first.
  - **Validate**: proof verifies (the binary verifies internally; the run completes without panic).

**Exit criterion:** end-to-end prove passes on this hardware with one GPU, with measured per-segment timings and heap profile.

### Phase 2 — multi-GPU

**2.1** Pin each GL/LPP segment goroutine to a GPU. The GL/LPP goroutines are already spawned via `errgroup.Go`. Add a `GPUAffinity` struct passed to `RunGL`/`RunLPP`:
  - Goroutine `i` gets device `i % nGPUs`.
  - The goroutine sets `runtime.LockOSThread` and calls `cudaSetDevice(id)` once at entry (via a thin Go wrapper).
  - Replace any direct `gpu.GetDevice()` call inside `proveSegment` (and quotient/vortex hooks) with a context-carried `*gpu.Device`. Add a `proveSegment` parameter or a `context.Context` value.

**2.2** Raise `numConcurrentSubProverJobs` to 2 (= number of GPUs). Each GPU has 96 GiB; even at 2²⁶-domain Vortex commits it fits. We do not raise to 4 because each segment also needs CPU + host RAM headroom.

**2.3** Validate:
  - End-to-end proof verifies.
  - Both `nvidia-smi dmon` lines show ≥80% utilization during GL phase.
  - GL+LPP wall-clock approximately halves vs Phase 1.4.

**2.4** Confirm conglomeration is not the bottleneck. The hierarchical merge runs concurrently with proving (`numConcurrentMergeJobs=4`), and merges currently use CPU plonk. After Phase 1 the inner proves are faster, so merges may now be the wall-clock long pole. If so, Phase 4 below is the next mover.

**Exit criterion:** segment-phase time approximately halved vs single-GPU; both GPUs saturated; proof verifies.

### Phase 3 — host memory orchestration

After Phase 2, the segment hot paths are on GPU. The remaining host pressure comes from witness mmap + compiled-module mmap + intermediate Go-heap buffers around quotient and vortex.

**3.1** Profile with `runtime/pprof` heap snapshots taken at `glPhaseStart`, mid-GL, post-GL GC, post-LPP GC. Stash under `/scratch/runs/<ts>/heap-*.pprof`. Drive the heap budget to ≤200 GiB peak (leaving 100 GiB system slack on a 499 GiB box, with the rest budgeted for OS + pinned).

**3.2** Eliminate the duplicated CPU witness rows once they're on GPU:
  - In the rewired Vortex ColumnAssignment, after `Commit` succeeds, call `run.Columns.TryDel(colName)` for every committed column (mirrors what the existing `OpenSelectedColumnsProverAction.Run` already does at line 341-344). Move that delete into a hook right after Commit so the host buffer is freed immediately.

**3.3** Cap pinned host memory. `gpu/vortex.AllocPinned` and `gpu/plonk` pinned buffers can grow unbounded. Add an env-var-controlled cap (`LIMITLESS_PINNED_GIB=64`); when exceeded, force unpinned fallback for new allocations.

**3.4** Validate that peak heap fits and proof still verifies. Re-run end-to-end and snapshot heap peak.

**Exit criterion:** heap peak ≤ 200 GiB across the full run; proof verifies; no thrashing.

### Phase 4 — BLS plonk on GPU (final conglomeration + outer)

The hierarchical conglomeration's final merge produces a BLS12-377 plonk proof; `execCirc.MakeProof` produces another. Both use gnark's CPU `plonk.Prove`. The user reports outer is ~4–5 min today.

**4.1** Inspect how `cong.ProveSegmentBLS` and `circuits.ProveCheck` reach `plonk.Prove`. Identify the gnark `plonk.Prove` call site (likely under `protocol/compiler/plonkinwizard` for the conglomeration BLS path; gnark library for the outer). Decide where to inject the GPU plonk path:
  - Option A: replace `plonk.Prove` calls with `gpu/plonk.Prove` (requires building a `*GPUProvingKey` from the existing `plonk.ProvingKey`).
  - Option B: implement the small subset of `gnark` plonk-prover hooks needed to use GPU MSM/FFT, and keep the rest CPU.

**4.2** Implement Option A first — the gnark interface is stable, and `gpu/plonk` already produces a verifying proof (per its WORKLOG end-to-end at n=2²⁷ in 78s). Behind `LIMITLESS_GPU_OUTER=1`. Do not regress correctness — verify the outer proof against the verifying key.

**4.3** Run the final BLS conglomeration on GPU (whichever GPU is least busy at end-of-segments; GPU 1 by convention since GPU 0 typically owns more segments due to pinning skew).

**4.4** Run the outer `execCirc.MakeProof` on GPU once segments + conglomeration finish.

**Exit criterion:** outer + final-conglomeration combined wall-clock ≤ 2 min (from 4–5 min); proof verifies.

### Phase 5 — pipeline & overlap

**5.1** Overlap conglomeration with segments. The current `RunConglomerationHierarchical` already runs concurrently with proving but uses CPU plonk; with the 4.x BLS GPU prover we can hand each merge a GPU when it becomes free. Schedule policy: any free GPU picks up the next merge as soon as a pair of proofs is ready.

**5.2** Overlap the outer-proof setup load (`circuits.LoadSetup`, started in a background goroutine in `prove.go:95-99`) with the GPU work — already done by the existing code, but verify the GPU work doesn't starve it.

**5.3** Optional: batch identical-shape Vortex commits across GL segments for the same module type. A few module types (HUB-A, ARITH-OPS) appear multiple times in segment lists; same-shape encoded-matrix commits could share `GPUVortex` device buffers across rapid-fire calls (the cache in `commit_merkle.go:38` already does this, but verify it survives the multi-GPU pinning).

**Exit criterion:** wall-clock reduction beyond the naive 2× from doubling GPU count alone.

### Phase 6 — micro-tuning

Only if Phase 5 leaves headroom against the ≥4× target.

- Batched-affine MSM in BLS plonk (already documented as multi-day work in `gpu/plonk/WORKLOG.md` Step 7).
- Quotient `BatchInvert` chunk size revalidated for our segment domains (different from gpu/plonk's 2²⁷ regime).
- Symbolic eval bytecode improvements (loop unroll, slot reuse).

---

## Risk register

| Risk | Likelihood | Mitigation |
|---|---|---|
| GPU-resident `CommitState` doesn't survive the round-by-round flow inside Wizard (state map invalidated) | Medium | Phase 1.3 stores it under existing `VortexProverStateName(round)`; if that breaks, fall back to GPU snapshot per round (see `MEMORY.md` `SnapshotEncoded`). |
| Boards in segments exceed 8192 slot limit → CPU fallback dominates | Medium | Measure per-segment fallback rate in 1.2; if high, raise `MaxGPUSlots` or add bytecode chunking. |
| Multi-GPU goroutine pinning fights Go scheduler | Low | `runtime.LockOSThread` per goroutine isolates the cgo CUDA context. Document in code. |
| BLS plonk GPU path has correctness gaps not yet caught by existing tests | Medium | Phase 4 keeps env-var-gated CPU fallback; outer proof verifies on every run. |
| 200 GiB heap budget unreachable | Medium | Phase 3 has explicit pprof-driven loop; if we can't cap, cut concurrency to 1 (per-GPU serialized) and lean harder on cross-segment overlap. |
| `/scratch` not mounted | Low | Phase 0.1 is a hard prereq — the run won't fit on `/`. |

## Rollback discipline

- Every phase is a single PR / single commit pair. The prior commit is the rollback target.
- Each env var defaults off. Removing a phase = clearing the env var, not reverting code (until we promote a phase to default after several green runs).
- After a phase completes, the worklog gets a "Step N — DONE" section with measurements, decisions, and any reverted micro-experiments (mirroring the discipline of `gpu/plonk/WORKLOG.md`).

---

## Step 0.1 — mount /scratch and redirect witnessDir — DONE

Mounted nvme1n1 (3.5 TB instance store) at `/scratch`, persisted in `/etc/fstab`:

```
sudo mkfs.ext4 -F -L scratch /dev/nvme1n1
sudo mkdir -p /scratch && sudo mount /dev/nvme1n1 /scratch
sudo chown ubuntu:ubuntu /scratch
echo '/dev/nvme1n1 /scratch ext4 defaults,nofail 0 2' | sudo tee -a /etc/fstab
mkdir -p /scratch/{witnesses,runs,heaps}
```

`df -h /scratch` reports 3.3 TB free.

Code change: `backend/execution/limitless/prove.go` now reads `$LIMITLESS_WITNESS_DIR`
(falls back to `/tmp/witnesses`) and `$LIMITLESS_SUB_PROVER_JOBS`
(falls back to 4) at package init. Both go-builds clean. Run command for
this machine:

```
LIMITLESS_WITNESS_DIR=/scratch/witnesses \
LIMITLESS_SUB_PROVER_JOBS=2 \
GOGC=500 GOMEMLIMIT=450GiB \
./bin/prover prove --config ... --in ... --out ...
```

Note on instance-store volatility: nvme1n1 is reset on EC2 stop/start, so the
fstab entry uses `nofail` and Phase 0.1 must be re-run after a stop/start.

## Step 0.2 — GPU per-phase tracer — DONE

`gpu/trace.go` provides `gpu.TraceTime`, `gpu.TraceEvent`, `gpu.TraceEnabled`,
`gpu.TraceClose`. JSONL records are emitted to
`$LIMITLESS_GPU_PROFILE_PATH` (defaulting to
`$LIMITLESS_GPU_PROFILE_DIR/gpu_profile_<ts>.jsonl`, in turn defaulting to
`/scratch/runs`) when `LIMITLESS_GPU_PROFILE=1`. Otherwise calls are an
atomic-load no-op. `limitless.Prove` defers `gpu.TraceClose()`.

## Step 1.1 — multi-device Go API — DONE

`gpu.GetDeviceN(id) *Device` returns a per-id lazy singleton (id==0 reuses
the existing default singleton). `gpu.DeviceCount()` reads
`$LIMITLESS_GPU_COUNT` (default 1). Both gpu.Enabled false → safe nil/0.
Builds clean with and without the cuda tag.

## Step 1.2 — wire gpu/quotient.RunGPU — DONE

`globalcs.QuotientCtx.Run` now dispatches to `gpu/quotient.RunGPU` when
`LIMITLESS_GPU_QUOTIENT=1` and a GPU is available. CPU body renamed to
`runCPU` and runs on any GPU error or when the env var is unset.

`gpu.TraceEvent("quotient", 0, dur, {domain, ok, error?})` records each
attempt in the GPU profile JSONL.

Validation:
- `go build ./protocol/compiler/globalcs/...` clean (with and without cuda tag).
- `go test -tags cuda -run TestGPUNTTCosetEval ./gpu/quotient/` PASS (6.2s).
- `go test -tags cuda ./protocol/compiler/globalcs/` PASS (CPU path, 5 subtests).
- `LIMITLESS_GPU_QUOTIENT=1 go test -tags cuda ./protocol/compiler/globalcs/`
  PASS (5 subtests). Trace shows boards going through GPU bytecode with
  `slotFallback=0` on these inputs.

## Step 1.3 — wire gpu/vortex SIS commit — DONE

`protocol/compiler/vortex/prover.go` `ColumnAssignmentProverAction.Run` now
dispatches the Koala SIS-applied commit (line 127 path) to
`gpuvortex.CommitMerkleWithSIS` when `LIMITLESS_GPU_VORTEX=1` and a GPU is
available. The function returns the same shape as the CPU API
(`(EncodedMatrix, Commitment, *Tree, []field.Element)`) so downstream actions
(LinComb, OpenSelectedColumns) are unchanged for now.

Scope held narrow on purpose:
- Only SIS-applied Koala rounds are GPU-routed. NoSIS rounds and BLS rounds
  remain CPU.
- Drop-in mode: encoded matrix is D2H'd, host stays in charge of LinComb/Open.
  The device-resident `*CommitState` path (which would also accelerate
  LinComb and column extraction) is deferred — it requires teaching the
  prover state how to hold a GPU handle and is best done after Phase 2 when
  the segment goroutine knows its device.

`gpu.TraceEvent("vortex_commit_sis", 0, dur, {round, rows, cols})` records
each call.

Validation:
- `go build ./protocol/compiler/vortex/...` clean (with and without cuda tag).
- `go test -tags cuda -run TestCompiler ./protocol/compiler/vortex/` PASS
  (CPU path).
- `LIMITLESS_GPU_VORTEX=1 go test -tags cuda -run TestCompiler ./protocol/compiler/vortex/`
  PASS — all 9 subtests including 3 SIS configurations. First call ~3 s
  (pipeline init); subsequent calls reuse the cached `GPUVortex` per
  `getOrCreateGPUVortex`.

## Step 1.4 — end-to-end single-GPU validation — IN PROGRESS

First two run attempts failed during disk-asset deserialization
(`decodeSeqItem: offset 7881299347898368 out of bounds`). Both 7.0.1 and
7.0.7 assets failed identically; the source between asset-build commit
`0904632ed2` and `origin/main` is unchanged for the `prover/` tree, so the
mismatch is local to this box (assets weren't rebuilt for this binary).

Resolution: merged `Consensys/linea-monorepo#2751` (`perf/limitless-onthefly`)
which adds `ProveOnTheFly` — compile the entire LimitlessZkEVM in memory at
prover start when `serialization=false`. No disk assets needed.

Merge details:
- `crypto/vortex/prover_common.go` kept on HEAD's version (#2898 perf is
  newer than PR #2751's base).
- `backend/execution/limitless/prove.go` taken from PR with my Phase 0 layer
  on top: env-driven `witnessDir`, `gpu.TraceClose()` defer in both
  `Prove` and `ProveOnTheFly`.
- Env var renamed `LIMITLESS_SUB_PROVER_JOBS` → `LIMITLESS_SUBPROVER_JOBS`
  to match the PR's `getEnvPositiveInt` helper.
- Synthetic `TestConglomerationBasic` (3 subtests, full GL+LPP+Conglo flow)
  passes after the refactor: 574s.

First run aborted near the end of compile when RSS reached 497 GiB on a
499 GiB box — production uses r8a.24xlarge with 768 GiB. Added 256 GiB of
NVMe-backed swap on `/scratch/swapfile`, set `vm.swappiness=10` so the
kernel only swaps under genuine pressure. Total virtual ~755 GiB.

```
sudo fallocate -l 256G /scratch/swapfile && sudo chmod 600 /scratch/swapfile
sudo mkswap /scratch/swapfile && sudo swapon /scratch/swapfile
sudo sysctl vm.swappiness=10
```

## Step 2 — multi-GPU goroutine pinning — IN PROGRESS

Goal: each GL/LPP segment goroutine pins itself to one of the two GPUs so
they run in parallel.

Implementation:
- `gpu/threadlocal.go`: `SetCurrentDevice(*Device)` / `CurrentDevice()` and
  ID-only counterparts, keyed by `unix.Gettid()`. Goroutines call
  `runtime.LockOSThread()` first so the tid is stable.
- `backend/execution/limitless/gpu_pinning.go`: `pinGPU(slotIdx)` /
  `unpinGPU()`. The slot index is the *schedule slot* (place in the GL/LPP
  order list), not the witness index — this gives a flat round-robin across
  GPUs regardless of how the round-robin reorders by module type.
- `prove_onthefly.go`: each segment goroutine wraps its body with
  `pinGPU(slotIdx); defer unpinGPU()`.
- `globalcs.QuotientCtx.Run` and `vortex.ColumnAssignmentProverAction.Run`
  now consult `gpu.CurrentDevice()` instead of always going to device 0.
- `gpu/vortex` pipeline cache key extended with `deviceID` so two GPUs
  don't alias each other's cached `*GPUVortex`.

## Step 3 — vortex bottleneck and device-resident fix — DONE (2026-04-28)

After the on-the-fly compile run started OOMing on this 499 GiB box (the
production peak is ~500 GiB), pulled back to validate that the GPU paths
actually deliver wins at the unit level before pushing more end-to-end.

### Drop-in API was slower than CPU at production sizes

`go test -tags cuda -bench BenchmarkCommitMerkleWithSIS_*` head-to-head
at four sizes:

| size | n_cols × n_rows | CPU (AVX-512) | GPU drop-in | speedup |
|---|---|---:|---:|---:|
| Small | 4096 × 256 | 9 ms | 6 ms | 1.38× |
| Typical | 16384 × 256 | 23 ms | 23 ms | 1.02× |
| MedLarge | 65536 × 1024 | 105 ms | 136 ms | 0.77× |
| Large | 524288 × 2048 | 1199 ms | 1853 ms | **0.65×** |

Per-phase CUDA event breakdown @ Large:

```
H2D + RS encode      95.8 ms
SIS hash             47.8 ms
P2 MD hash           94.7 ms
Merkle + D2H + sync 173.6 ms
TOTAL on-device     411.9 ms
```

Wall-clock = 1853 ms ⇒ **~1442 ms of pure Go-side overhead**:
- `make([]koalabear.Element, 1024 × 524288)` = 8 GiB Go heap alloc + zero
- parallel copy from pinned host → Go-managed slice (8 GiB)
- 2048 × `smartvectors.NewRegular` to wrap rows
- `cloneSMTTreeFromRef` deep copy of Merkle tree
- `materializeRows` of input SmartVectors

The drop-in's contract (return `EncodedMatrix = []smartvectors.SmartVector`)
forced the full D2H + reconstruction. The kernel is fast; the API kills it.

### Device-resident path is fast

Same code, same matrix, but `gv.Commit(rows)` returns a `*CommitState`
that keeps the encoded matrix on device:

| size | CPU | GPU (resident) | speedup |
|---|---:|---:|---:|
| MedLarge | 100 ms | 22 ms | **4.5×** |
| Large | 1083 ms | 258 ms | **4.2×** |

The on-device kernel is identical to the drop-in's kernel — the speedup
came purely from skipping the Go-side reconstruction.

### Multi-GPU bug + fix

Initial 2-GPU bench gave 565 ms / 2 matrices = 1.05× over single. nvidia-smi
during run: GPU 0 = 29 GiB resident, GPU 1 = 585 MiB. **GPU 1 wasn't being
used.**

Root cause: `kb.cu` (vortex) never called `cudaSetDevice`. CUDA's "current
device" is per-OS-thread, so without an explicit set every alloc + launch
falls through to device 0 even when the caller "owns" device 1. The plonk
side already had `cudaSetDevice(ctx->device_id)` at every entry; vortex was
missing it.

Fix:
- `gnark_gpu_set_device(int)` C entry.
- `gpu.Device.Bind()` Go method.
- Called from `pinGPU(slot)` after `runtime.LockOSThread`, and from
  `NewGPUVortex` before any allocation.

After the fix:

|  | wall | per-matrix | speedup |
|---|---:|---:|---|
| 2-GPU concurrent | 498 ms / 2 | 249 ms | 1.64× over 1-GPU serial |
| 1-GPU single | 408 ms | 408 ms | — |

PCIe topology is `PIX` (both GPUs behind one switch on this box) — that
caps 2-GPU H2D bandwidth and prevents perfect 2× scaling. With 4× per-GPU
over CPU and ~1.6× from running on both GPUs, end-to-end vs the production
2-segment-concurrent CPU baseline is roughly **6× wall clock** at Large.

### Wired into the protocol compiler

`protocol/compiler/vortex/committed.go` introduces a `*committedHandle`
tagged union stored under `VortexProverStateName(round)`. Two variants:

- `host: vortex_koalabear.EncodedMatrix` — used for BLS rounds, NoSIS
  rounds, precomputeds, and SIS rounds when GPU is disabled.
- `gpu: *gpuvortex.CommitState` — used for SIS-applied Koala rounds when
  `LIMITLESS_GPU_VORTEX=1` and a GPU is bound to the calling goroutine.

Three actions updated:

- `ColumnAssignmentProverAction` calls `gpuvortex.CommitSIS` for the GPU
  path (no full D2H), wraps the returned `*CommitState` in a handle.
- `LinearCombinationComputationProverAction` computes the host portion
  via the existing parallel `vortex.LinearCombination`, then adds
  `α^(global_offset) · cs.LinComb(α)` for each GPU matrix. The GPU
  partial is a single device kernel; only the small UAlpha vector
  (sizeCodeword × E4) D2Hs.
- `OpenSelectedColumnsProverAction` extracts columns via
  `cs.ExtractColumns(entryList)` (small D2H of only the selected
  columns); Merkle proofs come from the host-side tree. Frees GPU
  buffers after extraction.

`asHandle()` accepts both `*committedHandle` and raw `EncodedMatrix` for
backward compatibility with callers that haven't been migrated.

### Validation

```
go test -tags cuda -run TestCompiler ./protocol/compiler/vortex/                 PASS
LIMITLESS_GPU_VORTEX=1 go test -tags cuda -run TestCompiler ./protocol/compiler/vortex/  PASS
```

All 9 mixed-config subtests pass (NoSIS, SIS, precomputed-NoSIS,
precomputed-SIS, mixed, empty round, multi-round). The end-to-end proof
verifies through `wizard.VerifyUntilRound` for all configurations.

### Microbench commands

```
LIMITLESS_GPU_COUNT=2 go test -tags cuda -bench '^BenchmarkCommitGPUResident' \
  -benchtime=3x ./gpu/vortex/

LIMITLESS_GPU_COUNT=2 go test -tags cuda -bench '^BenchmarkCommitGPUResidentVsCPU' \
  -benchtime=3x ./gpu/vortex/
```

### What's left for vortex

1. Per-segment end-to-end timing on a real (non-OOMing) workload. The
   on-the-fly compile peak is the blocker on this 499 GiB box; need to
   either swap-extend (256 GiB swap on /scratch is already there) or
   shrink the segment count.
2. Better cross-GPU PCIe utilisation: chunked H2D interleaving across
   the two GPUs to hide PIX-shared upstream contention.
3. Free the input pinned buffer faster (currently per-pipeline; would be
   nice to share one across GPUs).

## Step 4 — quotient bottleneck and pinned-buffer cache fix — DONE (2026-04-28)

Same exercise as Step 3, but for the quotient path.

### Validation

`TestGPUQuotientCorrectness` (new, in `protocol/compiler/globalcs/`)
proves a Fibonacci wizard at n=1024 and n=64K with both env-var on and
off. Both proofs verify. Correctness ✓.

### Microbench, before pinned-buffer cache

Two shapes, both at the wizard.Prove level so the comparison includes
all prover-step overhead (vortex CPU commit + the quotient swap):

```
BenchmarkGPUQuotient_*       — 1 Fibonacci-style constraint (thin)
BenchmarkGPUQuotientHeavy_*  — 16 base roots × 8 deg-2 mul-sub constraints
```

| size | shape  | CPU | GPU | speedup |
|---|---|---:|---:|---:|
| 64K | thin | 5 ms | 6 ms | 0.83× |
| 256K | thin | 11 ms | 15 ms | 0.71× |
| 1M | thin | 39 ms | 60 ms | 0.66× |
| 64K | heavy | 7 ms | 8 ms | 0.82× |
| 256K | heavy | 17 ms | 23 ms | 0.75× |
| 1M | heavy | 48 ms | 78 ms | 0.61× |

### Per-phase TIMING @ n=1M heavy (16 roots × 8 constraints)

```
pack         19.6 ms   ← cudaMallocHost on every call (NEW! 64 MiB pinned alloc)
h2d           1.4 ms
ifft          1.1 ms   GPU
cosetNTT      1.3 ms   GPU
symEval       6.1 ms total
                kernel: 2.7 ms  GPU
                post:   3.4 ms  host (ScalarMul of result by annulator)
```

GPU compute is fast (5 ms total kernel work) — comparable to CPU AVX-512
`board.Evaluate` on the same boards. **What kills it is the per-call
host-side overhead**, exactly the same shape as the vortex drop-in
problem from Step 3.

### Fix: pinned buffer cache

`gpu/vortex/pinned_cache.go` adds `GetPinned(deviceID, n)` /
`ReleasePinnedCache(deviceID)`. First call at a given (device, capacity)
pays the cudaMallocHost; subsequent calls reuse the cached buffer.
Keyed on (deviceID, capacity) so two GPUs each get their own pool.

`gpu/quotient/quotient.go` switches from `AllocPinned/FreePinned` to
`GetPinned`. Removes the per-call alloc cost.

### Microbench, after pinned-buffer cache

| size | shape | before | after |
|---|---|---:|---:|
| 64K | heavy | 0.82× | 0.82× (kernel-launch bound at this scale) |
| 256K | heavy | 0.75× | **1.19×** (first GPU win) |
| 1M | heavy | 0.61× | **0.85×** (warm calls; first call still cold) |

Per-phase TIMING after fix @ n=1M heavy, warm:

```
pack          1.4 ms   ← was 19.6 ms (-94%)
h2d           1.4 ms
ifft          1.1 ms
cosetNTT      1.3 ms
symEval       6.0 ms
TOTAL phases ~12 ms (vs CPU full Prove minus other-overhead ~10 ms)
```

The remaining gap at n=1M is dominated by:
- `post` 3.3 ms — host-side ScalarMul-by-annulator on a 1M-element
  vector. Trivially movable to GPU.
- `h2d` 1.4 ms running serial with `ifft` 1.1 ms — could overlap on
  separate streams.
- The synthetic constraint set: 1 board, 1 ratio, 2 cosets. Real
  segments have many ratios + many boards per ratio. The GPU symEval
  kernel scales linearly in board count *with* a fixed launch cost; CPU
  scales linearly in board count *without*. GPU should win clearly at
  ≥10–20 boards per ratio.

### Multi-GPU concurrent proving

`BenchmarkGPUQuotient2GPU_1M`: two goroutines, each pinned to one GPU,
each running `wizard.Prove` on the same compiled wizard.

```
Two proofs in parallel:    74.7 ms / op (37.4 ms per proof)
Single proof  (extrapolated): 59 ms (heavy 1M, after pinned cache)
Serial two proofs would be: 118 ms

Scaling: 1.58× over serial — confirms both GPUs do real work.
```

The 2-GPU bench is **bottlenecked by the *non-quotient* CPU prover steps**
(Vortex CPU commit + setup) which contend for the shared 48 cores; it's
not a clean GPU-only measurement. Real segments will benefit more
because more of their wall clock is on GPU.

### Bind() prerequisite

The 2-GPU bench is the second thing in the codebase that exercises
multi-GPU pinning end-to-end (after the vortex 2-GPU bench). The
`gpu.Device.Bind()` fix from Step 3 (cudaSetDevice per OS thread) was
non-optional — without it both goroutines silently land on device 0.

### Outstanding optimizations (planned, not yet landed)

1. **ScalarMul-by-annulator on GPU.** Saves ~3.3 ms/call at n=1M. Easy:
   the eval kernel already writes a device buffer; just call
   `kb_vec_scale` before D2H instead of host `vq.ScalarMul`.
2. **Stream H2D + IFFT.** Currently serial. Adding a transfer stream
   that runs while the previous chunk's IFFT executes saves the ~1.4 ms
   h2d time at n=1M.
3. **Cache device-side variables.X / PeriodicSample.** They're
   deterministic per (domainSize, ratio, cosetIdx); should be a hash-keyed
   device cache, not rebuilt per call. Currently free at small scale
   (`auxBuild=0s` in our bench because the synthetic constraint has none)
   but the real segments use them heavily.
4. **Multi-GPU within one segment.** Split independent ratio groups
   across the two GPUs when a segment has >=2 ratios. This is genuinely
   easy: pin one ratio loop to dev0, another to dev1, accumulate.
   Skipped for now because the cross-segment 2-GPU pinning (already
   in place via `pinGPU(slot)`) gives most of the win at the segment
   level.
5. **Move ext SoA marshaling to GPU.** The per-coset `extSOA :=
   make([]field.Element, len(rd.extIDs)*domainSize*4)` host-side
   transpose is also O(n) per coset; should be a GPU transpose kernel.

## Step 5 — full-GPU design and Vortex low-memory cleanup — DONE (2026-04-28)

`fullgpu.md` now captures the target architecture for the current
g7e.24xlarge host: four pinned segment workers, explicit CUDA device
affinity, GPU quotient and Vortex on every GL/LPP segment, conservative merge
parallelism while Vortex pipelines are cached, and BLS12-377 GPU PlonK as the
next required outer-proof integration.

### Build correction

`Makefile` now has `GO_BUILD_TAGS ?= debug` and a `bin/prover-cuda` target.
The GPU binary must be built with the `cuda` tag:

```
make bin/prover-cuda
# or: GO_BUILD_TAGS=debug,cuda make bin/prover
```

Bare `make bin/prover` still builds the debug-only binary unless
`GO_BUILD_TAGS` is provided.

### Vortex snapshot/recommit modes

The SIS Koala path now chooses between:

- snapshot mode: `CommitSIS` copies `d_encoded_col` into a per-round GPU
  buffer and downstream UAlpha/opening read the snapshot;
- recommit mode: `CommitSISRootOnly` stores only the Merkle tree/root, then
  `CommitSISLinComb` and `CommitSISExtractColumns` recommit from raw Wizard
  columns at the later phases.

Selection is adaptive:

- `LIMITLESS_GPU_VORTEX_SNAPSHOT_BUDGET_GIB` controls the default snapshot
  budget (48 GiB);
- `LIMITLESS_GPU_VORTEX_RECOMMIT=1` or `LIMITLESS_GPU_VORTEX_SNAPSHOT=0`
  forces low-memory recommit;
- self-recursed contexts still force snapshots because
  `recursion.ExtractWitness` can materialize encoded matrices from prover
  state before open.

### Pipeline race fix

`GPUVortex.CommitDirectAndThen` holds the cached pipeline mutex through commit
and the immediate dependent read. This matters for same-device concurrency:
locking only `CommitDirect` allowed another worker to overwrite
`d_encoded_col`, pinned SIS hashes, or the pinned Merkle tree before snapshot,
UAlpha, or selected-column extraction completed.

`CommitSIS`, `CommitSISRootOnly`, `CommitSISLinComb`, and
`CommitSISExtractColumns` now use that locked path.

### Hybrid UAlpha correctness fix

`LinearCombinationComputationProverAction.Run` now accumulates chunks with an
explicit global row offset in the protocol's real stack order:

1. NoSIS precomputeds
2. NoSIS committed rounds
3. SIS precomputeds
4. SIS committed rounds

That removes the old hybrid-ordering hazard where host SIS rows and GPU SIS
rows could be compacted separately and scaled by the wrong alpha powers.

### Validation

No `.cu`, `.cuh`, or CUDA headers changed in this step, so CUDA artifacts did
not need rebuilding.

```
go test ./protocol/compiler/vortex/ ./protocol/compiler/globalcs/                  PASS
LIMITLESS_GPU_COUNT=4 go test -tags cuda ./gpu/vortex ./gpu/quotient               PASS
LIMITLESS_GPU_VORTEX=1 LIMITLESS_GPU_QUOTIENT=1 LIMITLESS_GPU_COUNT=4 \
  go test -tags cuda ./protocol/compiler/vortex ./protocol/compiler/globalcs       PASS
LIMITLESS_GPU_VORTEX=1 LIMITLESS_GPU_VORTEX_RECOMMIT=1 LIMITLESS_GPU_COUNT=4 \
  go test -tags cuda -run TestCompiler ./protocol/compiler/vortex/                 PASS
make bin/prover-cuda                                                               PASS
gofmt -l <touched go files>                                                        PASS
golangci-lint run                                                                  PASS
```

The linter also required documenting intentional GPU-test deterministic RNG
usage and operator-supplied GPU trace paths with `nolint:gosec`.

### E2E attempt — failed after 8m23s

Command:

```
/usr/bin/time -v env \
  LIMITLESS_GPU_COUNT=4 \
  LIMITLESS_SUBPROVER_JOBS=4 \
  LIMITLESS_MERGE_JOBS=1 \
  LIMITLESS_GPU_QUOTIENT=1 \
  LIMITLESS_GPU_VORTEX=1 \
  LIMITLESS_GPU_PROFILE=1 \
  LIMITLESS_GPU_VORTEX_RECOMMIT=1 \
  GOGC=500 \
  GOMEMLIMIT=900GiB \
  ./bin/prover prove \
    --config prover-assets/7.0.1/config/config-mainnet-limitless.toml \
    --in /home/ubuntu/proverstuff/mainnet/execution/29994327-29994333-getZkProof.json \
    --out /tmp/proof.whatever
```

The provided config expected traces under
`/home/ubuntu/test_mainnet/traces/conflated`; for this local run that path was
linked to `/home/ubuntu/proverstuff/mainnet`.

Result:

- exit status: 2
- wall clock: 8m23.31s
- max RSS: 963,430,192 KiB
- GPU trace: `/scratch/runs/gpu_profile_20260428_030252.jsonl`
- proof output: not produced

The run progressed through real GPU quotient and Vortex work, but it missed
the target before failing. During the run VRAM peaked near the device limits
and the log reported `GPU vortex init failed` with a generic CUDA error before
the final crash in the recursive/PlonK solving path.

GPU trace summary:

```
quotient_count=18 total_ms=362757.426
vortex_count=30 total_ms=68453.521
snapshot_true=30
```

Worst GPU phases:

```
quotient device=2 domain=131072 ms=169886.233
quotient device=3 domain=262144 ms=97858.263
quotient device=0 domain=262144 ms=39199.799
quotient device=0 domain=262144 ms=28230.118
vortex_commit_sis device=3 rows=2048 cols=262144 ms=10304.936 snapshot=true
vortex_commit_sis device=0 rows=2304 cols=262144 ms=9855.681 snapshot=true
vortex_commit_sis device=2 rows=1536 cols=131072 ms=7375.865 snapshot=true
```

Immediate conclusions:

1. `LIMITLESS_GPU_VORTEX_RECOMMIT=1` is not enough for this case because
   self-recursed contexts currently force snapshots for correctness. That keeps
   30 snapshot buffers alive and pushes VRAM close to the 98 GiB/device limit.
   The next Vortex step is to make recursion witness extraction consume
   recomputed/sliced encoded matrices instead of requiring full snapshots.
2. Large quotient boards are the dominant wall-clock problem. The worst calls
   still report CPU fallback due to slot pressure (`maxSlots` above the
   current GPU board limit), making `symEval` minutes long despite GPU mode.
   The next quotient step is chunked GPU symbolic evaluation for large boards
   instead of falling back whole boards to CPU.
3. Host memory is also too close to the machine limit. At 963 GiB RSS, the
   current design leaves no margin for outer-proof work or transient solver
   allocations. The GPU path must release Vortex snapshots earlier and avoid
   duplicate host materialization of encoded matrices.

## 2026-04-28 — Step 6: Vortex-only baseline and focused tests

### Experiment shape

The incremental baseline is now env-gated as:

```bash
LIMITLESS_GPU_PIPELINE=vortex-only
```

In this mode Vortex uses the GPU path, but `globalcs.QuotientCtx.Run` stays on
CPU even when old `LIMITLESS_GPU_QUOTIENT` settings are present. The quotient
code logs an FFT-placement estimate before CPU quotient execution so we can
measure whether a GPU-only FFT handoff would be worth its transfers.

For the mainnet-limitless run I used:

```bash
/usr/bin/time -v env \
  LIMITLESS_GPU_PIPELINE=vortex-only \
  LIMITLESS_GPU_COUNT=4 \
  LIMITLESS_SUBPROVER_JOBS=2 \
  LIMITLESS_MERGE_JOBS=1 \
  LIMITLESS_GPU_PROFILE=1 \
  LIMITLESS_GPU_PROFILE_PATH=/scratch/runs/gpu_profile_vortex_only_2seg_20260428_031523.jsonl \
  LIMITLESS_GPU_VORTEX_RECOMMIT=1 \
  GOGC=500 \
  GOMEMLIMIT=900GiB \
  ./bin/prover prove \
    --config prover-assets/7.0.1/config/config-mainnet-limitless.toml \
    --in /home/ubuntu/proverstuff/mainnet/execution/29994327-29994333-getZkProof.json \
    --out /tmp/proof.vortex-only-2seg
```

Result:

- exit status: 2
- wall clock: 9m17.24s
- max RSS: 943,604,064 KiB
- proof output: not produced
- failure site: cgo SIGSEGV in `gpu/vortex.(*CommitState).ExtractSISHashes`
  while copying the pinned SIS hash buffer

This is better localized than the previous full-GPU attempt: quotient stayed on
CPU and was not the crash source. The crash exposed a Vortex cache lifetime bug.
`EvictPipelineCacheForDevice` could free a cached `GPUVortex` pipeline while
another goroutine was still inside `CommitDirectAndThen` and reading pinned SIS
hashes / Merkle buffers / `d_encoded_col`.

### FFT placement finding

The quotient estimator consistently chose CPU in this hybrid. Representative
logs from the run:

- domain 131072: estimated GPU-only FFT transfers 65.64 GiB, CPU quotient
  24.824s; ratio-4 FFT 3.834s, ratio-4 eval 14.381s.
- domain 262144: estimated GPU-only FFT transfers 10.02-31.51 GiB depending on
  the board; CPU quotient 3.7-7.8s.
- domain 524288: estimated GPU-only FFT transfers about 15.8-16.0 GiB; CPU
  quotient 4.2-5.6s.

Conclusion: GPU-only quotient FFT is not worth enabling while quotient
expression evaluation remains on CPU. It would add large H2D/D2H traffic and
VRAM pressure, because CPU eval still needs host SmartVectors for every
re-evaluated root.

### Fixes from the focused pass

- `LIMITLESS_GPU_PIPELINE=vortex-only` now explicitly enables GPU Vortex and
  disables GPU quotient.
- Conglomeration workers are pinned after sub-prover worker slots, so
  `LIMITLESS_SUBPROVER_JOBS=2` and `LIMITLESS_MERGE_JOBS=1` use a spare GPU for
  the merge worker instead of colliding with segment worker 0.
- `GPUVortex.Free` now takes the same per-pipeline mutex as commits.
- `EvictPipelineCache` and `EvictPipelineCacheForDevice` remove cache entries
  first, then free each victim only after its active commit/use section has
  returned.
- `CommitAndExtract` now also takes the per-pipeline mutex.
- `ExtractSISHashes` now copies into Go-managed memory instead of returning an
  untracked `mmap` slice. Production SIS slices can be GB-scale and must be
  visible to GC/GOMEMLIMIT.
- Added `TestEvictPipelineCacheForDeviceWaitsForActiveCommit`, a small CUDA
  regression test for the eviction race that caused the full-run crash.

No CUDA `.cu`, `.cuh`, or header files changed in this step, so no CUDA library
rebuild was needed. The Go CUDA binary still needs rebuilding after Go edits.

### Focused tests and benchmarks

The simpler test set for this stage is:

```bash
go test -tags cuda ./gpu/vortex -run \
  'TestEvictPipelineCacheForDeviceWaitsForActiveCommit|TestGPUEncodeAndSIS|TestGPUVortexCommit|TestGPUColumnExtraction' -count=1

go test -tags cuda ./gpu/vortex -count=1

LIMITLESS_GPU_PIPELINE=vortex-only LIMITLESS_GPU_COUNT=4 \
  go test -tags cuda ./protocol/compiler/vortex ./protocol/compiler/globalcs -count=1
```

Results:

```text
go test -tags cuda ./gpu/vortex -run '...' -count=1                       PASS 9.392s
go test -tags cuda ./gpu/vortex -count=1                                  PASS 8.933s
LIMITLESS_GPU_PIPELINE=vortex-only LIMITLESS_GPU_COUNT=4 \
  go test -tags cuda ./protocol/compiler/vortex ./protocol/compiler/globalcs -count=1
  ./protocol/compiler/vortex                                              PASS 20.485s
  ./protocol/compiler/globalcs                                            PASS 4.595s
```

A cheap Vortex benchmark run, still not the full prover:

```bash
go test -tags cuda ./gpu/vortex -run '^$' \
  -bench 'BenchmarkCommitGPUResident_MedLarge|BenchmarkCommitMerkleWithSIS_Typical' \
  -benchtime=1x -count=1
```

Results:

```text
BenchmarkCommitGPUResident_MedLarge-96                              65.000859ms/op 4129.72 MB/s
BenchmarkCommitMerkleWithSIS_Typical/GPU_CommitMerkleWithSIS-96     14.308475ms/op 1172.54 MB/s
BenchmarkCommitMerkleWithSIS_Typical/CPU_CommitMerkleWithSIS-96     25.620341ms/op  654.84 MB/s
BenchmarkCommitMerkleWithSIS_Typical/Speedup-96                     3.333 speedup_x
```

Post-change validation:

```text
go test ./protocol/compiler/vortex ./protocol/compiler/globalcs -count=1       PASS
make bin/prover-cuda                                                           PASS
gofmt -l <touched go files>                                                    PASS
golangci-lint run                                                              PASS
```

## 2026-04-28 - plonk2 curve-generic Fr and NTT foundation

Added `gpu/plonk2` as a clean foundation for generalizing the GPU PlonK stack
across BN254, BLS12-377, and BW6-761. The current `gpu/plonk` BLS12-377 prover
path remains untouched.

Implemented:

- curve-indexed plonk2 C ABI in `gpu/cuda/include/gnark_gpu.h`
- generic scalar-field CUDA kernels in `gpu/cuda/src/plonk2`
- Go wrappers in `gpu/plonk2`
- MSM memory planner for affine SW and the current BLS12-377 TE layouts
- package worklog and architecture notes in `gpu/plonk2/WORKLOG.md`

Validation:

```text
cmake --build gpu/cuda/build --target gnark_gpu -j2        PASS
go test ./gpu/plonk2                                      PASS 0.003s
go test -tags cuda ./gpu/plonk2 -count=1                  PASS 3.025s
```

Benchmarks for 1,048,576 scalar elements:

```text
BenchmarkFrVectorMul_CUDA/bn254-48          159766 ns/op
BenchmarkFrVectorMul_CUDA/bls12-377-48      178784 ns/op
BenchmarkFrVectorMul_CUDA/bw6-761-48        273116 ns/op
BenchmarkFFTForward_CUDA/bn254-48           615058 ns/op
BenchmarkFFTForward_CUDA/bls12-377-48       614738 ns/op
BenchmarkFFTForward_CUDA/bw6-761-48         972631 ns/op
```

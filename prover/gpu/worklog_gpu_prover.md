# GPU Limitless Prover — Worklog

Started: 2026-04-27. Owner: gautam.botrel.

Hardware (g7e.12xlarge):
- 2× NVIDIA RTX PRO 6000 Blackwell Server Edition, 96 GiB VRAM each, driver 590.48.01, CUDA 13.1.
- 48 vCPUs, 499 GiB host RAM.
- Storage: nvme0n1 (root, 1.1 TB, 88% used → 130 GB free); nvme1n1 (3.5 TB, **unmounted** — must mount before runs).

Reference run (CPU-only, r8a.24xlarge, ~700 GiB RAM): ~10–15 min inner proof + ~4–5 min outer proof = 15–20 min total. Goal: ≥4× on the inner proof using both GPUs, plus a meaningful cut on the outer.

This machine **cannot run the prover all-CPU** — peak host RAM is ~700–800 GiB. Every measurement here must run with GPU enabled and host pressure capped.

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

## Step 1.4 — end-to-end single-GPU validation — TODO

Next: run the actual limitless prover end-to-end with both env vars on, GPU
0 only, `LIMITLESS_SUB_PROVER_JOBS=2`. Capture JSONL phase timings, verify
proof passes, snapshot heap profile.

(Subsequent steps will be documented as they land.)

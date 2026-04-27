# GPU PlonK Optimization Worklog

Started: 2026-04-27. Hardware: RTX PRO 6000 Blackwell, CUDA 13.1.

## Method

Don't guess. Measure first, optimize the actual bottleneck, validate against existing reference results, document each step.

## Baseline (verified 2026-04-27)

```
FFT/256K   fwd 241µs    inv 252µs    coset 293µs
FFT/1M     fwd 531µs    inv 550µs    coset 689µs
FFT/4M     fwd 2040µs   inv 2122µs   coset 2820µs
MSM/16K    5.04 ms     (overhead-dominated)
MSM/64K    3.63 ms
MSM/256K   7.15 ms
MSM/1M     18.5 ms
ScaleByPow/4M  924µs (290 GB/s — near memory peak)
BatchInvert/1M 7.17 ms (4.7 GB/s — sequential prefix limited)
```

## Initial Hypothesis Check (against memory + back-of-envelope)

**GLV endomorphism revisit.** User correctly challenged the speedup claim:
- Without GLV (c=17, 253-bit): n × 15 windows = 15n assignments, 15 × 2^16 = 983K buckets.
- With GLV (c=17, 127-bit): 2n × 8 windows = 16n assignments, 8 × 2^16 = 524K buckets.

Phases scale with assignments (build_pairs, sort, accumulate) → +7%.
Reduce phase scales with buckets → -47%.
If reduce is ~20% of MSM, net = 0.20 × 0.47 - 0.07 ≈ +2% speedup. Marginal at best.

GLV is **not** the priority. The real bottleneck must be elsewhere.

**Memory-bandwidth check.** At n=1M:
- Scalar H2D: 32 MB (one-way) → ~1 ms at 30 GB/s registered DMA.
- Point loads in accumulate: each bucket entry loads 96 B from `d_points`. Avg bucket size = 1M/983K ≈ 1.04, so kernel reads ~1M × 96 B = 96 MB → ~50 µs at 1.9 TB/s D2D.
- Bucket writes: 983K × 192 B = 189 MB → ~100 µs.

Compute estimate: 1M × 9M (add) ≈ 9M field muls per phase. Each fp_mul ≈ 30 instr. At ~10 TFLOPS effective integer throughput, ~30 µs.

So accumulate phase should be ~150 µs at 1M, not the 8+ ms I sketched earlier. **My phase decomposition was wrong**. Measure properly before optimizing.

## Step 1 — Per-phase MSM profiling — DONE

Instrumentation: added `cudaEvent_t phase_event[10]` to MSMContext, recorded
across each stream phase, exposed via `gnark_gpu_msm_get_phase_timings` C API
and `MSM.LastPhaseTimings()` in Go. Validated with `BenchmarkMSMPhaseTimings`
on real SRS + scalar inputs.

### Results — phases in µs, average of 10 runs

| phase           | n=64K | n=256K | n=1M  | n=4M   |
|---              |---:   |---:    |---:   |---:    |
| h2d             |   157 |    588 |   614 |   2426 |
| build_pairs     |    28 |     91 |   286 |   1069 |
| sort            |    66 |    137 |   533 |   2386 |
| boundaries      |    10 |     18 |    63 |    237 |
| **accum_seq**   |  1181 |   3354 |  8928 |  29745 |
| accum_par       |     0 |      0 |     0 |    128 |
| reduce_partial  |   802 |    808 |  1670 |   1640 |
| reduce_finalize |   244 |    243 |   262 |    258 |
| d2h             |    13 |     12 |    12 |     13 |
| phase sum       |  2502 |   5252 | 12368 |  37902 |
| **wall**        |  3164 |   6039 | 14789 |  46127 |
| wall − sum      |   662 |    787 |  2421 |   8225 |

### What this tells us

1. **`accum_seq` is the dominant phase across all sizes** — 47% (64K), 56%
   (256K), 60% (1M), 64% (4M). Per-bucket sequential 9M-mult unified TE
   addition. Avg ~15 entries/bucket at n=1M, ~61 at n=4M.

2. **`reduce_partial` is the secondary cost** at small n — 25% of total at
   n=64K. Stable ~1.6 ms regardless of n (it scales with `total_buckets`,
   not `n`). At c=17 with 15 windows × 2¹⁶ = 983K buckets it's nearly fixed.

3. **`wall − sum` gap = 13–21% of wall.** This is host-side and CUDA-runtime
   overhead OUTSIDE the recorded GPU phases. Suspects:
   - `cudaHostRegister` / `cudaHostUnregister` of ~32n bytes scalars at n≥2²⁰
   - `msm_alloc_work_buffers` / `msm_free_work_buffers` per call (GBs of `cudaMalloc`/`cudaFree` every MSM!)
   - host Horner combination (15 windows × c=17 doublings + 14 adds in TE — ~120 µs)
   - host Montgomery R⁻¹ SCM (~250 µs)
   At n=4M the gap is 8.2 ms. Half a proof's MSMs at this size = ~30 ms wasted
   on alloc/free overhead alone.

4. `h2d` is non-trivial (5–8% of total) but is already at near-peak DMA
   bandwidth on the registered fast path. Not worth attacking.

5. `build_pairs`, `sort`, `boundaries`, `d2h` are all small. Skip.

### Hypothesis ranked by data, not guesses

**A. Persistent work buffers across MSM calls (Tier-2 lift).**
The prover does back-to-back commitments. Currently `msm_run_full` allocates
~few-GB sort buffers fresh and frees them on every call. Add an explicit
"keep buffers" mode and only free at quotient-phase boundary.
**Expected:** kill 5–10 ms of `wall − sum` gap per call → ~30–80 ms/proof.

**B. Replace 9M unified mixed-add with 7M precomputed-T mixed-add.**
Store points as (Y−X, Y+X, 2d·X·Y) = 144 bytes instead of (X, Y) = 96 bytes.
Saves 2 muls per add → ~22% on `accum_seq`.
**Expected at n=1M:** 22% × 8.9 ms ≈ 2 ms/MSM. **Tradeoff:** points memory
50% larger (12 GB → 18 GB at n=2²⁷) and DMA time grows correspondingly.
The ec.cuh comment says compact wins at 64M+ on memory bandwidth — but that
trade was made for accumulate-bound reasoning that we can now verify with
data. At n=4M (well below 64M), precomputed should be a net win.

**C. Compact `reduce_partial` time** — fixed ~1.6 ms cost is meaningful at
small n. Currently uses 9M extended-add for the running sum + the per-thread
HW prefix scan within block. Could use precomputed cumulative bucket sums or
better blocking. ~0.5–1 ms savings.

**D. Batched-affine bucket addition (sppark/cuZK style).**
Major refactor. Keeps unified semantics. Estimated 1.5–2× on `accum_seq`.
~3 day project. Saved for after A+B prove the kernel is malleable.

### Decision: do A first.

A is independent of formula changes, low-risk, easy to validate (existing
tests assert exact MSM output via reference results in msm_test.go). It's
also the single highest µs/effort ratio I can identify from the data.

After A lands cleanly, do B (precomputed point format). The combination
should give roughly 25–35% MSM speedup at production sizes.

## Step 2 — A: persistent work buffers across MSM calls — DONE

### Implementation

- `MSMContext.buffers_pinned` flag.
- `gnark_gpu_msm_pin_work_buffers` / `release_work_buffers` C API.
- `(*G1MSM).PinWorkBuffers()` / `ReleaseWorkBuffers()` Go API.
- `msm_run_full` skips `msm_unregister_host` + `msm_free_work_buffers` when
  pinned. Lazy alloc is already idempotent (no-op when buffers exist).

### Validation

`TestMSM*` correctness suite all passes with the change in place — same
reference results, no regression (full pre-existing 123 s test run).

### Bench results — `BenchmarkMSMPinned`, 10 iters

| n    | unpinned   | pinned     | savings   | speedup |
|---   |---:        |---:        |---:       |---:     |
| 64K  | 3 282 µs   | 2 634 µs   | -648 µs   | 1.25×   |
| 256K | 5 987 µs   | 5 370 µs   | -617 µs   | 1.11×   |
| 1M   | 14 981 µs  | 12 472 µs  | -2 509 µs | 1.20×   |
| 4M   | 46 789 µs  | 37 553 µs  | -9 236 µs | 1.25×   |

11–25% MSM speedup. The `wall − phase_sum` gap (alloc/free + host
register/unregister) is essentially eliminated.

**Caveat for prover integration.** Pinning holds VRAM constant. At n=2²⁷
sort buffers are ~50 GB — verify VRAM headroom before pinning, and always
call `ReleaseWorkBuffers()` before the quotient phase.

### Prover wiring

`prove.go`:
- After `NewG1MSMN` at instance setup → `PinWorkBuffers()`.
- Before quotient (around `OffloadPoints`) → `ReleaseWorkBuffers()`.
- After quotient (around `ReloadPoints`) → `PinWorkBuffers()` again.

Verified: `TestPlonkECMul{Basic,LazyPrepare,Negative,Concurrent}` all pass.
End-to-end `BenchmarkPlonkECMul353` (n=2²⁶, 48M constraints) completes
producing correct proofs: ~30 s per proof, of which the MSM phases
account for ~22% (LRO ≈ 1.7 s, Z ≈ 1.2 s, h1/h2/h3 ≈ 1.8 s, linPol ≈ 0.6 s,
batch open ≈ 1.5 s). Expected end-to-end gain ~3–5%.

## Step 3 — B: precomputed-T point format (7M mixed-add) — DONE

### Implementation

- `G1EdYZD` struct in ec.cuh (3 fp coords: y_minus_x, y_plus_x, two_d_xy).
- `ec_te_unified_mixed_add_yzd` (7M, drops on-the-fly `T_q = 2d·X·Y`).
- `ec_te_cnegate_yzd` branchless (swap y-x↔y+x; negate two_d_xy via fp_ccopy).
- `accumulate_buckets_kernel` + `accumulate_buckets_parallel_kernel` use new type.
- `MSMContext::d_points` typed `G1EdYZD*`; allocs use `sizeof(G1EdYZD)` (144 B).
- Go-side `g1TEPrecompPoint = [18]uint64` private type + `precompFromCompact`
  helper. `ConvertG1Points`, `PinG1TEPoints` allocate pinned buffers at the
  precomputed (144 B) size and convert during the copy.
- `ReadG1TEPointsPinned` streams 64K-point batches from the (compact, 96 B)
  on-disk format into the pinned (144 B) destination so we never heap-alloc
  the full SRS as a Go slice.
- Public API and disk dump format unchanged (G1TEPoint stays 96 B).

### Validation

- `TestMSM*` (full correctness suite, including reference G1Jac results
  at n=1024, 32K, 1M, 128M from msm_test.go:23-44) — all pass.
- `TestPlonkECMul*` end-to-end prover correctness — all pass.

### Bench — `BenchmarkMSMPinned`, 10 iters, pinned vs unpinned, with new format

| n    | A+B unpinned | A+B pinned | savings vs original | total speedup |
|---   |---:          |---:        |---:                 |---:           |
| 64K  | 3 063 µs     | 2 393 µs   | -889 vs 3 282       | 1.37×         |
| 256K | 5 287 µs     | 4 691 µs   | -1 296 vs 5 987     | 1.28×         |
| 1M   | 13 450 µs    | 11 020 µs  | -3 961 vs 14 981    | 1.36×         |
| 4M   | 45 060 µs    | 36 350 µs  | -10 439 vs 46 789   | 1.29×         |

### Phase-timing breakdown (B alone, unpinned, no Pin/Release effect)

| phase           | n=64K (Δ) | n=256K (Δ) | n=1M (Δ)  | n=4M (Δ)  |
|---              |---:       |---:        |---:       |---:       |
| accum_seq       | -240 µs   | -691 µs    | -1480 µs  | -1744 µs  |
| (relative)      | -20%      | -21%       | -17%      | -6%       |

Pattern matches the model: precomputed format wins more at small/mid n
where compute dominates; at n=4M the +50% point-memory traffic eats half
the compute saving. Still a net win at every measured size.

### Cumulative summary (A + B)

Combined MSM speedup: **1.28×–1.37× across n=64K…4M**.

Translates to roughly **6–10% end-to-end PlonK proof speedup** assuming
MSM is ~22% of proof time (measured at n=2²⁶ in BenchmarkPlonkECMul353).

## Step 4 — failed micro-experiment: launch_bounds(256, 3) — REVERTED

Tried bumping accumulate_buckets_kernel from `__launch_bounds__(256, 2)` to
`(256, 3)` to push more concurrent blocks per SM. The unified-add formula
needs ~80 registers per thread; pushing for 3 blocks/SM forces register
spilling. accum_seq regressed massively:

| n    | (256, 2)   | (256, 3)   | Δ     |
|---   |---:        |---:        |---:   |
| 64K  | 941 µs     | 1533 µs    | +62%  |
| 256K | 2 663 µs   | 4 311 µs   | +62%  |
| 1M   | 7 448 µs   | 15 317 µs  | +106% |
| 4M   | 28 001 µs  | 52 744 µs  | +88%  |

Reverted. (256, 2) is the correct setting for the current formula; raising
occupancy requires a smaller per-thread state — which is exactly what
batched-affine accumulation provides (next step).

## Step 5 — failed micro-experiment: FINALIZE_THREADS 32 → 64 — REVERTED

Bumped `FINALIZE_THREADS` from 32 to 64 to allow `reduce_blocks_per_window`
to reach the target of 50 (= 752/num_windows for c=17). Hypothesis: more
parallel range-blocks would shorten reduce_partial.

| phase           | (32) | (64) | Δ        |
|---              |---:  |---:  |---:      |
| reduce_partial  | 1670 | 1784 | +6.8% 1M |
| reduce_finalize | 262  | 369  | +41% 1M  |
| reduce_partial  | 1640 | 1726 | +5%   4M |
| reduce_finalize | 257  | 356  | +39%  4M |

Both phases regressed. finalize's prefix-scan grew log₂(64)=6 levels vs 5,
costing more shared-mem extended adds. partial got slower likely because
the kernel-launch overhead of 750 blocks vs 480 outweighed the smaller
per-block work — the partial kernel is launch-overhead-bound at this scale,
not parallelism-starved.

Reverted. Lesson: reduce_partial's ~1.6 ms is largely unavoidable launch
overhead with the current algorithm. Cracking it requires either a
different reduction algorithm (e.g., affine running-sum with batched
inverse) or fewer windows (= GLV — but only modest net gain per earlier
analysis).

## Step 6 — failed experiment: sub-group cooperative accumulate — REVERTED

Idea: K threads cooperatively accumulate one bucket. Each thread does ⌈B/K⌉
sequential mixed adds, then log₂(K) levels of shared-mem tree reduce.
Hoped to break the per-thread serial dependency at the cost of small reduce
overhead. Sweep K ∈ {4, 8, 16}.

`accum_seq` results (µs), env `GNARK_GPU_MSM_SUBGRP_K=K`:

|  K  |  64K  | 256K | 1M    | 4M    |
|---: |---:   |---:  |---:   |---:   |
|  0  |   946 | 2662 |  7444 | 28223 |  ← baseline (1 thread/bucket)
|  4  |  1640 | 2976 | 10742 | 28605 |
|  8  |  3119 | 4502 | 16130 | 33487 |
| 16  |  6854 | 8075 | 27589 | 42006 |

All K regress, monotonically worse with K. The extended-extended `ec_te_unified_add`
in the tree reduce (9M each, log₂K levels) plus shared-mem traffic is more
expensive than the saved per-thread serial chain (which was already
amortized by the 96K concurrent-thread saturation). Even at K=4 with avg
bucket = 60 (the "best" case), log₂4=2 levels of 9M add costs more than
shaving 45 mixed adds saves.

Reverted. **Hard lesson: any cooperative scheme using extended-coord
unified add for the reduce will lose**. The only path to faster
accumulate is batched-affine — a substantially larger refactor (point
format change + new pair-organization + cross-bucket batched inversion).

## Session summary

### Changes that landed (all validated against existing reference results
and end-to-end PlonK proofs)

1. **Per-phase MSM cudaEvent instrumentation** (msm.cu, api.cu, msm.go)
   — exposes precise phase breakdown for future optimization work.
2. **Persistent work buffers across MSM calls** (`PinWorkBuffers` /
   `ReleaseWorkBuffers`), wired into `prove.go` around the MSM wave and
   the quotient phase. Eliminates ~5–10 ms of cudaMalloc/Free +
   cudaHostRegister/Unregister per call.
3. **Precomputed-T point format** (G1EdYZD: Y-X, Y+X, 2d·X·Y in 144 B)
   replacing compact (X, Y) 96-byte format. Mixed-add drops from 9M to 7M.
   Conversion happens transparently at pin time; disk dump format and
   public Go API unchanged.

### Final MSM numbers (real SRS, 10 iters, RTX PRO 6000 Blackwell)

| n     | original (unpinned, compact) | A+B (pinned, precomputed) | speedup |
|---    |---:                           |---:                        |---:     |
| 64K   |   3 282 µs                    |   2 401 µs                 | 1.37×   |
| 256K  |   5 987 µs                    |   4 711 µs                 | 1.27×   |
| 1M    |  14 981 µs                    |  10 969 µs                 | 1.37×   |
| 4M    |  46 789 µs                    |  36 171 µs                 | 1.29×   |

### Negative results (documented and reverted)

| Experiment                                  | Result   |
|---                                          |---       |
| `__launch_bounds__(256, 3)` on accumulate   | -88% (register spill) |
| `FINALIZE_THREADS = 64` for reduce parallel | -7%   (launch overhead beats parallelism) |
| Sub-group cooperative kernel K∈{4,8,16}     | -73% to -206% (extended-add reduce too costly) |

These three failures all converge on the same insight: the existing
unified TE-extended addition is so good (7M with the new precomputed
input format) that any algorithmic change which adds even a small
constant of extended-extended adds will lose. The next step beyond this
ceiling **must** drop to affine arithmetic with batched inversion to win.

### What's the ceiling without batched-affine?

Per-phase breakdown at n=4M after A+B (pinned):

| phase           | µs     | % of MSM |
|---              |---:    |---:      |
| accum_seq       | 27 969 |   77%    |
| h2d             |  2 426 |    7%    |
| sort            |  2 379 |    7%    |
| reduce_partial  |  1 608 |    4%    |
| build_pairs     |  1 038 |    3%    |
| reduce_finalize |    254 |   <1%    |
| boundaries      |    237 |   <1%    |
| accum_par       |    125 |   <1%    |
| d2h             |     11 |   <1%    |

`accum_seq` is now 77% of the MSM. Without an algorithmic change to the
bucket accumulator (i.e., batched-affine), there's no further significant
optimization left in MSM at this scale. 

### Recommended next initiatives (for future sessions)

1. **Batched-affine bucket addition** in Short Weierstrass coordinates
   (sppark / cuZK / Yrrid style). Estimated multi-day kernel effort. Could
   give another 1.5–2× on `accum_seq` (= 1.4–1.7× on total MSM).
   Requires: SW affine point format alongside or instead of TE; new
   pair-organization kernels; cross-bucket batched inversion; multi-wave
   halving until each bucket reduces to one point.
2. **Multi-GPU sharding**. Linear scaling at n≥2²³, easy to validate.
3. **Fused inverse-coset NTT kernel**. Small NTT-side win
   (~30% of `CosetFFTInverse`, ~1 ms saved per call at large n).

## Step 7 — SW affine foundation primitives (in progress)

Goal: lay the groundwork for batched-affine bucket accumulation. Even though
the full integration is multi-day work, the primitives can be built and
validated incrementally.

### Implemented and validated

- **`fp_inv`** in `fp.cuh` — modular inverse via Fermat's little theorem
  (`a^(p-2) mod p`). Square-and-multiply over the 377-bit exponent, ~565
  fp_mul calls per inversion. Marked `__noinline__`. Used sparingly: only
  for the single global inversion in batched-invert contexts.
- **`G1AffineSW`** struct in `ec.cuh` — 96-byte SW affine point matching
  gnark's `bls12377.G1Affine` memory layout (12 limbs Montgomery form).
  Identity encoded as (0,0) (off-curve sentinel since 0² ≠ 1 = 0³+1).
- **`g1sw_neg`, `g1sw_cnegate`** — branchless conditional negate.
- **`g1sw_pair_add_with_inv_dx`** — non-unified affine pair add given
  precomputed 1/(x1-x0). Cost: 1S + 3M (the λ multiply is one of the 3M).
- **`g1sw_double_with_inv2y`** — affine doubling with precomputed 1/(2y).
  Included for completeness; near-zero hit rate in random-scalar MSM.
- **`g1sw_to_te_extended`** — SW affine → TE extended conversion at the
  output boundary of the new kernel. Mirrors the `convertToEdMSM` math
  from g1_te.go but for a single point. Uses one fp_inv internally
  (recovers two denominators with a single inversion via the standard
  `inv(a*b) → a, b` trick).

### Test infrastructure

- `gnark_gpu_test_sw_pair_add` / `gnark_gpu_test_sw_to_te` C entrypoints
  (msm.cu) running 1-thread test kernels and copying back to host.
- Go wrappers: `plonk.TestSWPairAddGPU`, `plonk.TestSWToTEGPU`
  (sw_affine_test_helpers.go — unfortunate name; cgo wrappers must live in
  the package, not in `_test.go`).
- `sw_affine_test.go`:
  - `TestGPUSWPairAdd` — 10 random pair-adds match `bls12377.G1Affine`
    addition exactly.
  - `TestGPUSWToTEExtended` — 5 random conversions match the reference
    formulas (constants and math from g1_te.go) exactly.

Both tests pass. The `fp_inv` correctness is implicitly validated end-to-end
via these tests since both rely on it.

### Built and validated

1. **Block-local batched-affine accumulate kernel** —
   `accumulate_buckets_batched_affine_kernel`. One block per bucket,
   256 threads. Reads SW affine points into shared mem, runs pairwise
   reduction with single-lane batched invert, converts result to TE
   extended at the kernel boundary.
2. **`MSM.LoadPointsSW`** API for uploading SW affine points to a
   dedicated `d_points_sw` buffer. Kept in parallel with existing
   TE-precomp `d_points`; new kernel uses one, legacy uses the other.
3. **Env-var gate** `GNARK_GPU_MSM_BATCHED_AFFINE=1` re-read on each
   `msm_run` so tests can A/B without process restart.
4. **Validation** — `TestGPUBatchedAffineReduce` (multi-wave reduce in
   isolation, N up to 256) AND `TestGPUBatchedAffineMSM` (full MSM
   pipeline at n=1024, 32K, 1M against reference G1Jac results) — all
   pass.

### Race condition found and fixed (2026-04-27)

Initial implementation failed non-deterministically at large N. Root
cause: in the pair-merge wave, pair-add at `tid=t` reads `pts[2t]` and
`pts[2t+1]`, while passthrough at `tid=half` writes `pts[half]`. For odd
active values, `pts[half]` is in the read range of pair-add at
`tid=half/2`, but in a different warp — no synchronization between read
and write. Different scheduling outcomes gave different results.

Fix: read all required inputs (lhs, rhs, last) into thread registers in
the first sub-phase, `__syncthreads()`, then write outputs in the second
sub-phase. Eliminates the data race entirely.

After the fix, `TestGPUBatchedAffineReduce` passes deterministically at
all N from 1 to 256 (verified across multiple runs).

### Bench result — `BenchmarkMSMBatchedAffine`, 10 iters

| n     | legacy `accum_seq` | batched-affine `accum_seq` | ratio |
|---    |---:                 |---:                          |---:   |
| 64K   |    944 µs           |  2 978 190 µs                |  3155× SLOWER |
| 256K  |  2 682 µs           |  5 300 524 µs                |  1976× SLOWER |
| 1M    |  7 014 µs           | 18 606 954 µs                |  2653× SLOWER |
| 4M    | 24 797 µs           | 26 388 558 µs                |  1064× SLOWER |

The kernel is **catastrophically slow**. Why:

- One block per bucket = ~983K blocks at n=4M.
- With `__launch_bounds__(256, 1)` only 188 blocks active at a time.
- Inside each block, the Montgomery batched-invert is single-lane
  (lane 0 does the entire forward scan + Fermat exp + backward scan,
  ~720 fp_mul per wave for avg bucket size).
- Effective concurrency: ~188 threads doing real work at any time,
  vs ~96K in the legacy kernel. **~500× less parallelism.**

The single Fermat exponentiation (565 fp_muls) per wave per block also
adds a fixed cost. log₂(B) waves × 565 = 3400+ fp_muls per bucket,
just for inversions. Compare to ~7B fp_muls in the legacy 7M
mixed-add chain. At B=15, that's 105 vs 3400 — 32× more inversion
work alone.

### What sppark/Yrrid actually do

The real win in batched-affine MSM is **cross-bucket batched
inversion**, not block-local. Across all wave-0 pairs (~30M at n=4M),
ONE Montgomery batch invert is performed: forward scan on 30M elements
(parallelizable), single Fermat exp once, backward scan on 30M elements
(parallelizable). The Fermat exp cost is amortized to ~0 per pair, and
the prefix product runs at full GPU memory-bandwidth throughput.

This requires:

- A pair-organization kernel that lays out all wave-`w` pairs in a flat
  buffer keyed by bucket-then-pair-position.
- A parallel prefix-product kernel (work-efficient scan, e.g.
  Blelloch).
- A single inversion kernel.
- A parallel backward scan that produces individual `inv_dx` for each
  pair.
- A finalize kernel that applies the pair-add per pair.

Per wave: 4–5 kernel launches. Across log₂(B) waves: 4–5× log₂(60) ≈
30 kernel launches. Substantially more orchestration than the current
1-kernel accumulate, but the parallel scan throughput should make it
worthwhile.

Effort estimate: 1 dedicated week of kernel work, given the code is
fresh (validated SW affine primitives, working block-local kernel as
testbed, validated SW→TE conversion). The next session should pick up
from this state.

### Disposition for THIS session

Keep the validated infrastructure (primitives, foundation kernel,
`LoadPointsSW`, env var, tests) on the branch. Default remains the
legacy 1-thread-per-bucket sequential kernel — batched-affine is
gated behind `GNARK_GPU_MSM_BATCHED_AFFINE=1` and only useful for
correctness validation of the SW arithmetic and the multi-wave
algorithm. **Do NOT enable in production paths** until cross-bucket
inversion lands.







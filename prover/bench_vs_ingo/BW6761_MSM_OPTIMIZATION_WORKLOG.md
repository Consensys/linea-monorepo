# BW6-761 MSM Optimization Worklog

Date: 2026-04-30

This note is the reset point for BW6-761 MSM optimization. It records the
measured bottleneck, the failed tactical experiment, and the clean algorithmic
plan to implement next.

## Scope

- Target: `gpu/plonk2` BW6-761 G1 MSM.
- Important sizes: `1 << 23`, `1 << 24`, and `1 << 25`.
- Comparison target: `/home/ubuntu/dev/cpp/icicle-gnark`.
- Current benchmark artifact folder: `bench_vs_ingo/`.

## Measured Baseline

Stable comparison in `gpu_perf_summary.csv` shows BW6-761 MSM is the main
regression:

- `1 << 15`: `plonk2 / ICICLE = 2.27x`.
- `1 << 20`: `plonk2 / ICICLE = 6.19x`.
- `1 << 23`: `plonk2 / ICICLE = 7.75x`.

Per-phase timings show the dominant cost is bucket accumulation:

- At `1 << 20`, `accum_seq` is about 500 ms of about 560 ms.
- At `1 << 22`, `accum_seq` is about 1.5 s of about 1.6 s.
- At `1 << 23`, `accum_seq` is about 2.65 s of about 2.75 s.
- A window sweep over `w=18..21` did not fix this; `w=18` remained best for
  `512Ki..8Mi`. Window tuning alone is not the right lever.

The current `plonk2` CUDA MSM shape is:

1. Split signed scalar windows into `(bucket, point)` pairs.
2. CUB radix sort by bucket key.
3. Detect bucket boundaries.
4. Accumulate each bucket with one CUDA thread.
5. Reduce buckets into window sums.
6. Final Horner-style window accumulation.

The bad step is 4. One thread performs all elliptic-curve additions for one
bucket. That leaves the GPU underused when BW6-761 buckets contain many points,
and each addition is expensive because BW6-761 uses 12-limb base-field
arithmetic.

## Failed Tactical Experiment

A quick prototype was attempted: keep the one-thread path for small buckets,
mark buckets above a threshold, and run one CUDA block per large bucket with an
in-block reduction.

Result after rebuilding `libgnark_gpu.a`:

- Sequential baseline, `16Mi`: about `4.89 s`, `accum_seq ~= 4.72 s`.
- Prototype, `16Mi`, 64 lanes / threshold 128: about `7.20 s`,
  `accum_seq ~= 4.72 s`, `accum_par ~= 2.31 s`.

Why it failed:

- It added a second pass for large buckets instead of replacing the expensive
  work shape.
- It still launched the sequential bucket kernel across all buckets and used a
  serial host copy of the large-bucket count as an orchestration barrier.
- One block per large bucket is too coarse and too expensive for buckets around
  128-256 points.
- It did not sort buckets by size or use segmented work descriptors, so the GPU
  still saw poor load balance.

The prototype was intentionally backed out. Do not revive it by threshold
tuning.

## Segmented Scheduler Experiment

A cleaner segmented populated-bucket implementation was tested next. It split
every non-empty BW6-761 bucket into 32-point segments, accumulated each segment
in parallel, and reduced the segment partials per bucket.

Correctness:

- `TestG1MSMPippengerRaw_CUDA/bw6-761` passed.
- Targeted CUDA MSM/commit tests passed.

First performance read, `PLONK2_BW6_MSM_SCALAR_MODE=full`,
`-benchtime=1x -count=1`:

| Size | Previous baseline | Segmented path | Result |
|---:|---:|---:|---:|
| `1Mi` | ~0.56 s | ~0.56 s | neutral |
| `4Mi` | ~1.61 s | ~1.98 s | slower |
| `8Mi` | ~2.79 s | ~3.45 s | slower |
| `16Mi` | ~4.89 s | ~6.60 s | slower |

Interpretation:

- The segmented path still performs essentially the same number of BW6 mixed
  additions, then adds Jacobian reductions over segment partials.
- The original one-thread-per-bucket kernel already has enough independent
  buckets to expose GPU parallelism at these sizes.
- The dominant gap is therefore not simply bucket-level parallelism. It is
  likely the per-addition cost and/or point representation/arithmetic quality
  relative to ICICLE.

Do not keep the segmented path unless it is paired with a reduction in the
number or cost of group operations.

## SOTA / Reference Design Notes

Primary local reference: ICICLE CUDA MSM in:

- `/home/ubuntu/dev/cpp/icicle-gnark/icicle/backend/cuda/src/msm/cuda_msm.cuh`
- `/home/ubuntu/dev/cpp/icicle-gnark/icicle/include/icicle/msm.h`
- `/home/ubuntu/dev/cpp/icicle-gnark/icicle/include/icicle/backend/msm_config.h`

Relevant ICICLE choices:

- It exposes `precompute_factor`, `c` window bits, `bitsize`, batching,
  device-resident inputs, and backend extensions.
- CUDA backend extension `large_bucket_factor` defaults to `10`.
- It uses `c = clamp(ceil(log2(msm_size)) - 4, 1, 20)` when `c` is automatic.
- It sorts populated buckets by size, computes a cutoff, and handles the large
  prefix separately.
- Large buckets are split into fixed-size segments of approximately average
  bucket size, accumulated into temporary bucket partials, reduced by a
  variable-size reduction, and then distributed back to bucket storage.
- Large-bucket work runs on its own CUDA stream and is synchronized before
  final bucket reduction.
- It supports chunking and overlaps copy/compute when data is host resident.

Literature/reference anchors:

- cuZK, ePrint 2022/1321: GPU MSM variants based on Pippenger; GPU
  implementation must avoid serial bucket accumulation bottlenecks.
  https://eprint.iacr.org/2022/1321.pdf
- PipeMSM, ePrint 2022/999: MSM performance is dominated by pipelined bucket
  accumulation/reduction and field arithmetic throughput.
  https://eprint.iacr.org/2022/999
- Pippenger performance thesis, Waterloo 2026: bucket method remains the
  standard large-MSM algorithm; optimizations target group-operation count and
  bucket scheduling.
  https://uwspace.uwaterloo.ca/items/43cc12b9-cb03-42bd-a01c-4b0ed5e12403

Important inference: ICICLE's `large_bucket_factor=10` tail path cannot by
itself explain the `1 << 23` gap. At `w=18`, average bucket size is about 64,
so a factor-10 cutoff would be about 640 points and only catch an extreme tail.
Therefore the next implementation must improve the general populated-bucket
work shape too. The segmented experiment showed that scheduling alone is still
not enough; the implementation must reduce the cost per bucket-addition or the
number of expensive group operations. Scheduling remains useful for true tail
buckets and host/device overlap, but the measured random-scalar gap is now
primarily an arithmetic/representation problem.

## Clean Design To Implement

### 1. Keep the Pippenger front-end but change bucket metadata

After sorting `(bucket, point)` pairs, build compact metadata for only populated
buckets:

- `bucket_key`: original global bucket id.
- `offset`: first point index in sorted point-index array.
- `size`: number of points in bucket.

Use CUB run-length encode or equivalent compact boundary output. The current
`bucket_offsets` / `bucket_ends` dense arrays are fine for correctness but poor
for scheduling because every later pass scans empty buckets.

### 2. Sort populated buckets by descending size

Sort `(size, bucket_key, offset)` by size descending. This gives a contiguous
large-bucket prefix and makes cutoff selection cheap. It also lets small-bucket
accumulation run over only populated small buckets.

### 3. Segment populated bucket accumulation

For BW6-761 large MSMs, do not assign one CUDA thread to one whole bucket.
Instead split each populated bucket into fixed-size point segments:

```text
segment_points = policy value, initially 32 or 64
segments_for_bucket = ceil(bucket_size / segment_points)
```

Each segment accumulates a short consecutive run of affine points into a
Jacobian partial. A second reduction pass combines the partials for each bucket.
This directly removes the long serial chain in the current kernel and gives a
regular amount of work per CUDA task.

The policy should be based on expected work per bucket, not an ad hoc benchmark
threshold. Start by enabling the segmented path when:

```text
average_bucket_size = count / (1 << window_bits)
average_bucket_size >= 16
```

That keeps small MSMs on the low-overhead path and switches BW6-761 large MSMs
to the scheduler that matches the measured bottleneck.

### 4. Keep a large-bucket stream for the true tail

Use:

```text
large_bucket_cutoff = large_bucket_factor * average_bucket_size
```

Start with ICICLE-compatible `large_bucket_factor = 10`, but keep it a policy
parameter in `MSMRunPlan`. The factor is a scheduling policy, not a benchmark
patch. It should have tests around memory planning and launch metadata. This
large-prefix path is not sufficient by itself; it is a tail optimization on top
of general segmentation.

For BW6-761 with `w=18`:

- `1 << 23`: average nonzero bucket size is about 64.
- `1 << 24`: average is about 128.
- `1 << 25`: average is about 256.

With factor 10, the large-bucket path only catches the true tail. That is good:
normal buckets should stay on a low-overhead path; only pathological long
buckets should be split.

### 5. Implement segmented bucket accumulation

Create descriptors for chunks of populated buckets:

```text
segment {
  bucket_rank
  segment_index
  offset
  length <= points_per_segment
}
```

Each segment is accumulated by one thread initially. This keeps the first clean
implementation simple and avoids expensive in-block Jacobian reductions for
short segments. A later version can use warp-per-segment if profiling shows
that one thread per segment underuses registers/SMs.

Then reduce segment partials per bucket. For normal buckets this reduction is a
short chain, e.g. 2-8 partials, rather than 64-256 mixed affine additions. The
output writes the final bucket sum into the normal bucket accumulator array.

### 6. Run normal and tail work concurrently

Launch the true large-bucket tail on a separate stream and the rest of the
segmented populated buckets on the main stream:

- large stream: segmented tail-bucket accumulation -> variable-size reduction
  -> scatter final large bucket sums.
- main stream: segmented accumulation/reduction for populated buckets after the
  large prefix.
- main stream waits on the large-stream event before bucket-to-window reduction.

This is the orchestration improvement the prototype lacked.

### 7. Revisit bucket-to-window reduction after accumulation is fixed

## 2026-04-30 Implementation Update

Implemented the first arithmetic optimization in `gpu/cuda/src/plonk2/field.cuh`:

- Added a `mul<BW6761FpParams>` specialization.
- The specialization represents the BW6-761 base field as 24 32-bit limbs
  internally.
- The hot multiply-add step uses explicit PTX carry-chain instructions
  (`mad.lo.cc.u32`, `madc.hi.u32`, `add.cc.u32`, `addc.u32`) instead of the
  generic 12-limb `unsigned __int128` CIOS path.
- The Pippenger algorithm, window policy, bucket memory shape, and public API
  are unchanged.

Measured result after rebuilding `gpu/cuda/build/libgnark_gpu.a`:

| Size | Previous plonk2 | Optimized plonk2 | Speedup |
|---:|---:|---:|---:|
| `1Mi` | ~0.555 s | 0.157 s | ~3.5x |
| `8Mi` | ~2.78 s | 0.689 s | ~4.0x |
| `16Mi` | ~4.89 s | 1.206 s | ~4.1x |
| `32Mi` | not previously run | 2.254 s | n/a |

Large-size comparison to ICICLE on the same machine:

| Size | Optimized plonk2 | ICICLE | plonk2 / ICICLE |
|---:|---:|---:|---:|
| `8Mi` | 0.689 s | ~0.355 s | ~1.9x |
| `16Mi` | 1.206 s | 0.525 s | ~2.3x |
| `32Mi` | 2.254 s | 0.879 s | ~2.6x |

This is a major improvement but does not yet meet the target of beating
ICICLE. The remaining gap is still mostly field-kernel quality inside bucket
accumulation, not window choice or bucket scheduling.

Follow-up experiments:

- Partial loop unrolling of the 24-limb multiply compiled but regressed badly
  (`16Mi` about 6.97 s). It was reverted.
- Moving the 32-bit modulus table to CUDA constant memory regressed slightly
  (`16Mi` about 1.36 s). It was reverted.
- A BW6-specific mixed-add path that skipped some generic degeneracy handling
  was made correct, but did not improve runtime (`16Mi` about 1.28 s). It was
  reverted.
- A post-optimization `c` check at `16Mi` found `c=18` still best for this
  implementation:
  - `c=18`: about 1.21-1.24 s.
  - `c=19`: about 1.28 s.
  - `c=20`: about 1.38 s.

Next required changes to beat ICICLE:

1. Replace the C++ looped 32-bit CIOS multiply with a fully generated
   ICICLE/Matter-Labs-style field backend: fixed register layout, explicit
   32-bit PTX add/mad chains, no dynamic limb access in the hot path, and
   specialized square.
2. Add a generated BW6-761 square path; current `square` still routes through
   multiplication and leaves a material amount of EC formula work on the table.
3. Revisit bucket accumulation only after the arithmetic backend is closer to
   ICICLE. Scheduling-only segmented paths were already shown to regress.
4. Then reassess true large-bucket splitting for skewed buckets and optional
   fixed-base precomputation under a memory budget. These are second-order
   until per-add field arithmetic is competitive.

Raw output for this implementation is saved in:

- `raw/bw6761_msm_32bit_ptx_20260430.txt`
The current reduction is small compared with `accum_seq` today, but once bucket
accumulation is fixed it may become visible. Keep it measured separately:

- `reduce_partial`
- `reduce_finalize`
- final Horner accumulation

Do not optimize this first.

### 8. Consider fixed-base precomputation only after bucket scheduling

ICICLE supports `precompute_factor`; our `MSMRunPlan` already has
`PrecomputeFactor` but it is not wired into CUDA. For persistent KZG SRS bases,
precomputation may help, but it increases resident point memory, which is
already expensive for BW6-761:

- affine BW6-761 point: 192 bytes.
- `1 << 25` SRS points: about 6.0 GiB before precompute.

Precomputation should be a second phase after the segmented scheduler is
correct and benchmarked.

## Implementation Tasks

1. Add compact populated-bucket metadata buffers to the resident MSM handle and
   memory planner.
2. Replace dense boundary-only scheduling with RLE or compact boundary output.
3. Add size-descending bucket sort.
4. Add large-prefix cutoff selection and device/host metadata reporting for
   benchmark diagnostics.
5. Implement segmented populated-bucket accumulation and per-bucket partial
   reduction.
6. Add the large-prefix stream as a tail optimization, with one event wait
   before existing bucket-to-window reduction.
7. Add correctness tests:
   - small BW6-761 known-value MSM;
   - randomized deterministic BW6-761 vs gnark-crypto for sizes around cutoff;
   - forced large-bucket test using repeated bucket digits;
   - all-zero and one-point edge cases.
8. Add benchmarks:
   - `1 << 20` through `1 << 25` BW6-761;
   - baseline sequential path if preserved behind an internal test switch;
   - phase metrics including populated buckets, large buckets, large segments,
     and stream wait time.
9. Only then tune `window_bits`, `large_bucket_factor`, and segment size.

## Acceptance Criteria

- Correctness matches gnark-crypto for all targeted BW6-761 tests.
- No regression for `1 << 20` and `1 << 23`.
- Clear speedup at `1 << 24` and `1 << 25`.
- Phase timing shows `accum_seq + accum_par` reduced, not merely moved between
  labels.
- Memory estimate includes all new buffers and remains feasible on the
  97 GiB RTX PRO 6000 for `1 << 25`.

## Resume Prompt

Start from this file. Do not implement a threshold-only branch. Implement the
segmented, size-sorted BW6-761 bucket scheduler described above, preserving the
existing Pippenger front-end and existing bucket-to-window reduction until the
accumulation phase is no longer dominant.

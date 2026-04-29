# plonk2 GPU Prover Backend Design

See `GPU_PLONK_LIBRARY_DESIGN.md` for the broader branch assessment,
drop-in-prover target architecture, and milestone roadmap. This file is the
shorter primitive-backend design note.

`gpu/plonk2` is the curve-generic GPU backend for PlonK prover acceleration.
It should reuse gnark's PlonK logic and replace only the low-level polynomial
and commitment kernels that dominate proving time.

## Principles

- Keep one Go/C ABI for all curves. Curve differences belong in compact curve
  descriptors and templated CUDA arithmetic, not in duplicated wrapper logic.
- Use gnark-crypto host layouts at API boundaries. MSM input points are affine
  short-Weierstrass points for BN254, BLS12-377, and BW6-761.
- Treat the current BLS12-377 twisted-Edwards MSM as an optional specialized
  backend. It is a useful performance baseline, but it should not define the
  generic architecture.
- Make peak memory explicit. BW6-761 buckets are large enough that window size
  and chunking are algorithm parameters, not incidental constants.
- Validate every GPU algorithm against gnark-crypto before optimizing it.
  Benchmarks are useful only after field arithmetic, NTTs, MSMs, and KZG
  commitments agree with CPU references.

## Layers

1. `Curve` and `CurveInfo` define field limb counts, scalar bit sizes, and
   supported IDs.
2. Fr vectors store limbs in structure-of-arrays form on device. Host transfer
   uses gnark-crypto's raw element layout and converts at upload/download.
3. FFT domains own per-curve twiddle tables and expose forward, inverse, coset,
   and bit-reversal operations through one API.
4. MSM handles own resident SRS points and reusable work buffers. `CommitRaw`
   is the KZG commitment boundary used by setup and, later, proof-time
   commitments.
5. PlonK orchestration should be adapted from `gpu/plonk`; the PlonK protocol
   itself should not be forked.

## MSM Direction

The current generic affine MSM is correct but not performance-ready. The ICICLE
comparison showed that an affine input contract can be competitive, while the
current `plonk2` implementation is dominated by sequential bucket and window
reductions.

The performance gap is not caused by affine inputs alone. It comes from the
shape of the reduction pipeline:

- `plonk2` creates one signed-window assignment per point/window and radix-sorts
  those assignments every commitment. This is a reasonable baseline, but it is
  a large fixed cost for PlonK where many commitments reuse the same resident
  SRS shape.
- bucket accumulation launches one thread per bucket. If a bucket owns many
  points, that thread walks the whole bucket serially.
- window reduction launches one block per window with one active thread. That
  thread performs the complete running-sum reduction over all buckets in the
  window, so a 16-bit window does roughly 32K bucket additions serially.
- the final Horner combination also runs on one thread. This is a smaller cost
  than window reduction, but it confirms that the current kernel is a
  correctness baseline rather than a GPU occupancy-oriented implementation.
- the Go `CommitRaw` path returns to the host for Montgomery-layout correction.
  That cost is minor compared with the current reductions at large sizes, but
  it should disappear in the optimized pipeline.

The ICICLE CUDA sources in `/tmp/icicle-gnark` confirm that their speed comes
from a materially different backend architecture:

- `MSMConfig` exposes `precompute_factor`, window size `c`, scalar `bitsize`,
  `batch_size`, shared-base batching, host/device residency flags,
  Montgomery-form flags, device-result output, async execution, and
  backend-specific extension knobs.
- the Go wrapper derives batch size and shared-base mode from the input/output
  slice lengths, then marks scalars, bases, and results as host or device
  resident before dispatching.
- the CUDA backend chooses `c` automatically from MSM size, clamps it for large
  buckets, and can override chunk count through `CUDA_MSM_NOF_CHUNKS`.
- memory planning uses `cudaMemGetInfo`, estimates scalar/index/point/bucket
  memory, adjusts `c` when bucket memory is too high, and chunks work when data
  does not fit comfortably.
- scalar assignment grouping uses CUB radix sort, CUB run-length encoding, and
  CUB scans rather than a hand-written boundary pass only.
- buckets are sorted by descending size. Large buckets are split into multiple
  segments, accumulated on a separate stream, reduced with variable-size
  reductions, and then distributed back into the main bucket array.
- bucket-module reduction uses an iterative parallel reduction path controlled
  by `single_stage_multi_reduction_kernel`; a simpler big-triangle path remains
  available behind a backend extension flag.
- host-to-device point transfers and large-bucket work can overlap with scalar
  processing through multiple streams and events.
- base precomputation is a first-class API, including shared-base batched MSMs.

This does not mean `plonk2` should import ICICLE's design wholesale. It does
mean the next `plonk2` MSM should copy the useful architecture boundaries:
configurable windows and chunks, bucket-size-aware scheduling, large-bucket
segmentation, batched shared-base commitments, device-resident inputs/results,
and source-level control over memory pressure.

The next MSM backend should keep the existing `G1MSM` API and replace only the
private reduction pipeline. The target design is:

- keep signed-window decomposition and radix-sort grouping;
- add an internal `MSMRunConfig` with `WindowBits`, `ScalarBits`,
  `PrecomputeFactor`, `BatchSize`, `SharedPoints`, `ChunkCount`,
  `LargeBucketFactor`, and device-residency flags;
- derive conservative defaults from `CurveInfo`, SRS length, and
  `PlanMSMMemory`, while allowing benchmark-only overrides;
- represent each non-empty bucket as one or more fixed-size work items
  `(bucket, begin, end)`, so large buckets occupy many blocks and small buckets
  do not monopolize a thread;
- accumulate bucket work items into partial Jacobian sums using block-local
  reductions;
- reduce bucket partials with a second compact kernel that is specialized by
  curve but not by large template expressions;
- replace single-thread window running sums with a parallel suffix-scan or
  segmented reduction over bucket sums;
- move the Montgomery correction and affine/projective normalization decisions
  into device code;
- batch PlonK commitments when possible, so setup/proof phases can amortize
  scalar upload, sort workspace, and launch overhead across `Ql`, `Qr`, `Qm`,
  `Qo`, `Qk`, `S1`, `S2`, `S3`, and `Qcp`;
- avoid template patterns that make `ptxas` impractically slow;
- preserve memory planning for BW6-761 before increasing bucket counts.

Any BLS12-377 twisted-Edwards path should be selected explicitly as a backend
optimization after the affine backend is competitive enough to justify the
extra maintenance surface.

### MSM Execution Plan

Progress should happen in small, measurable stages:

1. **Config and planner.** Add internal run configuration and extend
   `PlanMSMMemory` to estimate assignment, sort, bucket, partial, precompute,
   and batch memory. No kernel behavior changes.
2. **Device result normalization.** Move Montgomery correction into the CUDA
   MSM path and keep the Go result API unchanged. This removes a bad boundary
   before larger refactors.
3. **Bucket metadata.** Replace the hand-written boundary-only metadata with
   size-aware metadata: non-empty bucket IDs, bucket sizes, offsets, and an
   ordering by descending bucket size. This can initially still feed the old
   accumulation kernels.
4. **Large bucket path.** Split buckets above
   `LargeBucketFactor * average_bucket_size` into fixed-size segments, reduce
   their partials, and distribute results back to the bucket array.
5. **Window reduction.** Done for the first production path: per-window bucket
   ranges are reduced by a compile-friendly two-stage kernel, and the resident
   MSM handle keeps the partial buffers pinned with the rest of the work
   buffers. This intentionally does not add public MSM configuration knobs.
6. **Shared-base batching.** Add a private batched commitment entrypoint for
   PlonK setup/proof commitments using the resident SRS. Expose it publicly
   only after CPU/GPU equality and benchmarks are stable.
7. **Precomputation.** Evaluate precomputed bases for fixed PlonK SRS handles.
   Enable only when the memory planner shows that static memory is acceptable,
   especially for BW6-761.
8. **Prover integration.** Adapt `gpu/plonk` orchestration after the generic
   MSM and NTT layers are competitive on the reference benchmarks.

Each stage must leave `G1MSM.CommitRaw` valid so correctness tests and
benchmarks remain comparable.

The next MSM bottleneck is bucket accumulation, not window reduction. The
current accumulator still lets one thread own each bucket, so uniformly random
large MSMs are now close to ICICLE on BLS12-377, while BW6-761 remains slow on
small PlonK setup traces because every commitment pays for many wide-field
bucket/window operations. The next code stage should add size-aware bucket
metadata and large-bucket segmentation before adding a wider PlonK batching
surface.

For BW6-761, the current internal window policy is deliberately size-aware
without adding public MSM knobs: 13 bits below 256K points, 16 bits from 256K
to below 4M points, and 18 bits from 4M points upward. A 30,000,000 point
BW6-761 MSM is modeled at roughly 23.1 GB of dominant GPU memory and has been
validated on the current 98 GB GPU. Future chunking should preserve this API
while lowering the required peak memory on smaller cards.

### NTT Direction

The current `plonk2` NTT is much closer to useful performance than MSM, but
ICICLE still exposes design points worth copying:

- explicit input/output ordering, so callers can avoid standalone bit-reversal
  when adjacent polynomial operations are order-agnostic;
- batched transforms with row/column layout flags;
- host/device input and output flags, so transforms can stay resident across
  prover phases;
- backend extensions for fast twiddle tables and algorithm selection.

`plonk2` should not add all of these knobs to the public Go API immediately.
The right next step is an internal NTT plan that records ordering and
residency, then teaches PlonK orchestration to consume whichever order avoids
extra kernels.

### MSM Acceptance Bar

An MSM optimization is not ready just because one synthetic benchmark improves.
It must pass:

- CPU/GPU equality against gnark-crypto `MultiExp` for random and structured
  scalars on BN254, BLS12-377, and BW6-761;
- KZG commitment equality against gnark PlonK setup traces on all three curves;
- the 16-constraint and 1K-constraint setup-commitment benchmarks in
  `BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA`;
- large SRS benchmarks at 16K, 64K, 1M where memory permits;
- BW6-761 OOM checks from `PlanMSMMemory`, including the selected window size
  and chunking policy.

## Benchmark Contract

Required comparison points:

- gnark CPU PlonK setup and prove for BN254, BLS12-377, and BW6-761 on the
  same benchmark circuits;
- current `gpu/plonk` full BLS12-377 GPU prove on the same benchmark circuit,
  used as the first end-to-end target for generic orchestration;
- `gpu/plonk2` setup-commitment acceleration for BN254, BLS12-377, and
  BW6-761 against CPU MSM commitments on the same selector/permutation
  polynomials;
- `gpu/plonk` BLS12-377 NTT and MSM;
- `gpu/plonk2` BN254, BLS12-377, and BW6-761 NTT;
- `gpu/plonk2` affine MSM for all target curves;
- ICICLE MSM/NTT when a matching CUDA backend can be built reproducibly.

Each optimization pass should update `WORKLOG.md` with exact commands,
hardware, timings, and any correctness tests run.

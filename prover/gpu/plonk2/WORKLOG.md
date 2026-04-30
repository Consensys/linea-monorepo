# plonk2 GPU Generalization Worklog

`gpu/plonk2` is a curve-generic GPU foundation for PlonK proving across
**BN254**, **BLS12-377**, **BW6-761**.

The existing `gpu/plonk` package is a tightly tuned BLS12-377 path that uses
twisted Edwards point coordinates and BLS12-377-specialised field arithmetic.
Generalising it in place would mean duplicating the entire pipeline three
times. `gpu/plonk2` instead introduces one templated CUDA layer parameterised
by a `Params` struct (modulus, INV, R, limb count) and one curve-indexed C
ABI on top of it.

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│  Go: gpu/plonk2 (build tag: cuda)                                │
│    Curve   { BN254, BLS12-377, BW6-761 }                         │
│    FrVector / FFTDomain / quotient kernels / G1MSM / KZG commit  │
└──────────────────────────────────────────────────────────────────┘
                              ↓ cgo
┌──────────────────────────────────────────────────────────────────┐
│  C ABI: gnark_gpu.h (gnark_gpu_plonk2_*)                         │
└──────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────┐
│  CUDA: gpu/cuda/src/plonk2                                       │
│    field.cuh  : templated Montgomery for Fr / Fp                 │
│    ec.cuh     : templated short-Weierstrass affine + Jacobian    │
│    kernels.cu : Fr arithmetic, NTT (DIF/DIT), bit-reverse, scale │
│    g1.cu      : single-point validation kernels                  │
│    msm.cu     : curve-generic affine Pippenger MSM               │
└──────────────────────────────────────────────────────────────────┘
```

### Key design choices

- **One templated arithmetic, three field instantiations.** `field.cuh`
  expresses Montgomery operations in terms of `Params::LIMBS` and a constexpr
  `INV`. BN254 (4 limbs), BLS12-377 (4 / 6 limbs for Fr / Fp), BW6-761
  (6 / 12 limbs) reuse the same templated kernels.
- **Short-Weierstrass affine inputs for every curve.** Avoids duplicating the
  entire MSM pipeline per coordinate system. The current TE path in
  `gpu/plonk` stays as the optimised BLS12-377 baseline; we only revisit TE
  here if a measured delta justifies it.
- **Host layout = gnark-crypto layout.** Scalars and points are uploaded as
  raw AoS Montgomery limbs (`unsafe.Slice` over the gnark-crypto element
  arrays), so callers do not have to materialise extra buffers.
- **Device layout = SoA for scalars, AoS for points.** Field vectors use
  per-limb arrays for coalesced NTT/elementwise accesses. Affine points are
  stored AoS — each MSM bucket-accumulator thread reads a different point, so
  cross-thread coalescing is impossible regardless.
- **Curve dispatch is a switch in a thin launcher.** No per-curve copy of
  kernels in different translation units; one templated kernel, three
  explicit instantiations per launcher.
- **Multi-GPU is implicit via `gpu.Device`.** Every plonk2 call carries a
  device handle; allocations and launches go through the per-device CUDA
  stream. Concurrent operations on different `*gpu.Device` values run on
  separate GPUs.

## Status

| Component                                           | State    |
|-----------------------------------------------------|----------|
| Curve descriptor + Go enum                          | done     |
| Templated Fr / Fp Montgomery                        | done     |
| Fr vectors: copy/add/sub/mul/addmul/scalar ops      | done     |
| Forward / inverse NTT + bit-rev                     | done     |
| Coset FFT (forward + inverse)                       | done     |
| PlonK quotient backend kernels                      | done     |
| Z ratio factors + Z prefix product                  | done     |
| Real PlonK toy proof baseline for all target curves | done     |
| PlonK setup commitments via GPU MSM, all curves     | done     |
| Single-point G1 add/double                          | done     |
| Curve-generic G1 affine MSM                         | done     |
| Production curve-generic Pippenger MSM              | correctness-first |
| MSM memory planner                                  | done     |
| KZG commit/open quotient validation (3 curves)      | done     |
| Generic prover prepared state (3 curves)            | done     |
| Fixed selector/permutation GPU preparation          | done     |
| Solved wire Lagrange commitment phase (3 curves)    | done     |
| Multi-GPU sanity test                               | planned  |
| Comprehensive benchmarks                            | partial  |
| Full GPU PlonK proof generation on all curves       | partial: BLS12-377 legacy bridge + generic prep |

## MSM design (Pippenger, signed digits, CUB-sort)

Let `c` be the window width and `W = ceil(scalarBits/c)` the number of windows.
Each scalar is split into signed digits `d_w ∈ [-2^(c-1), 2^(c-1)]` per
window. The MSM is

    Σᵢ sᵢ Pᵢ = Σ_w 2^(w·c) · ( Σ_{d=1..2^(c-1)} d · B_{w,d} )

where `B_{w,d} = Σᵢ Pᵢ` over points whose signed window-`w` digit equals `±d`
(negative digits flip the point's Y coordinate).

**Pipeline**

1. **Decompose scalars.** One thread per scalar emits `W` `(key, value)`
   pairs.
   - `value` packs `(point_idx << 1) | sign_bit`.
   - `key` packs `(window_id, |digit|-1)` for non-zero digits, or a sentinel
     past `total_buckets` for zero digits.
2. **CUB radix sort.** Sort pairs by key. Sentinel keys cluster at the end
   and are skipped by the bucket boundary detector.
3. **Bucket boundaries.** One thread per assignment marks the start/end of
   each bucket run in the sorted array.
4. **Bucket accumulation.** One thread per bucket walks its assignments and
   accumulates them into a Jacobian point via the mixed `affine + Jacobian`
   addition formula. Negation of the affine summand is a single field
   negation of `Y`.
5. **Window reduction.** One thread per window computes
   `T_w = Σ_d d·B_{w,d}` via the running-sum trick:
   `T = 0; sum = 0; for d = 2^(c-1) downto 1: T += B_{w,d}; sum += T`.
   Linear in buckets, parallelisable across windows.
6. **Combine windows on host.** `final = T_{W-1}; for w = W-2 downto 0:
   final = (final << c) + T_w`. `W ≤ 30`, so this is microseconds done in
   gnark-crypto Jacobian.

**Window/bucket sizing.**

| Curve      | scalarBits | c  | W  | buckets/window | total buckets |
|------------|-----------:|---:|---:|---------------:|--------------:|
| BN254      |       254  | 16 | 16 |          32768 |       524288  |
| BLS12-377  |       253  | 16 | 16 |          32768 |       524288  |
| BW6-761    |       377  | 13 | 30 |           4096 |       122880  |

BW6-761's scalar is 377 bits and its base field is 12 limbs, so each
Jacobian bucket is 288 bytes. We pick a smaller `c=13` to keep peak bucket
memory around 36 MB even at large `N`. BN254 / BLS12-377 buckets are 96 bytes,
so `c=16` is comfortable.

## Validation strategy

Every implemented kernel is validated against gnark-crypto:

- raw host/device roundtrips (4-limb and 6-limb scalars)
- Fr add / sub / mul / addmul / scalar-mul / add-scalar-mul element-wise vs
  `bnfr` / `blsfr` / `bwfr`
- device-to-device Fr copy for all three curves
- Forward NTT vs `fft.NewDomain(n).FFT(_, fft.DIF)` for all three curves
- Forward+inverse NTT roundtrip
- Coset forward / inverse NTT vs `fft.OnCoset()` round-trips
- quotient primitives vs gnark-crypto field arithmetic:
  `ComputeL1Den`, `ReduceBlindedCoset`, `Butterfly4Inverse`,
  `PlonkGateAccum`, `PlonkPermBoundary`, `PlonkZComputeFactors`,
  and `ZPrefixProduct`
- Single G1 affine add/double for `inf+P`, `P+(-P)`, `2P`, `P+Q`
- affine Pippenger MSM vs gnark-crypto `G1Jac.MultiExp` for all three curves
- KZG commit and KZG opening quotient commitments vs gnark-crypto
- real gnark PlonK setup/prove/verify toy circuit on BN254, BLS12-377, and
  BW6-761, plus an invalid-witness negative test
- PlonK setup trace commitments recomputed through the generic GPU KZG
  backend for all target curves

The remaining validation target is full proof-generation wiring: replacing the
KZG/FFT/quotient backend calls inside the existing `gpu/plonk` prover
orchestration with the `plonk2` curve-generic surfaces. gnark's public PlonK
API does not expose a KZG callback, so this is a prover integration task, not
an option flip.

## Multi-GPU

The CUDA primitives are stateless beyond the per-device context provided by
`gpu.New(...)`. Multi-GPU validation is planned once MSM exists; the test
should allocate one `*gpu.Device` per visible GPU, run the same MSM
concurrently, and compare each result to the single-GPU CPU baseline. On a
single-GPU host it should self-skip.

## Benchmarks (RTX PRO 6000 Blackwell, n=2^20)

Field arithmetic and NTT:

```
BenchmarkFrVectorMul_CUDA/bn254          ~160 µs/op  (previous run)
BenchmarkFrVectorMul_CUDA/bls12-377       ~74 µs/op
BenchmarkFrVectorMul_CUDA/bw6-761        ~273 µs/op  (previous run)
BenchmarkFFTForward_CUDA/bn254           ~615 µs/op  (previous run)
BenchmarkFFTForward_CUDA/bls12-377       ~480 µs/op
BenchmarkFFTForward_CUDA/bw6-761         ~973 µs/op  (previous run)
BenchmarkCosetFFTForward_CUDA/bls12-377  ~791 µs/op
```

Comparison against the existing specialized `gpu/plonk` BLS12-377 kernels,
measured with `-benchtime=10x -count=3`:

| Operation       | `gpu/plonk` mean | `gpu/plonk2` mean | Delta |
|-----------------|-----------------:|------------------:|------:|
| Fr mul, n=2^20  |          86.5 µs |           73.7 µs | -14.8% |
| FFT fwd, n=2^20 |         355.1 µs |          480.3 µs | +35.3% |

Coset FFT was measured separately with `-benchtime=5x -count=3`:

| Operation             | `gpu/plonk` mean | `gpu/plonk2` mean | Delta |
|-----------------------|-----------------:|------------------:|------:|
| Coset FFT fwd, n=2^20 |         509.4 µs |          790.6 µs | +55.2% |

The generic Montgomery multiplication is already competitive. The FFT and
coset gaps are the first performance targets before moving to MSM.

MSM comparison against the existing specialized BLS12-377 twisted-Edwards
path, measured with `-benchtime=1x -count=1` on the same SRS/scalar dump:

| Size | `gpu/plonk` TE MSM | `gpu/plonk2` affine MSM | Delta |
|------|-------------------:|------------------------:|------:|
| 16K  |        5.5-8.1 ms  |             1.7-3.6 s   | hundreds-x slower |
| 64K  |           5.75 ms  |                2.33 s   | 404x slower |

This confirms that the current generic MSM is a correctness-first backend,
not yet production competitive. The immediate bottleneck is the deliberately
simple sequential bucket/window reduction and per-run scratch allocation.
An attempted templated parallel reduction in `msm.cu` was backed out because
NVCC/ptxas compile time became impractical, especially with BW6-761's 12-limb
base-field instantiation. Future optimization should split the reduction into
smaller translation units or specialize the heavy reduction kernels by base
field width.

## Out of scope (followup work)

- Full PlonK prover ported off `gpu/plonk` — large project; this package
  delivers the compute foundation it would build on, plus a verified KZG
  commit path which is the heart of the prover.
- Twisted-Edwards fast path on top of plonk2 (only if a measured delta
  justifies the duplicated surface area).
- Ingonyama ICICLE comparison — needs a separate clone+build job.

---

## 2026-04-28 — Original Fr / NTT foundation

Initial curve-generic foundation (Fr arithmetic and NTT for BN254 /
BLS12-377 / BW6-761) was landed and validated against gnark-crypto in this
session. See git history for the original commit; this worklog now tracks
the full plonk2 effort.

## 2026-04-28 — BW6-761 G1 Validation Fix

`TestG1AffinePrimitives_CUDA` exposed an incorrect BW6-761 base-field
Montgomery inverse in `field.cuh`. The constant is now aligned with
gnark-crypto's `fp.qInvNeg`, and both CUDA and non-CUDA package tests pass:

```
go test ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk2 -count=1
```

## 2026-04-28 — Curve-Generic Coset FFT

Added `FrVector.ScaleByPowersRaw`, `FFTDomain.CosetFFT`, and
`FFTDomain.CosetFFTInverse`. The implementation is decomposed:

```
CosetFFT        = ScaleByPowersRaw(g)    → FFT        → BitReverse
CosetFFTInverse = BitReverse             → FFTInverse → ScaleByPowersRaw(g⁻¹)
```

This matches `gpu/plonk`'s natural-order coset API and validates against
gnark-crypto on BN254, BLS12-377, and BW6-761. Scale-by-powers now uses a
per-call 256-entry device power table rather than per-element exponentiation.

## 2026-04-30 — Generic Prover Orchestration and Large-Bucket MSM

The generic prepared prover now has a clearer memory lifecycle:

1. resident proving-key state: canonical SRS, Lagrange SRS, FFT domain,
   permutation table, and fixed polynomials
2. explicit MSM scratch pinning for commitment waves
3. scratch release before later quotient/permutation phases

`NewProver` and `Prove` bind the current OS thread to the selected CUDA device
before CUDA work. This preserves correctness on multi-GPU hosts, where CUDA's
current device is thread-local.

The generic `gnark_gpu_plonk2_msm_create` C API now uploads points only; it no
longer allocates sort/key/bucket work buffers eagerly. Work buffers are
allocated by `PinWorkBuffers` or lazily by the first commitment. The memory
planner now counts both canonical and Lagrange SRS residency and includes the
overflow-bucket metadata used by the large-bucket path.

The major performance fix is a bounded two-phase accumulator in
`gpu/cuda/src/plonk2/msm.cu`:

- serial phase processes at most a dynamic cap per bucket
- overflow buckets are recorded in a compact device list
- a block-parallel tree reduction accumulates each large bucket tail

This fixes the solved-wire pathology where random MSM benchmarks were fast but
real PlonK L/R/O commitments could spend seconds in one-thread-per-bucket
serial accumulation.

Large solved-wire L/R/O wave benchmarks on RTX PRO 6000 Blackwell:

| Curve | Domain | Wave Time | Planned Peak |
|---|---:|---:|---:|
| BN254 | `1<<23` | 447.968 ms | 10.740 GiB |
| BN254 | `1<<24` | 871.679 ms | 21.428 GiB |
| BLS12-377 | `1<<23` | 1140.263 ms | 11.264 GiB |
| BLS12-377 | `1<<24` | 2144.233 ms | 22.451 GiB |
| BW6-761 | `1<<23` | 4872.834 ms | 17.457 GiB |
| BW6-761 | `1<<24` | 9131.384 ms | 34.144 GiB |

Raw data and analysis are preserved in
`bench_vs_ingo/GENERIC_PLONK_ORCHESTRATOR_WORKLOG.md` and
`bench_vs_ingo/generic_plonk_orchestrator_summary.csv`.

## 2026-04-28 — Direction Correction

`gpu/plonk` remains the reference for PlonK proving logic. The goal of
`gpu/plonk2` is not to fork a second prover, but to provide the curve-generic
backend surface that the existing prover flow needs: field vectors, FFT/coset
FFT, MSM/KZG commitment, and the fused quotient kernels. Validation should
compare these backend primitives against gnark-crypto and then plug them under
the existing `gpu/plonk` orchestration.

## 2026-04-28 — Fr Backend Surface Parity

Added the next `gpu/plonk`-style Fr vector operations to the curve-generic
backend:

- device-to-device copy
- `AddMul`: `v += a*b`
- `ScalarMulRaw`: `v *= c`
- `AddScalarMulRaw`: `v += a*c`

These are implemented once in `gpu/cuda/src/plonk2/kernels.cu`, dispatched by
curve id through the existing C ABI, and validated for BN254, BLS12-377, and
BW6-761 in `TestFrVectorOps_CUDA`. This keeps the effort focused on replacing
the specialized backend under the proven `gpu/plonk` orchestration.
The BLS12-377 benchmark is currently ~1.55x slower than `gpu/plonk`'s fused
coset path, so a fused generic scale+NTT+bit-reverse kernel remains the next
FFT optimization candidate.

## 2026-04-28 — PlonK E2E Baseline and Quotient Backend

Added a real PlonK toy proof test covering all target curves:

```
go test ./gpu/plonk2 -run TestPlonkE2E_AllTargetCurves -count=1
```

The test uses gnark's PlonK backend and `unsafekzg` SRS generation to compile,
setup, prove, and verify a small multiplication circuit over BN254,
BLS12-377, and BW6-761. It also checks that an invalid witness fails proof
generation. This is the end-to-end correctness baseline for the PlonK logic;
`gpu/plonk2` remains focused on replacing the curve-specialized CUDA backend
under that prover flow.

Added the curve-generic CUDA quotient surface required by the existing
`gpu/plonk` orchestration:

- `ReduceBlindedCoset`
- `ComputeL1Den`
- `PlonkZComputeFactors`
- `ZPrefixProduct`
- `PlonkPermBoundary`
- `PlonkGateAccum`
- `Butterfly4Inverse`

These kernels are templated over BN254, BLS12-377, and BW6-761 scalar fields,
validated against gnark-crypto field arithmetic, and exposed through one C ABI
and one Go API. The Z prefix product keeps the proven chunked scan structure:
GPU local chunk products, CPU scan of the small chunk-product vector, GPU
global fixup and exclusive shift.

Validation after this step:

```
cmake --build gpu/cuda/build --target gnark_gpu -j2
go test ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk2 -count=1
```

Remaining end-to-end GPU prover blockers are now concentrated in the
curve-generic KZG commitment path:

- general MSM over short-Weierstrass affine SRS points
- KZG commit and batch-open wrappers for BN254, BLS12-377, and BW6-761
- wiring those wrappers under the existing `gpu/plonk` prover flow without
  duplicating prover logic

## 2026-04-28 — Validation MSM Hook

Added a small validation-only affine MSM entrypoint:

```
gnark_gpu_plonk2_test_msm_naive
```

It runs a single-thread double-and-add MSM on GPU using the same curve-generic
short-Weierstrass formulas as the G1 add/double validation kernels. This is
not the production MSM design and should not be used for prover performance
measurements. Its purpose is narrower: prove that the field/curve templates,
affine point layout, scalar layout, and C/Go ABI can compute a KZG-style
commitment for BN254, BLS12-377, and BW6-761.

The test compares against gnark-crypto `G1Jac.MultiExp` for all three curves.
Because the GPU test kernel decomposes Montgomery-form scalar limbs directly,
the expected CPU result is multiplied by the Montgomery radix before
comparison, matching the correction pattern used by the current BLS12-377 MSM.

Validation:

```
go test -tags cuda ./gpu/plonk2 -run 'TestG1MSMNaive_CUDA|TestG1AffinePrimitives_CUDA' -count=1
```

Next step for real prover throughput is replacing this test hook with the
planned Pippenger pipeline over affine short-Weierstrass points:

- signed-window scalar decomposition for 4-limb and 6-limb scalar fields
- memory-bounded bucket accumulation, especially for BW6-761's 12-limb base
  field
- batched MSM calls for PlonK commitments and openings

## 2026-04-28 — Curve-Generic Affine Pippenger MSM

Added the first real `gpu/plonk2` MSM backend:

```
gnark_gpu_plonk2_msm_pippenger
```

The backend accepts gnark-crypto short-Weierstrass affine G1 points for all
supported curves and dispatches through the same C ABI by curve id. It uses
the production dataflow shape expected by the prover backend:

1. signed-window scalar decomposition over raw Montgomery scalar limbs
2. CUB radix sort of `(bucket, point)` assignments
3. bucket boundary detection
4. bucket accumulation in curve-generic Jacobian coordinates
5. per-window running-sum reduction
6. Horner recombination into one projective commitment

`CommitRaw` now calls this Pippenger backend instead of the single-thread
validation MSM hook. The validation hook remains available as
`gnark_gpu_plonk2_test_msm_naive` because it is useful for isolating EC formula
or ABI bugs.

The current bucket accumulation and window reduction kernels are deliberately
simple: one thread per bucket for accumulation and one thread per window for
the running-sum reduction. This is correct and keeps the API/review surface
small, but it is not the final performance target. The next MSM optimization
work should parallelize large buckets and window reductions without changing
the public Go/C surfaces.

Validation:

```
cmake --build gpu/cuda/build --target gnark_gpu -j2
go test -tags cuda ./gpu/plonk2 -run 'TestG1MSMPippengerRaw_CUDA|TestCommitRaw_CUDA' -count=1
go test -tags cuda ./gpu/plonk2 -count=1
go test ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk -run 'TestFr|TestFFTSmall|TestFFTRoundtrip|TestCosetFFT$' -count=1
```

Follow-up correction: the private MSM backend still decomposes raw
Montgomery-form scalar limbs directly, so its projective output is multiplied
by the scalar Montgomery radix. `CommitRaw` now applies the same host-side
`R^-1` correction used by the current `gpu/plonk` MSM before returning. Its
public semantics are therefore a true KZG-style commitment, while
`g1MSMPippengerRaw` remains the uncorrected diagnostic layer.

Added KZG-level validation against gnark-crypto:

```
go test -tags cuda ./gpu/plonk2 -run 'TestCommitRawMatchesKZG' -count=1
```

The tests compare `CommitRaw` against `kzg.Commit` and compare quotient
commitments used by `kzg.Open` for BN254, BLS12-377, and BW6-761.

Added a reusable MSM handle:

```
NewG1MSM(dev, curve, affinePointsRaw)
(*G1MSM).CommitRaw(scalarsRaw)
```

This keeps the canonical SRS points resident on the GPU across commitments,
which is the shape needed by the PlonK prover. The first implementation still
allocates per-run sort/bucket/scalar buffers; only point upload is amortized.
That keeps the implementation small while giving the prover integration a
stable ownership model. The next performance step is to pin/reuse sort and
bucket buffers across a wave of commitments, matching the current `gpu/plonk`
MSM lifecycle.

Validation:

```
cmake --build gpu/cuda/build --target gnark_gpu -j2
go test -tags cuda ./gpu/plonk2 -run 'TestG1MSMCommitRaw_CUDA|TestCommitRawMatchesKZG' -count=1
go test -tags cuda ./gpu/plonk2 -count=1
go test ./gpu/plonk2 -count=1
```

Remaining end-to-end GPU prover blockers:

- replace one-shot KZG commitment calls in the prover with the resident MSM
  handle
- wire those wrappers under the existing `gpu/plonk` prover orchestration
- benchmark generic BLS12-377 affine MSM against the current
  twisted-Edwards-specialized `gpu/plonk` MSM
- tune BW6-761 windowing/chunking so bucket memory stays below GPU VRAM limits

## 2026-04-28 — Resident MSM Work-Buffer Lifecycle

The resident `G1MSM` handle now owns reusable scratch buffers for scalar
staging, CUB sorting, bucket offsets, bucket accumulators, per-window results,
and the projective output. `CommitRaw` uses a preallocated device-points path,
so repeated commitments no longer allocate/free those buffers inside the hot
call.

Added explicit lifecycle methods:

```
(*G1MSM).PinWorkBuffers()
(*G1MSM).ReleaseWorkBuffers()
```

`ReleaseWorkBuffers` synchronizes the MSM stream and frees only scratch
buffers; SRS points remain resident. `CommitRaw` lazily reallocates scratch if
needed. This mirrors the current `gpu/plonk` memory lifecycle and gives the
future prover integration a way to reclaim bucket/sort memory during quotient
phases.

Validation now covers releasing and repinning work buffers in the all-curve
resident MSM tests:

```
cmake --build gpu/cuda/build --target gnark_gpu -j2
go test -tags cuda ./gpu/plonk2 -run 'TestG1MSMCommitRaw_CUDA|TestCommitRawMatchesKZG|TestPlonkE2EGPUSetupCommitments_AllTargetCurves_CUDA' -count=1
go test -tags cuda ./gpu/plonk2 -count=1
go test ./gpu/plonk2 -count=1
golangci-lint run ./gpu/plonk2
golangci-lint run --build-tags cuda ./gpu/plonk2
```

Benchmarks still show the same main conclusion: allocation reuse does not fix
the dominant cost. The bottleneck is the sequential bucket/window reduction,
not per-run allocation churn.

## 2026-04-28 — PlonK Setup E2E Through GPU KZG Backend

Added a CUDA E2E setup test:

```
go test -tags cuda ./gpu/plonk2 -run TestPlonkE2EGPUSetupCommitments_AllTargetCurves_CUDA -count=1
```

For BN254, BLS12-377, and BW6-761 the test now:

1. compiles the toy multiplication circuit;
2. builds an unsafe test SRS;
3. runs gnark PlonK setup/prove/verify end to end;
4. rebuilds the typed PlonK setup trace;
5. uploads the actual Lagrange SRS from the typed proving key to `G1MSM`;
6. recomputes every setup commitment (`Ql`, `Qr`, `Qm`, `Qo`, `Qk`,
   `S1`, `S2`, `S3`, and all `Qcp`) with the generic GPU MSM; and
7. compares the GPU commitments to the gnark verifying key.

This is the first all-curve E2E path where PlonK-specific material flows
through the generic GPU KZG backend. It is not yet full GPU proof generation:
proof-time commitments are still hidden behind gnark's curve-specific prover
internals and must be wired by adapting the existing `gpu/plonk`
orchestration rather than by trying to hook the public gnark API.

## 2026-04-28 — ICICLE MSM/NTT Comparison Pass

Cloned ICICLE into `/tmp/icicle` at `625532a624e5` and attempted to use the
current tree for CUDA comparisons. The current ICICLE repository no longer
carries the open CUDA backend in-tree, so `-cuda=local` cannot build the CUDA
backend from the clone. For executable numbers on this host, built the cached
`github.com/ingonyama-zk/icicle-gnark/v3 v3.2.2` CUDA libraries for BN254,
BLS12-377, and BW6-761 into `/tmp/icicle-gnark-install`.

Added an opt-in benchmark harness behind `cuda && iciclebench`:

```
gpu/plonk2/icicle_compare_cuda_test.go
gpu/plonk2/icicle_roots_cuda.go
```

The harness is intentionally outside default test tags. It compares:

- current `gpu/plonk` BLS12-377 twisted-Edwards MSM
- generic `gpu/plonk2` BLS12-377 affine MSM
- ICICLE BLS12-377 MSM
- generic `gpu/plonk2` NTT for BN254, BLS12-377, and BW6-761

MSM benchmark command:

```
CGO_LDFLAGS='-L/tmp/icicle-gnark-install/lib -Wl,-rpath,/tmp/icicle-gnark-install/lib' \
LD_LIBRARY_PATH=/tmp/icicle-gnark-install/lib \
ICICLE_BACKEND_INSTALL_DIR=/tmp/icicle-gnark-build-bls12_377/backend/cuda \
go test -tags 'cuda iciclebench' ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkCompareMSMBLS12377SRS' -benchtime=3x -count=1
```

Results on the RTX PRO 6000 Blackwell host:

| MSM | 16K | 64K |
| --- | ---: | ---: |
| current `gpu/plonk` TE MSM | 4.79 ms | 3.31 ms |
| ICICLE affine MSM | 5.79 ms | 6.09 ms |
| `gpu/plonk2` affine MSM | 1.69 s | 2.33 s |

The current generic affine MSM is therefore only a correctness/reference
pipeline. Its sequential bucket accumulation and per-window running-sum
reduction dominate runtime. ICICLE is already close to the specialized
BLS12-377 path on this benchmark shape, so the next `plonk2` MSM work should
parallelize large buckets and window reductions before adding prover plumbing.

NTT benchmark commands:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkFFTForward_CUDA$' -benchtime=5x -count=1

go test -tags cuda ./gpu/plonk -run '^$' \
  -bench '^BenchmarkBLSFFTForward/n=1M$' -benchtime=5x -count=1
```

Results for 1M forward transforms:

| NTT | Time |
| --- | ---: |
| current `gpu/plonk` BLS12-377 | 402.9 us |
| `gpu/plonk2` BN254 | 538.7 us |
| `gpu/plonk2` BLS12-377 | 537.6 us |
| `gpu/plonk2` BW6-761 | 867.9 us |

The same `plonk2` NTT run through the ICICLE comparison harness reports
approximately 60 GB/s for BN254/BLS12-377 and 57 GB/s for BW6-761 at 1M.

ICICLE CUDA NTT remains blocked in this local comparison setup. The v3.2.2 Go
wrapper declares `get_root_of_unity` with the wrong ABI; the installed C++
symbol returns an error code and writes the root into an out-parameter. The
local harness works around that with a tiny cgo ABI shim. After that fix, and
after setting the CUDA `fast_twiddles` domain-init extension, `InitDomain`
returns success but the subsequent CUDA NTT aborts inside ICICLE with:

```
NTT size=1048576 is too large for the domain (size=0).
```

Because the process aborts from C++ rather than returning an error, there is no
defensible ICICLE CUDA NTT timing from this host/library combination yet. The
next useful step is to build a matching current ICICLE release with its CUDA
backend, or obtain the private/current CUDA backend for the `/tmp/icicle`
checkout and rerun the same harness.

## 2026-04-28 — MSM Architecture Tightening

The ICICLE comparison clarified the MSM design target. `gpu/plonk2` should keep
a generic affine short-Weierstrass input contract for every curve. The current
BLS12-377 twisted-Edwards path remains valuable as a specialized
`gpu/plonk` baseline, but it should be opt-in rather than the default shape of
the new backend.

Updated the MSM memory planner accordingly:

- `DefaultMSMPlanConfig` now defaults BLS12-377 to affine short-Weierstrass,
  like BN254 and BW6-761.
- `BLS12377CompactTEPlanConfig` explicitly models the existing compact
  twisted-Edwards path when we want to compare memory and throughput against
  `gpu/plonk`.
- `DESIGN.md` records the intended `plonk2` layering and benchmark contract.

Also tried a first bucket-range MSM reduction kernel. The idea was sound at the
algorithm level, but the CUDA template instantiation made `ptxas` impractically
slow on this codebase. That experiment was backed out. The next reduction pass
should be more compile-friendly before it is allowed to replace the current
correctness-first sequential reduction. In practice, that likely means a
specialized reduction module with fewer instantiated formulas, not another
large template block included through `api.cu`.

## 2026-04-28 — CPU vs GPU PlonK Reference Benchmarks

Added reference benchmarks for the all-curve PlonK path:

```
gpu/plonk2/bench_plonk_reference_test.go
gpu/plonk2/bench_plonk_reference_cuda_test.go
```

The CPU benchmark measures gnark PlonK setup and prove for the same private
multiplication-chain circuit on BN254, BLS12-377, and BW6-761. The CUDA
benchmark measures the exact setup-commitment slice that `plonk2` currently
accelerates: selector and permutation polynomial commitments from gnark's
trace, using CPU `MultiExp` versus the resident generic affine `G1MSM`.

The benchmark split is intentional. `plonk2` does not yet implement full GPU
proof generation; full GPU prover timings remain available only through the
current BLS12-377-specific `gpu/plonk` package. For all curves, the correct
comparison today is full CPU PlonK versus the setup-commitment acceleration
boundary that is already wired and validated.

CPU full PlonK benchmark command:

```
go test ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceCPU(Prove|Setup)/.*/constraints=1K$' \
  -benchtime=1x -count=1
```

Results on the RTX PRO 6000 Blackwell host CPU:

| Curve | CPU setup, 1K constraints | CPU prove, 1K constraints |
| --- | ---: | ---: |
| BN254 | 14.6 ms | 41.2 ms |
| BLS12-377 | 20.0 ms | 49.8 ms |
| BW6-761 | 58.1 ms | 129.5 ms |

CUDA setup-commitment benchmark command:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/.*/constraints=1K/(cpu|gpu)$' \
  -benchtime=1x -count=1
```

Results:

| Curve | CPU setup commitments, 1K | `plonk2` GPU setup commitments, 1K |
| --- | ---: | ---: |
| BN254 | 14.9 ms | 3.27 s |
| BLS12-377 | 19.4 ms | 9.31 s |
| BW6-761 | 52.9 ms | 7.02 s |

Small 16-constraint reference command:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/.*/constraints=16/(cpu|gpu)$' \
  -benchtime=1x -count=1
```

Results:

| Curve | CPU setup commitments, 16 | `plonk2` GPU setup commitments, 16 |
| --- | ---: | ---: |
| BN254 | 2.29 ms | 2.87 s |
| BLS12-377 | 3.32 ms | 8.64 s |
| BW6-761 | 8.99 ms | 4.55 s |

These numbers are deliberately poor for the generic GPU MSM. They confirm that
the benchmark catches the known bottleneck: per-commitment signed-window sort
plus mostly sequential bucket/window reduction. The next MSM implementation
work should target this benchmark directly and must keep the same CPU/GPU
commitment checks before reporting speedups.

## 2026-04-28 — Why ICICLE Is Much Faster

The ICICLE comparison should not be interpreted as evidence that affine inputs
are inherently too slow. ICICLE's BLS12-377 affine MSM is close to the current
`gpu/plonk` twisted-Edwards path:

| MSM | 16K | 64K |
| --- | ---: | ---: |
| current `gpu/plonk` TE MSM | 4.79 ms | 3.31 ms |
| ICICLE affine MSM | 5.79 ms | 6.09 ms |
| `gpu/plonk2` affine MSM | 1.69 s | 2.33 s |

The large gap is mostly implementation architecture:

- `plonk2` sorts signed-window assignments and then assigns one thread to each
  bucket. A large bucket is accumulated serially by one thread.
- `plonk2` reduces each window with one active thread. With a 16-bit window,
  that means roughly 32K running-sum bucket additions per window on a single
  thread.
- `plonk2` finalizes the MSM with one active thread. This is smaller than the
  bucket/window cost but confirms the current path is a correctness reference,
  not a throughput design.
- `plonk2` runs one commitment at a time. PlonK setup and proof phases need
  many commitments over the same SRS, so a production backend should batch
  work or at least reuse more scheduling/state across commitments.
- `plonk2` still performs Montgomery-layout correction after the GPU result is
  copied back to the host. This is not the main cost today, but it is the wrong
  boundary for a polished backend.

Updated note: the `icicle-gnark` checkout does include CUDA backend sources
under `icicle/backend/cuda`. A later source pass below replaces this inferred
diagnosis with a direct source audit.

Updated `DESIGN.md` with the MSM acceptance bar and the next backend shape:

1. keep the existing `G1MSM` Go/C ABI and affine short-Weierstrass input;
2. split non-empty buckets into fixed-size work items;
3. accumulate bucket partials with block-level reductions;
4. reduce partials per bucket with a compile-friendly second stage;
5. replace the one-thread window running sum with a parallel suffix-scan or
   segmented reduction;
6. move Montgomery correction onto the device;
7. add batched PlonK commitment APIs only after the single-commitment path is
   correct and materially faster.

The next code change should therefore not touch PlonK orchestration yet. It
should replace the private MSM reduction pipeline while keeping the reference
benchmarks and all gnark-crypto equality checks intact.

## 2026-04-28 — ICICLE CUDA Source Audit

Cloned `icicle-gnark` into `/tmp/icicle-gnark` and inspected commit `1a967ec`.
The actual CUDA backend is present under:

```
/tmp/icicle-gnark/icicle/backend/cuda/src/msm/cuda_msm.cuh
/tmp/icicle-gnark/icicle/backend/cuda/src/msm/cuda_msm.cu
/tmp/icicle-gnark/icicle/backend/cuda/include/msm/cuda_msm_config.cuh
/tmp/icicle-gnark/icicle/include/icicle/msm.h
/tmp/icicle-gnark/wrappers/golang/core/msm.go
/tmp/icicle-gnark/icicle/include/icicle/ntt.h
/tmp/icicle-gnark/icicle/backend/cuda/include/ntt/cuda_ntt_config.cuh
```

Important source-level findings:

- ICICLE's public `MSMConfig` has the knobs `plonk2` is missing:
  `precompute_factor`, window size `c`, scalar `bitsize`, `batch_size`,
  shared-base batching, host/device residency for scalars/points/results,
  Montgomery-form flags, async execution, stream selection, and backend
  extensions.
- The Go wrapper derives `BatchSize`, `ArePointsSharedInBatch`, and
  host/device flags from the input/output slice shapes before dispatch. This
  means batching and residency are part of the normal API rather than a
  prover-specific afterthought.
- CUDA-specific extension keys include `large_bucket_factor`,
  `is_big_triangle`, and `nof_chunks`. Defaults are in
  `backend/msm_config.h`; `large_bucket_factor` defaults to 10.
- ICICLE computes a default window size as approximately `log2(msm_size)-4`,
  capped at 20 in the current source because of large-bucket pressure.
- ICICLE estimates scalar/index/point/bucket memory using `cudaMemGetInfo`,
  reserves only about 70% of available memory, lowers `c` when bucket memory
  is too large, and computes a chunk count when the problem does not fit.
- When all data already lives on device, ICICLE can run a single chunk. When
  host data is involved, it may still choose 4 chunks to overlap transfer and
  computation.
- Scalar assignment grouping uses CUB radix sort, CUB run-length encoding, and
  CUB scans. `plonk2` currently sorts and then detects boundaries with a
  simple custom pass.
- ICICLE sorts buckets by descending bucket size. It computes how many buckets
  qualify as "large", splits large buckets into multiple segment work items,
  accumulates those segments on a separate stream, reduces variable-size
  partials, and distributes the reduced result back into the bucket array.
- ICICLE has two bucket-module reduction modes: a simple big-triangle mode and
  an iterative parallel reduction using `single_stage_multi_reduction_kernel`.
  `plonk2` currently performs one-thread running-sum reduction per window.
- ICICLE precomputes bases as a first-class operation and tests both batched
  MSM and batched shared-base MSM. This maps directly to PlonK commitments over
  the same SRS.
- ICICLE NTT exposes ordering, batch size, column-batch layout,
  host/device input and output flags, async execution, fast twiddles, and
  algorithm selection. `plonk2` can use these ideas to avoid extra bit-reversal
  and host/device transitions when integrating PlonK orchestration.

The practical design implication is sharper now. The first performance target
is not a full prover rewrite and not a coordinate-system switch. It is an MSM
pipeline refactor that copies the useful ICICLE boundaries while keeping
`plonk2` small and reviewable:

1. Add internal MSM run configuration and memory planning for windows, chunks,
   precomputation, batching, large-bucket threshold, and residency.
2. Move Montgomery correction onto the device.
3. Replace boundary-only bucket metadata with size-aware metadata and
   descending bucket-size ordering.
4. Add a large-bucket segmented accumulation path.
5. Replace one-thread window running sums with a parallel bucket-module
   reduction.
6. Add a private batched shared-SRS commitment entrypoint for PlonK setup and
   proof commitments.
7. Evaluate base precomputation only after the memory planner can protect
   BW6-761 from excessive static memory use.

Updated `DESIGN.md` with this execution plan and an NTT direction section.

The source audit itself executes the first plan item at the design level:
remove guesswork, identify the minimum useful architecture from ICICLE, and
pin the next code changes to testable milestones.

## 2026-04-28 — MSM Window Reduction Upgrade

Replaced the one-thread-per-window running-sum reduction in the generic affine
Pippenger MSM with a two-stage bucket-range reduction:

1. each window is split into a small internal number of bucket ranges;
2. one block reduces each range with the running-sum trick and emits the range
   total plus the range sum;
3. a second block per window scans the range sums and applies the correction
   needed to recover the same weighted bucket total as the sequential
   algorithm.

The `G1MSM` Go API is unchanged. The extra partial buffers are internal work
buffers and are pinned/released together with the existing scalar, sort,
bucket, and window buffers. This keeps the public tuning surface small while
removing the biggest obvious serial kernel from the resident commitment path.

One implementation lesson: the first direct block-scan port made `ptxas` too
slow on the three-curve template instantiation, especially with BW6-sized
field arithmetic. The final version isolates Jacobian add/double into device
call helpers for the reduction kernels. Runtime remains good enough for this
stage, and the CUDA build stays practical.

Correctness:

```
cmake --build gpu/cuda/build --target gnark_gpu -j2
go test -tags cuda ./gpu/plonk2 -run 'TestG1MSMPippengerRaw_CUDA|TestG1MSMCommitRaw_CUDA|TestCommitRawMatchesKZG|TestPlonkE2EGPUSetupCommitments_AllTargetCurves_CUDA' -count=1
go test -tags cuda ./gpu/plonk2 -count=1
```

Both Go test commands passed.

BLS12-377 MSM comparison after the reduction change:

```
CGO_LDFLAGS='-L/tmp/icicle-gnark-install/lib -Wl,-rpath,/tmp/icicle-gnark-install/lib' \
LD_LIBRARY_PATH=/tmp/icicle-gnark-install/lib \
ICICLE_BACKEND_INSTALL_DIR=/tmp/icicle-gnark-build-bls12_377/backend/cuda \
go test -tags 'cuda iciclebench' ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkCompareMSMBLS12377SRS' -benchtime=3x -count=1
```

| MSM | 16K | 64K |
| --- | ---: | ---: |
| current `gpu/plonk` TE MSM | 4.84 ms | 3.24 ms |
| ICICLE affine MSM | 5.79 ms | 6.11 ms |
| `gpu/plonk2` affine MSM | 6.76 ms | 8.09 ms |

The previous `plonk2` numbers on the same benchmark were 1.69 s and 2.33 s, so
this is the first generic MSM path that is within striking distance of the
specialized BLS12-377 and ICICLE paths for medium-size shared-SRS MSMs.

Reference PlonK setup-commitment benchmark:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkG1MSMCommitRawBLS12377|^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA' \
  -benchtime=3x -count=1
```

| Benchmark | CPU | GPU |
| --- | ---: | ---: |
| BN254 setup commitments, 16 constraints | 2.22 ms | 16.05 ms |
| BLS12-377 setup commitments, 16 constraints | 2.89 ms | 35.78 ms |
| BW6-761 setup commitments, 16 constraints | 9.19 ms | 242.25 ms |
| BN254 setup commitments, 1K constraints | 13.96 ms | 50.45 ms |
| BLS12-377 setup commitments, 1K constraints | 19.10 ms | 146.91 ms |
| BW6-761 setup commitments, 1K constraints | 53.04 ms | 1.18 s |

The setup-commitment path improved by orders of magnitude compared with the
earlier seconds-long numbers, but it is still not competitive with CPU for
small reference circuits. That is expected because the benchmark performs many
small independent commitments and the generic MSM still sorts all assignments,
accumulates each bucket mostly serially, and launches one MSM per commitment.

Next MSM work:

1. add size-aware bucket metadata and segment large buckets into fixed-size
   work items;
2. add a private batched shared-SRS commitment path for PlonK selector and
   permutation commitments;
3. keep BW6-761 memory pressure explicit before increasing window size or
   adding base precomputation;
4. only then wire the full PlonK prover orchestration to compare GPU against
   gnark CPU proving end to end.

## 2026-04-28 — Full-Prover Benchmark Harness

Added `BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA` as the first
end-to-end prover comparison harness. It deliberately benchmarks the existing
`gpu/plonk` BLS12-377 prover, not a new `plonk2` prover. This gives the
generic backend a concrete target while avoiding a speculative PlonK rewrite.

The benchmark builds the same `benchChainCircuit` used by the all-curve CPU
benchmarks, creates a gnark unsafe test SRS, runs gnark CPU PlonK prove, and
compares it with the current BLS12-377 GPU prover using the canonical SRS
converted to compact twisted-Edwards points.

Command:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA' \
  -benchtime=1x -count=1
```

Results:

| Prover | 16 constraints | 1K constraints |
| --- | ---: | ---: |
| gnark CPU BLS12-377 PlonK prove | 7.94 ms | 50.73 ms |
| current `gpu/plonk` BLS12-377 GPU prove | 16.53 ms | 40.41 ms |

The current GPU prover already wins at 1K on this small benchmark shape and
loses at 16 constraints where fixed GPU overhead dominates. `plonk2` should use
this harness as the full-prover comparison contract: first reproduce the
BLS12-377 current-GPU result with generic wrappers, then extend the same shape
to BN254 and BW6-761 once the prover orchestration is factored by curve.

Validation after adding the harness:

```
go test -tags cuda ./gpu/plonk2 -count=1
golangci-lint run --build-tags cuda ./gpu/plonk2
```

Both commands passed.

## 2026-04-28 — BW6-761 MSM Size Sweep

Focused on BW6-761 MSM sizing and memory pressure. The public `G1MSM` API
still exposes no extra config object, but the internal default window policy is
now size-aware for BW6-761:

| BW6-761 point count | window bits |
| ---: | ---: |
| `< 256K` | 13 |
| `256K .. < 4M` | 16 |
| `>= 4M` | 18 |

The motivation is simple: the original 13-bit BW6-761 default minimized bucket
memory but paid for 30 scalar windows. At larger MSM sizes, sort and assignment
traffic dominate more than bucket memory, and 16/18-bit windows are materially
better while still fitting comfortably on the 98 GB Blackwell GPU used here.

Added:

- `BenchmarkBW6761MSMCommitRawSizes_CUDA`
- `PLONK2_BW6_MSM_BENCH_SIZES`, for example `16K,64K,1M,30000000`
- `PLONK2_BW6_MSM_WINDOW_BITS`, for benchmark-only window overrides
- `TestPlanMSMMemory_BW6761ThirtyMillionPoints`

The 30,000,000 point memory plan selects an 18-bit window. Its modeled dominant
terms are roughly:

- affine points + scalars: 7.2 GB
- signed-window assignments: 630M
- assignment/sort workspace estimate: 15.12 GB
- Jacobian bucket accumulators: 0.79 GB
- total modeled dominant memory: 23.1 GB

The model is intentionally conservative but still well below the available
device memory on the current 98 GB GPU.

Benchmark commands:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA' -benchtime=2x -count=1

PLONK2_BW6_MSM_BENCH_SIZES=4M go test -tags cuda ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA' \
  -benchtime=1x -count=1

PLONK2_BW6_MSM_BENCH_SIZES=8M go test -tags cuda ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA' \
  -benchtime=1x -count=1

PLONK2_BW6_MSM_BENCH_SIZES=16M go test -tags cuda ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA' \
  -benchtime=1x -count=1

PLONK2_BW6_MSM_BENCH_SIZES=30000000 go test -tags cuda ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA' \
  -benchtime=1x -count=1
```

Results:

| Points | Window bits | Time |
| ---: | ---: | ---: |
| 16K | 13 | 62.15 ms |
| 64K | 13 | 138.32 ms |
| 256K | 16 | 200.61 ms |
| 1M | 16 | 613.63 ms |
| 4M | 18 | 2.02 s |
| 8M | 18 | 3.87 s |
| 16M | 18 | 7.85 s |
| 30,000,000 | 18 | 14.35 s |

Window override spot checks at 1M:

| Window bits | Time |
| ---: | ---: |
| 13 | 1.55 s |
| 16 | 617 ms |
| 18 | 601 ms |
| 20 | 678 ms |

The 30M run completed successfully. This validates that the current generic
BW6-761 MSM can handle the requested scale on this machine, but the throughput
is still only about 100 MB/s of scalar input. The next BW6 performance target
is therefore not window tuning; it is bucket accumulation and sort architecture:
size-aware bucket metadata, large-bucket segmentation, and eventually chunking
for GPUs with less memory.

Validation after the BW6 sweep:

```
go test -tags cuda ./gpu/plonk2 -run 'TestPlanMSMMemory_BW6761|TestPlanMSMMemory_BW6761ThirtyMillionPoints|TestG1MSMCommitRaw_CUDA/bw6-761' -count=1
go test -tags cuda ./gpu/plonk2 -count=1
golangci-lint run --build-tags cuda ./gpu/plonk2
```

All commands passed.

## 2026-04-29 — Prompt 00 Non-CUDA Thread-Local Baseline

Split `gpu/threadlocal.go` into a Linux implementation using `unix.Gettid`
and a non-Linux fallback that compiles without thread-local device pinning.
Linux keeps the per-OS-thread `SetCurrentDevice`, `CurrentDevice`, and device
ID maps used by multi-GPU workers. Non-Linux builds now return the default
device and no-op the setters, which is sufficient for non-CUDA tests.

Inspected `gpu/cuda/src/plonk2/msm.cu` and nearby `plonk2` CUDA cleanup
blocks. The expected duplicate `cudaFree(d_buckets)` was already absent in
this checkout; no repeated free was found in the local `plonk2` cleanup paths.

Commands:

```
gofmt -w gpu/threadlocal_linux.go gpu/threadlocal_other.go
go test ./gpu ./gpu/plonk2 ./gpu/plonk
```

Result:

```
?    github.com/consensys/linea-monorepo/prover/gpu        [no test files]
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.053s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA validation was not run on this host in this pass.

## 2026-04-29 — Generic Three-Curve Prover Preparation

Started the actual curve-generic prover port behind `plonk2.Prover` rather
than the temporary BLS12-377 bridge.

Implemented:

- CUDA `genericProverState` for BN254, BLS12-377, and BW6-761.
- Typed gnark constraint-system trace extraction into raw plonk2 field words.
- Resident canonical and Lagrange SRS MSMs for each curve.
- Resident FFT domain twiddles and permutation table.
- GPU preparation of fixed selector/permutation polynomials by uploading their
  Lagrange evaluations and converting them to canonical form with the generic
  FFT path.
- Direct solved-wire Lagrange commitment through the resident generic Lagrange
  SRS MSM.
- Constructor cleanup on strict-mode backend-preparation errors.

This is not yet full proof generation: quotient construction, Z construction,
linearized polynomial commitment, KZG openings, Fiat-Shamir phase ordering, and
typed proof assembly are still on the remaining integration path. For
BLS12-377, `WithEnabled(true)` can still use the older `gpu/plonk` bridge for
full-proof benchmarks; BN254 and BW6-761 currently prepare generic GPU state
and then use the configured CPU fallback for full proof generation.

Tests added:

- `TestGenericProverStatePreparesFixedCommitments_AllTargetCurves_CUDA`
- `TestGenericProverStateCommitsSolvedWireLagrangePolynomials_CUDA`
- `BenchmarkGenericSolvedWireCommitment_CUDA`

Commands:

```
gofmt -w gpu/plonk2/prove.go \
  gpu/plonk2/generic_prepare_cuda.go \
  gpu/plonk2/generic_prepare_cuda_test.go \
  gpu/plonk2/generic_prepare_stub.go \
  gpu/plonk2/bench_generic_prover_cuda_test.go

go test -tags cuda ./gpu/plonk2 -run 'TestGenericProverState' -count=1
go test ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk2 -count=1
golangci-lint run ./gpu/plonk2
golangci-lint run --build-tags cuda ./gpu/plonk2

PLONK2_PLONK_BENCH_CONSTRAINTS=1Ki go test -tags cuda ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkGenericSolvedWireCommitment_CUDA$' \
  -benchtime=1x -count=1
```

Results:

```
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.564s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.140s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 3.410s
0 issues.
0 issues.
```

Generic solved L-wire commitment benchmark, 1Ki constraints:

| Curve | Time/op |
|-------|--------:|
| BN254 | 2.964 ms |
| BLS12-377 | 73.509 ms |
| BW6-761 | 18.455 ms |

The BLS12-377 number is from the correctness-first generic affine MSM, not the
older specialized twisted-Edwards `gpu/plonk` MSM. The next implementation
step is to make the generic prover use the resident state for the full L/R/O
commitment phase with blinding and reduced-range MSMs, then wire Z and quotient
construction phase by phase.

## 2026-04-29 — Plonk2 Prover GPU Bridge and 17M BLS12-377 Benchmark

Added a temporary CUDA GPU backend hook behind `plonk2.Prover`:

- `WithEnabled(true)` on BLS12-377 prepares and proves through the existing
  `gpu/plonk` BLS12-377 GPU prover, using the public `plonk2.Prover` API.
- The bridge is intentionally labelled `legacy_bls12_377` in trace output. It
  is not the curve-generic prover port through `plonk2` primitives.
- BN254 and BW6-761 still return `plonk2: GPU prover not wired yet` in strict
  mode. With CPU fallback enabled, they continue through gnark CPU fallback.

Added:

- `BenchmarkPlonk2EnabledFullProverBLS12377_CUDA`
- `PLONK2_PLONK_BENCH_CONSTRAINTS`, accepted by CPU setup/prove, setup
  commitment, current GPU, and `plonk2` enabled full-prover benchmarks.
  Supported suffixes: decimal `K`/`M` and binary `Ki`/`Mi`.
- CUDA test coverage that verifies `WithEnabled(true)` produces and verifies a
  BLS12-377 proof and records the `gpu` trace phase.

Plan for the actual three-curve generic prover:

1. Port BLS12-377 from the legacy `gpu/plonk` orchestration to `plonk2`
   resident Fr/FFT/MSM/quotient primitives while preserving gnark verifier
   compatibility.
2. Isolate proof assembly and type conversions behind small curve adapters.
3. Enable BN254 on the same orchestration after BLS12-377 verifies.
4. Enable BW6-761 with the memory planner enforcing conservative chunking and
   per-phase buffer release.
5. Add large-size correctness/benchmark runs for all three curves once the
   generic path, not the legacy bridge, is wired.

Validation commands:

```
gofmt -w gpu/plonk2/prove.go gpu/plonk2/prove_gpu_cuda.go \
  gpu/plonk2/prove_gpu_stub.go gpu/plonk2/prover_cuda_test.go \
  gpu/plonk2/bench_plonk_reference_test.go \
  gpu/plonk2/bench_plonk_reference_cuda_test.go \
  gpu/plonk2/bench_full_prover_cuda_test.go

go test ./gpu/plonk2 \
  -run 'TestParsePlonkBenchConstraintCount|TestProver' -count=1

go test -tags cuda ./gpu/plonk2 \
  -run 'TestParsePlonkBenchConstraintCount|TestFullProverEnabledBLS12377GPUBackend_CUDA|TestFullProverDisabledFallbackTargetCurves_CUDA' \
  -count=1
```

Validation results:

```
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.095s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.377s
```

Small BLS12-377 enabled-prover benchmark:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonk2EnabledFullProverBLS12377_CUDA/bls12-377/constraints=16/plonk2-enabled$' \
  -benchtime=1x -count=1

BenchmarkPlonk2EnabledFullProverBLS12377_CUDA/bls12-377/constraints=16/plonk2-enabled-32  18622732 ns/op

PLONK2_PLONK_BENCH_CONSTRAINTS=1K go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonk2EnabledFullProverBLS12377_CUDA' \
  -benchtime=1x -count=1

BenchmarkPlonk2EnabledFullProverBLS12377_CUDA/bls12-377/constraints=1000/plonk2-enabled-32  35913431 ns/op
```

Large BLS12-377 enabled-prover benchmark:

```
PLONK2_PLONK_BENCH_CONSTRAINTS=17M go test -tags cuda ./gpu/plonk2 \
  -run '^$' \
  -bench '^BenchmarkPlonk2EnabledFullProverBLS12377_CUDA$' \
  -benchtime=1x -count=1 -timeout 60m
```

Result:

```
BenchmarkPlonk2EnabledFullProverBLS12377_CUDA/bls12-377/constraints=17000000/plonk2-enabled-32  21904620308 ns/op
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2  767.869s
```

The 17M run used a PlonK domain of 67,108,864 points. The timed proof phase was
about 21.9 seconds. The overall `go test` runtime was about 12m48s because
the benchmark builds the large circuit fixture and unsafe test SRS before the
timed proof iteration. Phase log excerpt from the timed iteration:

```
solve: 4.601s
iFFT L,R,O,Qk + blind: 6.241s
MSM commit L,R,O: 7.951s
build Z: 8.856s
iFFT+commit Z: 10.069s
quotient GPU: 14.807s
MSM commit h1,h2,h3: 16.601s
eval+linearize+open Z: 18.177s
MSM commit linPol: 18.763s
batch opening (GPU): 20.295s
```

Observed resource shape during setup: host RSS peaked above 50 GiB while the
GPU remained mostly idle until proof execution. Future 17M+ benchmark work
should reuse prebuilt fixtures/SRS or add a separate fixture-generation step so
repeated prover measurements are not dominated by setup.

## 2026-04-29 — Asset-backed CUDA Prover and MSM Tests

Narrowed SRS asset loading to the CUDA GPU prover and GPU MSM paths in
`gpu/plonk2`. Added a plonk2-local test SRS asset indexer for
`prover-assets/kzgsrs` that supports BN254, BLS12-377, and BW6-761
`kzg_srs_{canonical,lagrange}_*.memdump` files. It uses the same selection
policy as the existing `gpu/plonk` SRS store: canonical requests use the
smallest file with size at least the request, and lagrange requests require an
exact domain-size match.

Updated:

- CUDA setup-commitment E2E tests to use asset-backed PlonK setup.
- CUDA setup-commitment benchmarks to use asset-backed PlonK setup.
- BLS12-377 full GPU prover benchmarks/tests to use asset-backed setup and
  pinned TE SRS assets for the legacy `gpu/plonk` proving key.
- CUDA KZG/MSM correctness tests that previously generated small KZG SRS values
  with `kzg.NewSRS`.
- BN254 and BW6-761 CUDA MSM size benchmarks to use real canonical SRS points
  from assets instead of repeated synthetic base points.

Because the available lagrange assets start at domain 256, tiny CUDA prover
fixtures were moved from the old 4/64-point domains to benchmark-chain circuits
with at least a 256-point PlonK domain. CPU-only API tests remain unchanged
except for shared benchmark-size helpers.

Validation:

```
go test ./gpu/plonk2 \
  -run 'TestParsePlonkBenchConstraintCount' \
  -count=1

go test -tags cuda ./gpu/plonk2 \
  -run 'TestSRSAssetsContainTargetCurves|TestCommitRawMatchesKZGCommit_CUDA|TestG1MSMCommitRaw_CUDA|TestG1MSMCommitRawBatchRejectsMalformedInputs_CUDA' \
  -count=1

go test -tags cuda ./gpu/plonk2 \
  -run 'TestPlonkE2EGPUSetupCommitments_AllTargetCurves_CUDA|TestFullProverEnabledBLS12377GPUBackend_CUDA' \
  -count=1
```

Results:

```
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.004s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 1.682s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.643s
```

Full package checks after the asset-loader change:

```
go test ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk2 -count=1
```

Results:

```
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.142s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 3.193s
```

Benchmark smoke checks:

```
PLONK2_BN254_MSM_BENCH_SIZES=1Ki PLONK2_BW6_MSM_BENCH_SIZES=1Ki \
  go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench 'BenchmarkBN254MSMCommitRawSizes_CUDA|BenchmarkBW6761MSMCommitRawSizes_CUDA' \
  -benchtime=1x -count=1

PLONK2_BN254_MSM_DISABLE_CPU_FALLBACK=1 PLONK2_BW6_MSM_DISABLE_CPU_FALLBACK=1 \
  PLONK2_BN254_MSM_BENCH_SIZES=16Ki PLONK2_BW6_MSM_BENCH_SIZES=16Ki \
  go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench 'BenchmarkBN254MSMCommitRawSizes_CUDA|BenchmarkBW6761MSMCommitRawSizes_CUDA' \
  -benchtime=1x -count=1

PLONK2_PLONK_BENCH_CONSTRAINTS=128 go test -tags cuda ./gpu/plonk2 \
  -run '^$' \
  -bench '^BenchmarkPlonk2EnabledFullProverBLS12377_CUDA|^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254' \
  -benchtime=1x -count=1
```

Results:

```
BenchmarkBN254MSMCommitRawSizes_CUDA/n=1Ki-32   1367329 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=1Ki-32  6140770 ns/op

BenchmarkBN254MSMCommitRawSizes_CUDA/n=16Ki-32   3538781 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=16Ki-32 50428345 ns/op

BenchmarkPlonk2EnabledFullProverBLS12377_CUDA/bls12-377/constraints=128/plonk2-enabled-32 50515283 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254/constraints=128/cpu-32          2617299 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254/constraints=128/gpu-32          2575851 ns/op
```

## 2026-04-29 — Prompt 03 MSM Correctness Hardening

Centralized raw gnark-crypto layout assumptions in `raw.go`:
scalar word count, affine G1 word count, projective G1 word count, and shared
validation for exact and resident-SRS scalar buffers. `CommitRaw` and
`G1MSM.CommitRaw` now use these helpers instead of local word-count checks.

Montgomery correction remains host-side through `correctRawMontgomeryMSM`.
Moving it into CUDA was not done in this pass because it changes the point
normalization/arithmetic boundary and should be validated on a CUDA host.

Added CPU-only raw layout tests and CUDA-only MSM edge-case tests for zero
scalars, one-hot scalars, deterministic random-like scalars, repeated points,
short scalar slices against a longer SRS, structured large-bucket inputs,
malformed resident SRS/scalar buffers, and release/re-pin work-buffer reuse.

Commands:

```
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlan|Test.*Raw|Test.*Curve' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.106s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.115s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA validation and MSM benchmarks were not run on this host in this pass.

## 2026-04-29 — Prompt 04 MSM Run Plan Boundary

Added private `MSMRunPlan` and `MSMRunPlanConfig` for resident MSM handles.
The plan records curve, point count, scalar bit size, selected window, window
count, batch size, chunk point count, shared-base mode, precompute factor,
large-bucket factor, and the attached `MSMMemoryPlan`.

Default policy:

- BN254 and BLS12-377 keep 16-bit windows at representative setup sizes.
- BW6-761 keeps the existing size-aware 13/16/18-bit policy.
- Batch size defaults to 1 and shared-base batching remains internal only.
- Precomputation is disabled (`PrecomputeFactor=1`).
- Large-bucket factor is recorded as 4 for future segmentation.
- If an internal memory limit is supplied, the planner selects a chunk size
  whose chunk estimate fits the limit.

`NewG1MSM` now derives and stores this plan while preserving the existing
`CommitRaw` execution path. No CUDA bucket metadata or segmentation kernels
were changed in this pass; the existing accumulation path remains the only
runtime path.

Commands:

```
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlan|Test.*MSMRunPlan' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.003s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.120s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA correctness and benchmark validation were not run on this host in this
pass.

## 2026-04-29 — Prompt 05 Private Shared-Base Batch Commitments

Added private `(*G1MSM).commitRawBatch`, which accepts a wave of raw scalar
slices for one resident SRS and returns one raw projective commitment per
slice. The first implementation deliberately loops over the existing
`CommitRaw` path so the public API and CUDA ABI remain unchanged. Callers keep
work buffers pinned before the batch and release them after the whole wave.

Setup-commitment benchmark internals now use the private batch path for GPU
waves. `MSMMemoryPlan` records batch size metadata, and `MSMRunPlan` forwards
the internal batch size into the memory plan; peak memory remains modeled as
one active commitment because this implementation is still a sequential loop
over the resident handle.

Added CUDA-only malformed batch tests for empty batches, empty batch items,
and truncated scalar slices.

Commands:

```
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.118s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.119s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA correctness and setup-commitment benchmarks were not run on this host in
this pass.

## 2026-04-29 — Prompt 06 NTT Plan and Private Batch Surface

Added internal NTT planning types for order, residency, direction, and batch
count. The documented current order contract is:

- `FFT`: natural input to bit-reversed output.
- `FFTInverse`: bit-reversed input to natural output.
- `CosetFFT`: natural input to natural output.
- `CosetFFTInverse`: natural input to natural output.

Current residency is device-to-device because the public `FFTDomain` API
operates on `FrVector` values. Added a private `transformBatch` helper that
loops over the existing transform methods while checking plan curve, size, and
batch count. No CUDA ABI change was made.

Commands:

```
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'Test.*NTTPlan|Test.*FFTSpec' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.003s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.119s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA correctness, bit-reversal benchmarks, and old `gpu/plonk` NTT comparison
benchmarks were not run on this host in this pass.

## 2026-04-29 — Prompt 07 BLS12-377 Full-Prover API Coverage

Added BLS12-377 full-prover API coverage through the prepared `plonk2.Prover`
entrypoint:

- valid BLS12-377 witnesses prove and verify through CPU fallback;
- invalid BLS12-377 witnesses fail through the same API;
- CUDA-only tests assert that, with a CUDA device present, the BLS12-377
  prepared prover still uses CPU fallback by default and returns
  `plonk2: GPU prover not wired yet` when fallback is disabled.

No BLS12-377 full-prover phase was ported to `plonk2` GPU primitives in this
pass. Proof phases remain on gnark's CPU prover behind the fallback policy.
The old specialized `gpu/plonk` prover was not used as a substitute for the
generic `plonk2` path.

Commands:

```
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlonkE2E|Test.*Prover' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.114s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.120s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA full-prover tests and full-prover benchmarks were not run on this host in
this pass.

## 2026-04-29 — Prompt 09 Rollout Controls and Trace Metadata

Added rollout controls:

- `WithEnabled(bool)` to opt into attempting the GPU prover path.
- `WithCPUFallback(bool)` remains the CPU fallback control.
- `WithStrictMode(bool)` rejects fallback and returns
  `plonk2: GPU prover not wired yet`.
- `WithMemoryLimit` and `WithPinnedHostLimit` continue to be enforced during
  preparation.
- `WithTrace(path)` writes metadata-only JSONL events.

Trace events use `event=plonk2_prover` and include phase, curve, domain size,
commitment count, point count, estimated peak bytes, pinned host bytes, MSM
window bits, MSM chunk points, and fallback reason. They do not include witness
values, scalar contents, proof bytes, or transcript data.

Added package documentation for CUDA build tags and rollout options. No
downstream production call sites were changed.

Commands:

```
gofmt -w gpu gpu/plonk2
go test ./gpu/plonk2 ./gpu -run 'Test.*Trace|Test.*Fallback|Test.*Options' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.073s
?    github.com/consensys/linea-monorepo/prover/gpu        [no test files]
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 cached
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA rollout tests were not run on this host in this pass.

## 2026-04-29 — CUDA Validation and Benchmark Baseline

Validated the pending CUDA follow-up checklist from
`gpu/plonk2/prompts/README.md` on a CUDA host. The first `gpu/plonk2` CUDA
test run exposed a compile error in `gpu/plonk2/msm.go` where an existing
`err` variable was redeclared with `:=`; fixed it with a minimal assignment
change and reran the suite.

Hardware and toolchain:

```
GPU: NVIDIA RTX PRO 6000 Blackwell, 97887 MiB VRAM
Driver: 590.48.01
CUDA runtime: 13.1
CUDA compiler: nvcc 13.1.115
CPU: Intel(R) Xeon(R) Platinum 8559C, 32 vCPUs
Go: go1.26.0 linux/amd64
```

Correctness commands:

```
go test -tags cuda ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk -count=1
go test -tags cuda ./gpu/plonk2 -run 'TestCommitRaw|TestG1MSM|TestG1Affine|Test.*MSM.*CUDA' -count=1
go test -tags cuda ./gpu/plonk2 -run 'TestFrVectorOps_CUDA|TestFFT|TestCoset|Test.*NTT' -count=1
go test -tags cuda ./gpu/plonk2 -run 'TestG1MSMCommitRaw_CUDA|TestCommitRawMatchesKZG|TestPlonkE2EGPUSetupCommitments' -count=1
go test -tags cuda ./gpu/plonk2 -run 'Test.*Trace|Test.*Fallback|Test.*FullProver' -count=1
```

Correctness results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 4.658s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk  214.364s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.928s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.362s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.865s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.467s
```

Benchmark commands:

```
go test -tags cuda ./gpu/plonk2 -run '^$' -bench 'BenchmarkG1MSMCommitRaw|BenchmarkBW6761MSMCommitRawSizes' -benchtime=3x -count=1
go test -tags cuda ./gpu/plonk2 -run '^$' -bench 'BenchmarkFFTForward_CUDA|BenchmarkCosetFFTForward_CUDA' -benchtime=5x -count=1
go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA' -benchtime=3x -count=1
go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA|Benchmark.*Plonk2.*Full' -benchtime=3x -count=1
go test -tags cuda ./gpu/plonk -run '^$' -bench 'BenchmarkBLSFFT' -benchtime=5x -count=1
```

Selected benchmark results:

```
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=16Ki-32          64709428 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=64Ki-32         138101111 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=256Ki-32        201089041 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=1Mi-32          613584503 ns/op
BenchmarkG1MSMCommitRawBLS12377/n=16K-32                  6866653 ns/op
BenchmarkG1MSMCommitRawBLS12377/n=64K-32                  8112939 ns/op

BenchmarkFFTForward_CUDA/bn254-32                          547598 ns/op
BenchmarkFFTForward_CUDA/bls12-377-32                      528553 ns/op
BenchmarkFFTForward_CUDA/bw6-761-32                        845423 ns/op
BenchmarkCosetFFTForward_CUDA/bn254-32                     786581 ns/op
BenchmarkCosetFFTForward_CUDA/bls12-377-32                 794276 ns/op
BenchmarkCosetFFTForward_CUDA/bw6-761-32                  1532884 ns/op

BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254/constraints=16/cpu-32        2068663 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254/constraints=16/gpu-32       18019160 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bls12-377/constraints=16/cpu-32    3158458 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bls12-377/constraints=16/gpu-32   40338685 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bw6-761/constraints=16/cpu-32      9364079 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bw6-761/constraints=16/gpu-32    243241314 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254/constraints=1K/cpu-32       12762162 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254/constraints=1K/gpu-32       49255774 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bls12-377/constraints=1K/cpu-32   17980537 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bls12-377/constraints=1K/gpu-32  147508097 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bw6-761/constraints=1K/cpu-32     62931272 ns/op
BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bw6-761/constraints=1K/gpu-32   1182232508 ns/op

BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA/bls12-377/constraints=16/cpu-32           8117481 ns/op
BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA/bls12-377/constraints=16/current-gpu-32  17929198 ns/op
BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA/bls12-377/constraints=1024/cpu-32        52526646 ns/op
BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA/bls12-377/constraints=1024/current-gpu-32 40631478 ns/op

BenchmarkBLSFFTForward/n=16K-32    72439 ns/op
BenchmarkBLSFFTForward/n=64K-32   105601 ns/op
BenchmarkBLSFFTForward/n=256K-32  160821 ns/op
BenchmarkBLSFFTForward/n=1M-32    416954 ns/op
BenchmarkBLSFFTForward/n=4M-32   1757202 ns/op
BenchmarkBLSFFTvsCPU/GPU/n=1M-32  395423 ns/op
BenchmarkBLSFFTvsCPU/CPU/n=1M-32 9352127 ns/op
```

Observations:

- Setup-commitment GPU benchmarks are slower than CPU at the measured 16 and
  1K constraint sizes, so the current private Go-loop batch commitment path is
  still visible in same-host results.
- The current BLS12-377 full-prover GPU path is slower than CPU at 16
  constraints but faster at 1024 constraints in this run.
- The legacy `gpu/plonk` BLS FFT path continues to show strong GPU speedups
  versus CPU at the measured sizes.

## 2026-04-29 — BLS12-377 MSM Plonk vs Plonk2 Comparison

Added a focused CUDA benchmark comparing the existing `gpu/plonk` BLS12-377 MSM
backend against the generic `gpu/plonk2` BLS12-377 MSM backend on the same
canonical SRS and scalar dump. The benchmark excludes SRS/scalar loading and
MSM handle construction from timed regions; both paths pin reusable MSM work
buffers before timing.

Command:

```
go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA$' -benchtime=10x -count=1
```

Results:

```
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk/n=1K-32     1783733 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk2/n=1K-32    6007765 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk/n=4K-32     2786905 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk2/n=4K-32    6367651 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk/n=16K-32    4527971 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk2/n=16K-32   6745269 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk/n=64K-32    2856659 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk2/n=64K-32   7940113 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk/n=256K-32   5888187 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk2/n=256K-32 11619956 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk/n=1M-32    14157304 ns/op
BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA/gpu-plonk2/n=1M-32   26154869 ns/op
```

Observed `plonk2` slowdown versus `gpu/plonk`: 3.37x at 1K, 2.28x at 4K,
1.49x at 16K, 2.78x at 64K, 1.97x at 256K, and 1.85x at 1M.

## 2026-04-29 — BW6-761 MSM Extended Size Sweep

Extended the BW6-761 MSM benchmark sweep down to 1Ki and up to 4Mi. While
running the sweep, fixed the benchmark-size parser to accept the same `Ki` and
`Mi` suffixes printed by the benchmark names.

Command:

```
PLONK2_BW6_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki,256Ki,1Mi,2Mi,4Mi go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA$' -benchtime=5x -count=1
```

Results:

```
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=1Ki-32     41331261 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=4Ki-32     42416654 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=16Ki-32    62452214 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=64Ki-32   140092574 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=256Ki-32  200820709 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=1Mi-32    611811846 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=2Mi-32   1149738812 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=4Mi-32   2004696371 ns/op
```

Observed shape: BW6-761 MSM has a roughly 41 ms floor at tiny sizes, is about
63 ms at 16Ki, crosses 600 ms at 1Mi, and reaches about 2.0 s at 4Mi on this
host.

## 2026-04-29 — BW6-761 CPU MSM Baseline

Added a CPU benchmark for the same BW6-761 MSM size sweep. The CPU benchmark
uses the same repeated base-point shape and deterministic scalar elements as
the GPU benchmark, then calls gnark-crypto `G1Jac.MultiExp` directly. It
accepts the same `PLONK2_BW6_MSM_BENCH_SIZES` environment variable as the GPU
benchmark.

Command:

```
PLONK2_BW6_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki,256Ki,1Mi,2Mi,4Mi go test ./gpu/plonk2 -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizesCPU$' -benchtime=5x -count=1
```

Results:

```
BenchmarkBW6761MSMCommitRawSizesCPU/n=1Ki-32      1298380 ns/op
BenchmarkBW6761MSMCommitRawSizesCPU/n=4Ki-32      3723124 ns/op
BenchmarkBW6761MSMCommitRawSizesCPU/n=16Ki-32    14294844 ns/op
BenchmarkBW6761MSMCommitRawSizesCPU/n=64Ki-32    46172283 ns/op
BenchmarkBW6761MSMCommitRawSizesCPU/n=256Ki-32  184118900 ns/op
BenchmarkBW6761MSMCommitRawSizesCPU/n=1Mi-32    323333228 ns/op
BenchmarkBW6761MSMCommitRawSizesCPU/n=2Mi-32    496894365 ns/op
BenchmarkBW6761MSMCommitRawSizesCPU/n=4Mi-32    859800627 ns/op
```

After adding this benchmark, the GPU benchmark was aligned to use raw words
from the same deterministic field elements instead of arbitrary raw scalar
limbs.

Aligned GPU command:

```
PLONK2_BW6_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki,256Ki,1Mi,2Mi,4Mi go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA$' -benchtime=5x -count=1
```

Aligned GPU results:

```
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=1Ki-32      72497434 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=4Ki-32     193168454 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=16Ki-32    669701889 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=64Ki-32   2613069011 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=256Ki-32   279810700 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=1Mi-32     926451718 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=2Mi-32    1756388579 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=4Mi-32    1988564551 ns/op
```

Compared with the aligned same-host GPU sweep, CPU is faster at every measured
size: 55.8x at 1Ki, 51.9x at 4Ki, 46.8x at 16Ki, 56.6x at 64Ki, 1.5x at
256Ki, 2.9x at 1Mi, 3.5x at 2Mi, and 2.3x at 4Mi.

The aligned GPU sweep showed a large cliff below 256Ki. Rerunning small sizes
with the window override set to 16 reduced that cliff:

```
PLONK2_BW6_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki PLONK2_BW6_MSM_WINDOW_BITS=16 go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA$' -benchtime=5x -count=1

BenchmarkBW6761MSMCommitRawSizes_CUDA/n=1Ki-32    46205046 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=4Ki-32    50753700 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=16Ki-32   58433841 ns/op
BenchmarkBW6761MSMCommitRawSizes_CUDA/n=64Ki-32  106224768 ns/op
```

This points at the current BW6-761 default window policy as a major source of
the small-size slowdown.

## 2026-04-29 — BW6-761 MSM Slowness Investigation and Window Policy Fix

Investigated why BW6-761 MSM on `gpu/plonk2` was slow relative to CPU and to
BLS12-377 MSM. The main finding is that the previous BW6-761 window policy
(`13` below 256Ki points and `16` until 4Mi points) was a poor fit for the
generic short-Weierstrass bucket pipeline.

Important context from the implementation:

- `build_pairs_kernel` emits one key/value assignment per point per signed
  scalar window, including sentinel keys for zero digits.
- The pipeline sorts all `count * num_windows` assignments with CUB, detects
  bucket boundaries, then launches one sequential accumulator thread per
  bucket.
- BW6-761 uses 12-limb base-field Jacobian points, so each bucket addition is
  much more expensive than the 6-limb BLS12-377 path. Too few buckets creates
  long per-bucket sequential addition chains and poor GPU occupancy.

Hypotheses tried:

1. **Scalar distribution.** The earlier GPU benchmark used arbitrary raw scalar
   limbs. Added canonical scalar modes and made full-width field elements the
   default; the low64 sparse mode remains available through
   `PLONK2_BW6_MSM_SCALAR_MODE=low64`.
2. **CPU baseline.** Added `BenchmarkBW6761MSMCommitRawSizesCPU`, using the
   same repeated base points and deterministic scalar modes, to keep CPU/GPU
   comparisons local and reproducible.
3. **Window policy.** Swept BW6-761 CUDA window sizes. Windows `15` and `17`
   were pathological on this kernel. Window `14` was best through roughly
   256Ki points; window `18` was best from about 512Ki upward.

Window sweep excerpts, full-width canonical scalars:

```
# w=14
n=1Ki     40444065 ns/op
n=4Ki     39424997 ns/op
n=16Ki    51135779 ns/op
n=64Ki    81523920 ns/op
n=256Ki  196779398 ns/op
n=512Ki  346144814 ns/op
n=1Mi    661636960 ns/op
n=2Mi   1297117298 ns/op
n=4Mi   2521362340 ns/op

# w=15, aborted after confirming the cliff
n=16Ki    511757777 ns/op
n=64Ki   1798143388 ns/op
n=256Ki  7739387935 ns/op
n=1Mi   30042274549 ns/op

# w=16
n=16Ki    61714125 ns/op
n=64Ki   105619819 ns/op
n=256Ki  281536002 ns/op
n=1Mi    925790932 ns/op
n=2Mi   1750103644 ns/op
n=4Mi   3569846089 ns/op

# w=17, aborted after confirming the cliff
n=16Ki    287623728 ns/op
n=64Ki   1019102219 ns/op
n=256Ki  3846462173 ns/op

# w=18
n=1Ki     54350937 ns/op
n=4Ki     87163100 ns/op
n=16Ki    93140250 ns/op
n=64Ki    98678097 ns/op
n=256Ki  209173515 ns/op
n=512Ki  335921504 ns/op
n=1Mi    597421054 ns/op
n=2Mi   1080716396 ns/op
n=4Mi   1999371832 ns/op
```

Implemented policy:

```
BW6-761: window=14 below 512Ki points, window=18 at/above 512Ki points.
Other curves: unchanged at window=16.
```

Post-change default GPU command:

```
PLONK2_BW6_MSM_SCALAR_MODE=full PLONK2_BW6_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki,256Ki,512Ki,1Mi,2Mi,4Mi go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA$' -benchtime=3x -count=1
```

Post-change default GPU results:

```
n=1Ki     40383768 ns/op
n=4Ki     39093925 ns/op
n=16Ki    48016935 ns/op
n=64Ki    81888780 ns/op
n=256Ki  197213107 ns/op
n=512Ki  338048106 ns/op
n=1Mi    595628642 ns/op
n=2Mi   1079930507 ns/op
n=4Mi   1995584600 ns/op
```

Matching CPU command:

```
PLONK2_BW6_MSM_SCALAR_MODE=full PLONK2_BW6_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki,256Ki,512Ki,1Mi,2Mi,4Mi go test ./gpu/plonk2 -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizesCPU$' -benchtime=3x -count=1
```

Matching CPU results:

```
n=1Ki       4640555 ns/op
n=4Ki      13541662 ns/op
n=16Ki     29095165 ns/op
n=64Ki    101418801 ns/op
n=256Ki   415989542 ns/op
n=512Ki   625815386 ns/op
n=1Mi    1098905515 ns/op
n=2Mi    2088183583 ns/op
n=4Mi    4072872291 ns/op
```

Outcome:

- GPU is still slower than CPU at 1Ki, 4Ki, and 16Ki, where fixed kernel/sort
  overhead dominates.
- GPU overtakes CPU by 64Ki and is about 2x faster from 256Ki through 4Mi on
  this host.
- The main remaining optimization ideas are to compact non-zero window
  assignments before sorting, add a parallel large-bucket accumulator, and
  avoid short-Weierstrass bucket accumulation for BLS12-377-style specialized
  curves where a cheaper coordinate system exists.

## 2026-04-29 — BW6-761 MSM Phase Timing and Small-Size Hybrid Cutoff

Added per-phase timing for resident `plonk2` MSM handles, matching the older
`gpu/plonk` phase order:

```
h2d, build_pairs, sort, boundaries, accum_seq, accum_par,
reduce_partial, reduce_finalize, d2h
```

The CUDA library rebuild completed `libgnark_gpu.a`; the full CMake build then
failed while linking the existing sandbox executable due to a pre-existing
duplicate `fp_inv` symbol. Go CUDA tests linked against the rebuilt library
successfully.

Pure-GPU BW6-761 phase timing, full-width canonical scalars:

```
PLONK2_BW6_MSM_DISABLE_CPU_FALLBACK=1 PLONK2_BW6_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA$' -benchtime=3x -count=1

n=1Ki   40414020 ns/op  accum_seq=553us    reduce_partial=11672us d2h=24321us
n=4Ki   39105680 ns/op  accum_seq=2491us   reduce_partial=10769us d2h=22544us
n=16Ki  47855549 ns/op  accum_seq=11322us  reduce_partial=10563us d2h=22166us
n=64Ki  81656481 ns/op  accum_seq=45091us  reduce_partial=10741us d2h=22121us
```

At 1Mi with the measured window policy, the dominant phase is bucket
accumulation:

```
n=1Mi 595147417 ns/op accum_seq=516269us reduce_partial=47638us d2h=21242us
```

Implemented a conservative hybrid cutoff for BW6-761 resident and one-shot
commitments: below 64Ki points, use gnark-crypto CPU `MultiExp` instead of the
full GPU sort/bucket/finalize pipeline. This is controlled by
`PLONK2_BW6_MSM_DISABLE_CPU_FALLBACK=1` for pure-GPU benchmarking.

Hybrid default results:

```
PLONK2_BW6_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki go test -tags cuda ./gpu/plonk2 -run '^$' -bench '^BenchmarkBW6761MSMCommitRawSizes_CUDA$' -benchtime=3x -count=1

n=1Ki   4650069 ns/op
n=4Ki   13087958 ns/op
n=16Ki  29360300 ns/op
n=64Ki  81663335 ns/op
```

This gives more than 10x improvement against the original aligned default
GPU path at 1Ki, 4Ki, 16Ki, and 64Ki. Against the post-window-policy pure GPU
path, the hybrid cutoff is 8.7x faster at 1Ki, 3.0x faster at 4Ki, and 1.6x
faster at 16Ki; 64Ki remains on GPU because the GPU path is faster there.

Next pure-GPU ideas from the timing:

- Move the final Horner combination out of the one-thread GPU kernel or make it
  block-parallel; this is the ~22ms `d2h` floor at small sizes.
- Add a parallel bucket-accumulation path for BW6-761. At 1Mi, `accum_seq`
  is ~516ms of ~595ms, so a 10x pure-GPU large-size speedup requires replacing
  the current one-thread-per-bucket sequential accumulator.
- Explore precomputed per-base multiples or a distinct BW6 SRS benchmark. The
  current repeated-base benchmark is useful for phase isolation but is not a
  production SRS shape.

## 2026-04-29 — BN254 MSM Baseline, Phase Timing, and Small-Size Fast Paths

Added BN254-specific MSM size benchmarks matching the BW6-761 harness:

- `BenchmarkBN254MSMCommitRawSizes_CUDA`
- `BenchmarkBN254MSMCommitRawSizesCPU`
- `PLONK2_BN254_MSM_BENCH_SIZES`
- `PLONK2_BN254_MSM_WINDOW_BITS`
- `PLONK2_BN254_MSM_SCALAR_MODE`
- `PLONK2_BN254_MSM_DISABLE_CPU_FALLBACK=1`

Initial full-width scalar baseline with repeated base point:

```
PLONK2_BN254_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki,256Ki,512Ki,1Mi,2Mi,4Mi go test ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkBN254MSMCommitRawSizesCPU$' -benchtime=3x -count=1

n=1Ki      717600 ns/op
n=4Ki     1843313 ns/op
n=16Ki    4685088 ns/op
n=64Ki   15248746 ns/op
n=256Ki  47552970 ns/op
n=512Ki  94081270 ns/op
n=1Mi   165987251 ns/op
n=2Mi   269578113 ns/op
n=4Mi   503170340 ns/op
```

The initial pure-GPU BN254 path was already much faster than BW6-761:

```
PLONK2_BN254_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki,256Ki,512Ki,1Mi,2Mi,4Mi go test -tags cuda ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkBN254MSMCommitRawSizes_CUDA$' -benchtime=3x -count=1

n=1Ki      3156476 ns/op
n=4Ki      3393128 ns/op
n=16Ki     3622837 ns/op
n=64Ki     4292215 ns/op
n=256Ki    6549385 ns/op
n=512Ki    8364792 ns/op
n=1Mi     13114819 ns/op
n=2Mi     22581461 ns/op
n=4Mi     41966162 ns/op
```

Phase timing showed a small fixed GPU floor: roughly 1.2-1.4 ms in final
GPU combination / D2H reporting and roughly 1.3-1.5 ms in partial reduction
for the 16-bit window. At 1Mi, `accum_seq` was the largest phase at ~7.1 ms,
with H2D at ~1.8-1.9 ms.

Window sweep results:

- 8-bit windows improve the tiny pure-GPU repeated-base case
  (`1Ki` ~2.50 ms, `4Ki` ~3.24 ms), mostly by shrinking partial reduction.
- 8-bit windows are already worse by `8Ki` and much worse by `16Ki`.
- 10/12/14/18-bit windows were worse than 16-bit for representative medium
  and large sizes.

Implemented:

- BN254 CPU fallback below 8Ki points, disabled with
  `PLONK2_BN254_MSM_DISABLE_CPU_FALLBACK=1`.
- BN254 pure-GPU 8-bit window below 8Ki points for disabled-fallback runs.
- A repeated-base fast path inside the tiny BN254 CPU fallback:
  if all bases are equal, compute `base * sum(scalars)` instead of a full
  `MultiExp`. This is mathematically valid but mainly helps the current
  repeated-base diagnostic benchmark; real SRS-shaped inputs should not depend
  on it.

Default repeated-base benchmark after the fast paths:

```
PLONK2_BN254_MSM_BENCH_SIZES=1Ki,4Ki,16Ki,64Ki,256Ki,512Ki,1Mi,2Mi,4Mi go test -tags cuda ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkBN254MSMCommitRawSizes_CUDA$' -benchtime=3x -count=1

n=1Ki       69128 ns/op
n=4Ki      117644 ns/op
n=16Ki    3608380 ns/op
n=64Ki    4236136 ns/op
n=256Ki   6576333 ns/op
n=512Ki   8378526 ns/op
n=1Mi    13195801 ns/op
n=2Mi    22440327 ns/op
n=4Mi    42038666 ns/op
```

This is a 45.7x improvement at 1Ki and 28.8x at 4Ki versus the initial
BN254 GPU repeated-base baseline. Sizes at and above 16Ki stay on the GPU and
are essentially unchanged.

Forced pure-GPU after the size-aware window policy:

```
PLONK2_BN254_MSM_DISABLE_CPU_FALLBACK=1 PLONK2_BN254_MSM_BENCH_SIZES=1Ki,4Ki,8Ki,16Ki,64Ki,1Mi go test -tags cuda ./gpu/plonk2 \
  -run '^$' -bench '^BenchmarkBN254MSMCommitRawSizes_CUDA$' -benchtime=5x -count=1

n=1Ki    2470228 ns/op
n=4Ki    3205833 ns/op
n=8Ki    3474916 ns/op
n=16Ki   3516655 ns/op
n=64Ki   4187162 ns/op
n=1Mi   12972233 ns/op
```

The >10x result is therefore a default-path repeated-base result, not a
general pure-GPU algorithmic speedup.

Checked an SRS-shaped BN254 setup-commitment benchmark after the fallback:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254/constraints=16$' \
  -benchtime=3x -count=1

constraints=16/cpu   1761762 ns/op
constraints=16/gpu   2562612 ns/op

go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA/bn254/constraints=1K' \
  -benchtime=3x -count=1

constraints=1K/cpu  13564761 ns/op
constraints=1K/gpu  13242599 ns/op
```

Compared with the earlier setup-commitment baseline in this worklog, the
default BN254 setup path improved from ~16.05 ms to ~2.56 ms at 16 constraints
and from ~50.45 ms to ~13.24 ms at 1K constraints. That is useful, but it is
not yet 10x on SRS-shaped setup commitments.

Validation:

```
gofmt -w gpu/plonk2/commit.go gpu/plonk2/msm.go gpu/plonk2/msm_window.go gpu/plonk2/msm_plan_test.go \
  gpu/plonk2/bench_bn254_msm_cpu_test.go gpu/plonk2/bench_bn254_msm_cuda_test.go
go test ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk2 -run 'TestCommitRaw|TestG1MSM|TestPlanMSM|TestMSMRunPlan' -count=1
go test -tags cuda ./gpu/plonk2 -count=1
golangci-lint run
```

Results:

```
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.143s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.688s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 3.019s
0 issues.
```

Remaining BN254 optimization ideas:

- Add an actual BN254 SRS raw-size MSM benchmark, not just repeated-base and
  setup-commitment proxies.
- For SRS-shaped small commitments, consider a CPU batch path that computes
  the setup commitment wave in parallel instead of one `MultiExp` per
  commitment.
- For pure GPU, reduce the fixed small-size floor by specializing the partial
  reduction/finalization path for low point counts; the 16-bit path spends
  most tiny-input time reducing empty buckets.
- For large pure GPU, `accum_seq` is still the dominant phase. A parallel
  large-bucket accumulator remains the main algorithmic route to another
  large-size speedup.

## 2026-04-29 — FFT Size Sweep to 32Mi and BLS12-377 Plonk Comparison

Added CUDA FFT size-sweep benchmarks:

- `BenchmarkFFTForwardSizes_CUDA`
- `BenchmarkCosetFFTForwardSizes_CUDA`
- `BenchmarkCompareBLS12377FFTForwardPlonkVsPlonk2_CUDA`
- `BenchmarkCompareBLS12377CosetFFTPlonkVsPlonk2_CUDA`
- `PLONK2_FFT_BENCH_SIZES`
- `PLONK2_BLS_FFT_COMPARE_SIZES`

The default size ladder is now `1Mi,4Mi,16Mi,32Mi`. FFT domains require a
power-of-two size; `32Mi` is 33,554,432 points.

BN254 and BW6-761 forward/coset FFT sweep:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^Benchmark(FFTForwardSizes|CosetFFTForwardSizes)_CUDA$' \
  -benchtime=3x -count=1
```

Results:

| Operation | Curve | 1Mi | 4Mi | 16Mi | 32Mi |
|-----------|-------|-----|-----|------|------|
| FFT | BN254 | 613.9 us | 3.814 ms | 22.414 ms | 47.324 ms |
| FFT | BW6-761 | 1.010 ms | 7.346 ms | 34.120 ms | 72.508 ms |
| CosetFFT | BN254 | 860.8 us | 4.855 ms | 30.596 ms | 63.880 ms |
| CosetFFT | BW6-761 | 1.630 ms | 10.790 ms | 50.781 ms | 106.387 ms |

BLS12-377 direct comparison against the specialized `gpu/plonk` FFT:

```
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkCompareBLS12377(FFTForward|CosetFFT)PlonkVsPlonk2_CUDA$' \
  -benchtime=3x -count=1
```

Results:

| Operation | Backend | 1Mi | 4Mi | 16Mi | 32Mi |
|-----------|---------|-----|-----|------|------|
| FFT | `gpu/plonk` | 486.3 us | 1.838 ms | 8.557 ms | 20.374 ms |
| FFT | `gpu/plonk2` | 607.8 us | 4.035 ms | 22.259 ms | 47.360 ms |
| CosetFFT | `gpu/plonk` | 581.6 us | 2.448 ms | 14.765 ms | 32.067 ms |
| CosetFFT | `gpu/plonk2` | 856.0 us | 5.125 ms | 30.358 ms | 63.979 ms |

At 32Mi, the specialized BLS12-377 `gpu/plonk` path is 2.3x faster for forward
FFT and 2.0x faster for coset FFT. This matches the implementation shape:
`gpu/plonk2` currently launches one generic radix-2 kernel per stage, while
`gpu/plonk` uses a BLS12-377-specific radix-8 path, fused shared-memory tail,
and fused scale-plus-first-stage coset FFT.

Optimization ideas:

- Port the radix-8 stage grouping to the templated plonk2 field layer. This is
  the main forward FFT gap and should help BN254, BLS12-377, and BW6-761.
- Port the fused tail kernels generically. Large FFTs currently pay one global
  memory round trip per remaining stage in `plonk2`; the specialized path keeps
  the tail in shared memory.
- Add a plonk2 fused coset-forward launcher that combines
  `ScaleByPowersRaw(generator)` with the first DIF stage, then bit-reverses at
  the end. This should reduce the coset overhead visible at 32Mi.
- Add per-phase FFT timing to separate domain upload, scale, stages, bit
  reverse, and synchronization. Current benchmarks time only steady-state
  transform calls after setup.

Validation:

```
gofmt -l gpu/plonk2/bench_fft_sizes_cuda_test.go
go test ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk2 -run 'TestFFT|TestCosetFFT|Test.*NTTPlan' -count=1
go test -tags cuda ./gpu/plonk2 -count=1
golangci-lint run
```

Results:

```
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.141s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 2.314s
ok  github.com/consensys/linea-monorepo/prover/gpu/plonk2 3.025s
0 issues.
```

## 2026-04-29 — Prompt 10 Optional Specialization Evaluation

Evaluated BLS12-377 twisted-Edwards MSM as the single specialization
candidate. Added `gpu/plonk2/SPECIALIZATION.md` with the target bottleneck,
expected benefit, extra code surface, correctness risk, rollback plan, and
decision.

Decision: defer. No specialization was implemented because the generic
`plonk2` full prover is not wired yet and no same-hardware CUDA baseline was
taken in this pass. The generic affine short-Weierstrass path remains the only
`plonk2` MSM backend.

Commands:

```
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 cached
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA generic correctness and specialization benchmarks were not run on this
host in this pass.

## 2026-04-29 — Prompt 08 BN254 and BW6-761 Full-Prover API Coverage

Extended full-prover API coverage across BN254, BLS12-377, and BW6-761:

- valid witnesses prove and verify through `plonk2.Prover` CPU fallback;
- invalid witnesses fail through the same API for all target curves;
- CUDA-only tests cover default fallback and disabled-fallback errors for all
  target curves when a CUDA device is present.

No BN254 or BW6-761 full-prover phase was ported to `plonk2` GPU primitives in
this pass. The curve adapters currently remain the constructor-time type
switches and memory-plan metadata derived from typed gnark proving keys.

Commands:

```
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlonkE2E|Test.*Prover|Test.*Curve' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.119s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.127s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA full-prover tests and full-prover benchmarks were not run on this host in
this pass.

## 2026-04-29 — Prompt 02 Memory Planner Runtime Contract

Extended `MSMMemoryPlan` to match the current CUDA MSM buffers more closely:
resident affine SRS points, scalar staging, output staging, key/value pair
buffers, CUB sort estimate, bucket offsets/ends, bucket accumulators,
per-window results, and partial reduction buffers.

Added `ProverMemoryPlan` with prepared-key persistent memory, NTT domain
twiddles, quotient working vectors, MSM wave memory, pinned host buffers, and
total peak helpers. The peak estimate is intentionally conservative:
persistent bytes plus quotient scratch plus MSM wave scratch.

`NewProver` now derives domain size, commitment count, and SRS point count from
the typed gnark proving key, computes the plan, stores it on the prepared
prover, and enforces `WithMemoryLimit` and `WithPinnedHostLimit`.

Commands:

```
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlan|Test.*Memory|Test.*Prover' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.065s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.116s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA validation was not run on this host in this pass.

## 2026-04-29 — Prompt 01 Prepared Prover API Skeleton

Added the public `Prover`, `NewProver`, package-level `Prove`, and idempotent
`Close` skeleton for `gpu/plonk2`. The constructor validates that the gnark
constraint system and PlonK proving key are both one of BN254, BLS12-377, or
BW6-761 and that their curves match before any proof call can reach gnark's
type-switching dispatcher.

CPU fallback is enabled by default. With fallback enabled, `Prover.Prove`
delegates to `gnark/backend/plonk.Prove`. With fallback disabled, it returns
`plonk2: GPU prover not wired yet`.

Added CPU-only tests for supported curves, mismatched key/constraint curves,
unsupported curves, disabled fallback, idempotent close, and package-level
`Prove`.

Commands:

```
gofmt -w gpu/plonk2/options.go gpu/plonk2/prove.go gpu/plonk2/prover_api_test.go
go test ./gpu/plonk2 -run 'TestPlonkE2E|Test.*Prover|Test.*Fallback' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

Results:

```
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.105s
ok   github.com/consensys/linea-monorepo/prover/gpu/plonk2 0.115s
?    github.com/consensys/linea-monorepo/prover/gpu/plonk  [no test files]
```

CUDA validation was not run on this host in this pass.

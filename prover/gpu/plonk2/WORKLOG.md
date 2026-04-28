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
| Multi-GPU sanity test                               | planned  |
| Comprehensive benchmarks                            | partial  |
| Full GPU PlonK proof generation on all curves       | not wired |

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

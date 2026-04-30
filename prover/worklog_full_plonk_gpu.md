# Full Generic PlonK GPU Prover Worklog

Date: 2026-04-30

Scope: implement full generic GPU PlonK proof generation in `gpu/plonk2` for
BN254, BLS12-377, and BW6-761, then cross-validate every produced proof with
gnark's CPU verifier.

This is the next step after making the generic FFT/MSM kernels competitive and
after stabilizing the generic prepared-state orchestrator. The target is not a
partial acceleration path or CPU fallback wrapper. The target is a complete
generic GPU prover flow that returns typed gnark PlonK proofs accepted by the
existing gnark verifier for all supported curves.

## Non-Negotiable Acceptance Criteria

1. `NewProver(... WithEnabled(true), WithStrictMode(true))` must produce a GPU
   proof for BN254, BLS12-377, and BW6-761 without taking the CPU prover path.
2. Every GPU proof must pass `gnarkplonk.Verify` using the corresponding gnark
   CPU verifier and verifying key.
3. Tests must include both:
   - a plain arithmetic circuit with no BSB22 commitments
   - a commitment-bearing circuit, preferably the ECMul circuit shape from
     `gpu/plonk/plonk_test.go`, or a smaller equivalent that uses
     `frontend.Committer`
4. Commitment-bearing tests must assert that:
   - the compiled circuit has commitment metadata
   - the generated proof has non-empty `Bsb22Commitments`
   - CPU verification succeeds
5. Negative tests must prove the verifier catches invalid output:
   - invalid witness must fail proving
   - a tampered GPU proof must fail CPU verification
6. No phase may silently fall back to `gnarkplonk.Prove` in strict mode.
7. Memory lifecycle must be explicit:
   - resident state stays resident
   - MSM scratch is pinned only around commitment waves
   - MSM scratch is released before quotient work vectors are allocated
   - configured memory limits are enforced before large allocations
8. Benchmark and validation outputs must be preserved under `bench_vs_ingo`.

## Current State

Implemented and validated:

- Curve-generic Fr/Fp arithmetic for BN254, BLS12-377, BW6-761.
- Curve-generic FFT, coset FFT, and quotient helper kernels.
- Curve-generic short-Weierstrass G1 MSM and KZG commitments.
- Large-bucket parallel MSM accumulation for real PlonK wire distributions.
- Generic prepared state:
  - canonical SRS MSM
  - Lagrange SRS MSM
  - FFT domain
  - fixed selector polynomials
  - permutation table
- Solved-wire L/R/O commitment wave benchmarks at domains `1<<23` and
  `1<<24`.
- Legacy full GPU proof bridge for BLS12-377 only.

Completed in this worklog:

- Full generic proof orchestration for BN254, BLS12-377, and BW6-761.
- Typed proof assembly for all three curves, including L/R/O, Z, H, BSB22,
  shifted-Z opening, batched opening, and linearized polynomial commitments.
- Generic BSB22 commitment callback path using the generic Lagrange SRS MSM.
- Generic quotient/H path using the CUDA quotient kernels and generic
  canonical SRS MSM.
- All-curve CPU verifier cross-validation for plain and commitment-bearing
  circuits.
- Negative coverage for invalid witnesses and tampered GPU proofs.

Explicit non-default limitations:

- `backend.WithStatisticalZeroKnowledge()` returns an explicit unsupported
  error in the generic GPU backend. The default gnark prover configuration,
  solver options, custom Fiat-Shamir challenge hash, custom hash-to-field, and
  custom KZG folding hash are wired.
- Domain size `n < 6` still returns the existing GPU-not-wired error in the
  quotient phase. The validated smoke and benchmark circuits use domain sizes
  above that degenerate PlonK quotient layout.

## Progress Log

### 2026-04-30 — Baseline

- Confirmed the checked-out `gpu/plonk2` state already has generic CUDA
  preparation for all three target curves: canonical SRS MSM, Lagrange SRS
  MSM, FFT domain, fixed selector polynomials, permutation table, and Qcp
  uploads.
- Confirmed proof generation is still not generic. `NewProver(...,
  WithEnabled(true))` prepares generic state, then `newProverGPUBackend`
  selects only the legacy BLS12-377 `gpu/plonk` bridge; BN254 and BW6-761 fall
  back to CPU unless strict mode rejects fallback.
- Immediate implementation target: make the backend boundary explicit and
  traceable for the generic path, with strict mode never reaching
  `gnarkplonk.Prove`.

### 2026-04-30 — Phase 1 Backend Boundary

- Changed default enabled backend selection to `generic_gpu` for BN254,
  BLS12-377, and BW6-761.
- Kept the legacy BLS12-377 `gpu/plonk` bridge only behind the explicit
  `WithLegacyBLS12377Backend(true)` option.
- Added distinct trace phases:
  - `generic_gpu_prepare`
  - `generic_gpu_prove`
  - `generic_gpu_error`
  - `cpu_fallback`
- Updated strict CUDA tests to assert that the generic backend is selected for
  all three curves and that strict mode records no `cpu_fallback` event.
- Current limitation: `generic_gpu` proves nothing yet; it returns
  `errGPUProverNotWired` after the traced prove attempt. This is intentional
  scaffolding for the remaining proof phases and prevents the legacy BLS path
  from masking missing generic work.

### 2026-04-30 — Phase 2 Typed Operations Skeleton

- Added private typed proof-operation helpers for BN254, BLS12-377, and
  BW6-761 covering:
  - typed proof skeleton allocation with BSB22 slots
  - field slice to raw limb conversion
  - raw limb to typed field slice conversion
  - raw GPU projective commitment normalization into typed KZG digests
- Added CUDA unit tests for all three curves:
  - field raw round-trips
  - raw projective commitment to typed digest conversion
  - proof skeleton BSB22 allocation
- Focused validation passed:
  - `go test ./gpu/plonk2 -run 'TestProver|TestPlonkE2E' -count=1`
  - `go test ./gpu/plonk2 -tags cuda,nocorset -run 'TestFullProver|TestCurveProofOps|TestGenericPrepared|TestCommitRaw' -count=1 -timeout=30m`

### 2026-04-30 — Phase 3 Solver Preflight Partial

- Added a generic plain-circuit solver preflight before the current
  `errGPUProverNotWired` return. This is not the full BSB22 callback path yet,
  but it prevents invalid plain witnesses from being reported as a generic
  readiness failure in strict mode.
- Added strict CUDA coverage proving invalid plain witnesses fail before the
  unwired proof-phase error for BN254, BLS12-377, and BW6-761.
- BSB22 commitment callback work remains open; commitment-bearing generic
  proving is still not implemented.
- Full package validation after these edits:
  - `go test ./gpu/plonk2 -tags cuda,nocorset -count=1 -timeout=30m`
  - `golangci-lint run ./gpu/plonk2`

### 2026-04-30 — Phase 3/4/5 Artifact Pipeline Partial

- Replaced the plain preflight-only generic prove path with an internal
  artifact pass that:
  - solves BN254, BLS12-377, and BW6-761 witnesses;
  - overrides the BSB22 commitment hint for commitment circuits;
  - commits BSB22 Lagrange polynomials with the generic Lagrange SRS MSM;
  - stores solved L/R/O vectors in Lagrange form;
  - converts L/R/O to canonical form on GPU;
  - applies PlonK L/R/O blinding;
  - commits blinded L/R/O through the generic canonical SRS MSM;
  - derives `gamma` and `beta` from the gnark-compatible transcript inputs;
  - builds Z with `PlonkZComputeFactors`, batch inversion, and
    `ZPrefixProduct`;
  - converts/blinds/commits Z through the generic canonical SRS MSM;
  - derives `alpha` from BSB22 commitments plus Z using the gnark-compatible
    transcript ordering.
- Added a commitment-bearing all-curve CUDA artifact test using
  `frontend.Committer`. It asserts commitment metadata exists, BSB22 output is
  non-empty, L/R/O and Z commitments are non-zero, and intermediate raw
  artifacts are retained for later phases.
- Current limitation: quotient, H commitments, linearized polynomial, and KZG
  openings are still not wired, so `generic_gpu` still returns
  `errGPUProverNotWired` after producing these artifacts.
- Focused validation passed:
  - `go test ./gpu/plonk2 -tags cuda,nocorset -run 'TestGenericGPUBuildArtifactsCommitmentCircuit|TestFullProverStrictGeneric|TestCurveProofOps' -count=1 -timeout=30m`
- Full package validation after this artifact stage:
  - `go test ./gpu/plonk2 -tags cuda,nocorset -count=1 -timeout=30m`
  - `golangci-lint run ./gpu/plonk2`

### 2026-04-30 — Phase 6/7 Full Proof Assembly

- Added generic quotient assembly for BN254, BLS12-377, and BW6-761:
  - releases canonical and Lagrange MSM scratch before allocating quotient
    work vectors;
  - evaluates the PlonK permutation boundary, gate term, completed Qk, and
    BSB22 `Qcp/pi2` contribution on four cosets;
  - performs inverse coset FFT plus the decomposed size-4 inverse butterfly;
  - splits the quotient into `H[0]`, `H[1]`, and `H[2]`;
  - commits all three H shards through the generic canonical SRS MSM;
  - derives `zeta` from the verifier-compatible transcript.
- Added typed finalizers for all three curves:
  - evaluates L/R/O, S1/S2, Qcp, and shifted Z;
  - builds the linearized polynomial with the fixed Qk polynomial, while the
    completed Qk is used only for the quotient;
  - commits the linearized polynomial through generic GPU KZG;
  - builds shifted-Z and batched opening quotient polynomials on the host and
    commits those quotient polynomials through generic GPU KZG.
- Replaced the earlier CPU KZG opening calls in all curve finalizers. The path
  no longer calls `gnarkplonk.Prove`, `kzg.Open`, or `kzg.BatchOpenSinglePoint`
  for strict generic proofs.
- Added support for default gnark prover options, solver options, custom
  Fiat-Shamir challenge hash, custom BSB22 hash-to-field, and custom KZG
  folding hash. Statistical zero-knowledge remains explicitly unsupported
  because it changes quotient shard lengths and randomizer accounting.

### 2026-04-30 — Phase 8 Correctness Validation

- Updated strict generic CUDA tests so `WithEnabled(true)` and
  `WithStrictMode(true)` now prove and verify for BN254, BLS12-377, and
  BW6-761 without recording `cpu_fallback`.
- Added all-curve commitment-bearing tests using `frontend.Committer`. The
  tests assert:
  - compiled commitment metadata exists;
  - `Bsb22Commitments` length matches the compiled metadata;
  - BSB22, L/R/O, and Z commitments are non-zero;
  - gnark's CPU verifier accepts the GPU proof.
- Added negative strict-path coverage:
  - invalid witnesses fail during proving and are not hidden behind
    `errGPUProverNotWired`;
  - tampering a batched claimed value makes gnark's CPU verifier reject the
    GPU proof.
- Final validation commands passed:
  - `go test ./gpu/plonk2 -tags cuda,nocorset -run 'TestFullProverStrictGenericBackendDoesNotFallback|TestFullProverStrictGenericInvalidWitness|TestFullProverStrictGenericTamperedProofFailsVerification|TestGenericGPUProverVerifiesCommitmentCircuit|TestGenericGPUBuildArtifactsCommitmentCircuit' -count=1 -timeout=30m`
  - `go test ./gpu/plonk2 -tags cuda,nocorset -count=1 -timeout=30m`
  - `go test ./gpu/plonk2 -run 'TestProver|TestPlonkE2E' -count=1`
  - `golangci-lint run ./gpu/plonk2`

### 2026-04-30 — Phase 9 Benchmark Artifact

- Added a strict generic all-curve full-prover benchmark and preserved the
  latest raw output under:
  - `bench_vs_ingo/raw/plonk2_generic_full_prover_all_curves_128_20260430.txt`
- Added the parsed summary:
  - `bench_vs_ingo/generic_full_prover_summary.csv`
- Final smoke benchmark command:
  - `PLONK2_PLONK_BENCH_CONSTRAINTS=128 go test ./gpu/plonk2 -tags cuda,nocorset -run '^$' -bench '^BenchmarkPlonk2EnabledFullProverAllCurves_CUDA$' -benchtime=1x -count=1 -timeout=30m`
- Final smoke benchmark results:
  - BN254: `32.732 ms/proof`
  - BLS12-377: `78.386 ms/proof`
  - BW6-761: `78.634 ms/proof`

### 2026-04-30 — ECMul Domain `1<<24` Benchmark

- Added an all-curve ECMul benchmark modeled after `gpu/plonk`'s BN254
  emulated scalar-mul circuit.
- The benchmark compiles the BN254 ECMul circuit against BN254, BLS12-377, and
  BW6-761 outer scalar fields, adds an explicit BSB22 commitment, calibrates
  the instance count to a target domain, and proves in strict generic GPU mode.
- Default target domain is `1<<24`; it can be overridden with
  `PLONK2_ECMUL_TARGET_DOMAIN`.
- Final command:
  - `PLONK2_ECMUL_TARGET_DOMAIN=16Mi go test ./gpu/plonk2 -tags cuda,nocorset -run '^$' -bench '^BenchmarkPlonk2ECMulTargetDomainAllCurves_CUDA$' -benchmem -benchtime=1x -count=1 -timeout=180m`
- Raw output:
  - `bench_vs_ingo/raw/plonk2_ecmul_domain_1p24_all_curves_20260430.txt`
- Summary:
  - `bench_vs_ingo/ecmul_domain_1p24_summary.csv`
- Results for 121 ECMul instances, 16,764,643 constraints, two BSB22
  commitments, and domain 16,777,216:
  - BN254: `36.473 s/proof`
  - BLS12-377: `37.260 s/proof`
  - BW6-761: `66.847 s/proof`
- Validation after adding the benchmark:
  - `go test ./gpu/plonk2 -tags cuda,nocorset -count=1 -timeout=30m`
  - `golangci-lint run ./gpu/plonk2`

### 2026-04-30 — Legacy `gpu/plonk` ECMul Comparison

- Added `BenchmarkPlonkECMul121` to the legacy `gpu/plonk` benchmark list so
  the old BLS12-377-only prover can be tested on the same ECMul instance count
  as the plonk2 domain `1<<24` run.
- Command:
  - `go test ./gpu/plonk -tags cuda -run '^$' -bench '^BenchmarkPlonkECMul121$' -benchmem -benchtime=1x -count=1 -timeout=180m`
- Raw output:
  - `bench_vs_ingo/raw/gpu_plonk_legacy_ecmul_121_20260430.txt`
- Comparison summary:
  - `bench_vs_ingo/ecmul_domain_1p24_legacy_vs_plonk2_summary.csv`
- Legacy result:
  - BLS12-377-only `gpu/plonk`: `8.071 s/proof`, domain 16,777,216,
    16,764,640 constraints, 121 ECMul instances.
- Comparable plonk2 result:
  - BLS12-377 generic `gpu/plonk2`: `37.260 s/proof`, domain 16,777,216,
    16,764,643 constraints, 121 ECMul instances, two BSB22 commitments.
- Difference:
  - On this large ECMul circuit, the current generic plonk2 prover is about
    `4.62x` slower than the previous specialized BLS12-377 prover
    (`37.260 / 8.071`).
  - The setup times are similar for BLS12-377 (`25.12 s` plonk2 vs `25.04 s`
    legacy), and GPU preparation is also close (`9.017 s` plonk2 vs
    `7.043 s` legacy). The gap is therefore in the timed proof path.
  - The legacy prover is a specialized BLS12-377 implementation with a more
    fused and mature proof pipeline. The generic plonk2 path is curve-generic
    and verifier-compatible across BN254, BLS12-377, and BW6-761, but still
    pays more host-side allocation/copy/polynomial assembly overhead. The raw
    benchmark reflects this: plonk2 BLS12-377 reports about `49.36 GB/op`
    allocated versus about `5.15 GB/op` for legacy.
- Validation:
  - `golangci-lint run ./gpu/plonk`
  - A repeated cached `BenchmarkPlonkECMul121` run passed and measured
    `8.409 s/proof`.

### 2026-04-30 — Generic Prover Upgrade Plan

Baseline conclusion:

- The large BLS12-377 ECMul proof isolates the gap to the proof phase:
  `37.260 s` for generic `gpu/plonk2` versus `8.071 s` for legacy
  `gpu/plonk`.
- Compile/setup/GPU-prepare times are close enough that SRS preparation is not
  the first bottleneck.
- The proof-time allocation signal is not close: generic `gpu/plonk2` reports
  about `49.36 GB/op`, while legacy reports about `5.15 GB/op`.
- The current generic path repeatedly downloads GPU-resident fixed
  polynomials, converts raw limbs into typed field slices, scans large
  polynomials serially for evaluation/linearization/folding, and rebuilds
  temporary blinded coefficient slices through full typed conversions.

Target:

- First target: bring BLS12-377 generic `gpu/plonk2` ECMul `1<<24` proof time
  to at most `10 s`, which is within 25% of the previous specialized prover and
  gives the expected `>=4x` improvement over the current generic baseline.
- Secondary target: keep BN254 and BW6-761 correctness intact and avoid
  curve-specific proof logic beyond field/digest typing.
- Apples-to-apples constraint: the replacement must not move material proof
  work into setup/preprocessing. Setup and GPU preparation should remain
  comparable to the baseline; improvements must come from the timed proof path
  or from reusing data already resident in the prepared state.

Implementation plan:

1. Add a curve filter to the large ECMul benchmark.
   - `PLONK2_ECMUL_CURVES=bls12-377` must run only the BLS12-377 case so
     iteration time is low enough for repeated optimization.
   - Preserve the all-curve default for final benchmark artifacts.
2. Do not add setup-time fixed-polynomial precomputation.
   - A first attempted idea was to cache typed canonical selectors during
     `genericProverState` preparation. That would reduce timed proof cost but
     is not a valid drop-in comparison because it moves proof work into setup.
   - Correct direction: use the already-resident GPU fixed polynomials directly
     in proof-time kernels/evaluations, or keep any fixed-polynomial host work
     inside the timed proof until the GPU replacement is implemented.
3. Remove full-vector typed conversion from blinding.
   - Replace `blindCanonicalRaw` with a raw-limb path that copies canonical raw
     coefficients once, appends a few random blinders, and subtracts only the
     first two or three field elements through typed scalar operations.
   - Expected impact: eliminate several full length-`n` typed intermediate
     slices for L/R/O/Z blinding.
4. Parallelize the remaining host polynomial algebra.
   - Evaluate canonical polynomials with chunked parallel Horner, matching the
     legacy prover's CPU fallback structure.
   - Use parallel Horner quotient for shifted-Z and batched-opening quotient
     construction.
   - Parallelize linearized-polynomial assembly over coefficient chunks.
   - Parallelize batched-opening polynomial folding over coefficient chunks.
   - Expected impact: make the remaining unavoidable host scans scale with the
     100+ CPU environment instead of one goroutine.
5. Benchmark and inspect the residual gap.
   - Run focused BLS12-377 ECMul `1<<24` benchmark after each optimization
     batch and preserve raw output under `bench_vs_ingo/raw`.
   - If BLS12-377 remains above `10 s`, collect a CPU profile for the focused
     benchmark and move the largest remaining host loop to the existing generic
     GPU vector kernels.
6. Final validation.
   - Run focused correctness tests for strict generic proofs and commitment
     circuits.
   - Run full `go test ./gpu/plonk2 -tags cuda,nocorset`.
   - Run `golangci-lint run ./gpu/plonk2`.
   - Re-run all-curve ECMul `1<<24` benchmark and document the before/after
     results in this worklog.

### 2026-04-30 — Orchestration Review Before Further Optimization

Constraint from review:

- Do not get performance by hiding work in setup/preprocessing. A drop-in
  generic prover replacement must keep setup and GPU preparation comparable to
  the baseline. Proof-time wins should come from better orchestration, fewer
  redundant transfers, reuse of already-resident data, and kernels that replace
  proof-time host loops.

Measurement hygiene:

- Added `PLONK2_ECMUL_CURVES` so focused benchmark runs can select one curve.
- Added `PLONK2_ECMUL_INSTANCES` so the known 121-instance `1<<24` ECMul case
  can be rerun without expensive calibration compiles.
- Raw-blinding-only proof-time change measured:
  - Command:
    `PLONK2_ECMUL_TARGET_DOMAIN=16Mi PLONK2_ECMUL_INSTANCES=121 PLONK2_ECMUL_CURVES=bls12-377 go test ./gpu/plonk2 -tags cuda,nocorset -run '^$' -bench '^BenchmarkPlonk2ECMulTargetDomainAllCurves_CUDA$' -benchmem -benchtime=1x -count=1 -timeout=180m`
  - Raw output:
    `bench_vs_ingo/raw/plonk2_ecmul_domain_1p24_bls12377_raw_blinding_only_instances121_20260430.txt`
  - Result: `37.016 s/proof`, `45.07 GB/op`.
  - Baseline: `37.260 s/proof`, `49.36 GB/op`.
  - Conclusion: raw blinding removes several GiB/op but does not address the
    main latency gap.

Legacy `gpu/plonk` tricks that matter:

1. Persistent instance layout.
   - Legacy keeps both GPU-resident fixed polynomials (`dQl`, `dQr`, `dQm`,
     `dQo`, `dS1`, `dS2`, `dS3`, `dQkFixed`, `dQcp`) and typed host canonical
     views for CPU-side opening work.
   - Legacy pre-allocates per-proof host buffers (`l/r/o` canonical,
     blinded `l/r/o/z`, `qk`, `hFull`, `openZBuf`) and uses pinned host memory
     for hot H2D/D2H buffers.
   - Generic currently keeps fixed polynomials on GPU, but does not keep
     reusable proof buffers. It allocates large raw buffers and typed
     conversion buffers per proof.
   - Valid generic adaptation: add a reusable proof scratch object for
     proof-time transient buffers and pinned host staging. This does not move
     computation into setup; it only changes allocation ownership. Avoid adding
     setup-time fixed-polynomial D2H copies unless they replace existing setup
     work without increasing preparation time.

2. BSB22 handling.
   - Legacy converts each BSB22 polynomial to canonical form once during the
     commitment hint, commits it with the canonical MSM, and keeps the
     canonical `pi2` polynomial for quotient and opening phases.
   - Generic commits BSB22 in Lagrange form through the Lagrange SRS, then
     later canonicalizes the same polynomial for quotient/finalization. This
     duplicates proof-time work and requires a second resident MSM context.
   - Valid generic adaptation: canonicalize BSB22 once in the hint path, commit
     with the canonical MSM, and store canonical pi2 in the artifacts. This can
     remove the Lagrange MSM from the proof path and likely from generic setup.
     This is not cheating; it removes duplicated proof work and may reduce setup
     memory rather than increasing it.

3. Qk and per-proof polynomial residency.
   - Legacy patches Qk in Lagrange form, performs one GPU iFFT, and keeps the
     canonical Qk on the GPU when allocation succeeds. Quotient then uses D2D
     copies from that vector.
   - Generic canonicalizes Qk to host raw, then uploads it again at quotient
     start. Qk is not needed by the finalizer as the completed polynomial, so
     the host copy is avoidable.
   - Valid generic adaptation: return an optional `*FrVector` from Qk
     canonicalization and pass that device vector into quotient. Keep a host
     fallback only for tight VRAM.

4. Quotient coset loop.
   - Legacy offloads/reloads MSM points and releases MSM scratch during the
     quotient phase to reclaim VRAM.
   - Legacy opportunistically keeps `L/R/O/Z`, Qk, and BSB22 pi2 sources as
     device vectors, reducing them for each coset on GPU instead of reuploading
     host buffers.
   - Legacy opportunistically stores the first three quotient coset results on
     device and uses the fourth working vector for the last result, avoiding
     four D2H transfers followed by four H2D uploads before the decomposed
     inverse FFT.
   - Legacy has a pinned `hFull` fallback only when VRAM is too tight for the
     device-resident coset-result path.
   - Generic always starts quotient by uploading `L/R/O/Z/Qk/pi2` from host,
     always copies all four coset results D2H, always uploads them back for the
     inverse coset FFT, then copies the four blocks D2H again.
   - Valid generic adaptation, in priority order:
     - keep Qk/pi2 device-resident when possible;
     - keep L/R/O/Z blinded device-resident after canonicalization/blinding
       once the MSM path can consume device scalars or after one upload;
     - store coset results on device with host fallback;
     - add plonk2 MSM point offload/reload if quotient VRAM requires it;
     - add quotient phase profiling equivalent to legacy
       `GNARK_GPU_QUOTIENT_PROFILE`.

5. Finalization/openings.
   - Legacy evaluates GPU-resident fixed polynomials with `PolyEvalGPU`.
   - Legacy overlaps CPU Horner for host blinded wire polynomials with GPU
     evaluation of fixed polynomials.
   - Legacy builds the linearized polynomial on GPU using already-resident
     fixed polynomials plus two temporary `FrVector`s, then copies only the
     result back for commitment/opening.
   - Generic copies fixed polynomials D2H every proof, converts them to typed
     slices, evaluates all values on CPU, builds linearized polynomial on CPU,
     then converts it back to raw for a GPU MSM.
   - Valid generic adaptation:
     - port `PolyEvalGPU` to plonk2 for BN254, BLS12-377, and BW6-761;
     - compute `s1/s2/qcp` evaluations from resident fixed vectors;
     - build linearized polynomial with generic GPU vector operations over
       resident fixed polynomials;
     - keep a host result only where KZG opening assembly still requires it.

6. MSM API shape.
   - Legacy MSM has compact pinned SRS storage, chunking for very large SRS,
     point offload/reload, and variadic multi-commit waves.
   - Generic plonk2 MSM keeps raw affine points resident, has pin/release work
     buffers, but no point offload/reload and `commitRawBatch` currently loops
     over single commits.
   - Valid generic adaptation:
     - add device-scalar commitment entry points so a canonical `FrVector` can
       be committed without D2H plus H2D roundtrip;
     - add real batched shared-base commitment waves if the C API can amortize
       scalar upload/sort setup;
     - add point offload/reload only if quotient-resident vectors need the VRAM.

Current generic transfer map for the BLS12-377 ECMul `1<<24` case:

- One full BN254/BLS12-377 scalar vector at this domain is 512 MiB
  (`2^24 * 4 * 8`). One BW6-761 scalar vector is 768 MiB
  (`2^24 * 6 * 8`).
- The list below counts visible full-vector orchestration transfers and is a
  lower bound; it does not include smaller transcript copies, KZG digest
  downloads, GPU MSM internal staging, or typed host conversion allocations.
- BSB22 path with two commitments:
  - current generic path: two Lagrange MSM uploads, plus two uploads for later
    canonicalization, plus two canonical downloads;
  - lower bound: 6 full-vector transfers, about 3 GiB.
- L/R/O path:
  - upload solved L/R/O for iFFT, download canonical L/R/O, upload blinded
    L/R/O for KZG MSM;
  - lower bound: 9 full-vector transfers, about 4.5 GiB.
- Z path:
  - upload solved L/R/O again for Z, download Z in Lagrange form, upload Z for
    iFFT, download canonical Z, upload blinded Z for KZG MSM;
  - lower bound: 7 full-vector transfers, about 3.5 GiB.
- Completed Qk:
  - upload patched Qk for iFFT, download canonical Qk, then upload it again at
    quotient start;
  - lower bound: 3 full-vector transfers, about 1.5 GiB.
- Quotient input caching:
  - upload blinded L/R/O/Z and every canonical pi2 source before the coset
    loop;
  - with two BSB22 commitments, lower bound: 6 full-vector transfers, about
    3 GiB.
- Quotient coset staging:
  - current generic path downloads all four coset results, uploads all four
    blocks for inverse coset FFT, then downloads all four final blocks;
  - lower bound: 12 full-vector transfers, about 6 GiB;
  - legacy's device-resident coset-result path removes the first 8 of these
    12 transfers when VRAM allows it. The final H download remains needed
    until H commitment/opening assembly also becomes device-scalar aware.
- Finalization fixed-polynomial path:
  - current generic path downloads Ql/Qr/Qm/Qo/Qk/S1/S2/S3 and Qcp vectors
    every proof, converts them to typed field slices, evaluates them on CPU,
    and later uploads the derived linearized/opening polynomials for MSM;
  - with two Qcp vectors, the fixed-polynomial downloads alone are 10
    full-vector transfers, about 5 GiB.
- This makes the immediate memory-orchestration target clear: quotient coset
  staging and finalizer fixed-polynomial D2H dominate the avoidable vector
  traffic. BSB22 cleanup is still valid, but it is not the largest transfer
  win for the current two-commitment ECMul benchmark.

Immediate optimization plan from this review:

1. Add instrumentation first, not guesswork.
   - Add generic phase timings comparable to legacy labels:
     solve, completeQk, LRO canonical/blind, LRO MSM, build Z, Z canonical/MSM,
     quotient input caching, quotient coset loop, quotient inverse/split,
     H MSM, fixed eval, linearization, shifted-Z opening, linPol MSM, batch
     opening.
   - Add counters for D2H/H2D full-vector transfers by phase.
2. Port device-resident quotient coset-result storage.
   - Expected effect: removes the largest avoidable quotient transfer pair:
     four D2H coset chunks plus four H2D inverse-FFT uploads.
   - Keep the same best-effort/fallback structure as legacy so this remains a
     drop-in proof-time optimization under tight VRAM.
3. Implement device-resident Qk/pi2 sources in quotient.
   - Expected effect: removes the completed-Qk reupload and removes pi2
     reuploads inside the quotient path.
   - For BSB22 pi2, start by uploading canonical sources once before the coset
     loop, then use D2D copies per coset.
4. Port generic `PolyEvalGPU` and GPU linearized polynomial.
   - Expected effect: removes proof-time fixed-polynomial D2H conversions and
     shifts the biggest finalizer scans onto resident device vectors without
     changing setup/preprocessing.
5. Implement BSB22 canonical-once/device-scalar path.
   - Expected effect: removes Lagrange MSM dependence for commitment circuits
     and avoids re-uploading pi2 once generic MSM can consume device scalars.
   - This is lower priority than quotient/finalizer traffic for the current
     two-commitment ECMul benchmark, but it becomes important for circuits with
     many BSB22 commitments.

Implementation result: device-resident quotient coset results.

- Ported the legacy best-effort structure into generic plonk2 quotient
  assembly:
  - allocate three extra `FrVector`s when VRAM permits;
  - store coset results 0, 1, and 2 D2D;
  - keep coset result 3 in the working result vector;
  - run the four inverse coset FFTs directly from those device vectors;
  - preserve the host-staging fallback if the extra vectors cannot be
    allocated.
- This is a proof-time orchestration change only. It does not modify setup,
  preprocessing, proving keys, verifying keys, or circuit data.
- Validation:
  - `go test ./gpu/plonk2 -tags cuda,nocorset -run 'TestFullProverStrictGenericBackendDoesNotFallback|TestGenericGPUProverVerifiesCommitmentCircuit|TestFullProverStrictGenericTamperedProofFailsVerification' -count=1 -timeout=30m`
  - `golangci-lint run ./gpu/plonk2`
- Focused benchmark:
  - Command:
    `PLONK2_ECMUL_TARGET_DOMAIN=16Mi PLONK2_ECMUL_INSTANCES=121 PLONK2_ECMUL_CURVES=bls12-377 go test ./gpu/plonk2 -tags cuda,nocorset -run '^$' -bench '^BenchmarkPlonk2ECMulTargetDomainAllCurves_CUDA$' -benchmem -benchtime=1x -count=1 -timeout=180m`
  - Raw output:
    `bench_vs_ingo/raw/plonk2_ecmul_domain_1p24_bls12377_quotient_cosets_device_instances121_20260430.txt`
  - Result: `36.662 s/proof`, `45.07 GB/op`.
  - Previous raw-blinding result: `37.016 s/proof`, `45.07 GB/op`.
  - Original generic baseline: `37.260 s/proof`, `49.36 GB/op`.
  - Conclusion: the device-resident coset staging is correct and removes an
    obvious transfer pair, but it is only a small latency win in this benchmark.
    The remaining gap is dominated elsewhere, most likely finalizer host
    polynomial scans plus KZG/MSM scalar staging rather than this quotient
    staging alone.

Profile-guided result: BLS12-377 finalizer parallelization.

- CPU profile after device-resident quotient cosets:
  - Raw profile:
    `bench_vs_ingo/raw/plonk2_ecmul_domain_1p24_bls12377_after_cosets_20260430.cpu.pprof`
  - Raw output:
    `bench_vs_ingo/raw/plonk2_ecmul_domain_1p24_bls12377_after_cosets_profile_20260430.txt`
  - The proof path spent most remaining CPU time in `finalizeBLS12377`:
    fixed-polynomial D2H/typed conversion, repeated Horner evaluations,
    linearized-polynomial assembly, and batch-opening folding.
- Ported legacy CPU parallelization tricks into the generic BLS12-377
  finalizer:
  - chunked parallel polynomial evaluation;
  - chunked parallel Horner quotient for opening polynomials;
  - coefficient-parallel linearized-polynomial assembly;
  - coefficient-parallel batch-opening fold.
- This is still proof-time work only. It does not change setup/preprocessing.
- Validation:
  - `go test ./gpu/plonk2 -tags cuda,nocorset -run 'TestFullProverStrictGenericBackendDoesNotFallback|TestGenericGPUProverVerifiesCommitmentCircuit|TestFullProverStrictGenericTamperedProofFailsVerification' -count=1 -timeout=30m`
  - `golangci-lint run ./gpu/plonk2`
- Focused benchmark:
  - Raw output:
    `bench_vs_ingo/raw/plonk2_ecmul_domain_1p24_bls12377_parallel_finalizer_instances121_20260430.txt`
  - Result: `23.708 s/proof`, `45.07 GB/op`.
  - Previous quotient-coset result: `36.662 s/proof`.
  - Original generic baseline: `37.260 s/proof`.
  - Legacy BLS12-377 baseline: `8.071 s/proof`.
  - Conclusion: CPU finalizer work was a major bottleneck. The generic prover
    is now `1.57x` faster than the original generic baseline on this benchmark,
    but still `2.94x` slower than legacy. The next pass must target the
    remaining large pieces: fixed-polynomial D2H/typed conversion, Qk/BSB22
    duplicate canonicalization, and scalar staging through KZG/MSM.

## Architecture Target

`gpu/plonk2` should own one generic backend:

```text
Prover
  genericGPUBackend
    genericProverState
      resident canonical SRS MSM
      resident Lagrange SRS MSM
      resident FFT domain
      resident fixed polynomials
      resident permutation table
    curveProofOps
      typed proof allocation
      typed field conversions
      typed transcript bindings
      typed verifier-compatible KZG digests/openings
```

The generic backend should replace the current `newProverGPUBackend` behavior
that only wires BLS12-377 through `gpu/plonk`.

The design should keep curve-specific code small and mechanical. Curve-specific
code is acceptable for typed gnark data structures and field conversions, but
the proof phase logic should be shared.

## Phase 1 - Backend Boundary

### Actions

1. Replace the BLS-only backend selection in `gpu/plonk2/prove_gpu_cuda.go`.
2. Introduce `genericGPUBackend` implementing `proverGPUBackend`.
3. Keep the existing legacy BLS backend available only as an explicit fallback
   or comparison path during development.
4. Make strict mode fail if the generic backend cannot prove.
5. Add trace events that distinguish:
   - `generic_gpu_prepare`
   - `generic_gpu_prove`
   - `generic_gpu_error`
   - `cpu_fallback`

### Acceptance Criteria

- BN254, BLS12-377, and BW6-761 all select `genericGPUBackend`.
- Strict mode never calls `gnarkplonk.Prove`.
- Existing CPU fallback behavior remains available only when strict mode is
  disabled.
- A test can assert the backend label or trace event is generic for all curves.

## Phase 2 - Typed Curve Operations

### Actions

Define a private `curveProofOps` interface covering only typed operations that
cannot be cleanly generic:

- allocate an empty typed proof
- read/write L/R/O/Z/H commitments
- read/write BSB22 commitments
- read/write claimed values and opening proofs
- marshal transcript inputs exactly as gnark expects
- convert raw GPU projective commitments to typed `G1Affine`
- convert typed field slices to raw `[]uint64`
- convert raw `[]uint64` back to typed field slices where needed
- expose typed verifying-key metadata needed by the prover

Implement one `curveProofOps` for each curve:

- BN254
- BLS12-377
- BW6-761

### Acceptance Criteria

- Unit tests can round-trip:
  - field slices to raw and back
  - raw GPU commitment output to typed digest
  - typed proof skeleton allocation for all curves
- Transcript marshaling is byte-for-byte compatible with gnark for a small
  reference proof.
- No `any` or reflection is used in hot loops.

## Phase 3 - Solver and BSB22 Commitment Callback

### Actions

1. Port the witness solving flow from `gpu/plonk/prove.go`.
2. Handle `constraint.PlonkCommitments` generically.
3. Override the BSB22 commitment hint:
   - collect committed values from solver inputs
   - inject random blinding values at the commitment and last constraint slots
   - commit with the generic canonical SRS MSM
   - bind the commitment digest into the Fiat-Shamir transcript
   - return the derived commitment value to the solver
4. Store:
   - solved L/R/O vectors
   - commitment values
   - BSB22 commitment digests
   - `pi2` canonical polynomials needed for quotient/linearization

### Acceptance Criteria

- Plain circuits with no commitments solve correctly.
- Commitment-bearing circuits solve correctly.
- `len(proof.Bsb22Commitments) == len(ccs.CommitmentInfo)` for commitment
  circuits.
- Invalid witness returns an error before proof verification.
- Commitment callback output matches gnark CPU prover behavior for a small
  deterministic circuit, modulo expected prover randomness.

## Phase 4 - L/R/O Commitment Wave

### Actions

1. Convert solved L/R/O from Lagrange form to canonical form on GPU.
2. Apply PlonK blinding to L/R/O.
3. Commit blinded L/R/O with the generic canonical SRS MSM.
4. Pin MSM scratch for the wave and release after the wave.
5. Record per-wire timing metrics in trace/debug mode.

### Acceptance Criteria

- GPU L/R/O commitments match CPU KZG commitments for a deterministic small
  circuit when using fixed blinding.
- Produced proof has non-zero L/R/O commitments.
- Strict memory lifecycle test confirms MSM scratch can be released after the
  wave and re-pinned later.

## Phase 5 - Z Polynomial

### Actions

1. Reuse generic kernels for permutation ratio factors:
   - numerator
   - denominator
   - batch inversion
   - ratio multiplication
2. Build `Z` with the generic prefix-product path.
3. Convert `Z` to canonical form.
4. Blind and commit `Z`.
5. Derive `alpha` from transcript using exactly gnark-compatible inputs:
   - BSB22 commitments, if any
   - `Z` commitment

### Acceptance Criteria

- GPU `Z` values match CPU reference for a small circuit before blinding.
- GPU `Z` commitment matches CPU KZG for fixed blinding.
- Transcript-derived `alpha` matches gnark CPU prover for a deterministic
  harness where randomness is controlled.

## Phase 6 - Quotient Polynomial and H Commitments

### Actions

1. Reuse the existing generic quotient kernels for:
   - wire coset FFTs
   - selector coset FFTs
   - permutation boundary term
   - gate term
   - BSB22 `Qcp/pi2` contribution
2. Split the quotient into `h1`, `h2`, `h3`.
3. Commit `h1`, `h2`, `h3` with canonical SRS MSM.
4. Release quotient work vectors before post-quotient opening commitments.

### Acceptance Criteria

- For small circuits, GPU quotient polynomial matches CPU reference at random
  points.
- `H[0]`, `H[1]`, and `H[2]` commitments are non-zero and verifier-compatible.
- Memory peak stays within `ProverMemoryPlan`.

## Phase 7 - Openings and Linearized Polynomial

### Actions

1. Derive `zeta` and `zeta * omega`.
2. Evaluate:
   - L, R, O at `zeta`
   - S1, S2 at `zeta`
   - Qcp values at `zeta`
   - Z at `zeta * omega`
3. Build the linearized polynomial.
4. Commit the linearized polynomial.
5. Build KZG opening proofs:
   - batched opening at `zeta`
   - shifted Z opening at `zeta * omega`
6. Assemble claimed values and opening digests into the typed proof.

### Acceptance Criteria

- CPU verifier accepts the full GPU proof for plain circuits on all curves.
- Tampering any of L/R/O/Z/H/linearized/opening proof fields makes CPU
  verification fail.
- KZG opening commitments produced by GPU match gnark CPU KZG opening
  commitments in deterministic small tests.

## Phase 8 - Commitment-Bearing ECMul Validation

### Actions

1. Reuse or copy the ECMul circuit shape from `gpu/plonk/plonk_test.go`.
2. Compile it for each target outer curve where feasible:
   - BLS12-377 already exists
   - BN254 and BW6-761 should use the same non-native BN254 scalar-mul shape
     if gnark supports it cleanly for those scalar fields
3. If ECMul is too heavy or unsupported for one curve, add a smaller circuit
   that explicitly calls `frontend.Committer.Commit` and still exercises
   BSB22 commitment plumbing.
4. For each curve:
   - setup SRS and proving key
   - prove with generic GPU prover in strict mode
   - verify with gnark CPU verifier
   - assert non-empty proof commitments

### Acceptance Criteria

- At least one all-curve test uses a commitment-bearing circuit.
- BLS12-377 specifically uses the ECMul test shape from the legacy GPU tests,
  unless a smaller equivalent is justified in the test comment.
- `proof.Bsb22Commitments` is non-empty for the commitment test.
- CPU verifier accepts every GPU proof.
- Invalid witness fails proving.
- Tampered proof fails verification.

## Phase 9 - Performance and Stability Benchmarks

### Actions

1. Add full-prover benchmarks:
   - plain arithmetic circuit
   - commitment-bearing circuit
2. Sizes:
   - smoke: small domains up to `1<<16`
   - stability: `1<<20`
   - large: `1<<23`, `1<<24`
3. Record:
   - total prove time
   - solve time
   - L/R/O commitment wave
   - Z build/commit
   - quotient
   - H commitment wave
   - opening/linearization
   - peak planned memory
   - actual device memory before/after, when available
4. Save raw outputs and summaries under `bench_vs_ingo`.

### Acceptance Criteria

- Large benchmarks complete without memory leaks or device OOM on the RTX PRO
  6000 class GPU used for prior work.
- Repeated prove runs are stable: no monotonic VRAM growth after proof close.
- Results are preserved in:
  - `bench_vs_ingo/raw`
  - a parsed summary CSV
  - a short worklog update

## Phase 10 - Multi-GPU Follow-Up

This is not a blocker for single-GPU correctness, but the implementation must
not prevent it.

### Actions

1. Add an optional test skipped unless `LIMITLESS_GPU_COUNT>=2`.
2. Create one `Prover` per GPU.
3. Pin each goroutine to its OS thread.
4. Bind each goroutine to its selected `gpu.Device`.
5. Prove independent witnesses concurrently.
6. Verify both proofs with CPU verifier.

### Acceptance Criteria

- No CUDA work falls through to device 0 for the second prover.
- Both proofs verify.
- Closing one prover does not invalidate the other prover's resident state.

## Implementation Order

1. Build generic backend skeleton and typed `curveProofOps`.
2. Make BN254 plain circuit prove and verify.
3. Make BN254 commitment circuit prove and verify.
4. Repeat for BLS12-377.
5. Repeat for BW6-761.
6. Remove or quarantine legacy BLS backend from default strict path.
7. Add benchmarks only after all verifier tests pass.

This order avoids hiding proof-assembly bugs behind the existing BLS legacy
backend and avoids optimizing before verifier compatibility is established.

## Test Commands

Expected focused commands during implementation:

```bash
go test ./gpu/plonk2 -tags cuda,nocorset \
  -run 'TestGenericGPUProve.*CUDA|TestGenericGPUProveCommitments.*CUDA' \
  -count=1 -timeout=30m

go test ./gpu/plonk2 -tags cuda,nocorset -count=1 -timeout=30m

golangci-lint run ./gpu/plonk2
```

Large benchmarks should be run separately and saved:

```bash
PLONK2_PLONK_BENCH_CONSTRAINTS=4194302,8388606 \
go test ./gpu/plonk2 -tags cuda,nocorset -run '^$' \
  -bench 'BenchmarkGenericGPUProve' \
  -benchmem -benchtime=3x -count=1 -timeout=180m
```

## Risk Register

### Transcript Compatibility

Risk: a GPU proof can be internally consistent but rejected by gnark verifier if
Fiat-Shamir bindings differ by one byte or one ordering rule.

Mitigation: add deterministic small tests comparing transcript challenge
material with a gnark CPU proof harness.

### Typed Proof Assembly

Risk: raw GPU projective outputs are valid but assigned to the wrong typed proof
field or not normalized as expected.

Mitigation: add per-phase tests that compare each digest/opening against CPU
KZG output before running full verifier tests.

### BSB22 Commitments

Risk: plain circuits verify but commitment circuits fail because the solver
callback, `Qk` patch, `Qcp` quotient contribution, or transcript binding is
wrong.

Mitigation: make the commitment-bearing test mandatory before considering a
curve complete.

### Memory Overlap

Risk: pinning MSM scratch while quotient vectors are resident causes large
domains to OOM.

Mitigation: encode explicit phase pin/release calls and assert memory plan
limits before prove.

### Legacy Backend Masking

Risk: BLS12-377 tests pass through the old `gpu/plonk` path while BN/BW are not
actually generic.

Mitigation: strict tests must assert the generic backend path was used.

## Definition of Done

The work is done when all of the following are true:

- BN254 generic GPU proof verifies with gnark CPU verifier.
- BLS12-377 generic GPU proof verifies with gnark CPU verifier.
- BW6-761 generic GPU proof verifies with gnark CPU verifier.
- At least one commitment-bearing test verifies for every curve.
- Invalid witnesses fail.
- Tampered proofs fail verification.
- Strict mode does not call CPU prover.
- Focused CUDA tests pass.
- `golangci-lint run ./gpu/plonk2` passes.
- Raw benchmark data and summaries are saved under `bench_vs_ingo`.

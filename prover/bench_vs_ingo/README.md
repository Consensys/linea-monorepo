# plonk2 vs ICICLE GPU Benchmarks

This folder preserves the April 30, 2026 comparison between Linea
`gpu/plonk2` and `/home/ubuntu/dev/cpp/icicle-gnark`.

## Environment

- GPU: NVIDIA RTX PRO 6000 Blackwell Server Edition, 97,887 MiB
- CPU: Intel Xeon Platinum 8559C
- Curves: BN254, BLS12-377, BW6-761
- Operations: forward FFT/NTT and G1 MSM
- Sizes: `1 << 15` through `1 << 23`
- Benchmark shape: `go test -bench ... -benchtime=3x -count=5`
- Runs were serialized so benchmark processes did not share the GPU.

ICICLE install to `/usr/local/lib` failed because that directory is root-owned,
so the ICICLE benchmarks linked and loaded libraries from:

```text
/home/ubuntu/dev/cpp/icicle-gnark/icicle/build
/home/ubuntu/dev/cpp/icicle-gnark/icicle/build/backend
```

## Files

- `gpu_perf_summary.csv` contains parsed median/mean/CV data.
- `raw/plonk2_all_curves_bench.txt` contains raw `gpu/plonk2` output.
- `raw/icicle_bls12377_bench.txt` contains raw ICICLE BLS12-377 output.
- `raw/icicle_bn254_bench.txt` contains raw ICICLE BN254 output.
- `raw/icicle_bw6761_bench.txt` contains raw ICICLE BW6-761 output.
- `raw/bw6761_msm_optimization_attempts.txt` contains the raw correctness and
  benchmark output from the BW6-761 MSM optimization attempts.
- `raw/bw6761_msm_32bit_ptx_20260430.txt` contains the raw output from the
  implemented BW6-761 32-bit/PTX Montgomery multiply optimization.
- `BW6761_MSM_OPTIMIZATION_WORKLOG.md` records the BW6-761 MSM design reset
  and the clean optimization plan.
- `raw/plonk2_fft_optimized_bn_bw_20260430.txt` contains the optimized BN254
  and BW6-761 FFT sweep.
- `raw/plonk2_fft_optimized_bls12377_20260430.txt` contains the optimized
  BLS12-377 FFT sweep and the legacy `gpu/plonk` comparison.
- `fft_optimized_summary.csv` contains parsed median optimized FFT results and
  ratios against the original `plonk2` and ICICLE runs.
- `FFT_OPTIMIZATION_WORKLOG.md` records the FFT design, implementation,
  experiments, and remaining work.
- `GENERIC_PLONK_ORCHESTRATOR_WORKLOG.md` records the generic prover-state
  orchestration work, large-bucket MSM fix, memory model, and `1<<23` /
  `1<<24` circuit benchmarks.
- `generic_plonk_orchestrator_summary.csv` contains parsed large-circuit
  generic PlonK wave results.

## Median Comparison

Ratio is `plonk2 / ICICLE`. Values below `1.0x` mean `plonk2` was faster.

| Curve | Operation | 1<<15 | 1<<20 | 1<<23 | Conclusion |
|---|---:|---:|---:|---:|---|
| BLS12-377 | FFT/NTT | 1.37x | 1.34x | 2.60x | `plonk2` slower |
| BLS12-377 | MSM | 1.11x | 0.65x | 0.97x | competitive; faster mid-size |
| BN254 | FFT/NTT | 1.38x | 1.31x | 2.58x | `plonk2` slower |
| BN254 | MSM | 0.47x | 0.38x | 0.61x | `plonk2` faster |
| BW6-761 | FFT/NTT | 0.57x | 1.08x | 2.05x | `plonk2` slower at large sizes |
| BW6-761 | MSM | 2.27x | 6.19x | 7.75x | `plonk2` much slower |

## Findings

- FFT/NTT is the consistent weakness. `plonk2` currently uses generic radix-2
  stage-per-kernel NTTs. The older/specialized path and ICICLE use more
  optimized NTT strategies such as grouped radix stages and fused tails.
- BN254 MSM is already faster than ICICLE across the tested range.
- BLS12-377 MSM is competitive: slightly slower at the smallest/largest points,
  but faster in the middle of the tested range.
- BW6-761 MSM is the major gap. Phase timings show `accum_seq` dominates,
  caused by one-thread-per-bucket sequential bucket accumulation over expensive
  12-limb base-field Jacobian additions.

## BW6-761 MSM Optimization Update

Implemented a BW6-761 base-field Montgomery multiplication specialization in
`gpu/cuda/src/plonk2/field.cuh`. It uses 24 32-bit limbs internally and PTX
carry-chain multiply-adds for the hot CIOS step. This keeps the Pippenger
algorithm and memory policy unchanged.

Latest single-run large-size results:

| Size | Previous plonk2 | Optimized plonk2 | ICICLE | Status |
|---:|---:|---:|---:|---|
| `8Mi` | ~2.78 s | 0.689 s | ~0.355 s | ~1.9x slower |
| `16Mi` | ~4.89 s | 1.206 s | 0.525 s | ~2.3x slower |
| `32Mi` | n/a | 2.254 s | 0.879 s | ~2.6x slower |

Conclusion: this is a roughly 4x plonk2 improvement on the large BW6 MSM path,
but it does not yet beat ICICLE. The next required step is a generated
ICICLE-style BW6 field backend, including specialized square, rather than more
window or bucket-threshold tuning.

## FFT Optimization Update

Implemented a curve-generic radix-8 NTT path in `gpu/plonk2`, plus fused
shared-memory tail stages for large transforms and inverse-scale fusion. The
design is the generic version of the optimized local `gpu/plonk` BLS12-377 NTT
strategy.

Latest median forward FFT/NTT results:

| Curve | Size | Previous plonk2 | Optimized plonk2 | ICICLE | Status |
|---|---:|---:|---:|---:|---|
| BN254 | `1<<23` | 10.329 ms | 4.146 ms | 4.004 ms | on par, 3.5% slower |
| BLS12-377 | `1<<23` | 10.321 ms | 4.091 ms | 3.969 ms | on par, 3.1% slower |
| BW6-761 | `1<<23` | 15.745 ms | 8.785 ms | 7.686 ms | improved, still 14.3% slower |

The small and mid-size ladder is improved or unchanged for many points, but
BN254/BLS12-377 still trail ICICLE below large sizes because ICICLE's
mixed-radix implementation uses deeper 16/32/64 stage grouping and fast-twiddle
layouts. Details and rejected experiments are in `FFT_OPTIMIZATION_WORKLOG.md`.

## Follow-Up Plan

1. Add a generic mixed-radix NTT pass with 16/32/64 stage grouping and
   fast-twiddle layouts, matching the remaining ICICLE FFT advantage.
2. Add fused coset scale plus first-stage FFT for coset transforms.
3. Reduce BW6-761 bucket-add cost with generated/specialized 12-limb G1
   arithmetic. The generic MSM now has bounded parallel large-bucket
   accumulation, but BW6 remains dominated by expensive 12-limb additions.
4. Reduce the small-size BW6-761 floor by avoiding the current final
   Horner/device-to-host overhead where possible.

## Change Made During Investigation

The BW6-761 CPU fallback cutoff in `gpu/plonk2/commit.go` was lowered from
`1 << 16` to `1 << 14`, because stable reruns showed the GPU path was already
faster at 16Ki and 32Ki points for the SRS-shaped benchmark. A cutoff test was
added in `gpu/plonk2/msm_fallback_cuda_test.go`.

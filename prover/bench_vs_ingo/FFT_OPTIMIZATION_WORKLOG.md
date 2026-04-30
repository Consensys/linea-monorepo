# FFT Optimization Worklog

Date: 2026-04-30

## Goal

Improve `gpu/plonk2` forward FFT/NTT performance for BN254, BLS12-377, and
BW6-761 against the saved ICICLE comparison in this folder. The main target was
the large-size regression at `1 << 22` and `1 << 23`, where the original
`plonk2` implementation paid one global-memory pass and one kernel launch per
radix-2 stage.

## Design Notes

The relevant state-of-the-art and local references point to the same design
direction:

- ICICLE exposes radix-2 and mixed-radix CUDA NTTs, recommends mixed radix for
  larger NTTs, and uses fast-twiddle domain mode to trade memory for speed.
- ICICLE's CUDA mixed-radix source decomposes large transforms into 16/32/64
  stage groups with explicit twiddle layouts.
- The existing `gpu/plonk` BLS12-377 NTT already uses radix-8 stage grouping,
  fused shared-memory tail stages, and inverse-scale fusion.

The implemented step ports the proven local `gpu/plonk` strategy into the
curve-generic `gpu/plonk2` field layer:

1. Fuse three DIF stages into a generic radix-8 forward kernel.
2. Fuse three DIT stages into a generic radix-8 inverse kernel.
3. Fuse the last large-transform stages into shared memory, with tail size
   selected from the actual scalar limb count and CUDA opt-in shared memory.
4. Fuse inverse `1/n` scaling into the final inverse stage.
5. Keep radix-8 behind curve/size cutoffs because it regresses small BW6-761
   transforms where 6-limb register pressure dominates.

## Files

- `gpu/cuda/src/plonk2/kernels.cu`
- `bench_vs_ingo/raw/plonk2_fft_optimized_bn_bw_20260430.txt`
- `bench_vs_ingo/raw/plonk2_fft_optimized_bls12377_20260430.txt`
- `bench_vs_ingo/fft_optimized_summary.csv`

## Results

Median timings, milliseconds:

| Curve | Size | Original plonk2 | Optimized plonk2 | ICICLE | Opt / ICICLE |
|---|---:|---:|---:|---:|---:|
| BN254 | `1<<15` | 0.101842 | 0.100480 | 0.073983 | 1.36x |
| BN254 | `1<<20` | 0.586480 | 0.524533 | 0.448334 | 1.17x |
| BN254 | `1<<23` | 10.328827 | 4.145966 | 4.004068 | 1.04x |
| BLS12-377 | `1<<15` | 0.100973 | 0.097344 | 0.073678 | 1.32x |
| BLS12-377 | `1<<20` | 0.584245 | 0.519645 | 0.436850 | 1.19x |
| BLS12-377 | `1<<23` | 10.321469 | 4.090625 | 3.968979 | 1.03x |
| BW6-761 | `1<<15` | 0.147325 | 0.142433 | 0.257417 | 0.55x |
| BW6-761 | `1<<20` | 0.936314 | 0.931554 | 0.869755 | 1.07x |
| BW6-761 | `1<<23` | 15.745164 | 8.784728 | 7.686204 | 1.14x |

Large BN254 and BLS12-377 are now effectively on par with ICICLE. BW6-761 is
substantially faster than before but still 14% slower at `1 << 23`.

## Experiments Rejected

- Lowering the fused-tail threshold to small sizes reduced kernel launches but
  made 32K-512K transforms slower because each tail block uses high shared
  memory and many threads.
- A 512-thread regular NTT launch was invalid for the current register-heavy
  generic kernels.
- A 32-bit/PTX Montgomery multiplication specialization for `BW6761FrParams`
  was correct but slower than the current 64-bit generic multiply in FFTs, so it
  was not kept.
- A forward radix-16 kernel reduced the number of global passes but was slower
  at 4M and 8M points because local storage for 16 field elements caused too
  much register pressure in the generic field layer.

## Remaining Work

To beat ICICLE across the whole ladder, `plonk2` needs the next mixed-radix
step rather than more thresholds:

1. Add a generic mixed-radix NTT pass that groups 4-6 stages per global pass,
   matching ICICLE's 16/32/64 block strategy.
2. Generate per-field multiplication/square code for 4-limb and 6-limb scalar
   fields instead of relying on generic `unsigned __int128` CIOS in NTT hot
   loops.
3. Precompute fast-twiddle layouts for the grouped kernels so thread-level
   butterflies avoid strided twiddle gathers.
4. Add a dedicated small-NTT path that fuses several stages without the large
   shared-memory tail occupancy cost.
5. Add a plonk2 fused coset-forward entrypoint that combines coset scaling with
   the first DIF stage.

## Verification

Commands run:

```text
cmake --build gpu/cuda/build --target gnark_gpu -j
go test ./gpu/plonk2 -tags cuda,nocorset -run 'TestFrVectorOps_CUDA|TestFFT|TestCoset|Test.*NTTPlan|TestG1MSMPippengerRaw_CUDA' -count=1 -timeout=30m
PLONK2_FFT_BENCH_SIZES=32768,65536,131072,262144,524288,1048576,2097152,4194304,8388608 go test ./gpu/plonk2 -tags cuda,nocorset -run '^$' -bench '^BenchmarkFFTForwardSizes_CUDA$' -benchtime=3x -count=5 -timeout=60m
PLONK2_BLS_FFT_COMPARE_SIZES=32768,65536,131072,262144,524288,1048576,2097152,4194304,8388608 go test ./gpu/plonk2 -tags cuda,nocorset -run '^$' -bench '^BenchmarkCompareBLS12377FFTForwardPlonkVsPlonk2_CUDA$' -benchtime=3x -count=5 -timeout=60m
```

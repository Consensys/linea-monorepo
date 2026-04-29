# Prompt 06: NTT Plan and Batched Transforms

## Goal

Add internal NTT planning for order and residency, then implement the smallest
useful batched transform surface. The objective is to prevent unnecessary bit
reversals and host transfers before full prover integration.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk2/fft.go`
- `gpu/plonk2/fr_fft_cuda_test.go`
- `gpu/cuda/src/plonk2/kernels.cu`
- `gpu/cuda/src/plonk/ntt.cu`
- `gpu/plonk/fft.go`
- `gpu/plonk/prove.go`, especially iFFT and quotient coset usage.

## Constraints

- Keep current public `FFT`, `FFTInverse`, `BitReverse`, `CosetFFT`, and
  `CosetFFTInverse` behavior unchanged.
- Add planning as internal structure first.
- Do not optimize by changing mathematical order unless tests prove equality.
- Do not add public NTT tuning knobs.

## Implementation Tasks

1. Add internal enum types:
   order: natural or bit-reversed;
   residency: host or device;
   direction: forward, inverse, coset-forward, coset-inverse.
2. Add `NTTPlan` with curve, size, direction, input order, output order,
   input residency, output residency, and batch count.
3. Add pure Go tests for plan transitions.
4. Add assertions/helpers that document the current order contract:
   `FFT` natural to bit-reversed;
   `FFTInverse` bit-reversed to natural;
   `CosetFFT` natural to natural;
   `CosetFFTInverse` natural to natural.
5. Implement a private batched NTT entrypoint only if it can be done with small
   ABI changes.
6. If batched CUDA is too large, land the plan and tests first, then document
   the next exact CUDA change.
7. Add benchmarks that report bit-reversal cost separately.
8. Compare generic BLS12-377 NTT to the old `gpu/plonk` NTT using existing
   benchmark patterns.

## Validation

Non-CUDA:

```bash
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'Test.*NTTPlan|Test.*FFTSpec' -count=1
```

CUDA:

```bash
go test -tags cuda ./gpu/plonk2 \
  -run 'TestFrVectorOps_CUDA|TestFFT|TestCoset|Test.*NTT' \
  -count=1
```

Benchmarks:

```bash
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench 'BenchmarkFFTForward_CUDA|BenchmarkCosetFFTForward_CUDA' \
  -benchtime=5x -count=1
```

If comparing with old BLS12-377:

```bash
go test -tags cuda ./gpu/plonk -run '^$' \
  -bench 'BenchmarkBLSFFT' -benchtime=5x -count=1
```

## Expected Final Report

Report:

- NTT order and residency contract.
- Whether a batched transform was implemented.
- Bit-reversal cost measurements if available.
- Correctness status per curve.
- Remaining NTT work before full prover integration.


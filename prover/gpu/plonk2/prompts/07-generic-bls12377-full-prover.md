# Prompt 07: Generic BLS12-377 Full Prover

## Goal

Wire a full BLS12-377 PlonK prover path through `gpu/plonk2` primitives. The
first target is behavioral parity with gnark and a direct comparison against
the existing specialized `gpu/plonk` prover.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk/prove.go`
- `gpu/plonk2/prove.go` and API skeleton from earlier milestones
- `gpu/plonk2/fr.go`
- `gpu/plonk2/fft.go`
- `gpu/plonk2/msm.go`
- `gpu/plonk2/quotient.go`
- `gpu/plonk2/bench_full_prover_cuda_test.go`
- gnark BLS12-377 PlonK prover:
  `github.com/consensys/gnark/backend/plonk/bls12-377/prove.go`
  `github.com/consensys/gnark/backend/plonk/bls12-377/setup.go`

## Constraints

- Start with BLS12-377 only.
- Do not modify gnark vendored/module-cache code.
- Do not modify circuit definitions.
- Preserve proof verification with gnark verifier.
- Keep CPU fallback available.
- Do not optimize unrelated phases while porting.
- Keep the port phase-by-phase and testable.

## Implementation Strategy

Port the existing `gpu/plonk` orchestration to use `plonk2` generic primitives
behind the API skeleton. Avoid clever redesign during the first port.

Recommended phases:

1. Prepared instance:
   parse BLS12-377 `SparseR1CS`;
   build trace;
   create generic `FFTDomain`;
   create generic `G1MSM`;
   upload fixed polynomials;
   upload permutation table.
2. Witness solving and Qk completion:
   reuse gnark solving and transcript behavior.
3. L/R/O/Qk iFFT and commitments:
   use `plonk2` NTT and MSM;
   verify LRO commitments are accepted by the verifier.
4. Z construction and commitment:
   use `plonk2` Z kernels where already available.
5. Quotient and H commitments:
   port `computeNumeratorGPU` logic to `plonk2` vector types.
6. Opening and final proof assembly:
   initially keep CPU/GPU split from `gpu/plonk` where simpler.

## Implementation Tasks

1. Add BLS12-377-specific type narrowing only at the package boundary.
2. Keep generic primitive calls inside the implementation.
3. Add internal phase tests where possible.
4. Add tiny-circuit full proof test:
   compile;
   setup;
   prove with `plonk2`;
   verify with gnark.
5. Add invalid-witness test:
   CPU prove fails;
   GPU prove also fails.
6. Add benchmark comparing:
   gnark CPU;
   old `gpu/plonk`;
   new `gpu/plonk2`.
7. Add trace labels for major phases.
8. Document any temporary host-side fallbacks in code comments and worklog.

## Validation

Non-CUDA:

```bash
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlonkE2E|Test.*Prover' -count=1
```

CUDA:

```bash
go test -tags cuda ./gpu/plonk2 \
  -run 'Test.*BLS.*Prover|Test.*FullProver|TestPlonkE2E' \
  -count=1
```

Benchmarks:

```bash
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA|Benchmark.*Plonk2.*Full' \
  -benchtime=3x -count=1
```

## Expected Final Report

Report:

- Which prove phases are on `plonk2` GPU primitives.
- Which phases still use CPU fallback.
- Proof verification status.
- Comparison to old `gpu/plonk` and gnark CPU.
- Known correctness or performance gaps.


# Prompt 03: MSM Correctness Hardening

## Goal

Harden the curve-generic MSM and KZG commitment path before performance work.
Move result correction into CUDA if practical, centralize raw layout adapters,
and expand correctness tests for edge cases and bucket-skew inputs.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk2/msm.go`
- `gpu/plonk2/commit.go`
- `gpu/plonk2/g1.go`
- `gpu/plonk2/commit_cuda_test.go`
- `gpu/plonk2/g1_cuda_test.go`
- `gpu/plonk2/bench_bw6761_msm_cuda_test.go`
- `gpu/cuda/src/plonk2/msm.cu`
- `gpu/cuda/src/plonk2/ec.cuh`
- `gpu/cuda/src/plonk2/field.cuh`
- `gpu/cuda/include/gnark_gpu.h`

## Constraints

- Do not change the public affine short-Weierstrass input contract.
- Do not start throughput refactors in this milestone.
- Do not add public tuning knobs.
- Keep all existing KZG equality tests passing.
- If device-side Montgomery correction is too large for one patch, implement
  the adapter/tests first and document the remaining correction task.

## Implementation Tasks

1. Audit raw point and scalar layout conversions in `gpu/plonk2`.
2. Create or consolidate named helper functions for raw gnark-crypto layouts.
3. Remove duplicated unsafe layout assumptions where possible.
4. Add explicit tests for raw adapter word counts and curve widths.
5. Investigate the current host-side `correctRawMontgomeryMSM` path.
6. Move Montgomery correction into CUDA if the change is small and auditable.
7. If moving correction into CUDA, update the C ABI only as needed and remove
   the Go-side correction from hot `CommitRaw`.
8. Add CUDA tests for:
   zero scalars;
   one-hot scalars;
   repeated points;
   random scalars;
   short scalar slices against a longer resident SRS;
   structured scalars that force large buckets;
   release and re-pin of work buffers.
9. Add malformed input tests for word-count mismatches and unsupported curves.
10. Add a repeated-commit benchmark or test that exercises pinned work buffers.
11. Keep test sizes small enough for CI but representative enough to catch
    signed-window and carry bugs.

## Validation

Non-CUDA:

```bash
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlan|Test.*Raw|Test.*Curve' -count=1
```

CUDA:

```bash
go test -tags cuda ./gpu/plonk2 \
  -run 'TestCommitRaw|TestG1MSM|TestG1Affine|Test.*MSM.*CUDA' \
  -count=1
```

Benchmarks if CUDA is available:

```bash
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench 'BenchmarkG1MSMCommitRaw|BenchmarkBW6761MSMCommitRawSizes' \
  -benchtime=3x -count=1
```

## Expected Final Report

Report:

- Raw layout assumptions now centralized.
- Whether Montgomery correction is still host-side or moved to CUDA.
- New correctness cases added.
- Commands run and results.
- Any remaining correctness risks.


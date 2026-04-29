# Prompt 04: MSM Throughput Plan

## Goal

Introduce an internal MSM run plan and begin replacing the generic MSM's
throughput bottlenecks with size-aware, memory-aware execution. This milestone
should remain correctness-preserving and reviewable.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk2/msm_plan.go`
- `gpu/plonk2/msm.go`
- `gpu/cuda/src/plonk2/msm.cu`
- `gpu/cuda/src/plonk/msm.cu`
- `gpu/plonk2/bench_plonk_reference_cuda_test.go`
- `gpu/plonk2/icicle_compare_cuda_test.go`
- `gpu/plonk2/WORKLOG.md`

The design target is not to copy another library wholesale. Copy useful
boundaries: run config, memory planning, chunking, large-bucket segmentation,
and shared-base batching later.

## Constraints

- Do not expose internal MSM knobs publicly.
- Keep the existing `G1MSM.CommitRaw` API valid.
- Keep old accumulation path available behind a private switch until the new
  path passes all correctness tests.
- Do not attempt every optimization in one patch.
- No new dependencies beyond existing CUDA/CUB usage.

## Implementation Tasks

1. Add internal `MSMRunPlan` with curve, points, scalar bits, window bits,
   windows, batch size, chunk points, shared-base flag, precompute factor,
   large-bucket factor, and memory plan.
2. Derive `MSMRunPlan` from `CurveInfo`, point count, memory limit, and current
   `PlanMSMMemory`.
3. Log or expose plan details only through tests/benchmarks/trace helpers.
4. Add pure Go tests for plan decisions across BN254, BLS12-377, BW6-761.
5. Implement chunked execution if memory plan requires it.
6. Add CUDA-side bucket metadata needed for future segmentation:
   bucket sizes, non-empty bucket list, and offsets.
7. Initially feed this metadata into the existing accumulation path if that
   keeps the patch small.
8. Add a private environment variable or build-time switch only for comparing
   old and new internals in benchmarks. Do not document it as public API.
9. Add large-bucket segmentation in a separate small step if metadata lands
   cleanly.
10. Preserve correctness after each kernel change.

## Validation

Pure Go:

```bash
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlan|Test.*MSMRunPlan' -count=1
```

CUDA correctness:

```bash
go test -tags cuda ./gpu/plonk2 \
  -run 'TestG1MSMCommitRaw_CUDA|TestCommitRawMatchesKZG|TestPlonkE2EGPUSetupCommitments' \
  -count=1
```

CUDA benchmarks:

```bash
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench 'BenchmarkG1MSMCommitRawBLS12377|BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA' \
  -benchtime=3x -count=1
```

## Expected Final Report

Report:

- `MSMRunPlan` fields and default policy.
- Which internal throughput step was implemented.
- Correctness status for all curves.
- Benchmark deltas versus previous branch state.
- Any fallback path kept for safety.


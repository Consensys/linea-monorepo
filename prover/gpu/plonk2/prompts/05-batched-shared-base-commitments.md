# Prompt 05: Batched Shared-Base Commitments

## Goal

Add a private batched commitment path for PlonK commitment waves using one
resident SRS. This should amortize scalar staging, sort workspace, launch
overhead, and device synchronization while keeping public APIs stable.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk2/msm.go`
- `gpu/plonk2/commit.go`
- `gpu/plonk2/bench_plonk_reference_cuda_test.go`
- `gpu/plonk2/e2e_plonk_cuda_test.go`
- `gpu/plonk/msm.go`
- `gpu/plonk/prove.go`, especially `gpuCommitN`
- `gpu/cuda/src/plonk2/msm.cu`
- `gpu/cuda/include/gnark_gpu.h`

## Constraints

- Keep `CommitRaw` unchanged and correct.
- Make the batched API private or internal until stable.
- Do not expose batch tuning knobs publicly.
- Preserve all existing single-commit tests.
- Do not regress memory planning.

## Implementation Tasks

1. Define a private Go entrypoint, for example `commitRawBatch` or
   `(*G1MSM).commitRawBatch`, that accepts multiple scalar slices.
2. Validate all scalar slices against the resident SRS capacity and curve limb
   count.
3. Decide the first implementation shape:
   loop over existing `CommitRaw` but preserve pinned work buffers;
   or add a CUDA batch ABI if the loop cannot meet the milestone goal.
4. If adding CUDA ABI, keep it flat:
   resident MSM handle, scalar pointer array or contiguous scalar batch,
   count array, batch count, output buffer, stream.
5. Update `PlanMSMMemory` and `MSMRunPlan` for batch size.
6. Port setup-commitment benchmark internals to use the private batch path.
7. Add correctness tests comparing each batched output to CPU KZG commitments.
8. Add tests for mismatched scalar lengths and empty batches.
9. Ensure work buffers remain pinned across the whole wave.
10. Keep full prover work out of this milestone.

## Validation

CUDA correctness:

```bash
go test -tags cuda ./gpu/plonk2 \
  -run 'TestG1MSMCommitRaw_CUDA|TestCommitRawMatchesKZG|TestPlonkE2EGPUSetupCommitments' \
  -count=1
```

CUDA benchmarks:

```bash
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA' \
  -benchtime=3x -count=1
```

Non-CUDA:

```bash
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -count=1
```

## Expected Final Report

Report:

- Batch API shape and why it is private.
- Whether CUDA ABI changed.
- Memory-plan changes.
- Correctness status per curve.
- Benchmark change for setup commitments.


# Prompt 02: Memory Planner as Runtime Contract

## Goal

Make memory planning authoritative enough that later prover work can avoid
surprise OOMs. Extend planning beyond the current MSM estimate and wire the
selected plan into prepared-prover construction without yet changing CUDA
algorithms.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk2/msm_plan.go`
- `gpu/plonk2/msm_plan_test.go`
- `gpu/plonk2/msm.go`
- `gpu/plonk2/fft.go`
- `gpu/plonk2/quotient.go`
- `gpu/device.go`
- `gpu/plonk/prove.go`, especially `gpuInstance`, host buffers, quotient
  allocation comments, MSM pin/release scheduling.

## Constraints

- Do not expose MSM tuning knobs publicly.
- Do not change MSM or NTT kernel behavior.
- No new dependencies.
- Keep estimates conservative and explain assumptions.
- Do not require CUDA for pure planning tests.
- Keep BW6-761 as the stress case.

## Implementation Tasks

1. Inventory all major current `plonk2` allocations:
   Fr vectors, FFT twiddles, MSM resident points, scalar staging, assignment
   arrays, CUB temp, buckets, partials, and outputs.
2. Inventory major planned full-prover allocations from `gpu/plonk/prove.go`.
3. Extend `MSMMemoryPlan` if needed so it matches actual CUDA buffers in
   `gpu/cuda/src/plonk2/msm.cu`.
4. Add plan structs for:
   prepared-key persistent memory;
   NTT domain memory;
   quotient working memory;
   host pinned memory;
   total per-proof peak memory.
5. Add a single top-level plan type, for example `ProverMemoryPlan`.
6. Add helper methods for totals:
   persistent bytes;
   scratch bytes;
   per-wave bytes;
   host pinned bytes;
   estimated peak bytes.
7. Add a planner function that accepts curve, domain size, number of
   commitments, point count, configured memory limit, and pinned host limit.
8. Query `dev.MemGetInfo()` in CUDA builds when a device is available.
9. Keep non-CUDA planning functional without a device.
10. Wire memory-limit options from Prompt 01 into planning, returning a clear
    error when a plan exceeds a hard limit.
11. Add benchmark/test metadata output helpers that format plan details.
12. Add tests for BN254, BLS12-377, and BW6-761 at representative sizes.
13. Include boundary tests for BW6-761 window policy and chunk estimates.

## Validation

Run:

```bash
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlan|Test.*Memory|Test.*Prover' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

If CUDA is available:

```bash
go test -tags cuda ./gpu/plonk2 -run 'TestPlan|TestG1MSMCommitRaw_CUDA' -count=1
```

## Expected Final Report

Report:

- New plan structs and what each byte category means.
- How the plan maps to actual CUDA buffers.
- How memory limits are enforced.
- Any buffers still unplanned and why.
- Commands run and results.


# Prompt 10: Optional Specialization

## Goal

Evaluate optional specializations only after the generic prover is correct and
measured. Implement a specialization only when benchmark evidence justifies the
added maintenance surface.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk2/WORKLOG.md`
- `gpu/plonk/WORKLOG.md`
- `gpu/plonk/msm.go`
- `gpu/plonk/g1_te.go`
- `gpu/cuda/src/plonk/msm.cu`
- `gpu/cuda/src/plonk2/msm.cu`
- Current benchmark reports under `gpu/` if present.

Potential specializations:

- BLS12-377 Twisted-Edwards MSM backend as an opt-in implementation.
- SRS precomputation factors for fixed-base PlonK commitments.
- Multi-GPU proof-level scheduling.
- Moving more opening/folding/Horner work to GPU.
- More aggressive NTT fusion or batched quotient coset scheduling.

## Constraints

- Do not specialize before generic correctness and benchmarks are stable.
- Do not replace the generic path.
- Keep every specialization opt-in internally until proven.
- Do not add a specialization without a benchmark and rollback path.
- Keep code size and auditability as first-class criteria.

## Implementation Tasks

1. Choose exactly one specialization candidate.
2. Write a short design note before coding:
   what bottleneck it targets;
   why generic code is insufficient;
   expected speedup;
   extra code surface;
   correctness risk;
   rollback plan.
3. Add or identify a benchmark that isolates the bottleneck.
4. Measure baseline on the same hardware and command.
5. Implement the smallest opt-in version.
6. Add correctness tests that compare the specialized path to the generic path
   and CPU reference.
7. Measure again with identical benchmark settings.
8. Keep or revert based on evidence.
9. Update `WORKLOG.md` with both successful and failed experiments.

## Validation

Always run generic correctness tests:

```bash
go test -tags cuda ./gpu/plonk2 -count=1
```

Run specialization-specific benchmarks with fixed commands, for example:

```bash
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '<specific benchmark>' -benchtime=5x -count=1
```

If the specialization touches old `gpu/plonk`, also run:

```bash
go test -tags cuda ./gpu/plonk -count=1
```

## Expected Final Report

Report:

- Specialization evaluated.
- Baseline numbers.
- New numbers.
- Code surface added.
- Correctness tests run.
- Decision: keep, gate, or revert.


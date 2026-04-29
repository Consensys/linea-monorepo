# Prompt 00: Hygiene and Compile Baseline

## Goal

Establish a clean baseline before deeper GPU PlonK work. Fix known hygiene
issues and make non-CUDA package tests compile on macOS/Linux while preserving
existing CUDA behavior.

## Context

Read first:

- `AGENTS.md`
- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk2/DESIGN.md`
- `gpu/plonk2/WORKLOG.md`
- `gpu/threadlocal.go`
- `gpu/device_stub.go`
- `gpu/enabled_nocuda.go`
- `gpu/cuda/src/plonk2/msm.cu`

Known issues from prior inspection:

- `gpu/cuda/src/plonk2/msm.cu` has an error cleanup block that frees
  `d_buckets` twice.
- `go test ./gpu/plonk2 ./gpu/plonk` fails on macOS before reaching package
  tests because `gpu/threadlocal.go` calls Linux-only `unix.Gettid`.

## Constraints

- Do not change GPU algorithms.
- Do not refactor package structure.
- Do not modify public APIs except where required to restore portable builds.
- Do not remove multi-GPU thread-local behavior on Linux.
- Keep CUDA build behavior intact.
- No new dependencies.

## Implementation Tasks

1. Inspect the current build tags and file layout in `gpu/`.
2. Fix `gpu/threadlocal.go` portability.
3. Prefer a platform split if needed:
   `threadlocal_linux.go` for `unix.Gettid`, and a portable or Darwin stub for
   non-Linux behavior.
4. Preserve the semantics that Linux CUDA workers can bind a `*gpu.Device` and
   device ID to the current OS thread.
5. On non-Linux platforms, provide a safe implementation that compiles and does
   not panic for non-CUDA tests.
6. Fix the duplicate `cudaFree(d_buckets)` in `gpu/cuda/src/plonk2/msm.cu`.
7. Search nearby CUDA cleanup blocks for the same class of repeated free or
   missing free.
8. Add or adjust small non-CUDA tests only if they clarify the portability fix.
9. Update `gpu/plonk2/WORKLOG.md` with commands run and outcomes.

## Validation

Run on this host:

```bash
go test ./gpu/plonk2 ./gpu/plonk
gofmt -w gpu/*.go
```

If Go files were split by build tag, run:

```bash
go test ./gpu
```

If a CUDA machine is available, also run:

```bash
go test -tags cuda ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk -count=1
```

If no CUDA machine is available, do not fake CUDA results. State that CUDA
validation is pending.

## Expected Final Report

Report:

- Files changed.
- Exact root cause of the non-CUDA compile failure.
- Exact CUDA cleanup issue fixed.
- Commands run and their results.
- CUDA validation status.
- Any remaining hygiene risks found by inspection.


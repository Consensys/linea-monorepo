# Prompt 09: Production Rollout Controls

## Goal

Add operational controls, tracing, and safe fallback behavior so downstream
callers can enable the GPU prover without removing the CPU path.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/trace.go`
- `gpu/worklog_gpu_prover.md`
- `gpu/plonk/prove.go`
- `gpu/plonk2` full prover implementation from earlier prompts
- Downstream call sites that currently call gnark PlonK or `gpu/plonk`

## Constraints

- Do not force GPU use by default.
- Do not remove CPU paths.
- Do not change production config without explicit maintainer approval.
- Do not log secrets, witnesses, private scalars, or proof internals.
- Keep trace output metadata-only: timings, sizes, memory, curve, phase names.

## Implementation Tasks

1. Define rollout options:
   enabled/disabled;
   CPU fallback on error;
   memory limit;
   pinned host limit;
   trace path;
   strict mode that fails instead of falling back.
2. Decide whether options are Go options, environment variables, or both.
3. Use existing repository conventions for GPU env vars if present.
4. Add trace events for:
   prepare start/end;
   each major prove phase;
   memory plan;
   selected curve;
   selected MSM window/chunk policy;
   fallback reason.
5. Ensure trace events never contain witness values or scalar contents.
6. Add tests for fallback behavior:
   GPU disabled;
   unsupported curve;
   memory plan exceeds limit;
   strict mode.
7. Add documentation for build tags, CUDA requirements, and example commands.
8. If touching downstream call sites, keep the patch small and behind an
   explicit opt-in.
9. Update worklog with factual validation commands.

## Validation

Non-CUDA:

```bash
gofmt -w gpu gpu/plonk2
go test ./gpu/plonk2 ./gpu -run 'Test.*Trace|Test.*Fallback|Test.*Options' -count=1
```

CUDA:

```bash
go test -tags cuda ./gpu/plonk2 \
  -run 'Test.*Trace|Test.*Fallback|Test.*FullProver' \
  -count=1
```

If downstream call sites changed, run their package-specific tests.

## Expected Final Report

Report:

- Rollout controls added.
- Trace schema and example event.
- Fallback behavior.
- Tests run.
- Any production integration still pending.


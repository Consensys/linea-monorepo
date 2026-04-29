# Prompt 01: API Contract and Prepared Prover Skeleton

## Goal

Add a minimal drop-in prover API skeleton for `gpu/plonk2` without implementing
GPU acceleration yet. The wrapper must preserve gnark behavior through CPU
fallback and establish the public contract that later milestones will fill in.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- `gpu/plonk2/doc.go`
- `gpu/plonk2/curve.go`
- `gpu/plonk2/stub.go`
- `gpu/plonk2/e2e_plonk_test.go`
- gnark PlonK dispatcher in the module cache:
  `github.com/consensys/gnark/backend/plonk/plonk.go`
- curve-specific gnark proving key and proof packages:
  `backend/plonk/bn254`
  `backend/plonk/bls12-377`
  `backend/plonk/bw6-761`

## Constraints

- Do not implement partial GPU proving in this milestone.
- Do not fork or copy gnark's prover internals.
- Do not expose CUDA tuning knobs.
- Preserve CPU proof verification behavior.
- Keep API narrow and documented.
- Make unsupported curves fail with clear errors.
- Keep non-CUDA builds working.

## Target API

Introduce a small API equivalent to:

```go
type Prover struct {
    // private fields
}

func NewProver(
    dev *gpu.Device,
    ccs constraint.ConstraintSystem,
    pk plonk.ProvingKey,
    opts ...Option,
) (*Prover, error)

func (p *Prover) Prove(
    fullWitness witness.Witness,
    opts ...backend.ProverOption,
) (plonk.Proof, error)

func (p *Prover) Close() error

func Prove(
    dev *gpu.Device,
    ccs constraint.ConstraintSystem,
    pk plonk.ProvingKey,
    fullWitness witness.Witness,
    opts ...backend.ProverOption,
) (plonk.Proof, error)
```

Options for this milestone:

- `WithCPUFallback(enabled bool)`.
- `WithMemoryLimit(bytes uint64)`, stored but not enforced yet.
- `WithPinnedHostLimit(bytes uint64)`, stored but not enforced yet.
- `WithTrace(path string)`, stored but not emitted yet.

## Implementation Tasks

1. Decide file placement. Prefer new files under `gpu/plonk2`, such as
   `prove.go`, `options.go`, and `prover_api_test.go`.
2. Add a private curve detection helper that maps supported gnark proving key
   and constraint system types to `Curve`.
3. Verify key/constraint curve compatibility at `NewProver` time.
4. Store the CPU fallback policy in `Prover`.
5. Make `Prover.Prove` call gnark `plonk.Prove` when CPU fallback is enabled.
6. If CPU fallback is disabled, return a clear "GPU prover not wired yet" error.
7. Make package-level `Prove` construct a `Prover`, defer `Close`, and call
   `Prover.Prove`.
8. Make `Close` idempotent.
9. Add tests for:
   supported curves with CPU fallback;
   unsupported or mismatched key/constraint pairs;
   disabled CPU fallback;
   idempotent `Close`;
   package-level `Prove`.
10. Keep tests tiny and CPU-only.

## Validation

Run:

```bash
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlonkE2E|Test.*Prover|Test.*Fallback' -count=1
go test ./gpu/plonk2 ./gpu/plonk
```

If CUDA is available, run the existing CUDA suite to ensure the new files did
not break build tags:

```bash
go test -tags cuda ./gpu/plonk2 -count=1
```

## Expected Final Report

Report:

- Public API added.
- Supported curve detection logic.
- How CPU fallback behaves.
- Tests added.
- Commands run and results.
- Known limitations that remain for later milestones.


# Prompt 08: BN254 and BW6-761 Full Prover

## Goal

Extend the generic full prover from BLS12-377 to BN254 and BW6-761. Keep the
same orchestration and primitive interfaces, with curve-specific code isolated
to type adapters and constants.

## Context

Read first:

- `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`
- The BLS12-377 generic prover implementation from Prompt 07
- `gpu/plonk2/curve.go`
- `gpu/plonk2/fr.go`
- `gpu/plonk2/fft.go`
- `gpu/plonk2/msm.go`
- `gpu/plonk2/quotient.go`
- gnark curve-specific PlonK packages:
  `backend/plonk/bn254`
  `backend/plonk/bls12-377`
  `backend/plonk/bw6-761`

## Constraints

- Do not duplicate prover orchestration per curve.
- Do not expose curve-specific public APIs.
- Keep BW6-761 memory policy conservative.
- Keep CPU fallback available.
- Do not assume a 98 GiB GPU.
- Preserve gnark verifier compatibility.

## Implementation Tasks

1. Identify all BLS12-377-specific assumptions left after Prompt 07.
2. Move those assumptions into small curve adapters.
3. Add BN254 adapter:
   constraint type;
   proving key type;
   verifying key type;
   scalar raw layout;
   SRS raw layout;
   proof assembly details.
4. Add BW6-761 adapter with the same boundaries.
5. Use existing `CurveInfo` for limb sizes and scalar bits.
6. Apply memory planner before preparation, especially for BW6-761.
7. Add tiny full-prover E2E tests for BN254 and BW6-761.
8. Add invalid-witness tests for BN254 and BW6-761.
9. Add benchmarks for setup, prove, quotient, and opening if the harness exists.
10. Keep BLS12-377 tests running to prevent generic regressions.

## Validation

Non-CUDA:

```bash
gofmt -w gpu/plonk2
go test ./gpu/plonk2 -run 'TestPlonkE2E|Test.*Prover|Test.*Curve' -count=1
```

CUDA:

```bash
go test -tags cuda ./gpu/plonk2 \
  -run 'Test.*FullProver|Test.*BN254.*Prover|Test.*BW6.*Prover|TestPlonkE2E' \
  -count=1
```

Benchmarks:

```bash
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench 'Benchmark.*FullProver|BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA' \
  -benchtime=3x -count=1
```

## Expected Final Report

Report:

- Adapter boundaries added.
- Curves enabled.
- Memory-plan behavior for BW6-761.
- Proof verification status per curve.
- Benchmark results per curve.
- Any curve-specific gaps.


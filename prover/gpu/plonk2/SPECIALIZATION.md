# Optional Specialization Evaluation

## Candidate

BLS12-377 twisted-Edwards MSM backend as an internal opt-in specialization.

## Targeted Bottleneck

The generic `plonk2` affine MSM is still slower than the existing
`gpu/plonk` BLS12-377 twisted-Edwards MSM in historical setup-commitment
benchmarks. A TE backend could reduce BLS12-377 commitment latency by reusing
the specialized point layout and kernels from `gpu/plonk`.

## Why Not Implemented Now

The generic `plonk2` full prover is not yet wired. Proof generation through
`plonk2.Prover` currently uses gnark CPU fallback, and only the primitive
setup-commitment path exercises GPU MSM. Specializing before the generic prover
is correct would add a second backend surface without proving that the generic
orchestration, memory lifecycle, trace contract, and fallback controls are
stable.

## Expected Speedup

Historical worklog data shows the old BLS12-377 TE MSM can be materially faster
than the current generic affine MSM at setup sizes. No new baseline was taken
for this evaluation, so there is no current same-hardware speedup estimate.

## Extra Code Surface

A real opt-in backend would need:

- BLS12-377-only SRS conversion or TE SRS loading;
- backend selection in the resident MSM handle;
- duplicated correctness tests against generic affine MSM and CPU KZG;
- benchmark paths for setup commitments and full prover phases;
- rollback controls in the run plan and trace output.

## Correctness Risk

The TE path changes point representation and normalization boundaries. It must
prove equality against CPU KZG and the generic affine path for every commitment
wave before being considered for production.

## Rollback Plan

Keep the affine short-Weierstrass path as the default and gate any TE backend
behind a private internal switch. If benchmarks or correctness tests regress,
disable the switch and leave the generic path untouched.

## Decision

Defer. Do not add optional specialization until the generic full prover is
implemented, CUDA correctness tests pass, and same-hardware setup/full-prover
benchmarks identify a remaining BLS12-377 MSM bottleneck.

# GPU PlonK Agent Prompt Pack

These prompts decompose `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md` into
independent implementation tasks. They are written to be handed to an agent as
standalone instructions.

Use them in order unless a maintainer explicitly changes priority:

1. `00-hygiene-compile-baseline.md`
2. `01-api-contract-prepared-prover-skeleton.md`
3. `02-memory-planner-runtime-contract.md`
4. `03-msm-correctness-hardening.md`
5. `04-msm-throughput-plan.md`
6. `05-batched-shared-base-commitments.md`
7. `06-ntt-plan-batched-transforms.md`
8. `07-generic-bls12377-full-prover.md`
9. `08-bn254-bw6761-full-prover.md`
10. `09-production-rollout-controls.md`
11. `10-optional-specialization.md`

Common rules for every prompt:

- Read `AGENTS.md`, `gpu/plonk2/GPU_PLONK_LIBRARY_DESIGN.md`,
  `gpu/plonk2/DESIGN.md`, and the relevant code before editing.
- Keep the change surface small and directly tied to the prompt.
- Do not modify circuit definitions, `Define()` methods, gnark constraint
  expressions, or `Check(ifaces.Runtime) error` methods.
- Do not add external dependencies without maintainer approval.
- Preserve the CPU path and make CUDA behavior opt-in or build-tagged where
  appropriate.
- Prefer correctness and auditability before performance.
- Run `gofmt` for Go changes.
- Run package-specific tests. If no CUDA machine is available, run non-CUDA
  tests and state exactly which CUDA tests remain pending.
- Update `gpu/plonk2/WORKLOG.md` only with factual commands, hardware, and
  outcomes from work actually performed.

## CUDA Follow-Up Checklist

The non-CUDA work from the prompt pack has been applied, but several items
still require a CUDA-enabled machine before they can be considered complete.

### Required CUDA Validation

Run the generic CUDA suite first:

```bash
go test -tags cuda ./gpu/plonk2 -count=1
go test -tags cuda ./gpu/plonk -count=1
```

Then run the prompt-specific checks that were pending:

```bash
go test -tags cuda ./gpu/plonk2 \
  -run 'TestCommitRaw|TestG1MSM|TestG1Affine|Test.*MSM.*CUDA' \
  -count=1

go test -tags cuda ./gpu/plonk2 \
  -run 'TestFrVectorOps_CUDA|TestFFT|TestCoset|Test.*NTT' \
  -count=1

go test -tags cuda ./gpu/plonk2 \
  -run 'TestG1MSMCommitRaw_CUDA|TestCommitRawMatchesKZG|TestPlonkE2EGPUSetupCommitments' \
  -count=1

go test -tags cuda ./gpu/plonk2 \
  -run 'Test.*Trace|Test.*Fallback|Test.*FullProver' \
  -count=1
```

### Required CUDA Benchmarks

Measure current primitive and setup/full-prover baselines on the same host:

```bash
go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench 'BenchmarkG1MSMCommitRaw|BenchmarkBW6761MSMCommitRawSizes' \
  -benchtime=3x -count=1

go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench 'BenchmarkFFTForward_CUDA|BenchmarkCosetFFTForward_CUDA' \
  -benchtime=5x -count=1

go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA' \
  -benchtime=3x -count=1

go test -tags cuda ./gpu/plonk2 -run '^$' \
  -bench '^BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA|Benchmark.*Plonk2.*Full' \
  -benchtime=3x -count=1

go test -tags cuda ./gpu/plonk -run '^$' \
  -bench 'BenchmarkBLSFFT' -benchtime=5x -count=1
```

Record hardware, driver/CUDA versions, exact commands, and results in
`gpu/plonk2/WORKLOG.md`.

### CUDA Implementation Still Left

- Move MSM Montgomery result correction into CUDA only after validating the
  current host-side correction tests on CUDA.
- Add CUDA-side MSM bucket metadata and large-bucket segmentation behind the
  existing generic path; keep the current accumulation path as rollback until
  correctness and benchmark deltas are known.
- Replace the private Go-loop batch commitment path with a real flat CUDA batch
  ABI only if setup-commitment benchmarks show the loop remains a bottleneck.
- Add a real batched NTT CUDA ABI only if bit-reversal/launch/transfer
  measurements justify it.
- Port the actual generic full prover phase by phase:
  1. Prepared-key state is now resident for BN254, BLS12-377, and BW6-761:
     canonical/Lagrange SRS MSMs, FFT domain, permutation table, and fixed
     selector/permutation polynomials.
  2. The solved L-wire Lagrange commitment phase now has a direct generic GPU
     benchmark, `BenchmarkGenericSolvedWireCommitment_CUDA`.
  3. Next wire the full L/R/O commitment phase with PlonK blinding and
     gnark's reduced-range MSM correction.
  4. Then wire Z construction/commitment, quotient construction/commitment,
     linearized polynomial commitment, KZG openings, Fiat-Shamir ordering, and
     typed proof assembly.
  5. A temporary `WithEnabled(true)` CUDA bridge exists through the older
     `gpu/plonk` BLS12-377 prover so large full-prover benchmarks can run
     through the `plonk2.Prover` API, but this is not the generic prover port.
- Re-evaluate optional specialization, especially BLS12-377 twisted-Edwards
  MSM, only after the generic full prover passes CUDA correctness and has
  same-hardware benchmark baselines.

### Large Full-Prover Benchmarks

The PlonK setup/prove benchmark sizes default to small CI-friendly circuits.
Use `PLONK2_PLONK_BENCH_CONSTRAINTS` for large runs:

```bash
PLONK2_PLONK_BENCH_CONSTRAINTS=17M go test -tags cuda ./gpu/plonk2 \
  -run '^$' \
  -bench '^BenchmarkPlonk2EnabledFullProverBLS12377_CUDA$' \
  -benchtime=1x -count=1 -timeout 60m
```

Accepted suffixes are decimal `K`/`M` and binary `Ki`/`Mi`; for example,
`17M` means 17,000,000 constraints and `17Mi` means 17*2^20 constraints.

For the generic three-curve prover path currently under construction, use the
same size environment variable with the solved-wire commitment benchmark:

```bash
PLONK2_PLONK_BENCH_CONSTRAINTS=1Ki,16Ki go test -tags cuda ./gpu/plonk2 \
  -run '^$' \
  -bench '^BenchmarkGenericSolvedWireCommitment_CUDA$' \
  -benchtime=1x -count=1 -timeout 60m
```

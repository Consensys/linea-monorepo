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
- Wire the actual generic full prover for BLS12-377, then extend it to BN254
  and BW6-761. Current `plonk2.Prover` still proves through gnark CPU fallback
  unless fallback is disabled or strict mode is enabled.
- Re-evaluate optional specialization, especially BLS12-377 twisted-Edwards
  MSM, only after the generic full prover passes CUDA correctness and has
  same-hardware benchmark baselines.

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

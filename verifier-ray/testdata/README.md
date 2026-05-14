# verifier-ray testdata

Fixtures in this directory should be exported from `prover-ray` and consumed by
the Zig verifier tests. Keep files small and deterministic.

Suggested groups:

- `field/` for Koalabear and extension arithmetic vectors.
- `transcript/` for Fiat-Shamir transcript vectors.
- `vortex/` for commitment opening vectors.
- `generated/` for generated verifier golden files.


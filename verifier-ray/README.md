# verifier-ray

`verifier-ray` is the initial Zig verifier package for Ray proofs.

The package is intentionally independent from `prover-ray` at the directory
level. The fixed verifier runtime, field arithmetic, cryptographic primitives,
and zkVM precompile interface live here. Code generation can also live here and
import `prover-ray` structures when needed.

## Layout

- `src/` contains the Zig verifier library and executable entry point.
- `src/generated/` contains generated verifier stubs.
- `codegen/` contains the Go code generation tool skeleton.
- `test/` contains Zig unit tests.
- `testdata/` holds fixtures exported from `prover-ray`.

## Local Checks

```bash
zig build
zig build test

cd codegen
go test ./...
```

The current implementation is a scaffold. Cryptographic operations and Vortex
verification are placeholders until prover-ray fixtures and generated verifier
actions are wired in.


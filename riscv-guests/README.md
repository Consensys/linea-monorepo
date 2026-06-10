# RISC-V Guest Programs

This directory is the shared build and release boundary for guest programs that target the Linea RISC-V ZKC interpreter. Guest programs are anchored and released together, so they use one Zig toolchain, one Zig package manifest, and one `build.zig`.

## Layout

```text
riscv-guests/
  .zigversion     Required Zig development version
  build.zig       Shared Zig build for all guest programs
  build.zig.zon   Shared Zig package manifest and dependency pins
  l2-execution/   Vanilla EVM execution guest
  tools/          Local tooling shared by guest packages
```

## Required Toolchain

- Zig `0.16.0-dev.3153+d6f43caad`. The exact version is recorded in `.zigversion` and enforced by `build.zig`.
- Go, for converting compiled ELFs into the JSON input consumed by the ZKC interpreter.
- `zkc` on `PATH`, for `make exec`, `make debug`, and fixture execution.
- Optional: `riscv64-unknown-elf-objdump` for compile-time disassembly output.

Set `ZIG=/path/to/zig` when the required Zig binary is not first on `PATH`.

## Local Dependencies

External Zig dependencies are pinned in `build.zig.zon` with URL and hash metadata: **Zesu** (the EVM/stateless execution library) and the **execution-spec-tests `tests-zkevm` fixtures** (a `lazy` dependency, fetched only when the native test that consumes it is built). Zig resolves them automatically on first build and reuses them afterwards; the dedicated `make fetch` target can pre-fetch them.

## Native test dependencies

`make test` (`zig build test`) runs the guest logic on the **host**, where Zesu's `default.zig` accelerator backend is linked against native crypto C libraries. They must be installed:

| Library | Provides |
| --- | --- |
| `libsecp256k1` | ecrecover / signature verification |
| OpenSSL (`libssl`, `libcrypto`) | secp256r1 (P-256) |
| `libblst` | BLS12-381 + KZG point evaluation |
| `libmcl` | BN254 |

They are expected under a single prefix — `/opt/homebrew` on macOS, `/usr/local` on Linux — overridable with the `-Dcrypto-prefix=<prefix>` build option. The simplest way to install them all is Zesu's helper, run from a Zesu checkout:

```bash
make install-deps   # brew install secp256k1 openssl; builds blst + mcl from source into <prefix>/lib
```

The freestanding guest object (`make compile`) needs **none** of these — its crypto symbols (`zkvm_*`) are left unresolved for the prover to supply. Only the native host test links the libraries.

## Development

Run commands from `riscv-guests/`:

```bash
make compile ZIG=/path/to/zig
make test ZIG=/path/to/zig
```

`make compile` builds the guest as a relocatable rv64im **object** under `zig-out/lib/` (its `zkvm_*` crypto symbols are left unresolved for the prover, so it is not a runnable ELF on its own). `make test` runs the native Zig unit tests (see [Native test dependencies](#native-test-dependencies)).

The guest input offset can be overridden when compiling:

```bash
make compile ZIG=/path/to/zig INPUT_OFFSET=0x08800000
```

### Spec tests (full EF zkevm fixture suite)

`make test` is a fast single-fixture smoke test. To run the guest against the **whole** EF execution-spec-tests zkevm stateless suite on the host (needs the same native crypto libs as `make test`):

```bash
make spec-test ZIG=/path/to/zig
# narrow / triage with SPEC_ARGS:
make spec-test ZIG=/path/to/zig SPEC_ARGS="--fork Amsterdam"   # one fork only
make spec-test ZIG=/path/to/zig SPEC_ARGS="--limit 50 -x"      # first 50 blocks, stop on first fail
make spec-test ZIG=/path/to/zig SPEC_ARGS="--report-only"      # print pass-rate, always exit 0
```

This builds `evm-execution-spec-runner` and walks the `blockchain_tests/` tree from the lazy `execution_spec_tests_zkevm` dependency (downloaded on demand — no curl, no embedding), running every block through the guest and failing if any block's output differs from the fixture's expected `statelessOutputBytes`. The corpus walking/reporting is guest-agnostic ([`spec_runner.zig`](l2-execution/src/spec_runner.zig)); each guest supplies a small input **adapter** ([`evm_spec_runner.zig`](l2-execution/src/evm_spec_runner.zig) for the vanilla guest), so a future extended guest reuses the same runner and only adapts the input.

## ZKC Interpreter Integration

Running a guest inside the Lineth proving system (memory layout, ELF→JSON conversion and the `zkc` interpreter) is proving-system-specific and is owned by [`arithmetization/src/test/examples/Makefile`](../arithmetization/src/test/examples/Makefile). The targets below build the guest and **delegate** the run to it (via its `elf-exec` / `elf-debug` external-ELF path), so the ELF→JSON and `zkc` recipes have a single source of truth:

```bash
make exec ZIG=/path/to/zig GUEST=l2-execution IN_BYTES=0x...
make debug ZIG=/path/to/zig GUEST=l2-execution IN_BYTES=0x...
make fixture-exec ZIG=/path/to/zig GUEST=l2-execution
make fixture-debug ZIG=/path/to/zig GUEST=l2-execution
```

These require `zkc` and `go` on `PATH`. `make compile` emits a relocatable object (`zkvm_*` unresolved, no entry point), but the ELF→JSON step needs a fully linked ELF with `PT_LOAD` segments. They are wired ahead of the prover-link toolchain so the path is ready once the guest can be linked into an ELF.

## Guest Packages

Each guest package owns its source and fixtures. The Zig toolchain and build logic stay at this directory level.

- `l2-execution/`: vanilla EVM execution guest. See `l2-execution/README.md`.

# RISC-V Guest Programs

This directory holds the RISC-V guest programs that target the Linea ZKC interpreter. Each guest is a **self-contained Zig package** — its own `build.zig`, `build.zig.zon` (its dependencies), `Makefile` (its compile/test lifecycle) and `src/`. A thin top-level `Makefile` orchestrates them all, and shared build logic lives in `build_common/`. They share one Zig toolchain (`.zigversion`).

## Layout

```text
riscv-guests/
  .zigversion        Required Zig development version (shared by all guests)
  Makefile           Top-level orchestrator — fans compile/test/… out to every guest in GUESTS
  build_common/      Shared build helpers, @import-ed by each guest's build.zig (path dependency)
  l2-execution/      Vanilla EVM execution guest: build.zig + build.zig.zon + Makefile + src/
```

**Add a guest:** create `riscv-guests/<name>/` (its own `build.zig`, `build.zig.zon`, `Makefile`, `src/`, depending on `../build_common`) and append `<name>` to `GUESTS` in the top-level `Makefile`. Future guests (Rollup, Aggregation) slot in this way — each with its own dependencies and compile/lint sequence.

## Required Toolchain

- Zig `0.16.0-dev.3153+d6f43caad`. Recorded in `.zigversion` and enforced by `build_common` (`requireZigVersion`).
- Go, for converting compiled ELFs into the JSON input consumed by the ZKC interpreter.
- `zkc` on `PATH`, for a guest's `exec` / `debug` / fixture targets.
- Optional: `riscv64-unknown-elf-objdump` for compile-time disassembly output.

Set `ZIG=/path/to/zig` when the required Zig binary is not first on `PATH`.

## Dependencies

Each guest pins its **own** external dependencies in its `build.zig.zon`. For `l2-execution`: **Zesu** (EVM/stateless execution), **Consensys/zesu-zkvm** (its pure-Zig precompile backend `stdlibs_accel`, which the guest's in-guest crypto delegates to), and the **execution-spec-tests `tests-zkevm` fixtures** (a `lazy` dependency, fetched only for the tests). Every guest also takes `../build_common` as a path dependency for the shared build helpers. `make fetch` pre-fetches a guest's tree.

## Native test dependencies

A guest's `make test` runs its logic on the **host**, where Zesu's `default.zig` accelerator backend links native crypto C libraries:

| Library | Provides |
| --- | --- |
| `libsecp256k1` | ecrecover / signature verification |
| OpenSSL (`libssl`, `libcrypto`) | secp256r1 (P-256) |
| `libblst` | BLS12-381 + KZG point evaluation |
| `libmcl` | BN254 |

Expected under a single prefix — `/opt/homebrew` on macOS, `/usr/local` on Linux — overridable with `-Dcrypto-prefix=<prefix>`. Install them all via Zesu's helper (from a Zesu checkout): `make install-deps`. The freestanding guest object (`make compile`) needs **none** of these — its `zkvm_*` symbols are left unresolved for the prover.

## Development

From `riscv-guests/`, the top-level Makefile builds/tests **every** guest in `GUESTS`:

```bash
make compile ZIG=/path/to/zig   # build each guest's relocatable rv64im object
make test    ZIG=/path/to/zig   # run each guest's native host tests
make clean   ZIG=/path/to/zig
make help
```

Work on a single guest by invoking its own Makefile directly:

```bash
make -C l2-execution compile ZIG=/path/to/zig
make -C l2-execution compile ZIG=/path/to/zig INPUT_OFFSET=0x08800000   # override the input offset
```

`make compile` builds the guest as a relocatable rv64im **object** under `<guest>/zig-out/lib/` (its `zkvm_*` crypto symbols are unresolved — not a runnable ELF on its own). `make test` runs the native Zig unit tests (see [Native test dependencies](#native-test-dependencies)).

### Spec tests (l2-execution only — full EF zkevm fixture suite)

The EF stateless-fixture suite is specific to the EVM-execution guest, so `spec-test` is an **l2-execution target**, not an orchestrated one (a rollup/aggregation guest has no equivalent). `make test` is the fast single-fixture smoke test; the full suite:

```bash
make -C l2-execution spec-test ZIG=/path/to/zig
make -C l2-execution spec-test ZIG=/path/to/zig SPEC_ARGS="--fork Amsterdam"
make -C l2-execution spec-test ZIG=/path/to/zig SPEC_ARGS="--match bal_self_transfer"
make -C l2-execution spec-test ZIG=/path/to/zig SPEC_ARGS="--report-only"
```

The runner walks the `blockchain_tests/` tree from the lazy `execution_spec_tests_zkevm` dependency and runs every block through the guest, failing if any output differs from the fixture's expected `statelessOutputBytes`. The corpus walking/reporting is reusable ([`spec_runner.zig`](l2-execution/src/spec_runner.zig)); a future extended-execution guest supplies its own input **adapter** ([`evm_spec_runner.zig`](l2-execution/src/evm_spec_runner.zig) is the vanilla one).

## Continuous Integration

[`riscv-guests-host-tests.yml`](../.github/workflows/riscv-guests-host-tests.yml) runs on every PR touching `riscv-guests/**`, with two parallel host-machine jobs:

- **Guest unit tests** — `zig fmt --check` plus the orchestrated `make test` (every guest in `GUESTS`).
- **l2-execution EF spec tests** — the full fixture suite via `make spec-test` (fail-hard; ~2,900 files / ~23k blocks, minutes on a warm cache).

The shared setup lives in [`.github/actions/setup-riscv-guests`](../.github/actions/setup-riscv-guests/action.yml): it installs the Zig pinned in `.zigversion` (via community mirrors — ziglang.org prunes dev builds), the apt crypto packages, and blst/mcl built from pinned upstream sources into `/usr/local`, with the builds and Zig package fetches cached. Running a guest **inside the ZKC interpreter** in CI is a separate, later stage — blocked on the prover link toolchain (see below).

## ZKC Interpreter Integration

Running a guest inside the Lineth proving system (memory layout, ELF→JSON, `zkc`) is owned by [`arithmetization/src/test/examples/Makefile`](../arithmetization/src/test/examples/Makefile). A guest's `exec`/`debug` build it and **delegate** the run there (single source of truth for ELF→JSON + `zkc`):

```bash
make -C l2-execution exec  ZIG=/path/to/zig IN_BYTES=0x...
make -C l2-execution debug ZIG=/path/to/zig IN_BYTES=0x...
make -C l2-execution fixture-exec ZIG=/path/to/zig
```

These need `zkc` and `go` on `PATH`. **Not yet runnable:** `make compile` emits a relocatable object (no entry point), but the ELF→JSON step needs a fully linked ELF with `PT_LOAD` segments — pending the prover-link toolchain (likely `zesu-zkvm/linea`'s harness).

## Guest Packages

Each guest folder is a complete package: its own dependencies (`build.zig.zon`), compile/test logic (`build.zig`), lifecycle (`Makefile`) and source (`src/`). Shared build helpers are factored into `build_common/`; the toolchain pin (`.zigversion`) is shared at this level.

- `l2-execution/`: vanilla EVM execution guest. See `l2-execution/README.md`.
```

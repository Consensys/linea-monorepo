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

External Zig dependencies, including Zesu, are pinned in `build.zig.zon` with URL and hash metadata. The shared Makefile passes `--fetch` to Zig so missing dependencies are resolved automatically, then reused by later builds.

## Development

Run commands from `riscv-guests/`:

```bash
make compile ZIG=/path/to/zig
make test ZIG=/path/to/zig
make write-fixtures ZIG=/path/to/zig
```

`make compile` builds the currently supported guest ELFs under `zig-out/bin/`. `make test` runs the native Zig unit tests. `make write-fixtures` regenerates checked-in fixture payloads.

The default memory layout can be overridden when compiling:

```bash
make compile ZIG=/path/to/zig PROGRAM_OFFSET=0x00000000 IN_BYTES_OFFSET=0x08800000 SP=0x08800000
```

## ZKC Interpreter Integration

The shared Makefile keeps the example-program integration path for the ZKC interpreter:

```bash
make json ZIG=/path/to/zig GUEST=l2-execution IN_BYTES=0x...
make exec ZIG=/path/to/zig GUEST=l2-execution IN_BYTES=0x...
make debug ZIG=/path/to/zig GUEST=l2-execution IN_BYTES=0x...
make fixture-exec ZIG=/path/to/zig GUEST=l2-execution
make fixture-debug ZIG=/path/to/zig GUEST=l2-execution
```

`make json` compiles the ELF and runs `tools/elf2json/main.go` to produce the JSON consumed by `arithmetization/src/main/riscv/main.zkc`. The Makefile runs that standalone Go helper with `GO_ENV=GO111MODULE=off`; override `GO_ENV` only if the helper is moved into a Go module.

## Guest Packages

Each guest package owns its source and fixtures. The Zig toolchain and build logic stay at this directory level.

- `l2-execution/`: vanilla EVM execution guest. See `l2-execution/README.md`.

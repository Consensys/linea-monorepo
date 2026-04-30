# Makefile

The `Makefile` in the present folder allows to compile and run RISC-V test programs against the Linea zkVM.
Specifically, it allows to compile programs written in assembly, Zig or Rust, convert the resulting ELF to JSON, and pass it to `zkc` as an input file that represents the program to be run.

## Requirements

- `riscv64-unknown-elf-as` — for assembly programs
- `zig` — for Zig programs
- `cargo` — for Rust programs
- `rustc` with `riscv64imac-unknown-none-elf` target — for Rust programs
- `go` — to convert ELF to JSON
- `zkc` — to execute/debug the JSON

## Usage

From the `Makefile` directory:

```bash
make TEST=<name>.<ext>
```

From anywhere using `-f`:

```bash
make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile TEST=<name>.<ext>
```

Or set up a shell alias to avoid repeating the path:

```bash
alias zkc-test='make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile'
```

Then simply:

```bash
zkc-test TEST=<name>.<ext>
```

Where `<ext>` is `.s`, `.zig`, or `.rs`. Source files are by default expected in the corresponding `asm/src/`, `zig/src/`, or `rust/src/` directory.

## Targets

| Target | Description |
|--------|-------------|
| `make TEST=foo.<ext>` | Compile and execute (default) |
| `make debug TEST=foo.<ext>` | Compile and debug |
| `make compile TEST=foo.<ext>` | Compile only |
| `make clean TEST=foo.<ext>` | Remove binary and JSON for this test |
| `make clean-all` | Remove all build artifacts |

## Options

| Variable | Default | Description |
|----------|---------|-------------|
| `SRC` | `asm/src/<TEST>`, `zig/src/<TEST>`, or `rust/src/<TEST>` depending on extension | Path to the source file, can be overridden |
| `BIN` | `asm/bin/<NAME>`, `zig/zig-out/bin/<NAME>`, or `rust/bin/<NAME>` depending on extension | Path to the output ELF binary, can be overridden |
| `JSON` | same directory as `BIN`, with `.json` extension | Path to the output JSON file, can be overridden |
| `STRIP` | `false` | Strip debug symbols from the ELF after compilation |
| `ZIG_STRIP` | `true` | Strip when compiling Zig (reduces binary size), ignored for `.s` and `.rs` |
| `INBYTES` | `""` | Input bytes written to memory at `INBYTES_OFFSET` before execution |
| `PROGRAM_OFFSET` | `0` | Memory offset where the program is loaded (up to 128 MB) |
| `INBYTES_OFFSET` | `0x8000000` | Memory offset where input bytes are written (up to 1 GB) |
| `ENTRY_POINT` | `0` | Entry point offset |

## Examples

```bash
# Run an assembly test
zkc-test TEST=test.s

# Run a Zig test without stripping
zkc-test TEST=test.zig ZIG_STRIP=false

# Run a Rust test with input bytes
zkc-test TEST=test.rs INBYTES="0xAABBCC"

# Compile only, don't execute
zkc-test compile TEST=test.s

# Debug a Zig program
zkc-test debug TEST=test.zig

# Override source and binary paths
zkc-test TEST=test.rs SRC=/path/to/test.rs BIN=/path/to/output/test
```

## Target ISA

All programs are compiled targeting `RV64IM` accordingly to the [Ethereum zkVM standards](https://github.com/eth-act/zkvm-standards/blob/main/standards/riscv-target/target.md).
Note that `Zicclsm` extension does not affect the generated ELF so it is omitted.
Moreover, ABI being `LP64` (soft-float) is relevant only for float numbers, which we do not use, so it can be omitted as well.

## Default memory layout

```
0x00000000  ──  program starts
    ↓  program grows up (up to 128 MiB)
0x07FFFFFF  ──  program ends at most
0x08000000  --  sp ends here
    ↑  stack grows downward
0x087fffff  ──  sp starts here (up to 8 MiB)
0x08800000  ──  input starts
    ↓  input grows up (up to 1 GiB)
0x48800000 ──  input ends at most
```

## Utils

Run the following command to disassemble the generated ELF:

```
riscv64-unknown-elf-objdump -d --line-numbers -S test
```

# Makefile

The `Makefile` in this folder has commands to compile and run RISC-V test programs written in assembly, Zig or Rust against the Linea zkVM.
Programs are compiled for the  `riscv64im_zicclsm-unknown-none-elf` architecture. The resulting ELF is converted to JSON, and passed to `zkc` as an input.
The output ELF is also disassembled, producing an explorable `<name>_disassembled.elf` file.

The executable, the json and the disassembled elf file all live in the `<ext>/bin/` folder.

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

and from anywhere using `-f`:

```bash
make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile TEST=<name>.<ext>
```

**Note:** The extension `<ext>` must be `.s`, `.zig`, or `.rs`. Source files are by default expected in the corresponding `asm/src/`, `zig/src/`, or `rust/src/` directory. Alternatively one can provide full paths.


## Alias

Useful shell function (add to `~/.zshrc` or `~/.bashrc`):

```bash
zkctest() {
    case "$1" in
        clean-all)
            make -f "path/to/linea-monorepo/arithmetization/src/test/examples/Makefile" clean-all
            ;;
        exec|debug|compile|clean)
            local target="$1"; shift
            make -f "path/to/linea-monorepo/arithmetization/src/test/examples/Makefile" "$target" TEST="$1" "${@:2}"
            ;;
        *)
            make -f "path/to/linea-monorepo/arithmetization/src/test/examples/Makefile" TEST="$1" "${@:2}"
            ;;
    esac
}

# Usage
zkctest <name>.<ext>
zkctest <name>.<ext> INBYTES="0xAABB"
zkctest <name>.<ext> INBYTES="0xAABB" INBYTES_OFFSET=0x8000000
zkctest debug <name>.<ext>
zkctest debug <name>.<ext> INBYTES="0xAABB"
zkctest compile <name>.<ext>
zkctest clean <name>.<ext>
zkctest clean-all
```

## Targets

| Target                        | Description                          |
|-------------------------------|--------------------------------------|
| `make TEST=foo.<ext>`         | Compile and execute (default)        |
| `make debug TEST=foo.<ext>`   | Compile and debug                    |
| `make compile TEST=foo.<ext>` | Compile only                         |
| `make clean TEST=foo.<ext>`   | Remove binary and JSON for this test |
| `make clean-all`              | Remove all build artifacts           |

## Options

| Variable         | Default                                                                                 | Description                                                                |
|------------------|-----------------------------------------------------------------------------------------|----------------------------------------------------------------------------|
| `SRC`            | `asm/src/<TEST>`, `zig/src/<TEST>`, or `rust/src/<TEST>` depending on extension         | Path to the source file, can be overridden                                 |
| `BIN`            | `asm/bin/<NAME>`, `zig/zig-out/bin/<NAME>`, or `rust/bin/<NAME>` depending on extension | Path to the output ELF binary, can be overridden                           |
| `JSON`           | same directory as `BIN`, with `.json` extension                                         | Path to the output JSON file, can be overridden                            |
| `STRIP`          | `false`                                                                                 | Strip debug symbols from the ELF after compilation                         |
| `ZIG_STRIP`      | `true`                                                                                  | Strip when compiling Zig (reduces binary size), ignored for `.s` and `.rs` |
| `INBYTES`        | `""`                                                                                    | Input bytes written to memory at `INBYTES_OFFSET` before execution         |
| `PROGRAM_OFFSET` | `0`                                                                                     | Memory offset where the program is loaded (up to 128 MB)                   |
| `INBYTES_OFFSET` | `0x8000000`                                                                             | Memory offset where input bytes are written (up to 1 GB)                   |
| `ENTRY_POINT`    | `0`                                                                                     | Entry point offset                                                         |

## Examples

```bash
# Run an assembly test
zkctest test.s

# Run a Zig test without stripping
zkctest test.zig ZIG_STRIP=false

# Run a Rust test with input bytes
zkctest test.rs INBYTES="0xAABBCC"

# Compile only, don't execute
zkctest compile test.s

# Debug a Zig program
zkctest debug test.zig

# Override source and binary paths
zkctest test.rs SRC=/path/to/test.rs BIN=/path/to/output/test
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

## Deprecated

**Note.** The following command is done by default.
Run the following command to disassemble the generated ELF:

```
riscv64-unknown-elf-objdump -d --line-numbers -S test
```

# Makefile

The `Makefile` in this folder has commands to compile and run RISC-V test programs written in assembly, Zig or Rust against the Linea zkVM.
Programs are compiled for the  `riscv64im_zicclsm-unknown-none-elf` architecture. The resulting ELF is converted to JSON, and passed to `zkc` as an input.
The output ELF is also disassembled, producing an explorable `<name>.objdump` file.

The executable, the JSON and the disassembled file live in `asm/bin/` for assembly, `zig/zig-out/bin/` for Zig, and `rust/target/riscv64im-unknown-none-elf/release/` for Rust.

## Requirements

- `riscv64-unknown-elf-as (>= 2.45)` — for assembly programs and Zig
- `zig (>= 0.16.0)` — for Zig programs
- `cargo (>= 1.88.0)` — for Rust programs
- `rustc (>= rustc 1.88.0)` with `riscv64imac-unknown-none-elf` target — for Rust programs
- `go (>= 1.26.1)` — to convert ELF to JSON
- `go-corset, zkc (>= 1.2.12)` — to execute/debug the JSON

## Usage

From the `Makefile` directory:

```bash
make TEST=<src_optional_subfolder>/<name>.<ext>
```

and from anywhere using `-f`:

```bash
make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile TEST=<src_optional_subfolder>/<name>.<ext>
```

**Note:** The extension `<ext>` must be `.s`, `.zig`, or `.rs`. Source files are by default expected in the corresponding `asm/src/`, `zig/src/`, or `rust/src/` directory or in subfolders.

## Alias and usage examples

Useful shell function (add to `~/.zshrc` or `~/.bashrc`):

```bash
riscv-test() {
    local makefile="path/to/linea-monorepo/arithmetization/src/test/examples/Makefile"
    case "$1" in
        clean-all|linker-script|blake-all|build-act4|run-act4)
            # targets that do NOT require TEST argument
            make -f "$makefile" "$1" "${@:2}"
            ;;
        exec|debug|compile|zkc-exec|zkc-debug|clean|verify-elf)
            # targets that require TEST argument
            make -f "$makefile" "$1" TEST="$2" "${@:3}"
            ;;
        *)
            # default target (riscv-test foo.<ext> is the same as riscv-test exec TEST=foo.<ext>)
            make -f "$makefile" TEST="$1" "${@:2}"
            ;;
    esac
}

# Usage examples

# Compile and execute (note that <name>.<ext> can be replaced by <src_optional_subfolder>/<name>.<ext>)
riscv-test <name>.<ext>
# Compile and execute with input bytes
riscv-test <name>.<ext> IN_BYTES="0xAABB"
# Compile and debug
riscv-test debug <name>.<ext>
# Compile and debug with input bytes
riscv-test debug <name>.<ext> IN_BYTES="0xAABB"
# Compile and execute with input bytes at a custom offset
riscv-test <name>.<ext> IN_BYTES="0xAABB" IN_BYTES_OFFSET=0x08800008
# Compile only
riscv-test compile <name>.<ext>
# Clean build artifacts for a specific test
riscv-test clean <name>.<ext>
# Clean all build artifacts
riscv-test clean-all
# Run all converted Blake test vectors
riscv-test blake-all
# Build ACT4 ELFs
riscv-test build-act4
# Run ACT4 ELFs through zkc
riscv-test run-act4
# Run blake_with_in_embedded.rs (input bytes are embedded in main())
riscv-test blake/blake_with_in_embedded.rs
# Run blake_with_in_bytes.rs with IN_BYTES="0x<213_bytes_input_hex><64_bytes_expected_output_hex>"
riscv-test blake/blake_with_in_bytes.rs IN_BYTES="0x0000000c48c9bdf267e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5d182e6ad7f520e511f6c3e2b8c68059b6bbd41fbabd9831f79217e1319cde05b61626300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000001ba80a53f981c4d0d6a2797b69f12f6e94c212f14685ac4b74b12bb6fdbffa2d17d87c5392aab792dc252d5de4533cc9518d38aa8dbf1925ab92386edd4009923"
# Generate the linker script with custom input bytes offset
riscv-test linker-script IN_BYTES_OFFSET=0x00000042
# Verify ELF offsets, entry point and sp match the default ones
riscv-test verify-elf <name>.<ext>
# Verify ELF offsets, entry point and sp match the custom ones
riscv-test verify-elf <name>.<ext> PROGRAM_OFFSET=0x10000000 IN_BYTES_OFFSET=0x18800000 SP=0x187fffff 
# Compile and verify generated ELF offsets, entry point and sp match the default ones
riscv-test compile <name>.<ext> VERIFY_ELF=true
```

## Targets

| Target                           | Description                                                           |
|----------------------------------|-----------------------------------------------------------------------|
| `make TEST=foo.<ext>`            | Compile and execute (default)                                         |
| `make debug TEST=foo.<ext>`      | Compile and debug                                                     |
| `make compile TEST=foo.<ext>`    | Compile only                                                          |
| `make zkc-exec TEST=foo.<ext>`   | Execute without recompiling                                           |
| `make zkc-debug TEST=foo.<ext>`  | Debug without recompiling                                             |
| `make clean TEST=foo.<ext>`      | Remove binary and JSON for this test                                  |
| `make clean-all`                 | Remove all build artifacts                                            |
| `make linker-script`             | Generate the linker script with the memory layout                     |
| `make verify-elf TEST=foo.<ext>` | Verify ELF offsets, entry point and sp match the ones in the Makefile |
| `make blake-all`                 | Run all blake test vectors in `rust/src/blake/blake.all`              |
| `make build-act4`                | Build ACT4 ELFs with the Linea ACT4 config                            |
| `make run-act4`                  | Run ACT4 ELFs through zkc and write results/logs under `act4/bin/`     |

## ACT4

ACT4 targets use the Linea config in `act4/config/linea-rv64im-zicclsm/`.

Build the ACT4 docker image once from a `riscv-arch-test` checkout:

```bash
cd /path/to/riscv-arch-test
docker build -t riscv-act4 .
```

Then build and run the Linea ACT4 ELFs:

```bash
make build-act4
make run-act4
```

The default output locations are:

- `act4/bin/work/linea-rv64im-zicclsm/elfs/` — generated ELFs
- `act4/bin/results.txt` — PASS/FAIL summary
- `act4/bin/logs/` — per-test JSON and filtered zkc output

Run a single ACT4 test after `make run-act4` has generated JSON logs:

```bash
zkc exec act4/bin/logs/<test>.json ../../main/riscv/main.zkc
```

When overriding paths, use the same variables as `run-act4`:

```bash
LOGS=act4/bin/logs
ZKC_MAIN=../../main/riscv/main.zkc
zkc exec "$LOGS/<test>.json" "$ZKC_MAIN"
```

Disassemble a generated ELF:

```bash
riscv64-unknown-elf-objdump -d act4/bin/work/linea-rv64im-zicclsm/elfs/rv64i/I/I-add-00.elf
```

Build a single extension by overriding the build and run inputs:

```bash
ACT4_EXTENSIONS=M make build-act4
ELF_DIR=act4/bin/work/linea-rv64im-zicclsm/elfs/rv64i/M make run-act4
```

Useful overrides:

| Variable | Default | Used by |
|---|---|---|
| `ACT4_CONFIG_DIR` | `act4/config` | `build-act4` |
| `ACT4_WORK_DIR` | `act4/bin/work` | both |
| `ELF_DIR` | derived from `ACT4_WORK_DIR` | `run-act4` |
| `ACT4_IMAGE` | `riscv-act4:latest` | `build-act4` |
| `ACT4_EXTENSIONS` | `I,M` | `build-act4` |
| `ACT4_JOBS` | `4` | `build-act4` |
| `ACT4_FAST` | `True` | `build-act4` |
| `ACT4_DEBUG` | empty | `build-act4` |
| `ELF2JSON` | `act4/bin/elf2json` | `run-act4` |
| `LOGS` | `act4/bin/logs` | `run-act4` |
| `RESULTS` | `act4/bin/results.txt` | `run-act4` |

## Options

| Variable         | Default                                                                                 | Description                                                                   |
|------------------|-----------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| `IN_BYTES`       | `""`                                                                                    | Input bytes written to memory at `IN_BYTES_OFFSET` before execution           |
| `PROGRAM_OFFSET` | `0x00000000`                                                                            | Memory address where the program is loaded (up to 128 MiB)                    |
| `IN_BYTES_OFFSET`| `0x08800000`                                                                            | Memory address where input bytes are written (up to 1 GiB)                    |
| `SP`             | `0x087fffff`                                                                            | Top of the stack region, stack grows downward from this address (8 MiB)       |
| `VERIFY_ELF`     | `false`                                                                                 | Set to `true` to verify offsets, entry point and sp match the ELF ones        |

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

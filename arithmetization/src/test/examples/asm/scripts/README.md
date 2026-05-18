# Atomic RV64IM opcode tests

Tiny single-opcode tests under `../src/op_*.s`, plus the script that run them.

## Files

| Path | Purpose | 
|---|---|
| `run_op_tests.sh`   | Compiles every `../src/op/op_*.s`, runs it through `zkc exec --ir`, and writes a summary. |
| `../src/op/op_*.s`     | The generated test programs. One opcode per file. |

## Usage

```bash
# Compile + run all tests, write report to ../bin/results.txt
./run_op_tests.sh
```

After `run_op_tests.sh` finishes, the last line of `../bin/results.txt`
gives a one-line summary, and `../bin/logs/<name>.out` holds the full
zkc trace per test.

## Test pattern

Every test follows the same minimal scaffold:

```asm
    la sp, _stack_start          # set up SP from the linker script
    li a0, <val1>
    li a1, <val2>
    <op> a2, a0, a1              # the instruction under test
    li a3, <expected>
    sub a0, a2, a3               # a0 = result − expected (0 ⇒ PASS)
    li a7, 93
    ecall                        # exit a0
```

The exit code is therefore the difference between observed and expected
result. `Program exited successfully (exit with code 0)` from zkc means
PASS; anything else is FAIL with the diff in the exit code.

## Required tools

The same toolchain the rest of `../../examples` already needs:

- `bash`
- GNU `make` — invoked as `make` on Linux, as `gmake` on macOS once you
  `brew install make`. The runner picks `gmake` when available.
- `go` (for the ELF→JSON helper, built once into `../bin/elf2json`)
- `riscv64-unknown-elf-gcc`
  - Ubuntu / Debian: `sudo apt install gcc-riscv64-unknown-elf`
  - macOS: `brew tap riscv-software-src/riscv && brew install riscv-tools`
    (or any other tap that ships an `riscv64-unknown-elf-gcc` binary)
- `zkc` (built from go-corset; see top-level `arithmetization/Makefile`)

The runner verifies all of these are on `PATH` and exits early with a
helpful message if anything is missing.
# ACT4 / `riscv-arch-test` driver for Linea

Two scripts to drive the upstream RISC-V Architectural Compliance Tests
(ACT4) against our zkc, end-to-end:

| Script | Purpose |
|---|---|
| `build_linea_elfs.sh` | Compiles the ACT4 self-checking ELFs for the Linea (`linea-rv64im-zicclsm`) config, inside the official `riscv-act4` docker image. |
| `run_linea_elfs.sh`   | Converts each ELF to a zkc JSON and runs `zkc exec --ir` on it. Writes a PASS/FAIL summary plus per-test logs. |

Both scripts resolve all paths relative to their own location (no
absolute paths baked in). External inputs (config dir, ELF dir, image
tag) are overridable via environment variables; defaults assume the
sibling-clone layout described below.

## One-time setup

1. **Clone the supporting repos alongside `linea-monorepo`** (the
   defaults assume this exact layout):

   ```
   <some-parent>/
   ├── linea-monorepo/       # this repo — zkc + the test pipeline
   ├── riscv-arch-test/      # upstream ACT4 sources           (branch: act4)
   └── zkevm-test-monitor/   # owns act4-configs/linea/        (our config)
   ```

2. **Build the ACT4 docker image** (one-shot; can take ~30 min the first time):

   ```bash
   cd ../../../../../../riscv-arch-test       # from this scripts/ dir
   docker build -t riscv-act4 .
   ```

3. **Linea ACT4 config** lives in `zkevm-test-monitor/act4-configs/linea/`.
   It must contain `linea/linea-rv64im-zicclsm/` with `link.ld`, `sail.json`,
   `linea-rv64im-zicclsm.yaml`, `rvmodel_macros.h`, `rvtest_config.h`,
   `test_config.yaml`. See `zkevm-test-monitor`'s own README for provenance.

4. **zkc** must be on `PATH`. The simplest install path is
   `go install github.com/consensys/go-corset/cmd/zkc@v1.2.14`.

## Reproduce the run

With the layout above, from this directory:

```bash
./build_linea_elfs.sh       # compile RV64I + RV64M ELFs (~2 min)
./run_linea_elfs.sh         # run each through zkc, write ../bin/results.txt (~5 min for I+M)
```

That's it. No env vars need to be set when the cloned repos are in
their default positions. Output:

- `../bin/results.txt` — one `PASS` / `FAIL` line per test, plus
  totals on the last line.
- `../bin/logs/<test>.out` — filtered zkc terminal output per test.
- `../bin/elf2json` — cached Go-built ELF→JSON helper.
- `../bin/work/linea-rv64im-zicclsm/elfs/` — the compiled ACT4 ELFs.

## Environment variables (only needed to override defaults)

| Variable | Default | Used by | Notes |
|---|---|---|---|
| `ACT4_CONFIG_DIR` | `../../../../../../zkevm-test-monitor/act4-configs/linea` | `build` | Directory that contains `linea-rv64im-zicclsm/`. |
| `ACT4_WORK_DIR`   | `../bin/work` | both | Receives the build output. `run_linea_elfs.sh` derives `ELF_DIR` from it. |
| `ELF_DIR`         | derived from `ACT4_WORK_DIR` | `run` | Point at a different ELF tree. |
| `ACT4_IMAGE`      | `riscv-act4:latest` | `build` | Docker image tag. |
| `ACT4_EXTENSIONS` | `I,M` | `build` | Comma-separated list (e.g. `I` for just RV64I). |
| `ACT4_JOBS`       | `4` | `build` | Parallel `make` jobs inside the container. |
| `ELF2JSON`        | `../bin/elf2json` (auto-built) | `run` | Path to the Go helper that converts ELF → zkc JSON. |
| `ZKC_MAIN`        | `../../../../main/riscv/main.zkc` (relative to scripts dir) | `run` | The zkc program that interprets the ELF. |
| `PER_TEST_TIMEOUT`| `300` (s) | `run` | Bail out on a stuck test. |
| `IN_BYTES`        | `""` (empty — ACT4 tests are self-contained) | `run` | Forwarded to `elf2json`; same meaning as in `../../Makefile`. |
| `PROGRAM_OFFSET`  | `0x00000000` | `run` | Forwarded to `elf2json`. |
| `IN_BYTES_OFFSET` | `0x08800000` | `run` | Forwarded to `elf2json` (unused by ACT4 but kept consistent with the Makefile). |
| `ENTRY_POINT`     | `$PROGRAM_OFFSET` | `run` | Forwarded to `elf2json` — ACT4 places `rvtest_entry_point` at the origin. |

## Required tools

- Docker (for `build_linea_elfs.sh`)
- `go` (auto-builds `elf2json` from `../../main.go` the first time)
- `zkc` on `PATH`
- bash, GNU coreutils

## Building just one extension

```bash
ACT4_EXTENSIONS=M ./build_linea_elfs.sh
ELF_DIR="../bin/work/linea-rv64im-zicclsm/elfs/rv64i/M" ./run_linea_elfs.sh
```

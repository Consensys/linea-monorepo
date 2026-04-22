# verifier-risc5

Small comparison harness for compiling the same verifier-style workload to several RISC-V targets and running the result in emulators.

The project has two main goals:

- compare bare-metal and hosted code generation for a small Go workload
- keep the bare-metal paths close to a zkVM guest model: fixed memory, no syscalls, and minimal runtime

## Dependency Setup

The Go toolchain version is taken from `go.mod`, so a recent `go` install is enough. The module will download the exact Go toolchain it needs.

Install the host packages used by the current Makefile:

```bash
sudo apt install clang-20 lld-20 llvm-20-tools llvm-20-dev libclang-20-dev \
  gcc-riscv64-unknown-elf binutils-riscv64-unknown-elf \
  qemu-system-misc qemu-user cmake g++
```

Tool management split:

- TamaGo is module-managed through `go.mod` and is bootstrapped with `go tool tamago`
- TinyGo is vendored under `third_party/tinygo` because the module archive is missing required submodules
- libriscv is vendored under `third_party/libriscv`

Vendor the external trees when needed:

```bash
make vendor-tinygo
make vendor-libriscv
```

`make vendor-tinygo` checks out TinyGo at the pinned commit `1f114b2acff247f52eb77a92363119e7327107b3` and applies `third_party/tinygo-riscv64im_zicclsm.patch`.

## Target Profiles

Current ISA and runtime targets:

| Path | Output | ISA / ABI target | Notes |
| --- | --- | --- | --- |
| `build-clang` | `build/verifier-clang.elf` | `rv64im_zicclsm`, `lp64`, `medany` | freestanding, no libc, no syscalls |
| `build-gcc` | `build/verifier-gcc.elf` | `rv64im`, `lp64`, `medany` | intended comparison against the same core profile, but the packaged GCC toolchain used here rejects `zicclsm` |
| `build-tinygo` | `build/verifier-tinygo.elf` | `rv64im_zicclsm`, `lp64`, `medany` | freestanding TinyGo target with a local patch set |
| `build-go-linux` | `build/verifier-go-linux-riscv64` | `linux/riscv64`, `GORISCV64=rva20u64` | hosted baseline only, not a zkVM-style guest |
| `build-tamago` | `build/verifier-tamago-sifive_u.elf` | `tamago/riscv64` on `sifive_u` | board-specific machine-mode image, not a generic `rv64im_zicclsm` guest |

Important caveat:

- `clang`, `TinyGo`, and the C guest path are the closest match to the intended bare-metal zkVM model
- `Go` is only a hosted Linux baseline
- `TamaGo` is useful as a bare-metal Go comparison, but it expects a `sifive_u` machine model with privileged CSRs and board state

## Build

Build all supported outputs explicitly:

```bash
make build-clang
make build-gcc
make build-go-linux
make build-tinygo
make build-tamago
```

Useful one-time preparation:

```bash
make vendor-tinygo
make vendor-libriscv
make tinygo-bootstrap
make build-libriscv-runner
```

## Emulation

QEMU system emulation for the generic bare-metal guests:

```bash
make emulate-clang
make emulate-gcc
make emulate-tinygo
```

Hosted Go baseline:

```bash
make emulate-go-linux
```

TamaGo on the board model it expects:

```bash
make emulate-tamago
```

libriscv runner examples:

```bash
make emulate-libriscv
make emulate-libriscv LIBRISCV_GUEST=build/verifier-gcc.elf
make emulate-libriscv LIBRISCV_GUEST=build/verifier-tinygo.elf
make emulate-libriscv LIBRISCV_GUEST=build/verifier-tamago-sifive_u.elf
```

Meaning of those commands:

- plain `make emulate-libriscv` runs the default `clang` bare-metal ELF
- the `gcc` and `TinyGo` ELFs are supported in the same runner
- the `TamaGo` command is intentionally rejected with a clear error because that image expects `sifive_u` machine-mode CSRs and board initialization

`libriscv` is not used for:

- `build/verifier-go-linux-riscv64`, because that is a hosted Linux userspace binary
- `build/verifier-tamago-sifive_u.elf`, because that is a board-specific machine-mode image

## Repository Layout

### `baremetal/`

- `baremetal/entry.S`: minimal freestanding entrypoint for the C guest
- `baremetal/guest.c`: no-libc guest workload and UART MMIO output
- `baremetal/linker.ld`: memory layout and ELF segment placement for the freestanding guest

### `toolchains/`

- `toolchains/tinygo/riscv64im_zicclsm-qemu-virt.json`: custom TinyGo target definition for `rv64im_zicclsm`
- `toolchains/tinygo/riscv64im_zicclsm-qemu-virt.ld`: TinyGo linker script for the QEMU `virt` machine
- `toolchains/tamago/sifive_u_bios.S`: tiny BIOS trampoline used to boot the TamaGo `sifive_u` guest under QEMU
- `toolchains/libriscv/CMakeLists.txt`: host build for the local libriscv runner
- `toolchains/libriscv/runner.cpp`: bare-metal ELF runner with UART MMIO trapping and compatibility diagnostics

### `cmd/verifier/`

- `cmd/verifier/core.go`: shared verifier-style computation used by every Go build
- `cmd/verifier/main_hosted.go`: hosted entrypoint used outside bare-metal builds
- `cmd/verifier/main_baremetal.go`: bare-metal entrypoint used with the `baremetal` build tag
- `cmd/verifier/announce_none.go`: no-op bare-metal announcement fallback
- `cmd/verifier/announce_qemu_virt.go`: UART MMIO output path for the `qemu_virt` bare-metal target
- `cmd/verifier/announce_tamago_sifiveu.go`: TamaGo-specific output path
- `cmd/verifier/tamago_sifiveu.go`: imports the TamaGo `qemu/sifive_u` board package

The build split is:

- hosted Go uses `main_hosted.go`
- bare-metal Go uses `main_baremetal.go`
- bare-metal output behavior is selected by build tags such as `qemu_virt` and `tamago_sifive_u`

## Notes

- `third_party/tinygo` and `third_party/libriscv` are ignored by Git on purpose
- the TinyGo fixes live in `third_party/tinygo-riscv64im_zicclsm.patch`
- if you want a fully fresh TinyGo checkout, remove `third_party/tinygo` and run `make vendor-tinygo` again

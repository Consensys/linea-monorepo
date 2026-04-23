# verifier-risc5

Small comparison harness for compiling the same verifier-style workload to several RISC-V targets and running the result in emulators.

The project has two main goals:

- compare bare-metal and hosted code generation for a small Go workload
- keep the bare-metal paths close to a zkVM guest model: fixed memory, no syscalls, and minimal runtime

The current guest ABI follows the zkVM shape:

- the host generates a raw input blob and preloads it into guest RAM at `0x80f00000`
- the guest reads that memory directly, computes the verifier-style hash, and checks it against the expected value stored in the same blob
- the guest writes its result code and computed value to a fixed status page at `0x80eff000`

For the generic `qemu-system-riscv64 -machine virt` runs, the guest also writes to QEMU's SiFive test finisher MMIO register at `0x00100000` so the VM stops without any guest-side stdout or syscalls.

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
| `build-tinygo` | `build/verifier-tinygo.elf` | `rv64im_zicclsm`, `lp64`, `medany` | freestanding TinyGo target with a local patch set, defaulting to a stripped single-hart `-gc=none` profile |
| `build-go-linux` | `build/verifier-go-linux-riscv64` | `linux/riscv64`, `GORISCV64=rva20u64` | hosted baseline only, not a zkVM-style guest |
| `build-tamago` | `build/verifier-tamago-sifive_u.elf` | `tamago/riscv64` on `sifive_u` | board-specific machine-mode image, not a generic `rv64im_zicclsm` guest |

Important caveat:

- `clang` and the freestanding C guest path are the closest match to the intended bare-metal zkVM model
- `TinyGo` can target the same memory layout and status ABI on QEMU `virt`, and the current minimal runtime is also accepted by the local `libriscv` runner
- `Go` is only a hosted Linux baseline
- `TamaGo` is useful as a bare-metal Go comparison, but it expects a `sifive_u` machine model with privileged CSRs and board state

## Build

Build all supported outputs explicitly:

```bash
make build-input
make build-clang
make build-gcc
make build-go-linux
make build-tinygo
make build-tamago
```

`make build-input` converts `inputs/default.json` into `build/verifier-input.bin`. Override the fixture with:

```bash
make build-input INPUT_JSON=/path/to/other.json
```

TinyGo collector selection:

```bash
make build-tinygo
make build-tinygo TINYGO_GC=leaking
make build-tinygo TINYGO_EXTRA_FLAGS=-nobounds
```

The default is `TINYGO_GC=none` because the current guest does not allocate and this removes a small amount of runtime state. If the guest starts allocating later, switch back to `TINYGO_GC=leaking`.

Current TinyGo size result with the checked-in defaults:

- `build/verifier-tinygo.elf`: about `656 B` of `.text`, one `2 KiB` stack, and no DWARF sections
- this uses a target-local minimal TinyGo runtime variant for `scheduler=none` plus non-scanning GC, which drops the multicore interrupt, spinlock, and CSR startup paths for this zkVM-style guest
- compared to the older 4-hart TinyGo profile, this removes roughly `48 KiB` of extra stack reservations, all debug sections, and most of the target-local runtime text

One optional extra flag exists for local experiments only:

- `TINYGO_EXTRA_FLAGS=-nobounds` trims a little more code size, but it disables bounds checks and is not the default

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

Those targets automatically preload `build/verifier-input.bin` into guest RAM using QEMU's generic loader device. The guest itself stays silent: it writes the outcome to the fixed status page and then exits QEMU through the finisher MMIO device instead of printing text. The Makefile interprets the resulting QEMU host exit code and prints a short host-side summary such as `emulate-clang: guest reported success` or `emulate-clang: guest reported failure via the QEMU finisher`. By default there is no host-side timeout for these `virt` guests because they are expected to exit on their own; if you want a safety kill, pass something like `QEMU_TIMEOUT=5s`. To try a different fixture:

```bash
make emulate-clang INPUT_JSON=inputs/default.json
make emulate-tinygo INPUT_JSON=inputs/verifier-mismatch.json
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
- the `gcc` ELF is supported in the same runner
- the current minimal `TinyGo` ELF is also supported in the same runner
- the `TamaGo` command is intentionally rejected with a clear error because that image expects `sifive_u` machine-mode CSRs and board initialization
- `emulate-libriscv` also preloads the same `build/verifier-input.bin` blob at `0x80f00000`
- `emulate-libriscv` validates the fixed status page at `0x80eff000` instead of watching guest stdout

`libriscv` is not used for:

- `build/verifier-go-linux-riscv64`, because that is a hosted Linux userspace binary
- `build/verifier-tamago-sifive_u.elf`, because that is a board-specific machine-mode image

## Repository Layout

### `baremetal/`

- `baremetal/entry.S`: minimal freestanding entrypoint for the C guest
- `baremetal/guest.c`: no-libc guest workload that reads the preloaded input blob, computes the result, writes the fixed status page, and exits QEMU through the finisher MMIO register
- `baremetal/linker.ld`: memory layout and ELF segment placement for the freestanding guest, with the top `4 KiB` reserved for status and the top `1 MiB` reserved for preloaded input

### `toolchains/`

- `toolchains/tinygo/riscv64im_zicclsm-qemu-virt.json`: custom TinyGo target definition for `rv64im_zicclsm`
- `toolchains/tinygo/riscv64im_zicclsm-qemu-virt.ld`: TinyGo linker script for the QEMU `virt` machine, reserving both the fixed status page and the input window and keeping the TinyGo stack small
- `toolchains/tamago/sifive_u_bios.S`: tiny BIOS trampoline used to boot the TamaGo `sifive_u` guest under QEMU
- `toolchains/libriscv/CMakeLists.txt`: host build for the local libriscv runner
- `toolchains/libriscv/runner.cpp`: bare-metal ELF runner with input preloading, QEMU finisher trapping, and status-page validation

### `cmd/verifier/`

- `cmd/verifier/core.go`: shared verifier-style computation used by every Go build
- `cmd/verifier/input.go`: shared Go-side input representation
- `cmd/verifier/input_default.go`: hosted fallback input source using the default fixture
- `cmd/verifier/input_baremetal.go`: bare-metal reader for the fixed guest input window at `0x80f00000`
- `cmd/verifier/status_baremetal.go`: bare-metal writer for the fixed guest status page at `0x80eff000`
- `cmd/verifier/halt_baremetal.go`: explicit bare-metal fallback halt helper used after reporting status
- `cmd/verifier/main_hosted.go`: hosted entrypoint used outside bare-metal builds
- `cmd/verifier/main_baremetal.go`: bare-metal entrypoint used with the `baremetal` build tag
- `cmd/verifier/announce_none.go`: generic bare-metal status-page reporting fallback
- `cmd/verifier/announce_qemu_virt.go`: `qemu_virt` status-page reporting plus the QEMU finisher MMIO exit path
- `cmd/verifier/announce_tamago_sifiveu.go`: TamaGo-specific status-page reporting path
- `cmd/verifier/tamago_sifiveu.go`: imports the TamaGo `qemu/sifive_u` board package

### `cmd/inputgen/`

- `cmd/inputgen/main.go`: host-side generator that turns a JSON fixture into the raw preloaded guest input blob

### `inputs/`

- `inputs/default.json`: default host-side fixture containing the words array and expected result

### `internal/`

- `internal/workload/workload.go`: shared computation and default fixture values used by the Go guest and the input generator
- `internal/guestabi/abi.go`: fixed-address guest ABI constants for both the input blob and the result status page

The build split is:

- hosted Go uses `main_hosted.go`
- bare-metal Go uses `main_baremetal.go`
- bare-metal output behavior is selected by build tags such as `qemu_virt` and `tamago_sifive_u`

## Notes

- `third_party/tinygo` and `third_party/libriscv` are ignored by Git on purpose
- the TinyGo fixes live in `third_party/tinygo-riscv64im_zicclsm.patch`
- if you want a fully fresh TinyGo checkout, remove `third_party/tinygo` and run `make vendor-tinygo` again
- the Makefile suppresses command echo consistently; use `make -n <target>` if you want to inspect the exact shell commands

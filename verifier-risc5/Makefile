.DEFAULT_GOAL := help

BUILD_DIR := build
BAREMETAL_DIR := baremetal
TINYGO_TARGET := toolchains/tinygo/riscv64im_zicclsm-qemu-virt.json
TINYGO_TARGET_LD := toolchains/tinygo/riscv64im_zicclsm-qemu-virt.ld
TINYGO_TARGET_BUILD := $(BUILD_DIR)/$(notdir $(TINYGO_TARGET))
TAMAGO_BIOS_SRC := toolchains/tamago/sifive_u_bios.S
LIBRISCV_TOOLCHAIN_DIR := toolchains/libriscv
LIBRISCV_CMAKE := $(LIBRISCV_TOOLCHAIN_DIR)/CMakeLists.txt
LIBRISCV_RUNNER_SRC := $(LIBRISCV_TOOLCHAIN_DIR)/runner.cpp
GO_MAIN := ./cmd/verifier

LLVM_VERSION ?= 20
LLVM_BIN_DIR ?= /usr/lib/llvm-$(LLVM_VERSION)/bin
LLVM_DEV_DIR ?= /usr/lib/llvm-$(LLVM_VERSION)
LLVM_INCLUDE_DIR ?= $(if $(wildcard $(LLVM_DEV_DIR)/include/llvm-c/Core.h),$(LLVM_DEV_DIR)/include,/usr/include/llvm-$(LLVM_VERSION))
CLANG ?= $(if $(wildcard $(LLVM_BIN_DIR)/clang),$(LLVM_BIN_DIR)/clang,clang)
LD_LLD ?= $(if $(wildcard $(LLVM_BIN_DIR)/ld.lld),$(LLVM_BIN_DIR)/ld.lld,ld.lld)

GO ?= go
CMAKE ?= cmake
HOST_GO_VERSION := $(shell $(GO) env GOVERSION)
TINYGO_ROOT ?= third_party/tinygo
TINYGO_BIN := $(abspath $(TINYGO_ROOT))/build/tinygo
TINYGO_LLVM_PROJECT_DIR := $(TINYGO_ROOT)/llvm-project
LIBRISCV_ROOT ?= third_party/libriscv
XDG_CACHE_HOME_LOCAL ?= $(abspath .cache/xdg)
TAMAGO_CACHE_HOME ?= $(XDG_CACHE_HOME_LOCAL)
TAMAGO ?= XDG_CACHE_HOME=$(TAMAGO_CACHE_HOME) $(GO) tool tamago
RISCV_GCC ?= riscv64-unknown-elf-gcc
RISCV_OBJCOPY ?= riscv64-unknown-elf-objcopy
READELF ?= readelf

QEMU_SYSTEM ?= qemu-system-riscv64
QEMU_USER ?= qemu-riscv64
QEMU_TIMEOUT ?= 1s
QEMU_MACHINE ?= virt
QEMU_CPU ?= rv64
QEMU_MEM ?= 256M

TARGET_TRIPLE ?= riscv64-unknown-none-elf
RV_MARCH ?= rv64im_zicclsm
RV_MARCH_GCC ?= $(patsubst %_zicclsm,%,$(RV_MARCH))
RV_ABI ?= lp64
RV_CODE_MODEL ?= medany
GO_RISCV_PROFILE ?= rva20u64

TAMAGO_TAGS ?= baremetal,tamago_sifive_u,tiny
TAMAGO_LDFLAGS ?= -T 0x80010000 -R 0x1000
TINYGO_VERSION ?= v0.41.0
TINYGO_SCHEDULER ?= none
TINYGO_GC ?= conservative
TINYGO_PANIC ?= trap
TINYGO_OPT ?= z
TINYGO_LLVM_PROJECT_BRANCH ?= tinygo_20.x
TAMAGO_TOOL_VERSION ?= v1.26.2
LIBRISCV_VERSION ?= v1.15
TAMAGO_GOROOT := $(TAMAGO_CACHE_HOME)/tamago-go/tamago-go$(patsubst v%,%,$(TAMAGO_TOOL_VERSION))
TAMAGO_BIN := $(TAMAGO_GOROOT)/bin/go
LIBRISCV_BUILD_DIR := $(BUILD_DIR)/libriscv-host
LIBRISCV_RUNNER := $(LIBRISCV_BUILD_DIR)/libriscv_runner
LIBRISCV_GUEST ?= $(CLANG_ELF)
LIBRISCV_MAX_INSTRUCTIONS ?= 50000000

CLANG_ELF := $(BUILD_DIR)/verifier-clang.elf
GCC_ELF := $(BUILD_DIR)/verifier-gcc.elf
GO_LINUX_BIN := $(BUILD_DIR)/verifier-go-linux-riscv64
TINYGO_ELF := $(BUILD_DIR)/verifier-tinygo.elf
TAMAGO_ELF := $(BUILD_DIR)/verifier-tamago-sifive_u.elf
TAMAGO_BIOS_ELF := $(BUILD_DIR)/tamago-sifive_u-bios.elf
TAMAGO_BIOS_BIN := $(BUILD_DIR)/tamago-sifive_u-bios.bin
TAMAGO_DTB := $(BUILD_DIR)/qemu-sifive_u.dtb

.PHONY: help \
	tools-sync tools-list clean-tool-cache vendor-tinygo tinygo-bootstrap \
	vendor-libriscv build-libriscv-runner \
	build-clang build-gcc build-go-linux build-tinygo build-tamago \
	emulate-clang emulate-gcc emulate-go-linux emulate-tinygo emulate-tamago emulate-libriscv \
	clean

help:
	@printf '%s\n' \
		'Targets:' \
		'  make build-clang      - bare-metal ELF with clang + ld.lld' \
		'  make build-gcc        - bare-metal ELF with riscv64-unknown-elf-gcc' \
			'  make build-go-linux   - linux/riscv64 ELF with upstream Go' \
			'  make vendor-tinygo    - clone the vendored TinyGo checkout, submodules, and llvm-project builtins sources' \
			'  make vendor-libriscv  - clone the vendored libriscv checkout and submodules' \
			'  make tinygo-bootstrap - build the local TinyGo binary from the vendored TinyGo checkout' \
			'  make build-libriscv-runner - build the local bare-metal libriscv host runner' \
			'  make build-tinygo     - bare-metal ELF with the vendored TinyGo binary' \
			'  make build-tamago     - bare-metal ELF for QEMU sifive_u with TamaGo' \
			'  make tools-sync       - re-pin TamaGo in go.mod via go get -tool' \
		'  make tools-list       - show builtin and module-managed Go tools' \
		'  make clean-tool-cache - remove the local TinyGo/TamaGo XDG cache' \
		'  make emulate-clang    - run the clang bare-metal ELF on QEMU virt' \
		'  make emulate-gcc      - run the gcc bare-metal ELF on QEMU virt' \
			'  make emulate-go-linux - run the upstream Go ELF under qemu-riscv64' \
			'  make emulate-tinygo   - run the TinyGo bare-metal ELF on QEMU virt' \
			'  make emulate-tamago   - run the TamaGo ELF on QEMU sifive_u' \
			'  make emulate-libriscv - run the bare-metal ELF in the vendored libriscv runner' \
			'' \
			'Useful overrides:' \
			'  TINYGO_ROOT=third_party/tinygo - location of the vendored TinyGo checkout used by build-tinygo' \
			'  LIBRISCV_ROOT=third_party/libriscv - location of the vendored libriscv checkout used by emulate-libriscv' \
			'  TINYGO_VERSION=v0.41.0      - TinyGo tag cloned by make vendor-tinygo' \
			'  LIBRISCV_VERSION=v1.15      - libriscv tag cloned by make vendor-libriscv' \
			'  TINYGO_LLVM_PROJECT_BRANCH=tinygo_20.x - llvm-project branch cloned for TinyGo compiler-rt builtins' \
			'  HOST_GO_VERSION=go1.26.2    - Go toolchain used to bootstrap the vendored TinyGo binary' \
			'  LIBRISCV_GUEST=build/verifier-clang.elf - bare-metal ELF passed to emulate-libriscv' \
			'  TINYGO_GC=conservative      - TinyGo GC mode used by build-tinygo' \
			'  TINYGO_SCHEDULER=none       - TinyGo scheduler mode used by build-tinygo' \
			'  LLVM_VERSION=20             - select the host LLVM major version used for TinyGo and the clang/lld bare-metal path' \
			'' \
			'Ubuntu packages usually needed for the tested paths:' \
			'  sudo apt install clang-20 lld-20 llvm-20-tools llvm-20-dev libclang-20-dev gcc-riscv64-unknown-elf binutils-riscv64-unknown-elf qemu-system-misc qemu-user cmake g++' \
			'' \
		'Go tool-managed compilers in this module:' \
		'  TamaGo: XDG_CACHE_HOME=.cache/xdg go tool tamago ...' \
		'' \
		'Notes:' \
		'  TinyGo is vendored instead of module-managed because the upstream module archive omits required git submodules such as src/net.' \
		'  The default checkout lives under third_party/ rather than vendor/ because vendor/ triggers Go module vendoring semantics and breaks normal go/tinygo builds.' \
		'  build-tinygo defaults to third_party/tinygo, bootstraps third_party/tinygo/build/tinygo locally, vendors tinygo-org/llvm-project for compiler-rt builtins, and still requires LLVM headers and libclang development files on the host.' \
		'  This Makefile defaults to LLVM $(LLVM_VERSION), injects headers/libs from $(LLVM_DEV_DIR) for the vendored TinyGo bootstrap, and redirects tool caches under .cache/xdg.' \
		'  go tool tamago bootstraps tamago-go into a cache directory; this Makefile keeps that cache under .cache/ so go test ./... will not recurse into it.' \
			'  build-tamago uses the installed tamago-go binary directly with GODEBUG=goindex=0 because tamago-go1.26.2 misclassifies std packages for tamago/riscv64 when the module index is enabled.' \
			'  emulate-tamago dumps a DTB from QEMU and builds a tiny BIOS trampoline locally, so no separate dtc installation is required for the sifive_u path.' \
			'  emulate-libriscv uses a tiny vendored host runner with UART MMIO traps for generic bare-metal guests, rather than the Linux-oriented rvlinux CLI.' \
			'  TamaGo sifive_u images are board-specific machine-mode guests and should be run with make emulate-tamago, not emulate-libriscv.' \
			'  Upstream Go cannot target unknown-none-elf; build-go-linux is only a hosted linux/riscv64 baseline.' \
			'  The bare-metal C/LLVM path uses riscv64-unknown-none-elf plus -march=rv64im_zicclsm because that is accepted more widely than an extension-bearing triple.' \
			'  QEMU_CPU defaults to rv64. Override it if your local QEMU supports an exact extension string.'

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

$(TINYGO_TARGET_BUILD): $(TINYGO_TARGET) $(TINYGO_TARGET_LD) | $(BUILD_DIR)
	sed 's|"linkerscript": "riscv64im_zicclsm-qemu-virt.ld"|"linkerscript": "$(abspath $(TINYGO_TARGET_LD))"|' "$(TINYGO_TARGET)" > "$@"

tools-sync:
	$(GO) get -tool github.com/usbarmory/tamago/cmd/tamago@$(TAMAGO_TOOL_VERSION)

tools-list:
	$(GO) tool

clean-tool-cache:
	rm -rf $(TAMAGO_CACHE_HOME) $(BUILD_DIR)/xdg-cache

vendor-tinygo:
	if [ "$(TINYGO_ROOT)" != "vendor/tinygo" ] && [ -d "vendor/tinygo" ] && [ ! -e "$(TINYGO_ROOT)" ]; then mkdir -p "$(dir $(TINYGO_ROOT))" && mv "vendor/tinygo" "$(TINYGO_ROOT)"; fi
	if [ "$(TINYGO_ROOT)" != "vendor/tinygo" ]; then rmdir "vendor" 2>/dev/null || true; fi
	test -d "$(TINYGO_ROOT)" || git clone --branch "$(TINYGO_VERSION)" --depth 1 --recursive https://github.com/tinygo-org/tinygo "$(TINYGO_ROOT)"
	test -f "$(TINYGO_ROOT)/go.mod"
	git -C "$(TINYGO_ROOT)" submodule update --init --recursive
	test -d "$(TINYGO_LLVM_PROJECT_DIR)" || git clone --branch "$(TINYGO_LLVM_PROJECT_BRANCH)" --depth 1 https://github.com/tinygo-org/llvm-project "$(TINYGO_LLVM_PROJECT_DIR)"

vendor-libriscv:
	test -d "$(LIBRISCV_ROOT)" || git clone --branch "$(LIBRISCV_VERSION)" --depth 1 --recursive https://github.com/libriscv/libriscv "$(LIBRISCV_ROOT)"
	test -f "$(LIBRISCV_ROOT)/CMakeLists.txt"
	git -C "$(LIBRISCV_ROOT)" submodule update --init --recursive

$(TINYGO_BIN): vendor-tinygo
	mkdir -p "$(dir $(TINYGO_BIN))"
	cd "$(TINYGO_ROOT)" && \
		XDG_CACHE_HOME="$(XDG_CACHE_HOME_LOCAL)" \
		CGO_CPPFLAGS="-I$(LLVM_INCLUDE_DIR)" \
		CGO_LDFLAGS="-L$(LLVM_DEV_DIR)/lib" \
		GOTOOLCHAIN="$(HOST_GO_VERSION)" \
		TINYGOROOT="$(abspath $(TINYGO_ROOT))" \
		$(GO) build -trimpath -o "$(TINYGO_BIN)" .

tinygo-bootstrap: $(TINYGO_BIN)

$(LIBRISCV_RUNNER): $(LIBRISCV_CMAKE) $(LIBRISCV_RUNNER_SRC) | vendor-libriscv $(BUILD_DIR)
	$(CMAKE) -S "$(LIBRISCV_TOOLCHAIN_DIR)" -B "$(LIBRISCV_BUILD_DIR)" \
		-DCMAKE_BUILD_TYPE=Release \
		-DLIBRISCV_SOURCE_DIR="$(abspath $(LIBRISCV_ROOT))" \
		-DRISCV_BINARY_TRANSLATION=OFF \
		-DRISCV_LIBTCC=OFF \
		-DRISCV_32I=OFF \
		-DRISCV_64I=ON \
		-DRISCV_128I=OFF \
		-DRISCV_MEMORY_TRAPS=ON
	$(CMAKE) --build "$(LIBRISCV_BUILD_DIR)" --target libriscv_runner

build-libriscv-runner: $(LIBRISCV_RUNNER)

$(BUILD_DIR)/entry.o: $(BAREMETAL_DIR)/entry.S | $(BUILD_DIR)
	$(CLANG) \
		--target=$(TARGET_TRIPLE) \
		-march=$(RV_MARCH) \
		-mabi=$(RV_ABI) \
		-mcmodel=$(RV_CODE_MODEL) \
		-c \
		-o $@ \
		$<

$(BUILD_DIR)/guest.o: $(BAREMETAL_DIR)/guest.c | $(BUILD_DIR)
	$(CLANG) \
		--target=$(TARGET_TRIPLE) \
		-ffreestanding \
		-fno-builtin \
		-march=$(RV_MARCH) \
		-mabi=$(RV_ABI) \
		-mcmodel=$(RV_CODE_MODEL) \
		-c \
		-o $@ \
		$<

$(CLANG_ELF): $(BUILD_DIR)/entry.o $(BUILD_DIR)/guest.o $(BAREMETAL_DIR)/linker.ld
	$(LD_LLD) \
		-melf64lriscv \
		-T $(BAREMETAL_DIR)/linker.ld \
		-o $@ \
		$(BUILD_DIR)/entry.o \
		$(BUILD_DIR)/guest.o

build-clang: $(CLANG_ELF)

build-gcc: | $(BUILD_DIR)
$(GCC_ELF): $(BAREMETAL_DIR)/entry.S $(BAREMETAL_DIR)/guest.c $(BAREMETAL_DIR)/linker.ld | $(BUILD_DIR)
	$(RISCV_GCC) \
		-ffreestanding \
		-fno-builtin \
		-nostdlib \
		-march=$(RV_MARCH_GCC) \
		-mabi=$(RV_ABI) \
		-mcmodel=$(RV_CODE_MODEL) \
		-T $(BAREMETAL_DIR)/linker.ld \
		-o $(GCC_ELF) \
		$(BAREMETAL_DIR)/entry.S \
		$(BAREMETAL_DIR)/guest.c

build-gcc: $(GCC_ELF)

$(GO_LINUX_BIN): | $(BUILD_DIR)
	GOOS=linux GOARCH=riscv64 GORISCV64=$(GO_RISCV_PROFILE) \
		$(GO) build -trimpath -o $(GO_LINUX_BIN) $(GO_MAIN)

build-go-linux: $(GO_LINUX_BIN)

$(TINYGO_ELF): $(TINYGO_BIN) $(TINYGO_TARGET_BUILD) $(TINYGO_TARGET_LD) | $(BUILD_DIR)
	XDG_CACHE_HOME="$(XDG_CACHE_HOME_LOCAL)" \
	GOTOOLCHAIN="$(HOST_GO_VERSION)" \
	TINYGOROOT="$(abspath $(TINYGO_ROOT))" \
	"$(TINYGO_BIN)" build \
		-target=$(abspath $(TINYGO_TARGET_BUILD)) \
		-scheduler=$(TINYGO_SCHEDULER) \
		-gc=$(TINYGO_GC) \
		-panic=$(TINYGO_PANIC) \
		-opt=$(TINYGO_OPT) \
		-o $(TINYGO_ELF) \
		$(GO_MAIN)

build-tinygo: $(TINYGO_ELF)

$(TAMAGO_BIN): | $(BUILD_DIR)
	if [ ! -x "$(TAMAGO_BIN)" ]; then rm -rf "$(TAMAGO_GOROOT)"; fi
	test -x "$(TAMAGO_BIN)" || env -u GOOS -u GOARCH -u GOOSPKG XDG_CACHE_HOME="$(TAMAGO_CACHE_HOME)" $(GO) tool tamago version >/dev/null
	test -x "$(TAMAGO_BIN)"

$(TAMAGO_ELF): $(TAMAGO_BIN) | $(BUILD_DIR)
	GODEBUG=goindex=0 GOWORK=off GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=riscv64 \
		GOROOT="$(TAMAGO_GOROOT)" PATH="$(TAMAGO_GOROOT)/bin:$$PATH" \
		"$(TAMAGO_BIN)" build \
		-tags $(TAMAGO_TAGS) \
		-ldflags '$(TAMAGO_LDFLAGS)' \
		-o $(TAMAGO_ELF) \
		$(GO_MAIN)

build-tamago: $(TAMAGO_ELF)

$(TAMAGO_DTB): | $(BUILD_DIR)
	$(QEMU_SYSTEM) -machine sifive_u,dumpdtb=$@ -display none -S

$(TAMAGO_BIOS_ELF): $(TAMAGO_BIOS_SRC) $(TAMAGO_ELF) | $(BUILD_DIR)
	ENTRY_ADDR=`$(READELF) -h "$(TAMAGO_ELF)" | awk '/Entry point address:/ {print $$4}'` && \
	$(RISCV_GCC) \
		-nostdlib \
		-march=rv64imac \
		-mabi=lp64 \
		-Wl,-Ttext=0x1000 \
		-Wl,--build-id=none \
		-DENTRY_ADDR=$$ENTRY_ADDR \
		-x assembler-with-cpp \
		-o "$(TAMAGO_BIOS_ELF)" \
		"$(TAMAGO_BIOS_SRC)"

$(TAMAGO_BIOS_BIN): $(TAMAGO_BIOS_ELF) | $(BUILD_DIR)
	$(RISCV_OBJCOPY) -O binary "$(TAMAGO_BIOS_ELF)" "$(TAMAGO_BIOS_BIN)"

emulate-clang: $(CLANG_ELF)
	@status=0; \
	timeout $(QEMU_TIMEOUT) \
		$(QEMU_SYSTEM) \
		-machine $(QEMU_MACHINE) \
		-cpu $(QEMU_CPU) \
		-m $(QEMU_MEM) \
		-nographic \
		-bios none \
		-kernel $(CLANG_ELF) || status=$$?; \
	test $$status -eq 0 -o $$status -eq 124

emulate-gcc: $(GCC_ELF)
	@status=0; \
	timeout $(QEMU_TIMEOUT) \
		$(QEMU_SYSTEM) \
		-machine $(QEMU_MACHINE) \
		-cpu $(QEMU_CPU) \
		-m $(QEMU_MEM) \
		-nographic \
		-bios none \
		-kernel $(GCC_ELF) || status=$$?; \
	test $$status -eq 0 -o $$status -eq 124

emulate-go-linux: $(GO_LINUX_BIN)
	$(QEMU_USER) $(GO_LINUX_BIN)

emulate-tinygo: $(TINYGO_ELF)
	@status=0; \
	timeout $(QEMU_TIMEOUT) \
		$(QEMU_SYSTEM) \
		-machine $(QEMU_MACHINE) \
		-cpu $(QEMU_CPU) \
		-m $(QEMU_MEM) \
		-nographic \
		-bios none \
		-kernel $(TINYGO_ELF) || status=$$?; \
	test $$status -eq 0 -o $$status -eq 124

emulate-tamago: $(TAMAGO_ELF) $(TAMAGO_BIOS_BIN) $(TAMAGO_DTB)
	@status=0; \
	timeout $(QEMU_TIMEOUT) \
		$(QEMU_SYSTEM) \
		-machine sifive_u \
		-m 512M \
		-nographic \
		-monitor none \
		-serial stdio \
		-net none \
		-dtb $(TAMAGO_DTB) \
		-bios $(TAMAGO_BIOS_BIN) \
		-kernel $(TAMAGO_ELF) || status=$$?; \
	test $$status -eq 0 -o $$status -eq 124

emulate-libriscv: build-libriscv-runner $(LIBRISCV_GUEST)
	@if [ "$(abspath $(LIBRISCV_GUEST))" = "$(abspath $(TAMAGO_ELF))" ]; then \
		printf '%s\n' 'emulate-libriscv: $(TAMAGO_ELF) is a TamaGo sifive_u image and expects machine-mode CSRs/board state.'; \
		printf '%s\n' 'emulate-libriscv: use make emulate-tamago for this guest, or pass a generic bare-metal ELF such as build/verifier-clang.elf.'; \
		exit 2; \
	fi
	"$(LIBRISCV_RUNNER)" "$(LIBRISCV_GUEST)" "$(LIBRISCV_MAX_INSTRUCTIONS)"

clean:
	rm -rf $(BUILD_DIR)

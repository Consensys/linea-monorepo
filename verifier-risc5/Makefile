.DEFAULT_GOAL := help
.SILENT:

GO_MAIN := ./cmd/verifier
INPUTGEN_MAIN := ./cmd/inputgen
GO_SOURCES := $(shell find cmd internal -name '*.go' -print)

BUILD_DIR := build
BAREMETAL_DIR := baremetal
INPUTS_DIR := inputs
LIBRISCV_TOOLCHAIN_DIR := toolchains/libriscv
TAMAGO_TOOLCHAIN_DIR := toolchains/tamago
TINYGO_TOOLCHAIN_DIR := toolchains/tinygo

LIBRISCV_CMAKE := $(LIBRISCV_TOOLCHAIN_DIR)/CMakeLists.txt
LIBRISCV_RUNNER_SRC := $(LIBRISCV_TOOLCHAIN_DIR)/runner.cpp
TAMAGO_BIOS_SRC := $(TAMAGO_TOOLCHAIN_DIR)/sifive_u_bios.S
TINYGO_TARGET := $(TINYGO_TOOLCHAIN_DIR)/riscv64im_zicclsm-qemu-virt.json
TINYGO_TARGET_LD := $(TINYGO_TOOLCHAIN_DIR)/riscv64im_zicclsm-qemu-virt.ld

GO ?= go
CMAKE ?= cmake
CLANG ?= clang
LD_LLD ?= ld.lld
QEMU_SYSTEM ?= qemu-system-riscv64
QEMU_USER ?= qemu-riscv64
READELF ?= readelf
RISCV_GCC ?= riscv64-unknown-elf-gcc
RISCV_OBJCOPY ?= riscv64-unknown-elf-objcopy

HOST_GO_VERSION := $(shell $(GO) env GOVERSION)
XDG_CACHE_HOME_LOCAL ?= $(abspath .cache/xdg)

LLVM_VERSION ?= 20
LLVM_BIN_DIR ?= /usr/lib/llvm-$(LLVM_VERSION)/bin
LLVM_DEV_DIR ?= /usr/lib/llvm-$(LLVM_VERSION)
LLVM_INCLUDE_DIR ?= $(if $(wildcard $(LLVM_DEV_DIR)/include/llvm-c/Core.h),$(LLVM_DEV_DIR)/include,/usr/include/llvm-$(LLVM_VERSION))
CLANG := $(if $(wildcard $(LLVM_BIN_DIR)/clang),$(LLVM_BIN_DIR)/clang,$(CLANG))
LD_LLD := $(if $(wildcard $(LLVM_BIN_DIR)/ld.lld),$(LLVM_BIN_DIR)/ld.lld,$(LD_LLD))

TINYGO_ROOT ?= third_party/tinygo
TINYGO_VERSION ?= v0.41.0
TINYGO_COMMIT ?= 1f114b2acff247f52eb77a92363119e7327107b3
TINYGO_PATCH ?= third_party/tinygo-riscv64im_zicclsm.patch
TINYGO_BIN := $(abspath $(TINYGO_ROOT))/build/tinygo
TINYGO_LLVM_PROJECT_BRANCH ?= tinygo_20.x
TINYGO_LLVM_PROJECT_DIR := $(TINYGO_ROOT)/llvm-project
TINYGO_TARGET_BUILD := $(BUILD_DIR)/$(notdir $(TINYGO_TARGET))
TINYGO_SCHEDULER ?= none
TINYGO_GC ?= none
TINYGO_PANIC ?= trap
TINYGO_OPT ?= z
TINYGO_NO_DEBUG ?= 1
TINYGO_EXTRA_FLAGS ?=
TINYGO_ELF := $(BUILD_DIR)/verifier-tinygo.elf
TINYGO_QEMU_MACHINE ?= virt,aclint=on
TINYGO_QEMU_SMP ?= 1

TAMAGO_TOOL_VERSION ?= v1.26.2
TAMAGO_CACHE_HOME ?= $(XDG_CACHE_HOME_LOCAL)
TAMAGO_GOROOT := $(TAMAGO_CACHE_HOME)/tamago-go/tamago-go$(patsubst v%,%,$(TAMAGO_TOOL_VERSION))
TAMAGO_BIN := $(TAMAGO_GOROOT)/bin/go
TAMAGO_TAGS ?= baremetal,tamago_sifive_u,tiny
TAMAGO_LDFLAGS ?= -T 0x80010000 -R 0x1000
TAMAGO_ELF := $(BUILD_DIR)/verifier-tamago-sifive_u.elf
TAMAGO_BIOS_ELF := $(BUILD_DIR)/tamago-sifive_u-bios.elf
TAMAGO_BIOS_BIN := $(BUILD_DIR)/tamago-sifive_u-bios.bin
TAMAGO_DTB := $(BUILD_DIR)/qemu-sifive_u.dtb

LIBRISCV_ROOT ?= third_party/libriscv
LIBRISCV_VERSION ?= v1.15
LIBRISCV_BUILD_DIR := $(BUILD_DIR)/libriscv-host
LIBRISCV_RUNNER := $(LIBRISCV_BUILD_DIR)/libriscv_runner
LIBRISCV_GUEST ?= $(CLANG_ELF)
LIBRISCV_MAX_INSTRUCTIONS ?= 50000000

QEMU_TIMEOUT ?=
QEMU_MACHINE ?= virt
QEMU_CPU ?= rv64
QEMU_MEM ?= 256M
QEMU_FINISHER_SUCCESS_EXIT ?= 0
QEMU_FINISHER_FAILURE_EXIT ?= 3
QEMU_TIMEOUT_CMD := $(if $(strip $(QEMU_TIMEOUT)),timeout $(QEMU_TIMEOUT),)

TARGET_TRIPLE ?= riscv64-unknown-none-elf
RV_MARCH ?= rv64im_zicclsm
RV_MARCH_GCC ?= $(patsubst %_zicclsm,%,$(RV_MARCH))
RV_ABI ?= lp64
RV_CODE_MODEL ?= medany
GO_RISCV_PROFILE ?= rva20u64

CLANG_ELF := $(BUILD_DIR)/verifier-clang.elf
GCC_ELF := $(BUILD_DIR)/verifier-gcc.elf
GO_LINUX_BIN := $(BUILD_DIR)/verifier-go-linux-riscv64
INPUT_JSON ?= $(INPUTS_DIR)/default.json
INPUT_BIN := $(BUILD_DIR)/verifier-input.bin
GUEST_INPUT_ADDR ?= 0x80f00000
TINYGO_DEBUG_FLAG := $(if $(filter 1 true yes,$(TINYGO_NO_DEBUG)),-no-debug,)

.PHONY: help \
	clean-tool-cache vendor-tinygo tinygo-bootstrap \
	vendor-libriscv build-libriscv-runner \
	build-input build-clang build-gcc build-go-linux build-tinygo build-tamago \
	emulate-clang emulate-gcc emulate-go-linux emulate-tinygo emulate-tamago emulate-libriscv \
	clean

define check_qemu_finisher_status
case $$status in \
	$(QEMU_FINISHER_SUCCESS_EXIT)) \
		printf '%s\n' '$(1): guest reported success' ;; \
	$(QEMU_FINISHER_FAILURE_EXIT)) \
		printf '%s\n' '$(1): guest reported failure via the QEMU finisher' >&2; \
		exit $$status ;; \
	124) \
		printf '%s\n' '$(1): timed out before the guest signaled completion' >&2; \
		exit $$status ;; \
	*) \
		printf '%s %s\n' '$(1): unexpected QEMU exit' "$$status" >&2; \
		exit $$status ;; \
esac
endef

help:
	printf '%s\n' \
		'Targets:' \
		'  make vendor-tinygo        - clone TinyGo at the pinned commit and apply the local RISC-V patch' \
		'  make vendor-libriscv      - clone libriscv and its submodules under third_party/' \
		'  make tinygo-bootstrap     - build the local TinyGo binary from third_party/tinygo' \
		'  make build-input          - generate the preloaded verifier input blob' \
		'  make build-clang          - build the freestanding clang/ld.lld ELF' \
		'  make build-gcc            - build the freestanding GCC ELF' \
		'  make build-go-linux       - build the hosted linux/riscv64 Go baseline' \
		'  make build-tinygo         - build the freestanding TinyGo ELF' \
		'  make build-tamago         - build the TamaGo sifive_u ELF' \
		'  make build-libriscv-runner - build the local libriscv bare-metal runner' \
		'  make emulate-clang        - run the clang ELF on QEMU virt' \
		'  make emulate-gcc          - run the GCC ELF on QEMU virt' \
		'  make emulate-go-linux     - run the hosted Go ELF under qemu-riscv64' \
		'  make emulate-tinygo       - run the TinyGo ELF on QEMU virt' \
		'  make emulate-tamago       - run the TamaGo ELF on QEMU sifive_u' \
		'  make emulate-libriscv     - run a generic bare-metal ELF in the vendored libriscv runner' \
		'  make clean-tool-cache     - remove local TinyGo/TamaGo caches under .cache/' \
		'' \
		'Useful overrides:' \
		'  LLVM_VERSION=20                   - host LLVM major version used by clang/lld and TinyGo bootstrap' \
		'  TINYGO_ROOT=third_party/tinygo    - TinyGo checkout location' \
		'  TINYGO_COMMIT=$(TINYGO_COMMIT) - pinned TinyGo commit expected by the local patch' \
		'  TINYGO_GC=none                    - TinyGo GC mode, defaults to the no-allocation profile for this guest' \
		'  TINYGO_NO_DEBUG=1                - strip DWARF/debug sections from the TinyGo ELF (default on)' \
		'  TINYGO_EXTRA_FLAGS=-nobounds     - pass extra TinyGo build flags for local experiments' \
		'  INPUT_JSON=inputs/default.json   - JSON fixture used to generate the preloaded input blob' \
		'  LIBRISCV_ROOT=third_party/libriscv - libriscv checkout location' \
		'  LIBRISCV_GUEST=build/verifier-gcc.elf - ELF passed to emulate-libriscv' \
		'  GUEST_INPUT_ADDR=0x80f00000      - guest RAM address where the input blob is preloaded' \
		'  QEMU_CPU=rv64                     - QEMU CPU string for the virt machine targets' \
		'  QEMU_TIMEOUT=5s                  - optional safety timeout for QEMU system emulation' \
		'' \
		'Ubuntu packages usually needed:' \
		'  sudo apt install clang-20 lld-20 llvm-20-tools llvm-20-dev libclang-20-dev gcc-riscv64-unknown-elf binutils-riscv64-unknown-elf qemu-system-misc qemu-user cmake g++' \
		'' \
		'Notes:' \
		'  TinyGo is vendored instead of module-managed because the upstream module archive omits required git submodules such as src/net.' \
		'  vendor-tinygo checks that the checkout is at the pinned commit before applying third_party/tinygo-riscv64im_zicclsm.patch.' \
		'  TamaGo remains pinned in go.mod and is bootstrapped via go tool tamago into .cache/xdg.' \
		'  emulate-libriscv is for the generic freestanding clang/gcc guests; the current TinyGo and TamaGo images still touch machine CSRs that this minimal runner does not model.' \
		'  Upstream Go cannot target unknown-none-elf; build-go-linux is only a hosted baseline.' \
		'  The freestanding clang and TinyGo paths target rv64im_zicclsm. The GCC path currently drops zicclsm because the packaged GCC toolchain only accepts rv64im.' \
		'  TinyGo now defaults to a single-hart, no-debug, -gc=none profile plus a minimal target-local runtime variant; use TINYGO_GC=leaking if your guest starts allocating.'

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

build-input: | $(BUILD_DIR)
	$(GO) run $(INPUTGEN_MAIN) -in $(INPUT_JSON) -out $(INPUT_BIN)

$(TINYGO_TARGET_BUILD): $(TINYGO_TARGET) $(TINYGO_TARGET_LD) | $(BUILD_DIR)
	sed 's|"linkerscript": "riscv64im_zicclsm-qemu-virt.ld"|"linkerscript": "$(abspath $(TINYGO_TARGET_LD))"|' "$(TINYGO_TARGET)" > "$@"

clean-tool-cache:
	rm -rf $(XDG_CACHE_HOME_LOCAL) $(BUILD_DIR)/xdg-cache

vendor-tinygo: $(TINYGO_PATCH)
	if [ "$(TINYGO_ROOT)" != "vendor/tinygo" ] && [ -d "vendor/tinygo" ] && [ ! -e "$(TINYGO_ROOT)" ]; then mkdir -p "$(dir $(TINYGO_ROOT))" && mv "vendor/tinygo" "$(TINYGO_ROOT)"; fi
	if [ "$(TINYGO_ROOT)" != "vendor/tinygo" ]; then rmdir "vendor" 2>/dev/null || true; fi
	test -f "$(TINYGO_PATCH)"
	test -d "$(TINYGO_ROOT)" || git clone --branch "$(TINYGO_VERSION)" --depth 1 --recursive https://github.com/tinygo-org/tinygo "$(TINYGO_ROOT)"
	test -f "$(TINYGO_ROOT)/go.mod"
	head="$$(git -C "$(TINYGO_ROOT)" rev-parse HEAD)"; \
	if [ "$$head" != "$(TINYGO_COMMIT)" ]; then \
		printf '%s\n' "vendor-tinygo: $(TINYGO_ROOT) is at $$head, expected $(TINYGO_COMMIT)."; \
		printf '%s\n' "vendor-tinygo: remove $(TINYGO_ROOT) and rerun if you want a clean repin."; \
		exit 1; \
	fi
	git -C "$(TINYGO_ROOT)" submodule update --init --recursive
	test -d "$(TINYGO_LLVM_PROJECT_DIR)" || git clone --branch "$(TINYGO_LLVM_PROJECT_BRANCH)" --depth 1 https://github.com/tinygo-org/llvm-project "$(TINYGO_LLVM_PROJECT_DIR)"
	if git -C "$(TINYGO_ROOT)" apply --reverse --check "$(abspath $(TINYGO_PATCH))" >/dev/null 2>&1; then \
		:; \
	elif git -C "$(TINYGO_ROOT)" apply --check "$(abspath $(TINYGO_PATCH))" >/dev/null 2>&1; then \
		git -C "$(TINYGO_ROOT)" apply "$(abspath $(TINYGO_PATCH))"; \
	else \
		printf '%s\n' "vendor-tinygo: patch $(TINYGO_PATCH) does not apply to $(TINYGO_COMMIT)."; \
		exit 1; \
	fi

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

$(GO_LINUX_BIN): $(GO_SOURCES) go.mod go.sum | $(BUILD_DIR)
	GOOS=linux GOARCH=riscv64 GORISCV64=$(GO_RISCV_PROFILE) \
		$(GO) build -trimpath -o $(GO_LINUX_BIN) $(GO_MAIN)

build-go-linux: $(GO_LINUX_BIN)

$(TINYGO_ELF): $(TINYGO_BIN) $(TINYGO_TARGET_BUILD) $(TINYGO_TARGET_LD) $(GO_SOURCES) go.mod go.sum | $(BUILD_DIR)
	XDG_CACHE_HOME="$(XDG_CACHE_HOME_LOCAL)" \
	GOTOOLCHAIN="$(HOST_GO_VERSION)" \
	TINYGOROOT="$(abspath $(TINYGO_ROOT))" \
	"$(TINYGO_BIN)" build \
		-target=$(abspath $(TINYGO_TARGET_BUILD)) \
		-scheduler=$(TINYGO_SCHEDULER) \
		-gc=$(TINYGO_GC) \
		-panic=$(TINYGO_PANIC) \
		-opt=$(TINYGO_OPT) \
		$(TINYGO_DEBUG_FLAG) \
		$(TINYGO_EXTRA_FLAGS) \
		-o $(TINYGO_ELF) \
		$(GO_MAIN)

build-tinygo: $(TINYGO_ELF)

$(TAMAGO_BIN): | $(BUILD_DIR)
	if [ ! -x "$(TAMAGO_BIN)" ]; then rm -rf "$(TAMAGO_GOROOT)"; fi
	test -x "$(TAMAGO_BIN)" || env -u GOOS -u GOARCH -u GOOSPKG XDG_CACHE_HOME="$(TAMAGO_CACHE_HOME)" $(GO) tool tamago version >/dev/null
	test -x "$(TAMAGO_BIN)"

$(TAMAGO_ELF): $(TAMAGO_BIN) $(GO_SOURCES) go.mod go.sum | $(BUILD_DIR)
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

emulate-clang: build-input $(CLANG_ELF)
	status=0; \
	$(QEMU_TIMEOUT_CMD) \
		$(QEMU_SYSTEM) \
		-machine $(QEMU_MACHINE) \
		-cpu $(QEMU_CPU) \
		-m $(QEMU_MEM) \
		-nographic \
		-bios none \
		-device loader,file=$(INPUT_BIN),addr=$(GUEST_INPUT_ADDR),force-raw=on \
		-kernel $(CLANG_ELF) || status=$$?; \
	$(call check_qemu_finisher_status,emulate-clang)

emulate-gcc: build-input $(GCC_ELF)
	status=0; \
	$(QEMU_TIMEOUT_CMD) \
		$(QEMU_SYSTEM) \
		-machine $(QEMU_MACHINE) \
		-cpu $(QEMU_CPU) \
		-m $(QEMU_MEM) \
		-nographic \
		-bios none \
		-device loader,file=$(INPUT_BIN),addr=$(GUEST_INPUT_ADDR),force-raw=on \
		-kernel $(GCC_ELF) || status=$$?; \
	$(call check_qemu_finisher_status,emulate-gcc)

emulate-go-linux: $(GO_LINUX_BIN)
	$(QEMU_USER) $(GO_LINUX_BIN)

emulate-tinygo: build-input $(TINYGO_ELF)
	status=0; \
	$(QEMU_TIMEOUT_CMD) \
		$(QEMU_SYSTEM) \
		-machine $(TINYGO_QEMU_MACHINE) \
		-smp $(TINYGO_QEMU_SMP) \
		-cpu $(QEMU_CPU) \
		-m $(QEMU_MEM) \
		-nographic \
		-bios none \
		-device virtio-rng-device \
		-device loader,file=$(INPUT_BIN),addr=$(GUEST_INPUT_ADDR),force-raw=on \
		-kernel $(TINYGO_ELF) || status=$$?; \
	$(call check_qemu_finisher_status,emulate-tinygo)

emulate-tamago: build-input $(TAMAGO_ELF) $(TAMAGO_BIOS_BIN) $(TAMAGO_DTB)
	status=0; \
	$(QEMU_TIMEOUT_CMD) \
		$(QEMU_SYSTEM) \
		-machine sifive_u \
		-m 512M \
		-nographic \
		-monitor none \
		-serial stdio \
		-net none \
		-device loader,file=$(INPUT_BIN),addr=$(GUEST_INPUT_ADDR),force-raw=on \
		-dtb $(TAMAGO_DTB) \
		-bios $(TAMAGO_BIOS_BIN) \
		-kernel $(TAMAGO_ELF) || status=$$?; \
	test $$status -eq 0 -o $$status -eq 124

emulate-libriscv: build-input build-libriscv-runner $(LIBRISCV_GUEST)
	if [ "$(abspath $(LIBRISCV_GUEST))" = "$(abspath $(TAMAGO_ELF))" ]; then \
		printf '%s\n' 'emulate-libriscv: $(TAMAGO_ELF) is a TamaGo sifive_u image and expects machine-mode CSRs/board state.'; \
		printf '%s\n' 'emulate-libriscv: use make emulate-tamago for this guest, or pass a generic bare-metal ELF such as build/verifier-clang.elf.'; \
		exit 2; \
	fi
	if [ "$(abspath $(LIBRISCV_GUEST))" = "$(abspath $(TINYGO_ELF))" ]; then \
		printf '%s\n' 'emulate-libriscv: $(TINYGO_ELF) currently touches machine CSRs such as mtvec during startup.'; \
		printf '%s\n' 'emulate-libriscv: use make emulate-tinygo for this guest, or pass a generic freestanding ELF such as build/verifier-clang.elf.'; \
		exit 2; \
	fi
	"$(LIBRISCV_RUNNER)" "$(LIBRISCV_GUEST)" "$(INPUT_BIN)" "$(LIBRISCV_MAX_INSTRUCTIONS)"

clean:
	rm -rf $(BUILD_DIR)

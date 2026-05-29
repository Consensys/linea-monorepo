# For the target we refer to https://github.com/eth-act/zkvm-standards/blob/main/standards/riscv-target/target.md.
# Note that Zicclsm extension does not affect the generated ELF so it is omitted.
# Moreover, ABI being LP64 (soft-float) is relevant only for float numbers, which we do not use, so it can be omitted as well.

# Declare all targets as phony to prevent conflicts with files of the same name
.PHONY: exec debug compile exec-elf elf-to-json install-zkc zkc-exec zkc-debug clean clean-all linker-script verify-elf vector-build vector-json vector-exec blake-rust-build blake-rust-json blake-rust-exec keccak-rust-build keccak-rust-json keccak-rust-exec act4-build act4-exec require-test require-src require-bin-ext require-vector-accepts require-vector-json-tools

# Keep validation, build, and execution targets running in order
.NOTPARALLEL:

######################################################
# Variables and support functions
######################################################

# Directory of this Makefile (used for relative paths)
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ARITHMETIZATION_DIR := $(abspath $(MAKEFILE_DIR)../../..)

# Paths to go main (to turn BIN into JSON), zkc main (to execute/debug the json) and linker script
GO_MAIN       := $(MAKEFILE_DIR)main.go
ZKC_MAIN      := $(MAKEFILE_DIR)../../main/riscv/main.zkc
LINKER_SCRIPT := $(MAKEFILE_DIR)linker_script.ld
# Assembly, Zig, or Rust test file to execute with exec target (TEST=src_optional_subfolder/your_test(.s|.zig|.rs)
TEST ?=

# Already compiled binary to execute with exec-elf target (BIN_EXT=path/to/your_bin)
BIN_EXT  ?=
JSON_EXT ?= $(BIN_EXT).json

# Generic test vector variables
VECTOR_ACCEPTS              ?=
VECTOR_N_VECTORS            ?= 0
# `VECTOR_JSON_MODE` is `dir` for one JSON per vector or `single` for one batched JSON
VECTOR_JSON_MODE            ?= dir
VECTOR_JSON                 ?= $(JSON)
VECTOR_JSON_DIR             ?= $(dir $(JSON))vector_json
VECTOR_IN_BYTES             ?= $(BIN).in_bytes.all
VECTOR_ELF2JSON             ?= $(BIN)_elf2json
VECTOR_ACCEPTS_TO_IN_BYTES  ?=
VECTOR_ACCEPTS2IN_BYTES     ?= $(BIN)_accepts2inbytes

# If TEST=src_optional_subfolder/your_test(.s|.zig|.rs), the variables below are set as follows:
# src_optional_subfolder/your_test
BASENAME     := $(basename $(TEST))

# your_test
NAME 		 := $(notdir $(BASENAME))

# src_optional_subfolder/
SUBDIR       := $(dir $(TEST))

# your_test(.s|.zig|.rs)
NOTSUBDIR    := $(notdir $(TEST))

# (.s|.zig|.rs)
EXT          := $(suffix $(TEST))

# Paths below are determined by TEST and the extension
SRC  := $(MAKEFILE_DIR)$(if $(filter .s,$(EXT)),asm,$(if $(filter .zig,$(EXT)),zig,rust))/src/$(TEST)
BIN  := $(MAKEFILE_DIR)$(if $(filter .s,$(EXT)),asm/bin/,$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,rust/target/riscv64im-unknown-none-elf/release/))$(NAME)
JSON := $(BIN).json
# Note: in the future the JSON file will be removed and main.zkc will receive directly the ELF file

# Inputs and offsets for the program
IN_BYTES        ?= ""

# Values to generate the linker script
# 128 MiB (0x00000000 -> 0x07FFFFFF)
PROGRAM_OFFSET  ?= 0x00000000

# 1 GiB   (0x08800000 -> 0x48FFFFFF)
IN_BYTES_OFFSET ?= 0x08800000

# Since program starts from _start, the entry point is the same as the program offset
ENTRY_POINT     := $(PROGRAM_OFFSET)

# 8 MiB   (0x08000000 <- 0x08800000)
SP              ?= 0x08800000

# Parsed values are just to verify that the generated ELF is correct (lazy evaluation allows to parse the ELF only after it is generated)
PROGRAM_OFFSET_FROM_ELF  = $(shell LANG=C riscv64-unknown-elf-readelf -S $(BIN) | grep '\.text' | awk '{print "0x"$$5}')
IN_BYTES_OFFSET_FROM_ELF = $(shell LANG=C riscv64-unknown-elf-nm $(BIN) | grep _input_start | awk '{print "0x"$$1}') 
ENTRY_POINT_FROM_ELF     = $(shell LANG=C riscv64-unknown-elf-readelf -h $(BIN) | grep 'Entry point' | awk '{print $$4}')
SP_FROM_ELF              = $(shell LANG=C riscv64-unknown-elf-nm $(BIN) | grep _stack_start | awk '{print "0x"$$1}')
VERIFY_ELF               ?= false             

# ACT4 testing framework specific variables
ACT4_CONFIG_DIR := $(MAKEFILE_DIR)act4/config
ACT4_WORK_DIR   := $(MAKEFILE_DIR)act4/bin/work
ACT4_CONFIG     := $(ACT4_CONFIG_DIR)/linea-rv64im-zicclsm/test_config.yaml
# The 3 parameters below must be consistent with the ones provided to arithmetization/Makefile
ACT4_REF        ?= 9798a554ce4139f472c9ccd3a18c9061d0f7024d
ACT4_REPO       ?= $(abspath $(shell git -C $(MAKEFILE_DIR) rev-parse --show-toplevel)/../riscv-arch-test)
ACT4_RISCV_DIR  ?= $(HOME)/riscv
ACT4_IMAGE      := riscv-act4:$(ACT4_REF)
ACT4_EXTENSIONS := I,M
ACT4_JOBS       := 4
ACT4_ELF_DIR    := $(ACT4_WORK_DIR)/linea-rv64im-zicclsm/elfs
ACT4_LOGS       := $(MAKEFILE_DIR)act4/bin/logs
ACT4_RESULTS    := $(MAKEFILE_DIR)act4/bin/results.txt
# ACT4_BUILD_MODE = (host|docker)
ACT4_BUILD_MODE ?= host
# Parameters passed to https://github.com/riscv/riscv-arch-test/blob/act4/Makefile:
# If true, disables objdump generation for faster builds
ACT4_FAST       ?= false
# If true, enables debug output (signature objdump, trace files, and trap report)
ACT4_DEBUG      ?= true

# String to determine if a program is exiting with exit code 0
EXIT_SUCCESS_STRING := "EXIT CODE = 0"

# Support function to generate JSON from an ELF
define ELF_TO_JSON
	@echo "\nGenerating $(2) from $(1)"
	@go run $(GO_MAIN) "$(1)" "$(IN_BYTES)" $(IN_BYTES_OFFSET) > "$(2)"
endef

######################################################
# Targets to handle generic tests
######################################################

# Default target (make)
exec: compile zkc-exec

debug: compile zkc-debug

compile: require-src linker-script
# Print all variables for debugging purposes
	@echo "TEST:      $(TEST)"
	@echo "BASENAME:  $(BASENAME)"
	@echo "NAME:      $(NAME)"
	@echo "SUBDIR:    $(SUBDIR)"
	@echo "NOTSUBDIR: $(NOTSUBDIR)"
	@echo "EXT:       $(EXT)"
	@echo "SRC:       $(SRC)"
	@echo "BIN:       $(BIN)"
	@echo "JSON:      $(JSON)"
# Generate ELF
	@if [ "$(EXT)" = ".s" ]; then \
		mkdir -p $(MAKEFILE_DIR)asm/bin; \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		riscv64-unknown-elf-gcc -march=rv64im -mabi=lp64 -nostdlib -T$(LINKER_SCRIPT) -o $(BIN) $(SRC); \
	elif [ "$(EXT)" = ".zig" ]; then \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		zig build --build-file $(MAKEFILE_DIR)zig/build.zig -Dpath=$(BASENAME); \
	elif [ "$(EXT)" = ".rs" ]; then \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		cargo +nightly rustc \
			--manifest-path $(MAKEFILE_DIR)rust/Cargo.toml \
			--target $(MAKEFILE_DIR)rust/riscv64im-unknown-none-elf.json \
			-Z build-std=core \
			-Z json-target-spec \
			--bin $(NAME) \
			--release \
			-- -C link-arg=-T$(LINKER_SCRIPT); \
	else \
		echo "Unsupported file extension: $(EXT)"; \
		exit 1; \
	fi
# Disassemble ELF
	@echo "\nDisassembling $(BIN)"
	@LANG=C riscv64-unknown-elf-objdump -D -s --line-numbers $(BIN) > $(BIN).objdump
# Verify that PROGRAM_OFFSET, IN_BYTES_OFFSET, ENTRY_POINT, SP match the ELF values
	@if [ "$(VERIFY_ELF)" = "true" ]; then \
		$(MAKE) -f $(MAKEFILE_DIR)Makefile verify-elf TEST=$(TEST); \
	fi
# Generate JSON from ELF
	$(call ELF_TO_JSON,$(BIN),$(JSON))

# Note: compiled Rust libraries may contain instructions from excluded extensions. For instance, the core library is compiled
# for the IMAC extension, and contains COMPRESSED instructions. The output ELF may thus contain such instructions, too, even
# if these extensions were explicitly exluded from the target.
#
# Note: we use opt-level=2 to remove unused code.

# Convert and execute an already compiled external ELF
exec-elf: elf-to-json
	zkc exec $(JSON_EXT) $(ZKC_MAIN)

# Convert an already compiled external ELF to a JSON
elf-to-json: require-bin-ext
# Generate JSON from ELF
	$(call ELF_TO_JSON,$(BIN_EXT),$(JSON_EXT))

install-zkc:
	$(MAKE) -C $(ARITHMETIZATION_DIR) install-zkc

zkc-exec: require-test install-zkc
	zkc exec $(JSON) $(ZKC_MAIN)

zkc-debug: require-test install-zkc
	zkc debug $(JSON) $(ZKC_MAIN)

clean: require-test
	rm -f $(BIN) $(JSON) $(BIN).objdump $(VECTOR_ELF2JSON) $(VECTOR_ACCEPTS2IN_BYTES) $(VECTOR_IN_BYTES) $(LINKER_SCRIPT)

clean-all:
	rm -rf $(MAKEFILE_DIR)asm/bin/
	rm -rf $(MAKEFILE_DIR)zig/zig-out/
	rm -rf $(MAKEFILE_DIR)zig/.zig-cache/
	rm -rf $(MAKEFILE_DIR)rust/target/
	rm -rf $(MAKEFILE_DIR)act4/bin/
	rm -f  $(LINKER_SCRIPT)

# Generate the linker script with the specified memory layout
linker-script:
	@echo "\nGenerating $(LINKER_SCRIPT)"
	@printf '%s\n' \
		'ENTRY(_start)' \
		'' \
		'MEMORY {' \
		'    PROGRAM (rx) : ORIGIN = $(PROGRAM_OFFSET), LENGTH = 0x8000000' \
		'    STACK   (rw) : ORIGIN = $(shell printf "0x%08x" $$(($(SP) - 0x800000))), LENGTH = 0x800000' \
		'    IN  (r)  : ORIGIN = $(IN_BYTES_OFFSET), LENGTH = 0x40000000' \
		'}' \
		'' \
		'SECTIONS {' \
		'    . = ORIGIN(PROGRAM);' \
		'' \
		'    .text : {' \
		'        *(.text .text.*)' \
		'    } > PROGRAM' \
		'' \
		'    .rodata : {' \
		'        *(.rodata .rodata.*)' \
		'    } > PROGRAM' \
		'' \
		'    .data : {' \
		'        *(.data .data.*)' \
		'    } > PROGRAM' \
		'' \
		'    .bss : {' \
		'        *(.bss .bss.*)' \
		'    } > PROGRAM' \
		'' \
		'    /* stack grows downward from the top of the stack region */' \
		'    _stack_start = ORIGIN(STACK) + LENGTH(STACK);' \
		'' \
		'    /* symbol marking the start of the input region */' \
		'    _input_start = ORIGIN(IN);' \
		'}' \
		> $(LINKER_SCRIPT)

# Verify that PROGRAM_OFFSET, IN_BYTES_OFFSET, ENTRY_POINT, SP match the ELF values
verify-elf: require-test
	@echo "\nVerify that PROGRAM_OFFSET, IN_BYTES_OFFSET, ENTRY_POINT, SP match the ELF values"
	@[ "$(shell printf '0x%016x' $$(($(PROGRAM_OFFSET_FROM_ELF))))" = "$(shell printf '0x%016x' $$(($(PROGRAM_OFFSET))))" ] || \
		(echo "ERROR: PROGRAM_OFFSET mismatch: expected $(PROGRAM_OFFSET), got $(PROGRAM_OFFSET_FROM_ELF)" && exit 1)
	@echo "  PROGRAM_OFFSET:  $(PROGRAM_OFFSET) (expected) = $(PROGRAM_OFFSET_FROM_ELF) (from ELF)"
	@[ "$(shell printf '0x%016x' $$(($(IN_BYTES_OFFSET_FROM_ELF))))" = "$(shell printf '0x%016x' $$(($(IN_BYTES_OFFSET))))" ] || \
		(echo "ERROR: IN_BYTES_OFFSET mismatch: expected $(IN_BYTES_OFFSET), got $(IN_BYTES_OFFSET_FROM_ELF)" && exit 1)
	@echo "  IN_BYTES_OFFSET: $(IN_BYTES_OFFSET) (expected) = $(IN_BYTES_OFFSET_FROM_ELF) (from ELF)"
	@[ "$(shell printf '0x%016x' $$(($(ENTRY_POINT_FROM_ELF))))" = "$(shell printf '0x%016x' $$(($(ENTRY_POINT))))" ] || \
		(echo "ERROR: ENTRY_POINT mismatch: expected $(ENTRY_POINT), got $(ENTRY_POINT_FROM_ELF)" && exit 1)
	@echo "  ENTRY_POINT:     $(ENTRY_POINT) (expected) = $(ENTRY_POINT_FROM_ELF) (from ELF)"
	@[ "$(shell printf '0x%016x' $$(($(SP_FROM_ELF))))" = "$(shell printf '0x%016x' $$(($(SP))))" ] || \
		(echo "ERROR: SP mismatch: expected $(SP), got $(SP_FROM_ELF)" && exit 1)
	@echo "  SP:              $(SP) (expected) = $(SP_FROM_ELF) (from ELF)"
	@echo "All values match\n"

###############################################################
# Targets to handle generic test vectors
###############################################################

vector-build: require-src
	$(MAKE) -f $(MAKEFILE_DIR)Makefile compile TEST=$(TEST)
	@echo "building $(VECTOR_ELF2JSON)..."
	go build -o $(VECTOR_ELF2JSON) $(GO_MAIN)
	@if [ -n "$(VECTOR_ACCEPTS_TO_IN_BYTES)" ]; then \
		echo "building $(VECTOR_ACCEPTS2IN_BYTES)..."; \
		go build -o $(VECTOR_ACCEPTS2IN_BYTES) $(VECTOR_ACCEPTS_TO_IN_BYTES); \
	fi

vector-json: require-src require-vector-accepts require-vector-json-tools
	@echo "selecting vectors from $(VECTOR_ACCEPTS) (N=$(VECTOR_N_VECTORS), 0=all if supported)"
	$(VECTOR_ACCEPTS2IN_BYTES) "$(VECTOR_ACCEPTS)" $(VECTOR_N_VECTORS) > $(VECTOR_IN_BYTES)
	@if [ "$(VECTOR_JSON_MODE)" = "single" ]; then \
		echo "rendering ZKC JSON to $(VECTOR_JSON)"; \
		$(VECTOR_ELF2JSON) $(BIN) @$(VECTOR_IN_BYTES) $(IN_BYTES_OFFSET) > $(VECTOR_JSON) || exit $$?; \
	elif [ "$(VECTOR_JSON_MODE)" = "dir" ]; then \
		mkdir -p $(VECTOR_JSON_DIR); \
		rm -f $(VECTOR_JSON_DIR)/*.json; \
		echo "rendering one ZKC JSON per vector under $(VECTOR_JSON_DIR)"; \
		n=0; while IFS= read -r in_bytes; do \
			n=$$((n + 1)); idx=$$(printf '%03d' $$n); \
			$(VECTOR_ELF2JSON) $(BIN) "$$in_bytes" $(IN_BYTES_OFFSET) > "$(VECTOR_JSON_DIR)/$$idx.json" || exit $$?; \
		done < $(VECTOR_IN_BYTES); \
		echo "generated $$n vector JSON(s) in $(VECTOR_JSON_DIR)"; \
	else \
		echo "error: VECTOR_JSON_MODE must be single or dir"; \
		exit 2; \
	fi

vector-exec: install-zkc
	@if [ "$(VECTOR_JSON_MODE)" = "single" ]; then \
		exit_line=$$(zkc exec $(VECTOR_JSON) $(ZKC_MAIN) 2>&1 | awk 'previous == "ECALL for exit" { print; exit } { previous = $$0 }'); \
		if [ "$$exit_line" = $(EXIT_SUCCESS_STRING) ]; then \
			echo "PASS  $(VECTOR_JSON)"; \
			echo "  $$exit_line"; \
		else \
			echo "FAIL  $(VECTOR_JSON)"; \
			echo "  $$exit_line"; \
			exit 1; \
		fi; \
	elif [ "$(VECTOR_JSON_MODE)" = "dir" ]; then \
		pass=0; fail=0; i=0; \
		for json in $(VECTOR_JSON_DIR)/*.json; do \
			[ -f "$$json" ] || { echo "error: no vector JSONs found in $(VECTOR_JSON_DIR)" >&2; exit 1; }; \
			i=$$((i + 1)); \
			echo "RUN   case $$i  ($$json)"; \
			exit_line=$$(zkc exec "$$json" $(ZKC_MAIN) 2>&1 | awk 'previous == "ECALL for exit" { print; exit } { previous = $$0 }'); \
			if [ "$$exit_line" = $(EXIT_SUCCESS_STRING) ]; then \
				echo "PASS  case $$i"; \
				echo "  $$exit_line"; \
				pass=$$((pass + 1)); \
			else \
				echo "FAIL  case $$i"; \
				echo "  $$exit_line"; \
				fail=$$((fail + 1)); \
			fi; \
		done; \
		echo "DONE: $$i vectors   PASS: $$pass   FAIL: $$fail"; \
		if [ "$$fail" -ne 0 ]; then echo "error: $$fail vector(s) failed" >&2; exit 1; fi; \
	else \
		echo "error: VECTOR_JSON_MODE must be single or dir"; \
		exit 2; \
	fi

###############################################################
# Keccak wrapper targets and benchmark workflow variables
###############################################################

KECCAK_ACCEPTS     ?= $(MAKEFILE_DIR)rust/src/keccak/keccak.accepts
KECCAK_N_VECTORS   ?= 10
KECCAK_JSON        ?= $(MAKEFILE_DIR)rust/target/riscv64im-unknown-none-elf/release/keccak.json

keccak-rust-build:
	cargo clean -p riscv-tests --manifest-path $(MAKEFILE_DIR)rust/Cargo.toml
	KECCAK_N_VECTORS=$(KECCAK_N_VECTORS) $(MAKE) -f $(MAKEFILE_DIR)Makefile vector-build \
		TEST=keccak/keccak.rs \
		IN_BYTES=0x00 \
		VECTOR_ACCEPTS_TO_IN_BYTES=$(MAKEFILE_DIR)scripts/keccak_accepts_to_in_bytes/main.go

keccak-rust-json: keccak-rust-build
	$(MAKE) -f $(MAKEFILE_DIR)Makefile vector-json \
		TEST=keccak/keccak.rs \
		VECTOR_ACCEPTS=$(KECCAK_ACCEPTS) \
		VECTOR_N_VECTORS=$(KECCAK_N_VECTORS) \
		VECTOR_JSON_MODE=single \
		VECTOR_JSON=$(KECCAK_JSON)

keccak-rust-exec: keccak-rust-json
	$(MAKE) -f $(MAKEFILE_DIR)Makefile vector-exec VECTOR_JSON_MODE=single VECTOR_JSON=$(KECCAK_JSON)

###############################################################
# Blake wrapper targets and benchmark workflow variables
###############################################################

BLAKE_ACCEPTS      ?= $(MAKEFILE_DIR)rust/src/blake/blake10.all
BLAKE_N_VECTORS    ?= 10
BLAKE_JSON_DIR     ?= $(MAKEFILE_DIR)rust/target/riscv64im-unknown-none-elf/release/blake_json

blake-rust-build:
	$(MAKE) -f $(MAKEFILE_DIR)Makefile vector-build \
		TEST=blake/blake_with_in_bytes.rs \
		VECTOR_ACCEPTS_TO_IN_BYTES=$(MAKEFILE_DIR)scripts/blake_accepts_to_in_bytes/main.go

blake-rust-json: blake-rust-build
	$(MAKE) -f $(MAKEFILE_DIR)Makefile vector-json \
		TEST=blake/blake_with_in_bytes.rs \
		VECTOR_ACCEPTS=$(BLAKE_ACCEPTS) \
		VECTOR_N_VECTORS=$(BLAKE_N_VECTORS) \
		VECTOR_JSON_MODE=dir \
		VECTOR_JSON_DIR=$(BLAKE_JSON_DIR)

blake-rust-exec: blake-rust-json
	$(MAKE) -f $(MAKEFILE_DIR)Makefile vector-exec VECTOR_JSON_MODE=dir VECTOR_JSON_DIR=$(BLAKE_JSON_DIR)

###############################################################
# ACT4 targets
###############################################################

# Build the ACT4 self-checking ELFs with the Linea ACT4 config 
# Note: build if needed
# Note: clone https://github.com/riscv/riscv-arch-test.git if needed
act4-build:
	@if [ -n "$$(find "$(ACT4_ELF_DIR)" -name '*.elf' -print -quit 2>/dev/null)" ]; then \
		echo "ACT4 ELFs already built in $(ACT4_ELF_DIR)"; \
		exit 0; \
	fi; \
	mkdir -p "$(ACT4_WORK_DIR)" || exit $$?; \
	work_dir=$$(cd "$(ACT4_WORK_DIR)" && pwd) || exit $$?; \
	[ -d "$(ACT4_CONFIG_DIR)/linea-rv64im-zicclsm" ] || { echo "error: missing $(ACT4_CONFIG_DIR)/linea-rv64im-zicclsm"; exit 2; }; \
	ensure_act4_repo() { \
		if [ ! -d "$(ACT4_REPO)" ]; then \
			git clone https://github.com/riscv/riscv-arch-test.git "$(ACT4_REPO)" || exit $$?; \
		fi; \
		if ! git -C "$(ACT4_REPO)" rev-parse "$(ACT4_REF)^{commit}" >/dev/null 2>&1; then \
			git -C "$(ACT4_REPO)" fetch --tags origin || exit $$?; \
		fi; \
		if [ "$$(git -C "$(ACT4_REPO)" rev-parse HEAD)" != "$$(git -C "$(ACT4_REPO)" rev-parse "$(ACT4_REF)^{commit}")" ]; then \
			git -C "$(ACT4_REPO)" checkout --detach "$(ACT4_REF)" || exit $$?; \
		fi; \
	}; \
	if [ "$(ACT4_BUILD_MODE)" = "docker" ]; then \
		config_dir=$$(cd "$(ACT4_CONFIG_DIR)" && pwd) || exit $$?; \
		if ! docker image inspect "$(ACT4_IMAGE)" >/dev/null 2>&1; then \
			ensure_act4_repo; \
			echo "building Docker image $(ACT4_IMAGE) from $(ACT4_REPO)"; \
			docker build -t "$(ACT4_IMAGE)" "$(ACT4_REPO)" || exit $$?; \
		fi; \
		echo "building ACT4 ELFs in $$work_dir with Docker"; \
		docker run --rm \
			-v "$$config_dir:/act4/config/cores/linea:ro" \
			-v "$$work_dir:/act4/work" \
			"$(ACT4_IMAGE)" \
			bash -c "rm -rf /act4/work/linea-rv64im-zicclsm && CONFIG_FILES=config/cores/linea/linea-rv64im-zicclsm/test_config.yaml EXTENSIONS=$(ACT4_EXTENSIONS) $(if $(filter true,$(ACT4_DEBUG)),DEBUG=true) $(if $(filter true,$(ACT4_FAST)),FAST=true) make --jobs $(ACT4_JOBS) --keep-going" || exit $$?; \
	elif [ "$(ACT4_BUILD_MODE)" = "host" ]; then \
		ensure_act4_repo; \
		echo "building ACT4 ELFs in $$work_dir on host"; \
		export PATH="$(ACT4_RISCV_DIR)/bin:$$PATH"; \
		$(MAKE) --jobs $(ACT4_JOBS) --keep-going -C "$(ACT4_REPO)" \
			CONFIG_FILES="$(ACT4_CONFIG)" \
			WORKDIR="$$work_dir" \
			EXTENSIONS="$(ACT4_EXTENSIONS)" \
			$(if $(filter true,$(ACT4_DEBUG)),DEBUG=true) \
			$(if $(filter true,$(ACT4_FAST)),FAST=true) || exit $$?; \
	else \
		echo "error: ACT4_BUILD_MODE must be docker or host"; \
		exit 2; \
	fi

# Build and run the ACT4 self-checking ELFs with the Linea ACT4 config.
act4-exec: install-zkc act4-build
	@rm -f "$(ACT4_LOGS)"/*; \
	mkdir -p "$(ACT4_LOGS)"; \
	> "$(ACT4_RESULTS)"; \
	count=0; pass=0; fail=0; \
	for elf in $$(find "$(ACT4_ELF_DIR)" -name '*.elf' | sort); do \
		name=$$(basename "$$elf" .elf); \
		base="$(ACT4_LOGS)/$$name"; \
		count=$$((count + 1)); \
		go run $(GO_MAIN) "$$elf" "" 0 > "$$base.json" 2> "$$base.json.err"; \
		[ -s "$$base.json.err" ] || rm -f "$$base.json.err"; \
		zkc exec "$$base.json" $(ZKC_MAIN) > "$$base.full" 2>&1; \
		awk '(previous == "ECALL for write" || previous == "ECALL for exit") && $$0 != "" { print } { previous = $$0 }' "$$base.full" > "$$base.out"; \
		exit_line=$$(awk 'previous == "ECALL for exit" { print; exit } { previous = $$0 }' "$$base.full"); \
		if [ "$$exit_line" = $(EXIT_SUCCESS_STRING) ]; then \
			echo "PASS  $$name"; \
			echo "PASS  $$name" >> "$(ACT4_RESULTS)"; \
			awk '{ print "  " $$0 }' "$$base.out"; \
			awk '{ print "  " $$0 }' "$$base.out" >> "$(ACT4_RESULTS)"; \
			pass=$$((pass + 1)); \
			rm -f "$$base.full"; \
		else \
			echo "FAIL  $$name"; \
			echo "FAIL  $$name" >> "$(ACT4_RESULTS)"; \
			awk '{ print "  " $$0 }' "$$base.out"; \
			awk '{ print "  " $$0 }' "$$base.out" >> "$(ACT4_RESULTS)"; \
			fail=$$((fail + 1)); \
		fi; \
	done; \
	echo "DONE: $$count tests   PASS: $$pass   FAIL: $$fail" | tee -a "$(ACT4_RESULTS)"; \
	echo "results: $(ACT4_RESULTS)"; \
	echo "logs:    $(ACT4_LOGS)/"; \
	if [ "$$count" -eq 0 ]; then echo "error: no ACT4 ELFs found in $(ACT4_ELF_DIR)" >&2; exit 1; fi; \
	if [ "$$fail" -ne 0 ]; then echo "error: $$fail of $$count ACT4 test(s) failed" >&2; exit 1; fi

######################################################
# Support targets to validate mandatory parameters
######################################################

require-test:
	@[ -n "$(TEST)" ] || { echo "TEST is mandatory. Usage: make TEST=src_optional_subfolder/your_test(.s|.zig|.rs)"; exit 1; }

require-src: require-test
	@[ -f $(SRC) ] || { echo "TEST file not found: $(SRC)"; exit 1; }

require-bin-ext:
	@[ -n "$(BIN_EXT)" ] || { echo "BIN_EXT is mandatory. Usage: make BIN_EXT=path/to/your_bin"; exit 1; }
	@[ -f $(BIN_EXT) ] || { echo "BIN_EXT file not found: $(BIN_EXT)"; exit 1; }

require-vector-accepts:
	@[ -n "$(VECTOR_ACCEPTS)" ] || { echo "VECTOR_ACCEPTS is mandatory. Usage: make vector-json TEST=src_optional_subfolder/your_test(.s|.zig|.rs) VECTOR_ACCEPTS=path/to/vectors"; exit 1; }
	@[ -f $(VECTOR_ACCEPTS) ] || { echo "VECTOR_ACCEPTS file not found: $(VECTOR_ACCEPTS)"; exit 1; }

require-vector-json-tools:
	@[ -x $(VECTOR_ELF2JSON) ] || { echo "VECTOR_ELF2JSON not found or not executable: $(VECTOR_ELF2JSON). Run vector-build first"; exit 1; }
	@[ -x $(VECTOR_ACCEPTS2IN_BYTES) ] || { echo "VECTOR_ACCEPTS2IN_BYTES not found or not executable: $(VECTOR_ACCEPTS2IN_BYTES). Run vector-build first"; exit 1; }

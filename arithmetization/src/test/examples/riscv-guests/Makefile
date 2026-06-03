# Shared RISC-V guest build and ZKC integration.

.PHONY: fetch compile json exec debug zkc-exec zkc-debug test write-fixtures fixture-exec fixture-debug clean

MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

#########################################################################
# Potentially new logic wrt /arithmetization/src/test/examples/Makefile
#########################################################################

GUEST  ?= l2-execution
ZIG    ?= zig
GO     ?= go
GO_ENV ?= GO111MODULE=off
ZKC    ?= zkc

BUILD_CACHE_DIR ?= $(MAKEFILE_DIR).zig-cache
ZIG_BUILD_FLAGS ?= --cache-dir $(BUILD_CACHE_DIR) --summary all
ZIG_FETCH_FLAGS ?= --fetch

ifeq ($(GUEST),l2-execution)
GUEST_BIN    := evm_execution_guest
FIXTURE_JSON := $(MAKEFILE_DIR)l2-execution/fixtures/contract_creation_then_ecrecover.json
else
$(error Unsupported GUEST '$(GUEST)'. Supported guests: l2-execution)
endif

fetch:
	$(ZIG) build $(ZIG_BUILD_FLAGS) --fetch

fixture-exec:
	$(MAKE) -f $(MAKEFILE_DIR)Makefile exec GUEST=$(GUEST) IN_BYTES="$$(sed -n 's/.*"payload": "\(0x[0-9a-fA-F]*\)".*/\1/p' "$(FIXTURE_JSON)")"

fixture-debug:
	$(MAKE) -f $(MAKEFILE_DIR)Makefile debug GUEST=$(GUEST) IN_BYTES="$$(sed -n 's/.*"payload": "\(0x[0-9a-fA-F]*\)".*/\1/p' "$(FIXTURE_JSON)")"

test:
	$(ZIG) build test $(ZIG_BUILD_FLAGS) $(ZIG_FETCH_FLAGS)

write-fixtures:
	$(ZIG) build write-fixtures $(ZIG_BUILD_FLAGS) $(ZIG_FETCH_FLAGS)

#########################################################################
# Mostly duplicated wrt /arithmetization/src/test/examples/Makefile
#########################################################################

BIN             := $(MAKEFILE_DIR)zig-out/bin/$(GUEST_BIN)
JSON            := $(MAKEFILE_DIR)zig-out/bin/$(GUEST_BIN).json
ELF_TO_JSON_GEN := $(MAKEFILE_DIR)../arithmetization/src/test/examples/scripts/elf_to_json_gen/main.go
ZKC_MAIN        := $(abspath $(MAKEFILE_DIR)../arithmetization/src/main/riscv/main.zkc)

IN_BYTES        ?=
IN_BYTES_OFFSET ?= 0x08800000
SP              ?= 0x08800000

exec: elf-to-json zkc-exec

debug: elf-to-json zkc-debug

compile:
	$(ZIG) build $(ZIG_BUILD_FLAGS) $(ZIG_FETCH_FLAGS) -Dprogram-offset=$(PROGRAM_OFFSET) -Dinput-offset=$(IN_BYTES_OFFSET) -Dsp=$(SP)
	@if command -v riscv64-unknown-elf-objdump >/dev/null 2>&1; then \
		echo "\nDisassembling $(BIN)"; \
		LANG=C riscv64-unknown-elf-objdump -D -s --line-numbers $(BIN) > $(BIN).objdump; \
	else \
		echo "\nSkipping objdump: riscv64-unknown-elf-objdump is not on PATH"; \
	fi

elf-to-json: compile
	@echo "\nGenerating $(JSON) from $(BIN)"
	@$(GO_ENV) $(GO) run $(ELF_TO_JSON_GEN) "$(BIN)" "$(IN_BYTES)" $(IN_BYTES_OFFSET) > "$(JSON)"

zkc-exec:
	$(ZKC) exec $(JSON) $(ZKC_MAIN)

zkc-debug:
	$(ZKC) debug $(JSON) $(ZKC_MAIN)

clean:
	rm -rf $(MAKEFILE_DIR)zig-out/
	rm -rf $(BUILD_CACHE_DIR)


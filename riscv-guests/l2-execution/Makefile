# l2-execution guest — self-contained lifecycle.
#
# This folder is its own Zig package (build.zig + build.zig.zon). The riscv-guests/ top-level
# Makefile fans `gp-compile`/`test`/`spec-test`/`fetch`/`clean` out to this file for every guest;
# guest-specific targets (the ZKC fixture-exec/fixture-debug below) live here, not in the orchestrator.
#
# Scope split: building + host-testing is proving-system-AGNOSTIC and lives here (`compile` links the
# statically-linked rv64im ELF — the zkvm-standards artifact — via build_common). Only the ZKC run
# itself (ELF->JSON via elf_to_json_gen, then zkc) is owned by
# arithmetization/src/test/examples/Makefile; fixture-exec/fixture-debug build that ELF and hand it over.

# guest-program = gp

.PHONY: fetch test spec-test clean gp-compile gp-linker-script gp-exec gp-debug 

MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

ZIG ?= zig

BUILD_CACHE_DIR ?= $(MAKEFILE_DIR).zig-cache
ZIG_BUILD_FLAGS ?= --cache-dir $(BUILD_CACHE_DIR) --summary all
# NOTE: `zig build --fetch` fetches the dependency tree and EXITS without building or running, so it
# must NOT be passed to build/test targets. Pre-fetching is the dedicated `fetch` target.
ZIG_FETCH_FLAGS ?=

# Folder of the Makefile containaining the logic for generating a linker scritpt,
# executing and debugging the guest program.
EXAMPLES_DIR := $(abspath $(MAKEFILE_DIR)../../arithmetization/src/test/examples)

# Where the ZKC run places the input — forwarded to the examples Makefile as IN_ORIGIN, and it must
# match the ELF's `_in_start` (build_common's linker script: 0x08800000). Also passed
# as -Din-origin, which the guest no longer consumes (it reads `_in_start` via read_input).
IN_ORIGIN ?= 0x08800000

# Hex big-endian input or `@path/to/in_bytes`
IN_BYTES ?= ""

GUEST_BIN    := evm_execution_guest
# Committed sample input for the fixture-exec/-debug convenience runs: a real fixture's SSZ framed
# [u64 LE len][SSZ] and pre-reversed for the converter (see "running the guest program in zkc.md",
# which also regenerates it). Lives under test/testdata/ with the other host-only data.
SAMPLE_INPUT := $(MAKEFILE_DIR)test/testdata/fixture_input.in.hex

# Guest artifact: the statically-linked rv64im ELF `compile` produces (entry _start + memory layout
# from build_common) — the zkvm-standards deliverable. The ZKC exec/debug runs feed it to ELF->JSON.
ELF_BIN := $(MAKEFILE_DIR)zig-out/bin/$(GUEST_BIN)

# ── Proving-system-agnostic guest lifecycle ─────────────────────────────────
fetch:
	$(ZIG) build $(ZIG_BUILD_FLAGS) --fetch

test:
	$(ZIG) build test $(ZIG_BUILD_FLAGS) $(ZIG_FETCH_FLAGS)

# Run the guest against the full EF zkevm stateless fixture set on the host. Fixtures come from the
# lazy execution_spec_tests_zkevm dependency. Narrow/triage with SPEC_ARGS, e.g.
# `make spec-test SPEC_ARGS="--fork Amsterdam -x"` or `SPEC_ARGS="--report-only"`.
SPEC_ARGS ?=
spec-test:
	$(ZIG) build spec-tests $(ZIG_BUILD_FLAGS) $(ZIG_FETCH_FLAGS) $(if $(strip $(SPEC_ARGS)),-- $(SPEC_ARGS))

clean:
	rm -rf $(MAKEFILE_DIR)zig-out/
	rm -rf $(BUILD_CACHE_DIR)

# Note: gp-compile uses custom logic to this Makefile.
# gp-linker-scritpt, gp-exec, gp-debug delegate logic
# to arithmetization/src/test/examples/Makefile.

gp-compile: gp-linker-script
	$(ZIG) build $(ZIG_BUILD_FLAGS) $(ZIG_FETCH_FLAGS) -Din-origin=$(IN_ORIGIN)
	@if command -v riscv64-unknown-elf-objdump >/dev/null 2>&1; then \
		echo "\nDisassembling $(ELF_BIN)"; \
		LANG=C riscv64-unknown-elf-objdump -D -s --line-numbers $(ELF_BIN) > $(ELF_BIN).objdump; \
	else \
		echo "\nSkipping objdump: riscv64-unknown-elf-objdump is not on PATH"; \
	fi

gp-linker-script:
	$(MAKE) -f $(EXAMPLES_DIR) linker-script LINKER_SCRIPT=$(MAKEFILE_DIR)../build_common/linker_script.ld IN_ORIGIN=$(IN_ORIGIN)

gp-exec: gp-compile
	$(MAKE) -f $(EXAMPLES_DIR) elf-exec BIN_EXT="$(ELF_BIN)" IN_BYTES="@$(SAMPLE_INPUT)" IN_ORIGIN=$(IN_ORIGIN) ZKC_EXEC_FLAGS=-q

gp-debug: gp-compile
	$(MAKE) -f $(EXAMPLES_DIR) elf-debug BIN_EXT="$(ELF_BIN)" IN_BYTES="@$(SAMPLE_INPUT)" IN_ORIGIN=$(IN_ORIGIN)

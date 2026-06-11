# l2-execution guest — self-contained lifecycle.
#
# This folder is its own Zig package (build.zig + build.zig.zon). The riscv-guests/ top-level
# Makefile fans `compile`/`test`/`spec-test`/`fetch`/`clean` out to this file for every guest;
# guest-specific targets (the ZKC exec/debug below) live here, not in the orchestrator.
#
# Scope split: building + host-testing is proving-system-AGNOSTIC and lives here (`compile` links the
# statically-linked rv64im ELF — the zkvm-standards artifact — via build_common). Only the ZKC run
# itself (ELF->JSON via elf_to_json_gen, then zkc) is owned by
# arithmetization/src/test/examples/Makefile; exec/debug build that ELF and hand it over.

.PHONY: fetch compile test spec-test exec debug fixture-exec fixture-debug clean

MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

ZIG ?= zig

BUILD_CACHE_DIR ?= $(MAKEFILE_DIR).zig-cache
ZIG_BUILD_FLAGS ?= --cache-dir $(BUILD_CACHE_DIR) --summary all
# NOTE: `zig build --fetch` fetches the dependency tree and EXITS without building or running, so it
# must NOT be passed to build/test targets. Pre-fetching is the dedicated `fetch` target.
ZIG_FETCH_FLAGS ?=

# Proving-system layer that owns the ZKC memory layout, ELF->JSON and zkc invocation. `exec`/`debug`
# delegate here so those recipes are never duplicated.
EXAMPLES_DIR := $(abspath $(MAKEFILE_DIR)../../arithmetization/src/test/examples)

# Where the ZKC run places the input — forwarded to the examples Makefile as IN_ORIGIN, and it must
# match the ELF's `_in_start` (build_common's linker script: 0x08800000). Also passed
# as -Dinput-offset, which the guest no longer consumes (it reads `_in_start` via read_input).
INPUT_OFFSET ?= 0x08800000

# Optional input bytes for a ZKC run (0x-prefixed; the converter byte-reverses it — see elf_to_json_gen).
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

compile:
	$(ZIG) build $(ZIG_BUILD_FLAGS) $(ZIG_FETCH_FLAGS) -Dinput-offset=$(INPUT_OFFSET)
	@if command -v riscv64-unknown-elf-objdump >/dev/null 2>&1; then \
		echo "\nDisassembling $(ELF_BIN)"; \
		LANG=C riscv64-unknown-elf-objdump -D -s --line-numbers $(ELF_BIN) > $(ELF_BIN).objdump; \
	else \
		echo "\nSkipping objdump: riscv64-unknown-elf-objdump is not on PATH"; \
	fi

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

# ── Lineth/ZKC runs: delegated to arithmetization/src/test/examples ──────────
# Build the ELF, then hand it to the ZKC layer (single source of truth for ELF->JSON + zkc). Requires
# `zkc` and `go` on PATH. elf_to_json_gen needs the PT_LOAD segments + entry point a linked ELF has.
exec: compile
	$(MAKE) -C $(EXAMPLES_DIR) elf-exec BIN_EXT="$(ELF_BIN)" IN_BYTES="$(IN_BYTES)" IN_ORIGIN=$(INPUT_OFFSET)

debug: compile
	$(MAKE) -C $(EXAMPLES_DIR) elf-debug BIN_EXT="$(ELF_BIN)" IN_BYTES="$(IN_BYTES)" IN_ORIGIN=$(INPUT_OFFSET)

# Guest-owned sample run: feed the committed, pre-framed sample input through ZkC (no IN_BYTES
# needed). The converter byte-reverses the 0x-hex back into [u64 LE len][SSZ] at _in_start.
fixture-exec:
	$(MAKE) -f $(MAKEFILE_DIR)Makefile exec IN_BYTES=@$(SAMPLE_INPUT)

fixture-debug:
	$(MAKE) -f $(MAKEFILE_DIR)Makefile debug IN_BYTES=@$(SAMPLE_INPUT)

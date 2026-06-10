# l2-execution guest — self-contained lifecycle.
#
# This folder is its own Zig package (build.zig + build.zig.zon). The riscv-guests/ top-level
# Makefile fans `compile`/`test`/`spec-test`/`fetch`/`clean` out to this file for every guest;
# guest-specific targets (the ZKC exec/debug below) live here, not in the orchestrator.
#
# Scope split: building + host-testing is proving-system-AGNOSTIC and lives here. Running inside the
# Lineth/ZKC proving system (memory layout, ELF->JSON, zkc) is owned by
# arithmetization/src/test/examples/Makefile; exec/debug only build the guest and hand it over.

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

# Input memory offset baked into the guest at compile time (-Dinput-offset). Proving-system-
# influenced (must match where the prover loads the input), so a plain build option with a default.
# For ZKC runs it is forwarded to the examples Makefile as IN_BYTES_OFFSET.
INPUT_OFFSET ?= 0x08800000

# Optional input bytes for a ZKC run (0x-prefixed, big-endian; see elf_to_json_gen).
IN_BYTES ?= ""

GUEST_BIN    := evm_execution_guest
FIXTURE_JSON := $(MAKEFILE_DIR)fixtures/contract_creation_then_ecrecover.json

# Guest artifact: a relocatable rv64im object (zkvm_* unresolved, no _start), same shape as
# zesu.rv64im.o. NOT a linked ELF, so the ZKC exec/debug wrappers cannot run until the prover
# toolchain links it (supplies the entry point and resolves zkvm_*).
BIN := $(MAKEFILE_DIR)zig-out/lib/$(GUEST_BIN).o

# ── Proving-system-agnostic guest lifecycle ─────────────────────────────────
fetch:
	$(ZIG) build $(ZIG_BUILD_FLAGS) --fetch

compile:
	$(ZIG) build $(ZIG_BUILD_FLAGS) $(ZIG_FETCH_FLAGS) -Dinput-offset=$(INPUT_OFFSET)
	@if command -v riscv64-unknown-elf-objdump >/dev/null 2>&1; then \
		echo "\nDisassembling $(BIN)"; \
		LANG=C riscv64-unknown-elf-objdump -D -s --line-numbers $(BIN) > $(BIN).objdump; \
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
# Build the guest, then hand the artifact to the ZKC layer (single source of truth for memory
# layout + ELF->JSON + zkc). Requires `zkc` and `go` on PATH. Blocked until the guest links to a
# full ELF (see BIN note) — elf_to_json_gen needs PT_LOAD segments, which a .o does not have.
exec: compile
	$(MAKE) -C $(EXAMPLES_DIR) elf-exec BIN_EXT="$(BIN)" IN_BYTES="$(IN_BYTES)" IN_BYTES_OFFSET=$(INPUT_OFFSET)

debug: compile
	$(MAKE) -C $(EXAMPLES_DIR) elf-debug BIN_EXT="$(BIN)" IN_BYTES="$(IN_BYTES)" IN_BYTES_OFFSET=$(INPUT_OFFSET)

# Guest-owned test input: extract the SSZ payload from the committed fixture and run it.
fixture-exec:
	$(MAKE) -f $(MAKEFILE_DIR)Makefile exec IN_BYTES="$$(sed -n 's/.*"payload": "\(0x[0-9a-fA-F]*\)".*/\1/p' "$(FIXTURE_JSON)")"

fixture-debug:
	$(MAKE) -f $(MAKEFILE_DIR)Makefile debug IN_BYTES="$$(sed -n 's/.*"payload": "\(0x[0-9a-fA-F]*\)".*/\1/p' "$(FIXTURE_JSON)")"

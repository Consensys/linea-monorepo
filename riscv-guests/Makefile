# RISC-V guest programs: build + host test (proving-system-agnostic), plus thin
# wrappers that DELEGATE proving-system-specific runs to the Lineth/ZKC layer.
#
# Scope split (see also arithmetization/src/test/examples/Makefile):
#   - Building and host-testing guests is proving-system-AGNOSTIC and lives here
#     (fetch / compile / test / clean). Guests are meant to be reused across provers.
#   - Running a guest inside the Lineth proving system (memory layout, ELF->JSON, zkc)
#     is proving-system-SPECIFIC. That logic has a single source of truth in
#     arithmetization/src/test/examples/Makefile; the exec/debug/fixture-* targets
#     below only build the guest and hand the artifact to it (no duplicated recipes).

.PHONY: fetch compile test exec debug fixture-exec fixture-debug clean

MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

GUEST ?= l2-execution
ZIG   ?= zig

BUILD_CACHE_DIR ?= $(MAKEFILE_DIR).zig-cache
ZIG_BUILD_FLAGS ?= --cache-dir $(BUILD_CACHE_DIR) --summary all
# NOTE: `zig build --fetch` fetches the dependency tree and EXITS without building or running.
# It must NOT be passed to build/test targets (that would make them silent no-ops). Pre-fetching
# is done by the dedicated `fetch` target, which passes `--fetch` literally.
ZIG_FETCH_FLAGS ?=

# Proving-system layer that owns the ZKC memory layout, ELF->JSON and zkc invocation.
# `exec`/`debug` below delegate here so those recipes are never duplicated.
EXAMPLES_DIR := $(abspath $(MAKEFILE_DIR)../arithmetization/src/test/examples)

# Input memory offset baked into the guest at compile time (-Dinput-offset). It is
# proving-system-influenced (it must match where the prover loads the input), so it is a plain
# build option with a default rather than ZKC-specific machinery. For ZKC runs it is forwarded
# to the examples Makefile as IN_BYTES_OFFSET.
INPUT_OFFSET ?= 0x08800000

# Optional input bytes for a ZKC run (0x-prefixed, big-endian; see elf_to_json_gen).
IN_BYTES ?= ""

# GUEST_BIN / FIXTURE_JSON must be set before BIN (immediate `:=` expansion references them).
ifeq ($(GUEST),l2-execution)
GUEST_BIN    := evm_execution_guest
FIXTURE_JSON := $(MAKEFILE_DIR)l2-execution/fixtures/contract_creation_then_ecrecover.json
else
$(error Unsupported GUEST '$(GUEST)'. Supported guests: l2-execution)
endif

# Guest artifact: a relocatable rv64im object (zkvm_* unresolved, no _start), same shape as
# zesu.rv64im.o. NOTE: it is NOT a linked ELF, so the ZKC exec/debug wrappers cannot run until
# the prover toolchain links it (supplies the entry point and resolves zkvm_*).
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
	$(MAKE) -f $(MAKEFILE_DIR)Makefile exec GUEST=$(GUEST) IN_BYTES="$$(sed -n 's/.*"payload": "\(0x[0-9a-fA-F]*\)".*/\1/p' "$(FIXTURE_JSON)")"

fixture-debug:
	$(MAKE) -f $(MAKEFILE_DIR)Makefile debug GUEST=$(GUEST) IN_BYTES="$$(sed -n 's/.*"payload": "\(0x[0-9a-fA-F]*\)".*/\1/p' "$(FIXTURE_JSON)")"

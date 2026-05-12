# ZKC toolchain for RISC-V arithmetization.
# Usage from linea-monorepo root:
#   make -C arithmetization check
# Or from this directory:
#   make check

MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
REPO_ROOT := $(abspath $(MAKEFILE_DIR)/..)

# Used to install zkc from go-corset repository (while zkc is not released yet).
GO_CORSET ?= $(REPO_ROOT)/go-corset
GO_CORSET_REPO ?= https://github.com/consensys/go-corset.git
# Optional: pin a commit or branch (e.g. 763f30878e5fc260f0d88863c9883c904f8fc056). Leave empty for default branch.
GO_CORSET_REF_FOR_ZKC ?=

# Used to install go-corset
GO_CORSET_VERSION ?= v1.2.14

ZKC_FILES := $(shell find $(MAKEFILE_DIR)src/main/riscv/* -name "*.zkc")
ZKC_MAIN := $(MAKEFILE_DIR)src/main/riscv/main.zkc

.PHONY: checkout-go-corset install-go-corset install-zkc zkc-compile check test-instructions test-xfail test-all

# Fetch go-corset if missing (no-op when CI already checked it out into GO_CORSET).
checkout-go-corset:
	@if [ -d "$(GO_CORSET)/.git" ]; then \
		echo "go-corset already present at $(GO_CORSET)"; \
	else \
		echo "Cloning $(GO_CORSET_REPO) -> $(GO_CORSET)"; \
		git clone "$(GO_CORSET_REPO)" "$(GO_CORSET)"; \
	fi
	@if [ -n "$(GO_CORSET_REF_FOR_ZKC)" ]; then \
		git -C "$(GO_CORSET)" fetch origin "$(GO_CORSET_REF_FOR_ZKC)" || true; \
		git -C "$(GO_CORSET)" checkout "$(GO_CORSET_REF_FOR_ZKC)"; \
		git -C "$(GO_CORSET)" pull --ff-only origin "$(GO_CORSET_REF_FOR_ZKC)" 2>/dev/null || true; \
	fi

# Install the zkc CLI from the go-corset repository.
install-zkc: checkout-go-corset
	cd "$(GO_CORSET)" && go install ./cmd/zkc

# Compile the main RISC-V ZKC program.
zkc-compile: $(ZKC_FILES)
	zkc compile $(ZKC_MAIN)

zkc-lint:
	zkc format --check $(ZKC_FILES)

zkc-lint-apply:
	zkc format $(ZKC_FILES)

riscv-check-compilation: install-zkc zkc-compile

riscv-check-lint: install-zkc zkc-lint

EXAMPLES_DIR := $(MAKEFILE_DIR)src/test/examples

# Instruction-type tests. Each test_<type>type_*.s file (type ∈ {r,i,s,b,u,j})
# is a self-checking guest that traps on wrong results via an illegal opcode
# → zkc exec exits non-zero. The *type_* glob naturally excludes test_xfail_*.s.
INSTR_TESTS := $(notdir $(wildcard $(EXAMPLES_DIR)/asm/src/test_*type_*.s))
test-instructions: install-zkc
	@echo "Cleaning instruction-test artifacts under $(EXAMPLES_DIR)/asm/bin/"
	@rm -f $(EXAMPLES_DIR)/asm/bin/test_*type_*
	@if [ -z "$(INSTR_TESTS)" ]; then \
	    echo "No tests matching test_*type_*.s — nothing to run"; \
	    exit 0; \
	fi
	@fail=0; \
	for t in $(INSTR_TESTS); do \
	    echo "=== $$t ==="; \
	    if $(MAKE) -C $(EXAMPLES_DIR) TEST=$$t exec >/tmp/$$t.out 2>&1; then \
	        echo "[PASS] $$t"; \
	    else \
	        cat /tmp/$$t.out; \
	        echo "[FAIL] $$t"; \
	        fail=1; \
	    fi; \
	    rm -f /tmp/$$t.out; \
	done; \
	exit $$fail

# Known-failure tests: each test_xfail_*.s exercises a behavior that is currently
# broken in zkC and is *expected* to make `zkc exec` exit non-zero. Pass/fail is
# inverted vs the per-type targets: a zkc exec failure is [XFAIL] (expected),
# a clean exit is [UNEXPECTED PASS] (the underlying bug was fixed and the test
# should be moved into the regular suite).
XFAIL_TESTS := $(notdir $(wildcard $(EXAMPLES_DIR)/asm/src/test_xfail_*.s))
test-xfail: install-zkc
	@echo "Cleaning XFAIL artifacts under $(EXAMPLES_DIR)/asm/bin/"
	@rm -f $(EXAMPLES_DIR)/asm/bin/test_xfail_*
	@if [ -z "$(XFAIL_TESTS)" ]; then \
	    echo "No tests matching test_xfail_*.s — nothing to run"; \
	    exit 0; \
	fi
	@fail=0; \
	for t in $(XFAIL_TESTS); do \
	    echo "=== $$t ==="; \
	    if $(MAKE) -C $(EXAMPLES_DIR) TEST=$$t exec >/tmp/$$t.out 2>&1; then \
	        cat /tmp/$$t.out; \
	        echo "[UNEXPECTED PASS] $$t — expected zkc exec to FAIL; was the underlying bug fixed?"; \
	        fail=1; \
	    else \
	        echo "[XFAIL] $$t (expected)"; \
	    fi; \
	    rm -f /tmp/$$t.out; \
	done; \
	exit $$fail

test-all: test-instructions test-xfail

# Install the go-corset CLI from a released version (skips if already at the target version).
install-go-corset:
	@if command -v go-corset >/dev/null 2>&1; then \
		installed=$$(go-corset --version | sed 's/go-corset//;s/^ *//'); \
		if [ "$$installed" = "$(GO_CORSET_VERSION)" ]; then \
			echo "go-corset $(GO_CORSET_VERSION) already installed"; \
		else \
			echo "go-corset version mismatch: installed=$$installed expected=$(GO_CORSET_VERSION)"; \
			echo "Installing go-corset@$(GO_CORSET_VERSION)..."; \
			go install github.com/consensys/go-corset/cmd/go-corset@$(GO_CORSET_VERSION); \
		fi; \
	else \
		echo "go-corset not found in PATH"; \
		echo "Installing go-corset@$(GO_CORSET_VERSION)..."; \
		go install github.com/consensys/go-corset/cmd/go-corset@$(GO_CORSET_VERSION); \
	fi

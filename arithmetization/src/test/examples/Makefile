# Usage (from anywhere):
#   make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile TEST=your_test(.s|.zig)
# Or:
#   cd arithmetization/src/test/examples && make TEST=your_test(.s|.zig|.rs)
# Optional shell alias:
#   alias zkc-test='make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile'
#
# For the target we refer to https://github.com/eth-act/zkvm-standards/blob/main/standards/riscv-target/target.md.
# Note that Zicclsm extension does not affect the generated ELF so it is omitted.
# Moreover, ABI being LP64 (soft-float) is relevant only for float numbers, which we do not use, so it can be omitted as well.

# Directory of this Makefile (used for relative paths)
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# Name of the test to run (with extension), mandatory unless clean-all is the target
ifeq ($(filter clean-all,$(MAKECMDGOALS)),)
ifndef TEST
$(error TEST is mandatory. Usage: make TEST=your_test(.s|.zig|.rs))
endif
endif
NAME := $(basename $(TEST))
EXT := $(suffix $(TEST))

# Paths below are determined by the extension
SRC ?= $(MAKEFILE_DIR)$(if $(filter .s,$(EXT)),asm,$(if $(filter .zig,$(EXT)),zig,rust))/src/$(TEST)
BIN ?= $(MAKEFILE_DIR)$(if $(filter .s,$(EXT)),asm/bin/,$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,rust/bin/))$(NAME)
JSON ?= $(MAKEFILE_DIR)$(if $(filter .s,$(EXT)),asm/bin/,$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,rust/bin/))$(NAME).json
# Note: the JSON file will be removed and main.zkc will receive directly the ELF file

# Set STRIP to strip the ELF (remove debug symbols)
STRIP ?= false

# Paths to go main (to turn bin into json) and zkc main (to execute/debug the json)
GO_MAIN ?= $(MAKEFILE_DIR)main.go
ZKC_MAIN ?= $(abspath $(MAKEFILE_DIR)../../main/riscv/main.zkc)

# Default target (make)
exec: compile zkc-exec

debug: compile zkc-debug

compile: $(SRC)
	@if [ "$(EXT)" = ".s" ]; then \
		mkdir -p $(MAKEFILE_DIR)asm/bin; \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		riscv64-unknown-elf-as -march=rv64im -o $(BIN) $^; \
	elif [ "$(EXT)" = ".zig" ]; then \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		zig build --build-file $(MAKEFILE_DIR)zig/build.zig -Dname=$(NAME); \
	elif [ "$(EXT)" = ".rs" ]; then \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		mkdir -p $(MAKEFILE_DIR)rust/bin && rustc --target riscv64imac-unknown-none-elf -C opt-level=0 -C debuginfo=0 -C target-feature=-a,-c $(SRC) -o $(BIN); \
	else \
		echo "Unsupported file extension: $(EXT)"; \
		exit 1; \
	fi
	@if [ "$(STRIP)" = "true" ]; then \
		echo "\nStripping $(BIN)"; \
		riscv64-unknown-elf-strip -S $(BIN); \
	fi
	@echo "\nGenerating $(JSON) from $(BIN)"
	@go run $(GO_MAIN) $(BIN) > $(JSON)

zkc-exec:
	zkc exec $(JSON) $(ZKC_MAIN)

zkc-debug:
	zkc debug $(JSON) $(ZKC_MAIN)

clean:
	rm -f $(BIN) $(JSON) $(BIN_STRIPPED)

clean-all:
	rm -rf $(MAKEFILE_DIR)asm/bin/
	rm -rf $(MAKEFILE_DIR)zig/zig-out/
	rm -rf $(MAKEFILE_DIR)rust/bin/
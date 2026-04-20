# Usage (from anywhere):
#   make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile TEST=your_test(.s|.zig)
# Or:
#   cd arithmetization/src/test/examples && make TEST=your_test(.s|.zig)
# Optional shell alias:
#   alias zkctest='make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile'

# Directory of this Makefile (used for relative paths)
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# Name of the test to run (with extension), mandatory unless clean-all is the target
ifeq ($(filter clean-all,$(MAKECMDGOALS)),)
ifndef TEST
$(error TEST is mandatory. Usage: make TEST=your_test(.s|.zig))
endif
endif
NAME := $(basename $(TEST))
EXT := $(suffix $(TEST))

# Define the assembler to use (useful in case TEST is .s)
AS = riscv64-unknown-elf-as
AS_FLAGS = -march=rv64im

# Paths below are determined by the extension
SRC ?= $(MAKEFILE_DIR)$(if $(filter .zig,$(EXT)),zig,asm)/src/$(TEST)
BIN ?= $(MAKEFILE_DIR)$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,asm/bin/)$(NAME)
JSON ?= $(MAKEFILE_DIR)$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,asm/bin/)$(NAME).json

# Set if strip should be applied to the binary
STRIP ?= false
BIN_STRIPPED = $(MAKEFILE_DIR)$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,asm/bin/)$(NAME).stripped
JSON_STRIPPED = $(MAKEFILE_DIR)$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,asm/bin/)$(NAME).stripped.json

# Paths to go main (to turn bin into json) and zkc main (to execute/debug the json)
GO_MAIN ?= $(MAKEFILE_DIR)main.go
ZKC_MAIN ?= $(abspath $(MAKEFILE_DIR)../../main/riscv/main.zkc)

# Default target (make)
exec: compile zkc-exec

debug: compile zkc-debug

compile: $(SRC)
	@if [ "$(EXT)" = ".s" ]; then \
		mkdir -p $(MAKEFILE_DIR)asm/bin; \
		$(AS) ${AS_FLAGS} -o $(BIN) $^; \
	elif [ "$(EXT)" = ".zig" ]; then \
		zig build --build-file $(MAKEFILE_DIR)zig/build.zig -Dname=$(NAME); \
	else \
		echo "Unsupported file extension: $(EXT)"; \
		exit 1; \
	fi
	@if [ "$(STRIP)" = "true" ]; then \
		cp $(BIN) $(BIN_STRIPPED); \
		riscv64-unknown-elf-strip -S $(BIN_STRIPPED); \
		echo "\nGenerating $(JSON_STRIPPED) from $(BIN_STRIPPED)"; \
		go run $(GO_MAIN) $(BIN_STRIPPED) > $(JSON_STRIPPED); \
	fi
	@echo "\nGenerating $(JSON) from $(BIN)"
	@go run $(GO_MAIN) $(BIN) > $(JSON)

zkc-exec:
	@if [ "$(STRIP)" = "true" ]; then \
		echo "\nExecuting $(JSON_STRIPPED)"; \
		zkc exec $(JSON_STRIPPED) $(ZKC_MAIN); \
	else \
		echo "\nExecuting $(JSON)"; \
		zkc exec $(JSON) $(ZKC_MAIN); \
	fi

zkc-debug:
	@if [ "$(STRIP)" = "true" ]; then \
		echo "\nDebugging $(JSON_STRIPPED)"; \
		zkc debug $(JSON_STRIPPED) $(ZKC_MAIN); \
	else \
		echo "\nDebugging $(JSON)"; \
		zkc debug $(JSON) $(ZKC_MAIN); \
	fi

clean:
	@rm $(BIN) $(JSON) 
	@if [ "$(STRIP)" = "true" ]; then \
		 rm $(BIN_STRIPPED) $(JSON_STRIPPED); \
	fi

clean-all:
	rm -rf $(MAKEFILE_DIR)asm/bin/
	rm -rf $(MAKEFILE_DIR)zig/zig-out/
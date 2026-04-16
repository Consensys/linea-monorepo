# Usage (from anywhere):
#   make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile TEST=your_test(.s|.zig)
# Or:
#   cd arithmetization/src/test/examples && make TEST=your_test(.s|.zig)
# Optional shell alias:
#   alias zkctest='make -f /path/to/linea-monorepo/arithmetization/src/test/examples/Makefile'

# Directory of this Makefile (used for relative paths)
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# Name of the test to run (with extension)
ifeq ($(filter clean-all,$(MAKECMDGOALS)),)
ifndef TEST
$(error TEST is mandatory. Usage: make TEST=your_test(.s|.zig))
endif
endif
NAME := $(basename $(TEST))
EXT := $(suffix $(TEST))

# define the assembler to use (useful in case TEST is .s)
AS = riscv64-unknown-elf-as
AS_FLAGS = -march=rv64im

# Paths below are determined by the extension
SRC ?= $(MAKEFILE_DIR)$(if $(filter .zig,$(EXT)),zig,asm)/src/$(TEST)
BIN ?= $(MAKEFILE_DIR)$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,asm/bin/)$(NAME)
JSON ?= $(MAKEFILE_DIR)$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,asm/bin/)$(NAME).json

# Paths to go main (to turn bin into json) and zkc main (to execute/debug the json)
GO_MAIN ?= $(MAKEFILE_DIR)main.go
ZKC_MAIN ?= $(abspath $(MAKEFILE_DIR)../../main/riscv/main.zkc)

# default target (make)
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
	go run $(GO_MAIN) $(BIN) > $(JSON)

zkc-exec:
	zkc exec $(JSON) $(ZKC_MAIN)

zkc-debug:
	zkc debug $(JSON) $(ZKC_MAIN)

clean:
	rm $(BIN) $(JSON)

clean-all:
	rm -rf $(MAKEFILE_DIR)asm/bin/
	rm -rf $(MAKEFILE_DIR)zig/zig-out/
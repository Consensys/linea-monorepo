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
EXT  := $(suffix $(TEST))

# Paths below are determined by the extension
SRC  ?= $(MAKEFILE_DIR)$(if $(filter .s,$(EXT)),asm,$(if $(filter .zig,$(EXT)),zig,rust))/src/$(TEST)
BIN  ?= $(MAKEFILE_DIR)$(if $(filter .s,$(EXT)),asm/bin/,$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,rust/target/riscv64im-unknown-none-elf/release/))$(NAME)
JSON ?= $(MAKEFILE_DIR)$(if $(filter .s,$(EXT)),asm/bin/,$(if $(filter .zig,$(EXT)),zig/zig-out/bin/,rust/target/riscv64im-unknown-none-elf/release/))$(NAME).json
# Note: in the future the JSON file will be removed and main.zkc will receive directly the ELF file

# Set STRIP to strip the ELF (remove debug symbols)
STRIP ?= false

# Set ZIG_STRIP to strip the ELF when compiling Zig (removing debug symbols and reducing size)
# Ignored for .s and .rs files
ZIG_STRIP ?= true

# Disassembler: prefer llvm-objdump (shows bytes in memory order, matching the JSON format).
# Falls back to the Homebrew keg-only path on macOS, then to riscv64-unknown-elf-objdump.
# Override with: make OBJDUMP=/path/to/objdump
OBJDUMP ?= $(or $(shell command -v llvm-objdump 2>/dev/null),$(shell command -v rust-objdump 2>/dev/null),$(shell test -x /opt/homebrew/opt/llvm/bin/llvm-objdump && echo /opt/homebrew/opt/llvm/bin/llvm-objdump),$(shell command -v riscv64-unknown-elf-objdump 2>/dev/null))

# Paths to go main (to turn bin into json) and zkc main (to execute/debug the json)
GO_MAIN  := $(MAKEFILE_DIR)main.go
ZKC_MAIN := $(MAKEFILE_DIR)../../main/riscv/main.zkc

# Inputs and offsets for the program
IN_BYTES        ?= ""
PROGRAM_OFFSET ?= 0         # (up to 128 MiB)
IN_BYTES_OFFSET ?= 0x8000000 # (up to 1 GiB)
ENTRY_POINT    ?= 0

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
		zig build --build-file $(MAKEFILE_DIR)zig/build.zig -Dname=$(NAME) -Dstrip=$(ZIG_STRIP); \
	elif [ "$(EXT)" = ".rs" ]; then \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		cargo +nightly rustc \
			--manifest-path $(MAKEFILE_DIR)rust/Cargo.toml \
			--target $(MAKEFILE_DIR)rust/riscv64im-unknown-none-elf.json \
			-Z build-std=core \
			-Z json-target-spec \
			--bin $(NAME) \
			--release; \
	else \
		echo "Unsupported file extension: $(EXT)"; \
		exit 1; \
	fi
	@if [ "$(STRIP)" = "true" ]; then \
		echo "\nStripping $(BIN)"; \
		riscv64-unknown-elf-strip -S $(BIN); \
	fi
	@echo "\nDisassembling $(BIN)"
  # the following regular expression maps
  #
  # <hex-pc>: <byte_1><byte_2><byte_3><byte_4><non hex><etc>
  #
  # to
  #
  # <hex-pc>: <byte_1><byte_2><byte_3><byte_4> (<byte_4><byte_3><byte_2><byte_1>)<non hex><etc>
  #
  # Claude also included the same type of operation for COMPRESSED instructions ... the 2nd -e expression
	@$(OBJDUMP) -d --line-numbers -S $(BIN) \
		| sed -E \
		-e 's/^([[:space:]]*[0-9a-f]+:[[:space:]]*)([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})([^0-9a-f])/\1\2\3\4\5 (\5\4\3\2)\6/' \
		-e 's/^([[:space:]]*[0-9a-f]+:[[:space:]]*)([0-9a-f]{2})([0-9a-f]{2})([^0-9a-f])/\1\2\3 (\3\2)\4/' \
		> $(BIN)_disassembled.elf
	@echo "\nGenerating $(JSON) from $(BIN)"
	@go run $(GO_MAIN) $(BIN) "$(IN_BYTES)" $(PROGRAM_OFFSET) $(IN_BYTES_OFFSET) $(ENTRY_POINT) > $(JSON)

# Note: compiled Rust libraries may contain instructions from excluded extensions. For instance, the core library is compiled
# for the IMAC extension, and contains COMPRESSED instructions. The output ELF may thus contain such instructions, too, even
# if these extensions were explicitly exluded from the target.
#
# Note: we use opt-level=2 to remove unused code.

zkc-exec:
	zkc exec $(JSON) $(ZKC_MAIN)

zkc-debug:
	zkc debug $(JSON) $(ZKC_MAIN)

clean:
	rm -f $(BIN) $(JSON) $(BIN_STRIPPED) $(BIN)_disassembled.elf

clean-all:
	rm -rf $(MAKEFILE_DIR)asm/bin/
	rm -rf $(MAKEFILE_DIR)zig/zig-out/
	rm -rf $(MAKEFILE_DIR)zig/.zig-cache/
	rm -rf $(MAKEFILE_DIR)rust/target/

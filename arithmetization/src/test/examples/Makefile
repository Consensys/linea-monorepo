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

# Paths to go main (to turn bin into json), zkc main (to execute/debug the json) and linker script
GO_MAIN  := $(MAKEFILE_DIR)main.go
ZKC_MAIN := $(MAKEFILE_DIR)../../main/riscv/main.zkc
LINKER_SCRIPT := $(MAKEFILE_DIR)linker_script.ld

# Inputs and offsets for the program
IN_BYTES        ?= ""

# Values to generate the linker script
PROGRAM_OFFSET  ?= 0x00000000        # 128 MiB (0x00000000 -> 0x07FFFFFF)
IN_BYTES_OFFSET ?= 0x08800000        # 1 GiB   (0x08800000 -> 0x48FFFFFF)
ENTRY_POINT     := $(PROGRAM_OFFSET) # Since program starts from _start, the entry point is the same as the program offset
SP              ?= 0x087fffff        # 8 MiB   (0x08000000 <- 0x087fffff)

# Parsed values are just to verify that the generate ELF is correct
PROGRAM_OFFSET_FROM_ELF  = $(shell LANG=C riscv64-unknown-elf-readelf -S $(BIN) | grep '\.text' | awk '{print "0x"$$5}')
IN_BYTES_OFFSET_FROM_ELF = $(shell LANG=C riscv64-unknown-elf-nm $(BIN) | grep _input_start | awk '{print "0x"$$1}') 
ENTRY_POINT_FROM_ELF     = $(shell LANG=C riscv64-unknown-elf-readelf -h $(BIN) | grep 'Entry point' | awk '{print $$4}')
SP_FROM_ELF              = $(shell LANG=C riscv64-unknown-elf-nm $(BIN) | grep _stack_start | awk '{print "0x"$$1}')
VERIFY_ELF               ?= false             

# Default target (make)
exec: compile zkc-exec

debug: compile zkc-debug

compile: $(SRC) $(LINKER_SCRIPT)
# Generate ELF
	@if [ "$(EXT)" = ".s" ]; then \
		mkdir -p $(MAKEFILE_DIR)asm/bin; \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		riscv64-unknown-elf-gcc -march=rv64im -mabi=lp64 -nostdlib -T$(LINKER_SCRIPT) -o $(BIN) $(SRC); \
	elif [ "$(EXT)" = ".zig" ]; then \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		zig build --build-file $(MAKEFILE_DIR)zig/build.zig -Dname=$(NAME); \
	elif [ "$(EXT)" = ".rs" ]; then \
		echo "\nGenerating $(BIN) from $(SRC)"; \
		cargo +nightly rustc \
			--manifest-path $(MAKEFILE_DIR)rust/Cargo.toml \
			--target $(MAKEFILE_DIR)rust/riscv64im-unknown-none-elf.json \
			-Z build-std=core \
			-Z json-target-spec \
			--bin $(NAME) \
			--release \
			-- -C link-arg=-T$(LINKER_SCRIPT); \
	else \
		echo "Unsupported file extension: $(EXT)"; \
		exit 1; \
	fi
# Disassemble ELF
	@echo "\nDisassembling $(BIN)"
	@LANG=C riscv64-unknown-elf-objdump -D -s --line-numbers $(BIN) > $(BIN)_disassembled.txt
# Verify that PROGRAM_OFFSET, IN_BYTES_OFFSET, ENTRY_POINT, SP match the ELF values
	@if [ "$(VERIFY_ELF)" = "true" ]; then $(MAKE) verify-elf; fi
# Generate JSON from ELF
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
	rm -f $(BIN) $(JSON) $(BIN)_disassembled.txt

clean-all:
	rm -rf $(MAKEFILE_DIR)asm/bin/
	rm -rf $(MAKEFILE_DIR)zig/zig-out/
	rm -rf $(MAKEFILE_DIR)zig/.zig-cache/
	rm -rf $(MAKEFILE_DIR)rust/target/
	rm -f $(LINKER_SCRIPT)

$(LINKER_SCRIPT):
	@echo "ENTRY(_start)" > $@
	@echo "" >> $@
	@echo "MEMORY {" >> $@
	@echo "    PROGRAM (rx) : ORIGIN = $(PROGRAM_OFFSET), LENGTH = 0x8000000" >> $@
	@echo "    STACK   (rw) : ORIGIN = $(shell printf '0x%08x' $$(($(SP) - 0x800000 + 1))), LENGTH = 0x800000" >> $@
	@echo "    IN  (r)  : ORIGIN = $(IN_BYTES_OFFSET), LENGTH = 0x40000000" >> $@
	@echo "}" >> $@
	@echo "" >> $@
	@echo "SECTIONS {" >> $@
	@echo "    . = ORIGIN(PROGRAM);" >> $@
	@echo "" >> $@
	@echo "    .text : {" >> $@
	@echo "        *(.text .text.*)" >> $@
	@echo "    } > PROGRAM" >> $@
	@echo "" >> $@
	@echo "    .rodata : {" >> $@
	@echo "        *(.rodata .rodata.*)" >> $@
	@echo "    } > PROGRAM" >> $@
	@echo "" >> $@
	@echo "    .data : {" >> $@
	@echo "        *(.data .data.*)" >> $@
	@echo "    } > PROGRAM" >> $@
	@echo "" >> $@
	@echo "    .bss : {" >> $@
	@echo "        *(.bss .bss.*)" >> $@
	@echo "    } > PROGRAM" >> $@
	@echo "" >> $@
	@echo "    /* stack grows downward from the top of the stack region */" >> $@
	@echo "    _stack_start = ORIGIN(STACK) + LENGTH(STACK);" >> $@
	@echo "" >> $@
	@echo "    /* symbol marking the start of the input region */" >> $@
	@echo "    _input_start = ORIGIN(IN);" >> $@
	@echo "}" >> $@

verify-elf: $(BIN)
	@echo "\nVerify that PROGRAM_OFFSET, IN_BYTES_OFFSET, ENTRY_POINT, SP match the ELF values"
	@[ "$(shell printf '0x%016x' $$(($(PROGRAM_OFFSET_FROM_ELF))))" = "$(shell printf '0x%016x' $$(($(PROGRAM_OFFSET))))" ] || \
		(echo "ERROR: PROGRAM_OFFSET mismatch: expected $(PROGRAM_OFFSET), got $(PROGRAM_OFFSET_FROM_ELF)" && exit 1)
	@echo "  PROGRAM_OFFSET:  $(PROGRAM_OFFSET) (expected) = $(PROGRAM_OFFSET_FROM_ELF) (from ELF)"
	@[ "$(shell printf '0x%016x' $$(($(IN_BYTES_OFFSET_FROM_ELF))))" = "$(shell printf '0x%016x' $$(($(IN_BYTES_OFFSET))))" ] || \
		(echo "ERROR: IN_BYTES_OFFSET mismatch: expected $(IN_BYTES_OFFSET), got $(IN_BYTES_OFFSET_FROM_ELF)" && exit 1)
	@echo "  IN_BYTES_OFFSET: $(IN_BYTES_OFFSET) (expected) = $(IN_BYTES_OFFSET_FROM_ELF) (from ELF)"
	@[ "$(shell printf '0x%016x' $$(($(ENTRY_POINT_FROM_ELF))))" = "$(shell printf '0x%016x' $$(($(ENTRY_POINT))))" ] || \
		(echo "ERROR: ENTRY_POINT mismatch: expected $(ENTRY_POINT), got $(ENTRY_POINT_FROM_ELF)" && exit 1)
	@echo "  ENTRY_POINT:     $(ENTRY_POINT) (expected) = $(ENTRY_POINT_FROM_ELF) (from ELF)"
	@[ "$(shell printf '0x%016x' $$(($(SP_FROM_ELF))))" = "$(shell printf '0x%016x' $$(($(SP))))" ] || \
		(echo "ERROR: SP mismatch: expected $(SP), got $(SP_FROM_ELF)" && exit 1)
	@echo "  SP:              $(SP) (expected) = $(SP_FROM_ELF) (from ELF)"
	@echo "All values match\n"

# define the assembler to use
AS = riscv64-unknown-elf-as

# NOTE: add to .zshrc (or .bashrc) an alias with the path to the makefile, for example:
# alias zkctest="make -C /Users/lorenzo/Desktop/linea-monorepo/arithmetization/src/test/examples/asm

# replace with the path to your local copy of the linea monorepo
REPO_DIR = /Users/lorenzo/Desktop/linea-monorepo

ifndef TEST
$(TEST variable is mandatory. Usage: make TEST=your_test)
endif
ASM ?= $(REPO_DIR)/arithmetization/src/test/examples/asm/src/$(TEST).s
BIN ?= $(REPO_DIR)/arithmetization/src/test/examples/asm/bin/$(TEST)
JSON ?= $(REPO_DIR)/arithmetization/src/test/examples/asm/bin/$(TEST).json
GO_MAIN ?= $(REPO_DIR)/arithmetization/src/test/examples/main.go
ZKC_MAIN ?= $(REPO_DIR)/arithmetization/src/main/riscv/main.zkc

# default target (make)
exec: compile zkc-exec

debug: compile zkc-debug

compile: $(ASM)
	@mkdir -p $(REPO_DIR)/arithmetization/src/test/examples/asm/bin
	$(AS) -o $(BIN) $^
	go run $(GO_MAIN) $(BIN) > $(JSON)

zkc-exec:
	zkc exec $(JSON) $(ZKC_MAIN)

zkc-debug:
	zkc debug $(JSON) $(ZKC_MAIN)

clean:
	rm $(BIN) $(JSON)

clean-all:
	rm -rf $(REPO_DIR)/arithmetization/src/test/examples/asm/bin/
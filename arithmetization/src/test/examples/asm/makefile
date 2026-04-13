# define the assembler to use
AS = riscv64-unknown-elf-as

# Add to your (.zshrc or .bashrc) enviroment:
# 	export LINEA_MONOREPO_DIR=/path/to/linea-monorepo
# 	alias zkctest="make -C \"$LINEA_MONOREPO_DIR/arithmetization/src/test/examples/asm\""
# Usage: 
# 	zkctest TEST=your_test

ifndef LINEA_MONOREPO_DIR
$(error LINEA_MONOREPO_DIR enviroment variable is mandatory. Usage: export LINEA_MONOREPO_DIR=/path/to/linea-monorepo)
endif
ifndef TEST
$(TEST variable is mandatory. Usage: make TEST=your_test)
endif
ASM ?= $(LINEA_MONOREPO_DIR)/arithmetization/src/test/examples/asm/src/$(TEST).s
BIN ?= $(LINEA_MONOREPO_DIR)/arithmetization/src/test/examples/asm/bin/$(TEST)
JSON ?= $(LINEA_MONOREPO_DIR)/arithmetization/src/test/examples/asm/bin/$(TEST).json
GO_MAIN ?= $(LINEA_MONOREPO_DIR)/arithmetization/src/test/examples/main.go
ZKC_MAIN ?= $(LINEA_MONOREPO_DIR)/arithmetization/src/main/riscv/main.zkc

# default target (make)
exec: compile zkc-exec

debug: compile zkc-debug

compile: $(ASM)
	@mkdir -p $(LINEA_MONOREPO_DIR)/arithmetization/src/test/examples/asm/bin
	$(AS) -o $(BIN) $^
	go run $(GO_MAIN) $(BIN) > $(JSON)

zkc-exec:
	zkc exec $(JSON) $(ZKC_MAIN)

zkc-debug:
	zkc debug $(JSON) $(ZKC_MAIN)

clean:
	rm $(BIN) $(JSON)

clean-all:
	rm -rf $(LINEA_MONOREPO_DIR)/arithmetization/src/test/examples/asm/bin/
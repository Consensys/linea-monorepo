# Simple makefile

# define the assembler to use
AS = riscv64-unknown-elf-as

# simple build rule
bin/test: src/test.s
	@mkdir -p bin
	$(AS) -o $@ $^

clean:
	rm -rf bin

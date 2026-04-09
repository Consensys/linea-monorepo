# define the assembler to use
AS = riscv64-unknown-elf-as
# define the name of the test file (by default it is "test")
INPUT ?= test

# default target (make)
exec: compile zkc-exec

debug: compile zkc-debug

compile: src/$(INPUT).s
	@mkdir -p bin
	$(AS) -o bin/$(INPUT) $^
	go run ../main.go bin/$(INPUT) > bin/$(INPUT).json

zkc-exec:
	zkc exec bin/$(INPUT).json ../../../main/riscv/main.zkc

zkc-debug:
	zkc debug bin/$(INPUT).json ../../../main/riscv/main.zkc

clean:
	rm bin/$(INPUT) bin/$(INPUT).json

clean-all:
	rm -rf bin
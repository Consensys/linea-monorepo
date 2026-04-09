# define the assembler to use
AS = riscv64-unknown-elf-as
INPUT ?= test # default input file (test.s)

# default target (make)
# target : dependency
bin/$(INPUT) : src/$(INPUT).s
	@mkdir -p bin
	$(AS) -o $@ $^

json:
	go run ../main.go bin/$(INPUT) > bin/$(INPUT).json

exec:
	zkc exec bin/$(INPUT).json ../../../main/riscv/main.zkc

debug:
	zkc debug bin/$(INPUT).json ../../../main/riscv/main.zkc

all: bin/$(INPUT) json exec

all-debug: bin/test json debug

clean:
	rm -rf bin/$(INPUT) bin/$(INPUT).json

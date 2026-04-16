zig-out/bin/arithmetic_test: src/arithmetic_test.zig
	zig build
	go run ../main.go zig-out/bin/arithmetic_test > zig-out/bin/arithmetic_test.json

exec:
	zkc exec zig-out/bin/arithmetic_test.json ../../../main/riscv/main.zkc

debug:
	zkc debug zig-out/bin/arithmetic_test.json ../../../main/riscv/main.zkc

clean:
	rm -rf zig-out
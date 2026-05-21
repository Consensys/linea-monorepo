SHELL := /bin/bash

.PHONY: ci help fmt fmt-zig fmt-codegen fmt-testdata-generate build build-zig build-release build-r5 run test test-zig test-generated-zig test-codegen generate-testdata verify-testdata zkc-build zkc-verify clean

R5_BIN := zig-out/bin/verifier-ray
R5_JSON := $(R5_BIN).json
ELF_TO_JSON := ../arithmetization/src/test/examples/main.go
ZKC_MAIN := ../arithmetization/src/main/riscv/main.zkc

IN_BYTES ?=
PROGRAM_OFFSET ?= 0x00000000
IN_BYTES_OFFSET ?= 0x08800000
ENTRY_POINT ?= $(PROGRAM_OFFSET)

ci: fmt build test zkc-verify

help:
	@printf '%s\n' \
		'Usage: make <target>' \
		'' \
		'Common targets:' \
		'  help                  Show this help message' \
		'  ci                    Run fmt, native build, tests, and zkc verifier smoke test' \
		'  fmt                   Run Zig and Go formatting checks' \
		'  build                 Build verifier-ray natively' \
		'  build-release         Build verifier-ray natively with ReleaseSmall and stripped symbols' \
		'  build-r5              Build verifier-ray for the Linea R5 zkVM target' \
		'  run                   Run verifier-ray natively' \
		'  test                  Run Zig tests and Go codegen tests' \
		'  test-zig              Generate/verify testdata and run Zig tests' \
		'  test-codegen          Run Go codegen tests' \
		'  zkc-build             Build the R5 ELF and generate the zkc JSON input' \
		'  zkc-verify            Run verifier-ray through the zkc RISC-V interpreter' \
		'  clean                 Remove Zig build artifacts' \
		'' \
		'zkc variables:' \
		'  IN_BYTES              Input bytes copied into the interpreter input region' \
		'  PROGRAM_OFFSET        Program load address, default 0x00000000' \
		'  IN_BYTES_OFFSET       Input bytes address, default 0x08800000' \
		'  ENTRY_POINT           Entry point, default PROGRAM_OFFSET'

fmt: fmt-zig fmt-codegen fmt-testdata-generate

fmt-zig:
	zig fmt --check build.zig src test testdata/generated

fmt-codegen:
	@if [[ -n "$$(gofmt -l codegen)" ]]; then \
		echo "please run gofmt"; \
		gofmt -l codegen; \
		exit 1; \
	fi

fmt-testdata-generate:
	@if [[ -n "$$(gofmt -l testdata/generate)" ]]; then \
		echo "please run gofmt"; \
		gofmt -l testdata/generate; \
		exit 1; \
	fi

build: build-zig

build-zig:
	zig build

build-release:
	zig build -Doptimize=ReleaseSmall -Dstrip=true

build-r5:
	zig build -Dr5=true --release=small -Dstrip=true

run:
	zig build run

test: test-zig test-codegen

test-zig: verify-testdata
	zig build test

test-generated-zig: test-zig

test-codegen:
	cd codegen && go test ./...

generate-testdata:
	cd testdata/generate && go run .

verify-testdata: generate-testdata
	git diff --exit-code -- testdata/generated/vectors.zig

zkc-build: build-r5
	@echo "Generating $(R5_JSON)"
	go run $(ELF_TO_JSON) $(R5_BIN) "$(IN_BYTES)" $(PROGRAM_OFFSET) $(IN_BYTES_OFFSET) $(ENTRY_POINT) > $(R5_JSON)

zkc-verify: zkc-build
	zkc exec $(R5_JSON) $(ZKC_MAIN)

clean:
	rm -rf .zig-cache zig-out

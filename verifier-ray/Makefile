SHELL := /bin/bash

.PHONY: ci fmt fmt-zig fmt-codegen build build-zig test test-zig test-generated-zig test-codegen zkc-verify clean

ci: fmt build test zkc-verify

fmt: fmt-zig fmt-codegen

fmt-zig:
	zig fmt --check build.zig src test

fmt-codegen:
	@if [[ -n "$$(gofmt -l codegen)" ]]; then \
		echo "please run gofmt"; \
		gofmt -l codegen; \
		exit 1; \
	fi

build: build-zig

build-zig:
	zig build

test: test-zig test-codegen

test-zig:
	zig build test

test-generated-zig: test-zig

test-codegen:
	cd codegen && go test ./...

zkc-verify:
	@echo "zkc verifier smoke test is not wired yet; skipping."

clean:
	rm -rf .zig-cache zig-out

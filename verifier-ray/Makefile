SHELL := /bin/bash

.PHONY: ci fmt fmt-zig fmt-codegen fmt-testdata-generate build build-zig build-release test test-zig test-generated-zig test-codegen generate-testdata verify-testdata generate-stub verify-stub zkc-verify clean

ci: fmt build test zkc-verify

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

test: test-zig test-codegen

test-zig: verify-testdata verify-stub
	zig build test

test-generated-zig: test-zig

test-codegen:
	cd codegen && go test ./...

generate-testdata:
	cd testdata/generate && go run .

verify-testdata: generate-testdata
	git diff --exit-code -- testdata/generated/vectors.zig

generate-stub:
	cd codegen && go run ./cmd/ray-zig-codegen/ > ../src/generated/stub.zig.tmp && mv ../src/generated/stub.zig.tmp ../src/generated/stub.zig

verify-stub: generate-stub
	git diff --exit-code -- src/generated/stub.zig

zkc-verify:
	@echo "zkc verifier smoke test is not wired yet; skipping."

clean:
	rm -rf .zig-cache zig-out

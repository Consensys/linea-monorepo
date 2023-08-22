# required for `source`
SHELL := /usr/bin/env bash
CORSET_ROOT := $(shell pwd)/../corset

CGO_CFLAGS := "-I${CORSET_ROOT}/target/"
CGO_LDFLAGS := "${CORSET_ROOT}/target/release/libcorset.a -lm -ldl"
CORSET_FLAGS := CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS}
VERSION := $(shell git rev-parse --short HEAD)

# Ensure that we are using corset from the monorepo
CORSET := PATH="${CORSET_ROOT}/target/release":$$PATH

# Provers environment variables used for test-case generation
PROVER_ROOT := $(pwd)
PROVER_PKEY_FILE := ${PROVER_ROOT}/setup-data/proving-key
PROVER_VKEY_FILE := ${PROVER_ROOT}/setup-data/verifying-key
PROVER_R1CS_FILE := ${PROVER_ROOT}/setup-data/r1cs
PROVER_SOL_VERIFIER := ${PROVER_ROOT}/setup-data/verifierContract.sol
LAYER2_CHAIN_ID := 58190
LAYER2_MESSAGE_SERVICE_CONTRACT := 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
PROVER_CONFLATED_TRACES_DIR := /some/path
VERSION := v0.0.0

# Flags to preprends to the prover to run in testmode
PROVER_FLAGS := PROVER_PKEY_FILE=${PROVER_PKEY_FILE} PROVER_VKEY_FILE=${PROVER_VKEY_FILE} PROVER_R1CS_FILE=${PROVER_R1CS_FILE} PROVER_SOL_VERIFIER=${PROVER_SOL_VERIFIER} PROVER_SKIP_TRACES=true LAYER2_CHAIN_ID=${LAYER2_CHAIN_ID} LAYER2_MESSAGE_SERVICE_CONTRACT=${LAYER2_MESSAGE_SERVICE_CONTRACT} PROVER_VERSION=${VERSION} PROVER_CONFLATED_TRACES_DIR=${PROVER_CONFLATED_TRACES_DIR}


# these targets are not defined as depending on source files,
# so they should be recompiled by default “just in case”.
.PHONY: bin/prover bin/checker zkevm/zkevm.bin corset docker bin/testcase-gen bin/contract-gen testdata

##
## This rule produces both the corset binary and libcorset.a & corset.h for the FFI
##
corset:
	cd ${CORSET_ROOT} && cargo build --release

##
## Build and bundle the Corset trace-expander dependency
##
zkevm/zkevm.bin: corset
	cd ../constraints && $(CORSET) make zkevm.bin && mv zkevm.bin ../prover/zkevm/

##
## Generate `define.go` based on the latest version of corset and the constraints
##
wizard: zkevm/define/define.go
zkevm/define/define.go: corset
	cd ../constraints && $(CORSET) make -B zkevm.go && mv zkevm.go ../prover/zkevm/define/define.go

##
## Generate the setup and the smart-contract (to be run with S3 access)
##
setup: bin/contract-gen
	source cmd/prover/conf.env && bin/contract-gen generate --full --trusted-srs -p ./docker/setup/full
	source cmd/prover/conf.env && bin/contract-gen-large generate --full --trusted-srs -p ./docker/setup/full-large

##
## Build the prover and the generator, generate a setup and run the prover over a sample input
##
sample-proof: bin/prover bin/generate setup
	source XXXXXXXX/env-light.sh && LAYER2_CHAIN_ID=1 bin/prover -in ./backend/testing/a/0-2-getZkProof.json -out ./backend/testing/0-2-zkProof.json

###
### Checker
###
bin/checker: zkevm/zkevm.bin
	mkdir -p bin/
	rm -f $@
	$(CORSET_FLAGS) go build -o $@ ./cmd/checker

##
##	Smart-contract and Setup generator
##
bin/contract-gen: zkevm/zkevm.bin
	mkdir -p bin
	rm -f $@
	$(CORSET_FLAGS) go build -o $@ ./cmd/contract-gen
	$(CORSET_FLAGS) go build -tags largesize -o $@-large ./cmd/contract-gen

###
### The prover
###
bin/prover: zkevm/zkevm.bin
	mkdir -p bin/
	rm -f $@
	$(CORSET_FLAGS) go build -o ./$@ -tags debug ./cmd/prover

###
### The prover with larger size
###
bin/prover-large: zkevm/zkevm.bin
	mkdir -p bin/
	rm -f $@
	$(CORSET_FLAGS) go build -tags largesize -o ./$@ ./cmd/prover

##
##	Generate a random test sample for the smart-contract
##
bin/testcase-gen:
	mkdir -p bin
	rm -f $@
	$(CORSET_FLAGS) go build -o ./$@ ./cmd/testcase-gen
	source cmd/testcase-gen/gen-batch1.sh
	source cmd/testcase-gen/gen-batch2.sh

##
##	Test the checker of the smart-contract
##
test-checker: bin/checker
	$^ --trace ./backend/testing/a/raw-trace.gz

##
##	Build the prover docker image
##
docker:
	echo "Building docker image for the prover version ${VERSION}"
	cd .. && docker buildx build -f prover/Dockerfile --build-arg VERSION=${VERSION} --build-context prover=prover/ --build-context corset=corset/ --build-context constraints=constraints/ -t consensys/linea-prover .

##
## Run all the unit-tests
##
test: zkevm/zkevm.bin
	$(CORSET_FLAGS) go test ./...

test-corset: test-corset-a test-corset-b
test-corset-a: corset zkevm/zkevm.bin zkevm/define/define.go
	$(CORSET_FLAGS) go test -timeout 2m -run TestSampleA ./backend/testing -v

test-corset-b: corset zkevm/zkevm.bin zkevm/define/define.go
	$(CORSET_FLAGS) go test -timeout 2m -run TestSampleB github.com/consensys/accelerated-crypto-monorepo/backend/testing -v

##
## Run the CI linting
##
ci-lint: zkevm/zkevm.bin
	$(CORSET_FLAGS) golangci-lint run

##
## Echo, the CGO flags. Usefull for testing manually
##
cgo-flags:
	@echo $(CORSET_FLAGS)

#
# Update the prover's testdata in ../testdata/prover
#
make update-testdata-for-coordinator:
	go test -timeout 30m -tags amd64,nocorset -run TestSamples github.com/consensys/accelerated-crypto-monorepo/backend/testing -v
	cp -f backend/testing/**/*-*-zkProof.json ../testdata/prover/output


# The version of go-corset as presented in the go.mod
GO_CORSET_VERSION := $(shell grep 'go-corset' go.mod | sed 's/.* //')
# The go-corset module name
GO_CORSET_MOD := github.com/consensys/go-corset
# The go path, usually $HOME/go
GOPATH := $(shell go env GOPATH)
# The directory where to find the zkc testdata
ZKC_TESTDATA_SYNC_SRC_DIR := $(GOPATH)/pkg/mod/$(GO_CORSET_MOD)@$(GO_CORSET_VERSION)/testdata/zkc
ZKC_TESTDATA_SYNC_DST_DIR := zkcdriver/testdata/synced
ZKC_BIN := $(GOPATH)/bin/zkc

# these targets are not defined as depending on source files,
# so they should be recompiled by default “just in case”.
.PHONY: \
	test \
	ci-lint \
	download-zkc-testdata \
	go-install-zkc-compiler \
	check-zkc-testdata

##
## Run all the unit-tests
##
test:
	go test -tags debug ./...

##
## Run the CI linting
##
ci-lint:
	golangci-lint run --timeout 5m

##
## Install the go-corset zkc compiler
##
go-install-zkc-compiler:
	go install $(GO_CORSET_MOD)/cmd/zkc@$(GO_CORSET_VERSION)

##
## Check that the committed zkc testdata is up to date (for CI)
##
check-zkc-testdata: download-zkc-testdata
	git diff --exit-code zkcdriver/testdata/synced/

##
## Download the zkc testdata
##
download-zkc-testdata: go-install-zkc-compiler
	rm -rf $(ZKC_TESTDATA_SYNC_DST_DIR)/*
	mkdir -p $(ZKC_TESTDATA_SYNC_DST_DIR)
	go get github.com/consensys/go-corset@$(GO_CORSET_VERSION)
	cp -rf $(ZKC_TESTDATA_SYNC_SRC_DIR)/* $(ZKC_TESTDATA_SYNC_DST_DIR)
	chmod -R u+w $(ZKC_TESTDATA_SYNC_DST_DIR)
	rm -rf $(ZKC_TESTDATA_SYNC_DST_DIR)/invalid
	find $(ZKC_TESTDATA_SYNC_DST_DIR) -name "*.zkc" | while read f; do \
		$(ZKC_BIN) compile --wir -o $${f%.zkc}.bin $$f &> /dev/null || true; \
	done




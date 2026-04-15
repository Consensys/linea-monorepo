package main

import "C"

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
)

//go:generate go build -tags nocorset -ldflags "-s -w" -buildmode=c-shared -o libdecompressor.so libdecompressor.go
func main() {}

var (
	dictStore dictionary.Store
	lastError error
	lock      sync.Mutex // probably unnecessary if coordinator guarantees single-threaded access
)

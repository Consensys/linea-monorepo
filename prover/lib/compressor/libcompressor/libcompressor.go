package main

import "C"

import (
	"sync"

	blob_v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
)

//go:generate go build -tags nocorset -ldflags "-s -w" -buildmode=c-shared -o libcompressor.so libcompressor.go
func main() {}

type instance struct {
	compressor *blob_v1.BlobMaker
	lastError  error
	mu         sync.Mutex
}

var (
	instances        = map[C.int]*instance{}
	nextHandle C.int = 1
	registryMu sync.Mutex
)

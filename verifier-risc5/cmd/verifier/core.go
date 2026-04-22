package main

import "github.com/consensys/linea-monorepo/verifier-risc5/internal/workload"

var Result uint64

func ComputeWords(words []uint64) uint64 {
	return workload.Compute(words)
}

func Compute() uint64 {
	return ComputeWords(workload.DefaultWords[:])
}

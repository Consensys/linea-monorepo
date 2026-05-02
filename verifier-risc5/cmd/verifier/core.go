package main

import "github.com/consensys/linea-monorepo/verifier-risc5/internal/workload"

var Result uint64

func ComputeWords(words []uint64) uint64 {
	result, _ := ComputeWordsChecked(words)
	return result
}

func ComputeWordsChecked(words []uint64) (uint64, bool) {
	return computeWordsImpl(words)
}

func Compute() uint64 {
	return ComputeWords(workload.DefaultWords[:])
}

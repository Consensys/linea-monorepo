//go:build !baremetal || !zkvm_precompile

package main

import "github.com/consensys/linea-monorepo/verifier-risc5/internal/workload"

func computeWordsImpl(words []uint64) (uint64, bool) {
	return workload.Compute(words), true
}

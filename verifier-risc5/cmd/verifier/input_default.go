//go:build !baremetal

package main

import "github.com/consensys/linea-monorepo/verifier-risc5/internal/workload"

func loadVerifierInput() (verifierInput, bool) {
	return verifierInput{
		Words:    workload.DefaultWords[:],
		Expected: workload.DefaultExpected,
	}, true
}

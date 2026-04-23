//go:build !baremetal

package main

import (
	"os"

	"github.com/consensys/linea-monorepo/verifier-risc5/internal/guestabi"
)

func main() {
	input, ok := loadVerifierInput()
	if !ok {
		os.Exit(int(guestabi.StatusCodeInputError))
	}

	var okCompute bool
	Result, okCompute = ComputeWordsChecked(input.Words)
	if !okCompute {
		os.Exit(int(guestabi.StatusCodeInputError))
	}

	if Result == input.Expected {
		os.Exit(int(guestabi.StatusCodeSuccess))
	}

	os.Exit(int(guestabi.StatusCodeMismatch))
}

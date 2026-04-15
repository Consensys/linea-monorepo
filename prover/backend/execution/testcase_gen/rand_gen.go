package testcase_gen

import (
	"math/rand/v2"
)

// Random number generator
type RandGen struct {
	rand.Rand
	Params struct {
		SupTxPerBlock         int
		SupL2L1LogsPerBlock   int
		SupMsgReceiptPerBlock int
	}
}

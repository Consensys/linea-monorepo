//go:build nocorset

package zkevm

import (
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

// This function and this tags exists so it is possible to test the entirety
// of the code of the repo without having to compile corset.
func AssignFromCorset(reader io.Reader, run *wizard.ProverRuntime) {
	panic("called corset in a non-corset build")
}

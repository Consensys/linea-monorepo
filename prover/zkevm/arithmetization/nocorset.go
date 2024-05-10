//go:build nocorset

package arithmetization

import (
	"io"

	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

// This function and this tags exists so it is possible to test the entirety
// of the code of the repo without having to compile corset.
func AssignFromCorset(reader io.Reader, run *wizard.ProverRuntime) {
	panic("called corset in a non-corset build")
}

//go:build nocorset

package arithmetization

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// This function and this tags exists so it is possible to test the entirety
// of the code of the repo without having to compile corset.
func AssignFromCorset(traceFile string, run *wizard.ProverRuntime) {
	panic("called corset in a non-corset build")
}

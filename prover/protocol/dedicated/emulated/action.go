package emulated

import "github.com/consensys/linea-monorepo/prover/protocol/wizard"

// ProverActionFn is a wrapper to register a function as a prover action
type ProverActionFn struct {
	fn func(run *wizard.ProverRuntime)
}

func (a *ProverActionFn) Run(run *wizard.ProverRuntime) {
	a.fn(run)
}

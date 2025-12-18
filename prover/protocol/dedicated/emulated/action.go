package emulated

import "github.com/consensys/linea-monorepo/prover/protocol/wizard"

// proverActionFn is a wrapper to register a function as a prover action
type proverActionFn struct {
	fn func(run *wizard.ProverRuntime)
}

func (a *proverActionFn) Run(run *wizard.ProverRuntime) {
	a.fn(run)
}

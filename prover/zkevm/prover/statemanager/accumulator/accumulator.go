package accumulator

import (
	"github.com/consensys/zkevm-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

func AccumulatorDefine(
	// compiled IOP
	comp *wizard.CompiledIOP,
	// name of the Accumulator proof check instance
	name string,
	// Accumulaor object
	am *Accumulator,
) {
	am.Define(comp, name)

}

func AccumulatorAssign(
	// Accumulator object
	am *Accumulator,
	// State Manager traces
	traces []statemanager.DecodedTrace,
	// run object
	run *wizard.ProverRuntime,
) {
	am.Assign(run, traces)
}

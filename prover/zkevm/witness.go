package zkevm

import (
	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
)

// Witness is a collection of prover inputs used to derive an assignment to the
// full proving scheme.
type Witness struct {
	// ExecTracesFPath is the filepath toward the execution traces to use for
	// proof trace generation.
	ExecTracesFPath string
	// StateManager traces
	SMTraces [][]statemanager.DecodedTrace
}

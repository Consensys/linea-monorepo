package zkevm

import (
	"io"

	"github.com/consensys/zkevm-monorepo/prover/backend/execution/statemanager"
)

// Witness is a collection of prover inputs used to derive an assignment to the
// full proving scheme.
type Witness struct {
	// Reader to be passed to corset to derive the assignment
	ExecTraces func() io.ReadCloser
	// StateManager traces
	SMTraces [][]statemanager.DecodedTrace
}

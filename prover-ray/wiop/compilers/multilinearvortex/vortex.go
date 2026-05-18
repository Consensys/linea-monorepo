package multilinearvortex

import (
	"sort"

	"github.com/consensys/linea-monorepo/prover-ray/wiop"
)

// Compile converts a many-evaluation-same-point Lagrange evaluation IOP
// protocol on a multi-round set of polynomials protocol into a ROM-secure.
func Compile(sys *wiop.System) {

}

// CommitmentPhase is the prover action responsible for committing to the
// columns of the current round.
type CommitmentPhase struct {
	// round is the round at which this commitment phase is committed.
	Round int
	// ColumnIDs is the list of committed columns to be committed to at the
	// current round.
	ColumnIDs []wiop.ObjectID
	// NbColumn is the static number of column matrix size
	NbColumn int
}

// Run is the prover
func (cp *CommitmentPhase) Run(rt wiop.Runtime) {

	// 1. Gather all the column assignment
	// 2. Layout the columns in the Vortex matrix
	// 3. Commit the Vortex matrix
	// 4. Saves the Vortex commitment in the runtime "blobs"
	// 5. Also saves the layout in the proof blob.

	columnSizes := make([]int, len(cp.ColumnIDs))
	for i := range columnSizes {
		module := rt.System.Modules[cp.ColumnIDs[i].Slot()]
		columnSizes[i] = module.RuntimeSize(rt)
	}

	sort.Stable()

}

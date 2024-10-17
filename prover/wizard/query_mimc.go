package wizard

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

var _ Query = &QueryMiMC{}

// A QueryMiMC query over a set of 3 columns (block, oldState, newState) enforces
// that newState is the result of applying the QueryMiMC compression function to
// block and oldState.
//
// We use the MiMC as specified in the following paper.
// https://eprint.iacr.org/2016/492.pdf
//
// And the compression function uses the Miyaguchi's construction
// https://en.wikipedia.org/wiki/One-way_compression_function#Miyaguchi.E2.80.93Preneel
type QueryMiMC struct {
	// The columns on which the query applies
	Blocks, OldState, NewState Column
	metadata                   *metadata
	*subQuery
}

// NewMiMC constructs a new [QueryMiMC] query
func (api *API) NewMiMC(block, oldState, newState Column) *QueryMiMC {

	// Sanity-check : All columns must have the same length
	if block.Size() != oldState.Size() || block.Size() != newState.Size() {
		utils.Panic("block, oldState and newState must have the same length %v %v %v", block.Size(), oldState.Size(), newState.Size())
	}

	var (
		round = max(block.Round(), oldState.Round(), newState.Round())
		res   = &QueryMiMC{
			OldState: oldState,
			NewState: newState,
			Blocks:   block,
			metadata: api.newMetadata(),
			subQuery: &subQuery{
				round: round,
			},
		}
	)

	api.queries.addToRound(round, res)
	return res
}

/*
The verifier checks that the permutation was applied correctly
*/
func (m QueryMiMC) Check(run Runtime) error {

	var (
		blocks    = m.Blocks.GetAssignment(run)
		oldStates = m.OldState.GetAssignment(run)
		newStates = m.NewState.GetAssignment(run)
	)

	for i := 0; i < newStates.Len(); i++ {
		var (
			block      = blocks.Get(i)
			oldState   = oldStates.Get(i)
			newState   = newStates.Get(i)
			recomputed = mimc.BlockCompression(oldState, block)
		)

		if recomputed != newState {
			return fmt.Errorf(
				"QueryMiMC compression check failed for row #%v : block %v, oldState %v, newState %v",
				i, block.String(), oldState.String(), newState.String(),
			)
		}
	}

	return nil
}

// Check the mimc relation in a gnark circuit
func (m QueryMiMC) CheckGnark(api frontend.API, run RuntimeGnark) {

	var (
		blocks    = m.Blocks.GetAssignmentGnark(api, run)
		oldStates = m.OldState.GetAssignmentGnark(api, run)
		newStates = m.NewState.GetAssignmentGnark(api, run)
	)

	for i := 0; i < len(newStates); i++ {
		var (
			block      = blocks[i]
			oldState   = oldStates[i]
			newState   = newStates[i]
			recomputed = mimc.GnarkBlockCompression(api, oldState, block)
		)

		api.AssertIsEqual(newState, recomputed)
	}
}

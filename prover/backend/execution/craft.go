package execution

import (
	"github.com/consensys/linea-monorepo/prover/backend/execution/bridge"
	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Craft prover's functional inputs
func CraftProverOutput(
	cfg *config.Config,
	req *Request,
) (Response, [][]statemanager.DecodedTrace) {

	var (
		po              Response
		smTraces        [][]statemanager.DecodedTrace
		l2BridgeAddress = cfg.Layer2.MsgSvcContract
	)

	// Extract the data from the block
	blocks := req.Blocks()
	po.BlocksData = make([]BlockData, len(blocks))
	for i := range blocks {

		// The assignment is not made in the loop directly otherwise the
		// linter will detect "memory aliasing in a for loop" even though
		// we do not modify the data of the block.
		block := &blocks[i]

		// Encode the transactions
		po.BlocksData[i].RlpEncodedTransactions = RlpTransactions(block)
		po.BlocksData[i].FromAddresses = utils.HexConcat(FromAddresses(block)...)

		// Fetch the transaction indices
		logs := req.LogsForBlock(i)
		po.BlocksData[i].BatchReceptionIndices = bridge.BatchReceptionIndex(
			logs,
			l2BridgeAddress,
		)

		// Fetch the timestamps
		po.BlocksData[i].TimeStamp = block.Time()

		// Filter the logs L2 to L1, and hash them before sending them
		// back to the coordinator.
		l2l1MessageHashes := bridge.L2L1MessageHashes(logs, l2BridgeAddress)
		po.BlocksData[i].L2ToL1MsgHashes = l2l1MessageHashes

		// Also filters the RollingHashUpdated logs
		events := bridge.ExtractRollingHashUpdated(logs, l2BridgeAddress)
		if len(events) > 0 {
			po.BlocksData[i].LastRollingHashUpdatedEvent = events[len(events)-1]
		}
	}

	// Value of the first blocks
	po.FirstBlockNumber = int(blocks[0].NumberU64())

	// Add into that the data of the state-manager
	// Run the inspector and pass the parsed traces back to the caller.
	// These traces may be used by the state-manager module depending on
	// if the flag `PROVER_WITH_STATE_MANAGER`
	smTraces = InspectStateManagerTraces(cfg, req, &po)

	// Hash everything into the prover inputs
	po.ComputeProofInput()

	return po, smTraces
}

// InspectStateManagerTraces parsed the state-manager traces from the given
// input and inspect them to see if they are self-consistent and if they match
// the parentStateRootHash. This behaviour can be altered by setting the field
// `tolerate_state_root_hash_mismatch`, see its documentation. In case of
// success, the function returns the decoded state-manager traces. Otherwise, it
// panics.
func InspectStateManagerTraces(
	cfg *config.Config,
	req *Request,
	resp *Response,
) (traces [][]statemanager.DecodedTrace) {
	// Extract the traces from the inputs
	traces = req.StateManagerTraces()
	firstParent := req.ZkParentStateRootHash
	parent := req.ZkParentStateRootHash

	for i := range traces {

		if len(traces[i]) > 0 {
			// Run the trace inspection routine
			old, new, err := statemanager.CheckTraces(traces[i])
			// The trace must have been validated
			if err != nil {
				utils.Panic("error parsing the state manager traces : %v", err)
			}

			// The "old of a block" must equal the parent
			if old != parent {
				utils.Panic("old does not match with parent root hash")
			}

			// Populate the prover's output with the recovered root hash
			resp.BlocksData[i].RootHash = new.Hex()
			parent = new
		} else {
			// This can happen when there are no transaction in a block
			// In this case, we do not need to do anything
			resp.BlocksData[i].RootHash = parent.Hex()
		}

	}

	resp.ParentStateRootHash = firstParent.Hex()

	// Returns the traces to be used by the state-manager prover. nil
	// if no traces are available.
	return traces
}

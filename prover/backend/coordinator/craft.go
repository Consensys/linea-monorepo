package coordinator

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/bridge"
	"github.com/consensys/accelerated-crypto-monorepo/backend/config"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// Craft prover's functional inputs
func CraftProverOutput(pi *ProverInput, l2BridgeAddress common.Address) (po ProverOutput, smTraces [][]any) {

	// Prover output
	po = ProverOutput{}

	// Extract the data from the block
	blocks := pi.Blocks()
	po.BlocksData = make([]BlockData, len(blocks))
	for i := range blocks {

		// The assignment is not made in the loop directly otherwise the
		// linter will detect "memory aliasing in a for loop" even though
		// we do not modify the data of the block.
		block := blocks[i]

		// Encode the transactions
		po.BlocksData[i].RlpEncodedTransactions = RlpTransactions(&block)
		po.BlocksData[i].FromAddresses = HexConcat(FromAddresses(&block)...)

		// Fetch the transaction indices
		logs := pi.LogsForBlock(i)
		po.BlocksData[i].BatchReceptionIndices = bridge.BatchReceptionIndex(
			logs,
			l2BridgeAddress,
		)

		// Sanity-check the alledged batch reception indices
		for _, index := range po.BlocksData[i].BatchReceptionIndices {
			bridge.MustLookLikeABatchReceptionTx(*block.Transactions()[index], l2BridgeAddress)
		}

		// Fetch the timestamps
		po.BlocksData[i].TimeStamp = block.Time()

		// Filter the logs L2 to L1, and hash them before sending them
		// back to the coordinator.
		l2l1MessageHashes := bridge.L2L1MessageHashes(logs, l2BridgeAddress)
		po.BlocksData[i].L2ToL1MsgHashes = l2l1MessageHashes
	}

	// Value of the first blocks
	po.FirstBlockNumber = int(blocks[0].NumberU64())

	// Add into that the data of the state-manager
	if pi.HasStateManagerTraces() {
		// Run the inspector and pass the parsed traces back to the caller.
		// These traces may be used by the state-manager module depending on
		// if the flag `PROVER_WITH_STATE_MANAGER`
		smTraces = InspectStateManagerTraces(pi, &po)

	}

	if !pi.HasStateManagerTraces() && config.MustGetProver().WithStateManager {
		// If passed the flags `PROVER_WITH_STATE_MANAGER`, we understand that
		// the intention of the user was to pass traces from the state-manager
		// but it was not found.
		logrus.Errorf("The flag `PROVER_WITH_STATE_MANAGER` is set, but could not" +
			"find the state manager traces. Check your configuration to make sure" +
			"that you are passing the flag on purpose or check the \"prover-input\"" +
			"format to ensure this is compatible with the prover's expectations.")
		smTraces = [][]any{}
	}

	if !pi.HasStateManagerTraces() {
		// Else, we use the keccak state-representation of Ethereum in the claims
		// in default of proving the state.
		PopulateWithKeccakRootHashes(pi, &po)
	}

	// Hash everything into the prover inputs
	po.ComputeProofInput()

	return po, smTraces
}

// Get the the state root hash a,d compute the root hash out of them
func InspectStateManagerTraces(pi *ProverInput, po *ProverOutput) (traces [][]any) {
	// Extract the traces from the inputs
	traces = pi.StateManagerTraces()
	firstParent := pi.ZKParentStateRootHash()
	parent := pi.ZKParentStateRootHash()

	for i := range traces {

		if len(traces[i]) > 0 {
			// Run the trace inspection routine
			old, new, err := eth.CheckTraces(traces[i])

			// The trace must have been validated
			if err != nil {
				utils.Panic("error parsing the state manager traces : %v", err)
			}

			// The "old of a block" must equal the parent
			if old != parent {
				utils.Panic("old does not match with parent root hash")
			}

			// Populate the prover's output with the recovered root hash
			po.BlocksData[i].RootHash = new.Hex()
			parent = new
		} else {
			// This can happen when there are no transaction in a block
			// In this case, we do not need to do anything
			po.BlocksData[i].RootHash = parent.Hex()
		}

	}

	po.ParentStateRootHash = firstParent.Hex()

	// Returns the traces to be used by the state-manager prover. nil
	// if no traces are available.
	return traces
}

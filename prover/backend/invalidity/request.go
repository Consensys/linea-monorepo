package invalidity

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// Request file for a forced transaction attempted to be included in the current aggregation
// The forcedTransactionNumbers from request files per aggregation, should create a consecutive sequence.
type Request struct {
	// RLP encoding of the forced transaction (hex encoded with 0x prefix).
	RlpEncodedTx string `json:"ftxRLP"`

	// Transaction number assigned by L1 contract (decimal encoding)
	ForcedTransactionNumber uint64 `json:"ftxNumber"`

	// Previous FTX rolling hash, i.e. the FTX stream hash of the previous forced transaction.
	PrevFtxRollingHash types.Bytes32 `json:"prevFtxRollingHash"`

	// The block number deadline before which one expects to see the transaction (decimal encoding)
	DeadlineBlockHeight uint64 `json:"ftxBlockNumberDeadline"`

	// The type of invalidity for the forced transaction.
	// Valid values: BadNonce, BadBalance, BadPrecompile, TooManyLogs, FilteredAddresses
	InvalidityType invalidity.InvalidityType `json:"invalidityType"`

	// Parent block hash
	ParentBlockHash types.Bytes32 `json:"parentBlockHash"`

	// ZK parent state root hash
	ZkParentStateRootHash types.Bytes32 `json:"zkParentStateRootHash"`

	// Path to conflated execution traces file (required for BadPrecompile, TooManyLogs cases)
	ConflatedExecutionTracesFile string `json:"conflatedExecutionTracesFile,omitempty"`

	// Account merkle proof from Shomei (rollup_getProof response)
	// Required for BadNonce, BadBalance cases
	AccountMerkleProof *AccountMerkleProof `json:"accountMerkleProof,omitempty"`

	// ZK state merkle proof (full Shomei trace)
	// Required for BadPrecompile, TooManyLogs cases
	// Requires Shomei to trace a block that does not exist
	ZkStateMerkleProof [][]statemanager.DecodedTrace `json:"zkStateMerkleProof,omitempty"`
	// case of FilteredAddresses, accountMerkleProof=null, zkStateMerkleProof=null

	// Simulated execution block number (ParentAggregationLastBlockNumber + 1)
	SimulatedExecutionBlockNumber uint64 `json:"simulatedExecutionBlockNumber,omitempty"`

	// Simulated execution block timestamp
	SimulatedExecutionBlockTimestamp uint64 `json:"simulatedExecutionBlockTimestamp,omitempty"`
	// for type of FilteredAddresses one of these two must be present
	FilteredAddressFrom types.EthAddress `json:"filteredAddressFrom,omitempty"`
	FilteredAddressTo   types.EthAddress `json:"filteredAddressTo,omitempty"`
}

// AccountMerkleProof represents the Shomei response from rollup_getProof(account address)
// Used for BadNonce, BadBalance invalidity cases
type AccountMerkleProof struct {
	// TODO: Define the structure based on Shomei's rollup_getProof response
	RawResponse []byte `json:"rawResponse"`
}

// AccountTrieInputs extracts the AccountTrieInputs from the AccountMerkleProof
// for the given sender address. Used for BadNonce and BadBalance cases.
// TODO: Implement once we reach a consensus on Request structure
func (req *Request) AccountTrieInputs() (invalidity.AccountTrieInputs, types.EthAddress, error) {
	return invalidity.AccountTrieInputs{}, types.EthAddress{}, fmt.Errorf("not implemented: AccountTrieInputs requires consensus on Request structure")
}

// Validate checks that the required fields are present based on the InvalidityType
func (req *Request) Validate() error {
	switch req.InvalidityType {
	case invalidity.BadNonce, invalidity.BadBalance:
		if req.AccountMerkleProof == nil {
			return fmt.Errorf("accountMerkleProof is required for %s invalidity type", req.InvalidityType)
		}
	case invalidity.BadPrecompile, invalidity.TooManyLogs:
		if req.ConflatedExecutionTracesFile == "" {
			return fmt.Errorf("conflatedExecutionTracesFile is required for %s invalidity type", req.InvalidityType)
		}
		if req.ZkStateMerkleProof == nil {
			return fmt.Errorf("zkStateMerkleProof is required for %s invalidity type", req.InvalidityType)
		}
	case invalidity.FilteredAddresses:
		if req.FilteredAddressFrom == (types.EthAddress{}) && req.FilteredAddressTo == (types.EthAddress{}) {
			return fmt.Errorf("one of filteredAddressFrom or filteredAddressTo is required for %s invalidity type", req.InvalidityType)
		}
	default:
		return fmt.Errorf("unknown invalidity type: %s", req.InvalidityType)
	}

	if req.SimulatedExecutionBlockNumber == 0 {
		return fmt.Errorf("simulatedExecutionBlockNumber is required for %s invalidity type", req.InvalidityType)
	}
	if req.SimulatedExecutionBlockTimestamp == 0 {
		return fmt.Errorf("simulatedExecutionBlockTimestamp is required for %s invalidity type", req.InvalidityType)
	}

	return nil
}

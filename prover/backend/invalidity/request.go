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
	PrevFtxRollingHash types.Bls12377Fr `json:"prevFtxRollingHash"`

	// The block number deadline before which one expects to see the transaction (decimal encoding)
	DeadlineBlockHeight uint64 `json:"ftxBlockNumberDeadline"`

	// The type of invalidity for the forced transaction.
	// Valid values: BadNonce, BadBalance, BadPrecompile, TooManyLogs, FilteredAddressFrom, FilteredAddressTo
	InvalidityType invalidity.InvalidityType `json:"invalidityType"`

	// ZK parent state root hash
	ZkParentStateRootHash types.KoalaOctuplet `json:"zkParentStateRootHash"`

	// Path to conflated execution traces file (required for BadPrecompile, TooManyLogs cases)
	ConflatedExecutionTracesFile string `json:"conflatedExecutionTracesFile,omitempty"`

	// Account merkle proof from Shomei (rollup_getProof response)
	// Required for BadNonce, BadBalance cases
	AccountMerkleProof statemanager.DecodedTrace `json:"accountMerkleProof,omitempty"`

	// ZK state merkle proof (full Shomei trace)
	// Required for BadPrecompile, TooManyLogs cases
	// Requires Shomei to trace a block that does not exist
	ZkStateMerkleProof [][]statemanager.DecodedTrace `json:"zkStateMerkleProof,omitempty"`
	// case of FilteredAddressFrom/FilteredAddressTo: accountMerkleProof=null, zkStateMerkleProof=null

	// Simulated execution block number (ParentAggregationLastBlockNumber + 1)
	SimulatedExecutionBlockNumber uint64 `json:"simulatedExecutionBlockNumber,omitempty"`

	// Simulated execution block timestamp
	SimulatedExecutionBlockTimestamp uint64 `json:"simulatedExecutionBlockTimestamp,omitempty"`

	// FilteredAddressFrom: the filtered "from" address (required for FilteredAddressFrom type)
	FilteredAddressFrom types.EthAddress `json:"filteredAddressFrom,omitempty"`
	// FilteredAddressTo: the filtered "to" address (required for FilteredAddressTo type)
	FilteredAddressTo types.EthAddress `json:"filteredAddressTo,omitempty"`
}

// AccountTrieInputs extracts the AccountTrieInputs from the AccountMerkleProof
// Used for BadNonce and BadBalance cases.
func (req *Request) AccountTrieInputs() (invalidity.AccountTrieInputs, types.EthAddress, error) {
	trace := req.AccountMerkleProof

	// The AccountMerkleProof should be a ReadNonZeroTrace for world state (account exists)
	readTrace, ok := trace.Underlying.(statemanager.ReadNonZeroTraceWS)
	if !ok {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, fmt.Errorf(
			"accountMerkleProof must be a ReadNonZeroTrace for world state, got type=%d location=%s",
			trace.Type, trace.Location,
		)
	}

	// Compute the leaf hash: Poseidon2(Prev, Next, HKey, HVal)
	leaf := readTrace.LeafOpening.Hash()

	return invalidity.AccountTrieInputs{
		Account:     readTrace.Value.Account,
		LeafOpening: readTrace.LeafOpening,
		Leaf:        leaf,
		Proof:       readTrace.Proof,
		Root:        readTrace.SubRoot,
	}, types.EthAddress(readTrace.Key), nil
}

// Validate checks that the required fields are present based on the InvalidityType
func (req *Request) Validate() error {
	switch req.InvalidityType {
	case invalidity.BadNonce, invalidity.BadBalance:
		if req.AccountMerkleProof.Underlying == nil {
			return fmt.Errorf("accountMerkleProof is required for %s invalidity type", req.InvalidityType)
		}
	case invalidity.BadPrecompile, invalidity.TooManyLogs:
		if req.ConflatedExecutionTracesFile == "" {
			return fmt.Errorf("conflatedExecutionTracesFile is required for %s invalidity type", req.InvalidityType)
		}
		if req.ZkStateMerkleProof == nil {
			return fmt.Errorf("zkStateMerkleProof is required for %s invalidity type", req.InvalidityType)
		}
	case invalidity.FilteredAddressFrom:
		if req.FilteredAddressFrom == (types.EthAddress{}) {
			return fmt.Errorf("filteredAddressFrom is required for %s invalidity type", req.InvalidityType)
		}
	case invalidity.FilteredAddressTo:
		if req.FilteredAddressTo == (types.EthAddress{}) {
			return fmt.Errorf("filteredAddressTo is required for %s invalidity type", req.InvalidityType)
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

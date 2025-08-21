package invalidity

import (
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Request file for  a  forced transaction attempted to be included in the current aggregation
// The forcedTransactionNumbers from request files per aggregation, should create a consecutive sequence.
type Request struct {
	// the forcedTransaction attempted to be included in the current aggregation,
	// @azam replace its type with a payload type since we do'nt need the signature part.
	ForcedTransactionPayLoad *ethtypes.Transaction `json:"forcedTransactionPayload"`
	// transaction number assigned by L1 contract
	ForcedTransactionNumber uint64 `json:"forcedTransactionNumber"`
	// from address as given by L1 contract
	FromAddresses types.EthAddress `json:"fromAddresses"`
	// the type of invalidity for each forced transaction;
	// for the executed valid transaction it is set to [invalidity.BadNonce]
	InvalidityTypes invalidity.InvalidityType `json:"invalidityTypes"`
	// account tri from the final state of the current aggregation.
	AccountTrie invalidity.AccountTrieInputs `json:"accountTrie"`
	// expected Block number as assigned by the L1 contract
	ExpectedBlockHeight uint64 `json:"expectedBlockHeight"`
	// state root hash of the current aggregation, so this is the same for all the request files relevant to the same aggregation
	StateRootHash types.Bytes32 `json:"stateRootHash"`
	// RollingHash associated with the transaction
	FtxStreamHash types.Bytes32 `json:"ftxStreamHash"`
	// Parent block hash
	ParentBlockHash types.Bytes32 `json:"parentBlockHash"`
}

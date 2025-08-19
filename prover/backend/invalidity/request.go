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
	ForcedTransactionPayLoad *ethtypes.Transaction
	// transaction number assigned by L1 contract
	ForcedTransactionNumber uint64
	// from address as given by L1 contract
	FromAddresses types.EthAddress
	// the type of invalidity for each forced transaction;
	// for the executed valid transaction it is set to [invalidity.BadNonce]
	InvalidityTypes invalidity.InvalidityType
	// account tri from the final state of the current aggregation.
	AccountTri invalidity.AccountTrieInputs
	// expected Block number as assigned by the L1 contract
	ExpectedBlockHeights uint64
	// state root hash of the current aggregation, so this is the same for all the request files relevant to the same aggregation
	StateRootHash types.Bytes32
	// RollingHash associated with the transaction
	RollingHashTx types.Bytes32
}

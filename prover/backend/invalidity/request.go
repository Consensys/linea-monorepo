package invalidity

import (
	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// Request file for a forced transaction attempted to be included in the current aggregation
// The forcedTransactionNumbers from request files per aggregation, should create a consecutive sequence.
type Request struct {
	ExecutionCtx execution.Request `json:"executionCtx"`
	// RLP encoding of the transaction see [ethereum.EncodeTxForSigning()]
	RlpEncodedTx []byte `json:"forcedTransactionRLP"`
	// transaction number assigned by L1 contract
	ForcedTransactionNumber uint64 `json:"forcedTransactionNumber"`
	// the height before which one expect to see the transaction.
	ExpectedBlockHeight uint64 `json:"expectedBlockHeight"`

	// previous FTX stream hash, i.e. the FTX stream hash of the previous forced transaction.
	PrevFtxRollingHash types.Bytes32 `json:"prevFtxRollingHash"`

	// the type of invalidity for each forced transaction;
	// for the executed valid transaction it is set to [invalidity.BadNonce]
	InvalidityTypes invalidity.InvalidityType `json:"invalidityTypes"`
}

// AccountTrieInputs extracts the AccountTrieInputs from the ZkStateMerkleProof traces
// for the given sender address. It looks for a world state ReadNonZero trace (type=0, location="0x")
// that matches the sender address.
func (req *Request) AccountTrieInputs() (invalidity.AccountTrieInputs, types.EthAddress, error) {
	panic("not implemented, we first have a consensus on Request structure")
}

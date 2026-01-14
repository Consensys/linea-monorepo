package invalidity

import (
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	circuitInvalidity "github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/ethereum/go-ethereum/core/types"
)

// FuncInput are all the relevant fields parsed by the prover that
// are functionally useful to contextualize what the proof is proving. This
// is used by the aggregation circuit to ensure that the invalidity proofs
// relate to consecutive Linea forced transactions.
func (req *Request) FuncInput() *public_input.Invalidity {

	// Decode the signed transaction to recover the sender address
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(req.RlpEncodedTx); err != nil {
		utils.Panic("could not decode the RlpEncodedTx: %v", err)
	}
	fromAddress := ethereum.GetFrom(tx)

	// Compute the signing hash (same as signer.Hash(tx))
	signer := ethereum.GetSigner(tx)
	txHash := signer.Hash(tx)

	// Compute the FtxRollingHash from the previous rolling hash
	ftxRollingHash := circuitInvalidity.UpdateFtxRollingHash(
		req.PrevFtxRollingHash,
		tx,
		int(req.DeadlineBlockHeight),
		fromAddress,
	)

	fi := &public_input.Invalidity{
		TxHash:              txHash,
		TxNumber:            uint64(req.ForcedTransactionNumber),
		FromAddress:         fromAddress,
		ExpectedBlockHeight: uint64(req.DeadlineBlockHeight),
		StateRootHash:       req.ExecutionCtx.ZkParentStateRootHash,
		FtxRollingHash:      ftxRollingHash,
	}
	return fi

}

package invalidity

import (
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	circuitInvalidity "github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// FuncInput are all the relevant fields parsed by the prover that
// are functionally useful to contextualize what the proof is proving. This
// is used by the aggregation circuit to ensure that the invalidity proofs
// relate to consecutive Linea forced transactions.
func (req *Request) FuncInput() *public_input.Invalidity {

	tx, err := ethereum.RlpDecodeWithSignature(req.RlpEncodedTx)
	if err != nil {
		utils.Panic("could not decode the RlpEncodedTx: %v", err)
	}
	fromAddress := ethereum.GetFrom(tx)

	// Compute the unsigned transaction hash
	txHash := ethereum.GetTxHash(tx)

	// Compute the FtxRollingHash from the previous rolling hash
	ftxRollingHash := circuitInvalidity.UpdateFtxRollingHash(
		req.PrevFtxRollingHash,
		txHash,
		req.DeadlineBlockHeight,
		fromAddress,
	)

	fi := &public_input.Invalidity{
		TxHash:              txHash,
		TxNumber:            uint64(req.ForcedTransactionNumber),
		FromAddress:         fromAddress,
		ExpectedBlockHeight: uint64(req.DeadlineBlockHeight),
		StateRootHash:       req.ZkParentStateRootHash,
		FtxRollingHash:      ftxRollingHash,
	}
	return fi

}

package invalidity

import (
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// FuncInput are all the relevant fields parsed by the prover that
// are functionally useful to contextualize what the proof is proving. This
// is used by the aggregation circuit to ensure that the invalidity proofs
// relate to consecutive Linea forced transactions.
func (req *Request) FuncInput() *public_input.Invalidity {

	var (
		txHash = crypto.Keccak256(req.RlpEncodedTx)
		fi     = &public_input.Invalidity{
			TxHash:              common.Hash(txHash),
			TxNumber:            uint64(req.ForcedTransactionNumber),
			FromAddress:         req.FromAddresses,
			ExpectedBlockHeight: uint64(req.ExpectedBlockHeight),
			StateRootHash:       req.StateRootHash,
			FtxRollingHash:      req.FtxRollingHash,
		}
	)
	return fi

}

package invalidity

import (
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
)

// FuncInput are all the relevant fields parsed by the prover that
// are functionally useful to contextualize what the proof is proving. This
// is used by the aggregation circuit to ensure that the invalidity proofs
// relate to consecutive Linea forced transactions.
func (req *Request) FuncInput() *public_input.Invalidity {

	var (
		fi = &public_input.Invalidity{
			TxHash:              req.ForcedTransactionPayLoad.Hash(),
			TxNumber:            uint64(req.ForcedTransactionNumber),
			FromAddress:         req.FromAddresses,
			ExpectedBlockHeight: uint64(req.ExpectedBlockHeights),
			StateRootHash:       req.StateRootHash,
			RollingHashTx:       req.RollingHashTx,
		}
	)

	return fi
}

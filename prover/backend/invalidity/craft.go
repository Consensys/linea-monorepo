package invalidity

import (
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
)

// FuncInput are all the relevant fields parsed by the prover that
// are functionally useful to contextualize what the proof is proving. This
// is used by the aggregation circuit to ensure that the execution proofs
// relate to consecutive Linea block execution.
func (rsp *Response) FuncInput(i int) *public_input.Invalidity {

	var (
		// req = rsp.Request
		fi = &public_input.Invalidity{
			/*TxHash:              req.TxHash,
			TxNumber:            uint64(rsp.TxNumber),
			FromAddress:         rsp.FromAddress,
			ExpectedBlockHeight: uint64(rsp.ExpectedBlockHeight),
			StateRootHash:       rsp.StateRootHash,*/
		}
	)

	return fi
}

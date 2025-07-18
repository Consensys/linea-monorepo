package public_input

import (
	"hash"

	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
)

// Invalidity represents the functional public inputs for the invalidity circuit
// The mimc hash over functional inputs is set as the public input of the circuit.
type Invalidity struct {
	TxHash        common.Hash      // hash of the transaction
	FromAddress   types.EthAddress // address of the sender
	BlockHeight   uint64           // block number for the current virtual block,
	StateRootHash types.Bytes32    // state-root-hash on which the invalidity is based
}

// Sum compute the mimc hash over the functional public inputs
func (pi *Invalidity) Sum(hsh hash.Hash) []byte {
	if hsh == nil {
		hsh = mimc.NewMiMC()
	}

	hsh.Reset()
	hsh.Write(pi.TxHash[:])
	hsh.Write(pi.FromAddress[:])
	writeNum(hsh, pi.BlockHeight)
	hsh.Write(pi.StateRootHash[:])

	return hsh.Sum(nil)
}

func (pi *Invalidity) SumAsField() field.Element {

	var (
		sumBytes = pi.Sum(nil)
		sum      = new(field.Element).SetBytes(sumBytes)
	)

	return *sum
}

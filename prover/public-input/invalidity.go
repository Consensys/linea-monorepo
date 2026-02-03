package public_input

import (
	"hash"

	mimc "github.com/consensys/linea-monorepo/prover/crypto/mimc_bls12377"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
)

// Invalidity represents the functional public inputs for the invalidity circuit
// The mimc hash over functional inputs is set as the public input of the circuit.
type Invalidity struct {
	TxHash              common.Hash // hash of the transaction
	TxNumber            uint64
	FromAddress         types.EthAddress // address of the sender
	ExpectedBlockHeight uint64           //  the max expected block number for the transaction to be executed.
	StateRootHash       [32]byte         // state-root-hash on which the invalidity is based
	FtxRollingHash      [32]byte         // the rolling hash of the forced transaction
}

// Sum compute the mimc hash over the functional public inputs
func (pi *Invalidity) Sum(hsh hash.Hash) []byte {
	if hsh == nil {
		hsh = mimc.NewMiMC()
	}

	hsh.Reset()
	hsh.Write(pi.TxHash[:16])
	hsh.Write(pi.TxHash[16:])
	types.WriteInt64On32Bytes(hsh, int64(pi.TxNumber))
	hsh.Write(pi.FromAddress[:])
	types.WriteInt64On32Bytes(hsh, int64(pi.ExpectedBlockHeight))
	hsh.Write(pi.StateRootHash[:])
	hsh.Write(pi.FtxRollingHash[:])

	return hsh.Sum(nil)
}

func (pi *Invalidity) SumAsField() field.Element {

	var (
		sumBytes = pi.Sum(nil)
		sum      = new(field.Element).SetBytes(sumBytes)
	)

	return *sum
}

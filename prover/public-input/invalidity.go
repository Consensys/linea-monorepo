package public_input

import (
	"hash"

	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
)

type Invalidity struct {
	TxHash               common.Hash
	FromAddress          types.EthAddress
	BlockHeight          uint64
	InitialStateRootHash [32]byte
	TimeStamp            uint64
}

func (pi *Invalidity) Sum(hsh hash.Hash) []byte {
	if hsh == nil {
		hsh = mimc.NewMiMC()
	}

	hsh.Reset()
	hsh.Write(pi.TxHash[:16])
	hsh.Write(pi.TxHash[16:])
	hsh.Write(pi.FromAddress[:])
	writeNum(hsh, pi.BlockHeight)
	hsh.Write(pi.InitialStateRootHash[:])
	writeNum(hsh, pi.TimeStamp)

	return hsh.Sum(nil)

}

func (pi *Invalidity) SumAsField() field.Element {

	var (
		sumBytes = pi.Sum(nil)
		sum      = new(field.Element).SetBytes(sumBytes)
	)

	return *sum
}

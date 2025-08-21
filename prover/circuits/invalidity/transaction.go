package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const MaxNbKeccakF = 32 // for hashing the transaction

type TxPayloadGnark struct {
	ChainID    frontend.Variable
	Nonce      frontend.Variable
	GasTipCap  frontend.Variable // a.k.a. maxPriorityFeePerGas
	GasFeeCap  frontend.Variable // a.k.a. maxFeePerGas
	Gas        frontend.Variable
	To         frontend.Variable `rlp:"nil"` // nil means contract creation
	Value      frontend.Variable
	Data       []frontend.Variable
	AccessList AccessListGnark
}

// AccessList is an EIP-2930 access list.
type AccessListGnark []AccessTupleGnark

// AccessTuple is the element type of an access list.
type AccessTupleGnark struct {
	Address     frontend.Variable   `json:"address"     gencodec:"required"`
	StorageKeys []frontend.Variable `json:"storageKeys" gencodec:"required"`
}

func (p *TxPayloadGnark) Sum(api frontend.API, hash keccak.BlockHasher) frontend.Variable {
	sum := hash.Sum(nil,
		utils.ToBytes(api, p.ChainID),
		utils.ToBytes(api, p.Nonce),
		utils.ToBytes(api, p.GasTipCap),
		utils.ToBytes(api, p.GasFeeCap),
		utils.ToBytes(api, p.Gas),
		utils.ToBytes(api, p.To),
		utils.ToBytes(api, p.Value),
		utils.ToBytes(api, p.Data),
		utils.ToBytes(api, p.AccessList),
	)
	return sum
}

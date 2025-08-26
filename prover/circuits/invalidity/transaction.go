package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
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

func Sum(api frontend.API, hash keccak.BlockHasher, b []frontend.Variable) frontend.Variable {
	var (
		v    [][32]frontend.Variable
		temp [32]frontend.Variable
	)
	for i := range b {
		copy(temp[:], b[i*32:i*32+32])
		v = append(v, temp)
	}
	sum := hash.Sum(nil, v...)
	return sum
}

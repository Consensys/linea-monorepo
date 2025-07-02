package public_input

import (
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

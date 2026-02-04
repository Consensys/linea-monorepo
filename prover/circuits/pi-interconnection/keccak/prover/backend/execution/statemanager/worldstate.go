package statemanager

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
)

const WS_LOCATION = "0x"

var MIMC_CONFIG = &smt.Config{
	HashFunc: hashtypes.MiMC,
	Depth:    40,
}

// Legacy keccak code hash of an empty account
// 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
var LEGACY_KECCAK_EMPTY_CODEHASH = common.FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")

type WorldState struct {
	AccountTrie  *AccountTrie
	StorageTries collection.Mapping[Address, *StorageTrie]
	Config       *smt.Config
}

// Returns an empty code hash : the hash of an empty bytes32
func EmptyCodeHash(config *smt.Config) Digest {
	hasher := config.HashFunc()
	hasher.Write(make([]byte, 32))
	return types.AsBytes32(hasher.Sum(nil))
}

func NewStorageTrie(config *smt.Config, address Address) *StorageTrie {
	return accumulator.InitializeProverState[FullBytes32, FullBytes32](config, address.Hex())
}

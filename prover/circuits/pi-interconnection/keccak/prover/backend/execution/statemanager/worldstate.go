package statemanager

import (
	"math/big"

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

func NewWorldState(config *smt.Config) *WorldState {
	return &WorldState{
		AccountTrie:  accumulator.InitializeProverState[Address, types.Account](config, WS_LOCATION),
		StorageTries: collection.NewMapping[Address, *accumulator.ProverState[FullBytes32, FullBytes32]](),
		Config:       config,
	}
}

// Returns an empty code hash : the hash of an empty bytes32
func EmptyCodeHash(config *smt.Config) Digest {
	hasher := config.HashFunc()
	hasher.Write(make([]byte, 32))
	return types.AsBytes32(hasher.Sum(nil))
}

// Returns an EOA account
func NewEOA(config *smt.Config, nonce int64, balance *big.Int) Account {
	return types.Account{
		Nonce:          nonce,
		Balance:        balance,
		StorageRoot:    EmptyStorageTrieHash(config), // The eth
		MimcCodeHash:   EmptyCodeHash(config),
		KeccakCodeHash: types.AsFullBytes32(LEGACY_KECCAK_EMPTY_CODEHASH),
		CodeSize:       0,
	}
}

// Returns an empty storage contract
func NewContractEmptyStorage(
	config *smt.Config,
	nonce int64,
	balance *big.Int,
	codeHash Digest,
	keccakCodeHash FullBytes32,
	codeSize int64,
) types.Account {
	return types.Account{
		Nonce:          nonce,
		Balance:        balance,
		StorageRoot:    EmptyStorageTrieHash(config),
		MimcCodeHash:   codeHash,
		KeccakCodeHash: keccakCodeHash,
		CodeSize:       codeSize,
	}
}

var MIMC_EMPTY_STORAGE = EmptyStorageTrieHash(MIMC_CONFIG)

func NewAccountTrie(config *smt.Config) *AccountTrie {
	return accumulator.InitializeProverState[Address, Account](config, WS_LOCATION)
}

func NewStorageTrie(config *smt.Config, address Address) *StorageTrie {
	return accumulator.InitializeProverState[FullBytes32, FullBytes32](config, address.Hex())
}

func EmptyStorageTrieHash(config *smt.Config) Digest {
	// the EthAddress does not contribute to the hash so
	// it is fine to send an empty one.
	trie := NewStorageTrie(config, Address{})
	return trie.TopRoot()
}

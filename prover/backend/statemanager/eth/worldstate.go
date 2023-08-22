package eth

import (
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/accumulator"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
)

const WS_LOCATION = "0x"

var MIMC_CONFIG = &smt.Config{
	HashFunc: hashtypes.MiMC,
	Depth:    40,
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

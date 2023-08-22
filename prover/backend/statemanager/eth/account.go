package eth

import (
	"io"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/ethereum/go-ethereum/common"
)

// Legacy keccak code hash of an empty account
// 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
var LEGACY_KECCAK_EMPTY_CODEHASH = common.FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")

// An Ethereum account represented with the zkTrie representation
type Account struct {
	// Implicitly we assume that the nonce fits into
	// a int64 but the obtained encoding is consistent
	// with how it would be if it was a big.Int
	Nonce          int64
	Balance        *big.Int
	StorageRoot    Digest
	CodeHash       Digest
	KeccakCodeHash FullBytes32
	CodeSize       int64
}

// Returns an empty code hash : the hash of an empty bytes32
func EmptyCodeHash(config *smt.Config) Digest {
	hasher := config.HashFunc()
	hasher.Write(make([]byte, 32))
	return hashtypes.BytesToDigest(hasher.Sum(nil))
}

// Write the account into a hasher
func (a Account) WriteTo(w io.Writer) (int64, error) {
	n0, _ := hashtypes.WriteInt64To(w, a.Nonce)
	n1, _ := hashtypes.WriteBigIntTo(w, a.Balance)
	n2, _ := a.StorageRoot.WriteTo(w)
	n3, _ := a.CodeHash.WriteTo(w)
	n4, _ := a.KeccakCodeHash.WriteTo(w)
	n5, _ := hashtypes.WriteInt64To(w, a.CodeSize)
	return n0 + n1 + n2 + n3 + n4 + n5, nil
}

// Returns an EOA account
func NewEOA(config *smt.Config, nonce int64, balance *big.Int) Account {
	return Account{
		Nonce:          nonce,
		Balance:        balance,
		StorageRoot:    EmptyStorageTrieHash(config), // The eth
		CodeHash:       EmptyCodeHash(config),
		KeccakCodeHash: FullBytes32(hashtypes.BytesToDigest(LEGACY_KECCAK_EMPTY_CODEHASH)),
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
) Account {
	return Account{
		Nonce:          nonce,
		Balance:        balance,
		StorageRoot:    EmptyStorageTrieHash(config),
		CodeHash:       codeHash,
		KeccakCodeHash: keccakCodeHash,
		CodeSize:       codeSize,
	}
}

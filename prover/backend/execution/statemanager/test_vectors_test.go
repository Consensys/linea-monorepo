package statemanager_test

import (
	"io"
	"math/big"
	"testing"

	eth "github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/assert"
)

// Construct a dummy address from an integer
func DummyAddress(i int) (a eth.Address) {
	a[0] = byte(i)
	return a
}

// Constructs a dummy full bytes from an integer
func DummyFullByte(i int) (f eth.FullBytes32) {
	f[0] = byte(i)
	return f
}

// Returns a dummy digest
func DummyDigest(i int) (d eth.Digest) {
	return types.DummyKoalaOctuplet(i)
}

func Hash(t io.WriterTo) types.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	t.WriteTo(hasher)
	return types.MustBytesToKoalaOctuplet(hasher.Sum(nil))
}

/*
Hash of an account with zero in every fields
*/
func TestHashZeroAccountMiMC(t *testing.T) {
	// Test the hash of an emptyAccount with zeroes everywhere
	emptyAccount := types.Account{}
	emptyAccount.Balance = big.NewInt(0)

	assert.Equal(
		t,
		"0x0f170eaef9275fd6098a06790c63a141e206e0520738a4cf5cf5081d495e8682",
		Hash(emptyAccount).Hex(),
	)
}

/*
- Test the empty storage trie hash
*/
func TestEmptyStorageTrieHash(t *testing.T) {
	assert.Equal(t, "0x07977874126658098c066972282d4c85f230520af3847e297fe7524f976873e5", eth.ZKHASH_EMPTY_STORAGE.Hex())
}

/*
- Root hash of the empty world state
*/
func TestEmptyWorldStateMiMC(t *testing.T) {
	worldstate := eth.NewWorldState()
	// should be the top root of an empty accumulator
	assert.Equal(t, "0x07977874126658098c066972282d4c85f230520af3847e297fe7524f976873e5", worldstate.AccountTrie.TopRoot().Hex())
}

/*
  - Insert EOA A
    -> check root hash
*/
func TestWorldStateWithAnAccountMiMC(t *testing.T) {

	account := eth.NewEOA(65, big.NewInt(835))
	// This gives a non-zero dummy address to the account
	address := DummyAddress(36)

	worldstate := eth.NewWorldState()
	worldstate.AccountTrie.InsertAndProve(address, account)

	// check that the hash of the inserted account matches
	assert.Equal(t, "0x11314cf80cdd63a376e468ea9e6c672109bcfe516f0349382df82e1a876ca8b2", Hash(account).Hex(), "inserted EOA hash")
	// check that the hash of the address matches
	assert.Equal(t, "0x0b9887ed089160e457c4078941214f313dacfb71a8ed1818da3468ef1fdbe282", Hash(address).Hex(), "hash of address")

	// finally check the root hash after insertion
	assert.Equal(t, "0x04c3c6de7195a187bc89fb4f8b68e93c7d675f1eed585b00d0e1e6241a321f86", worldstate.AccountTrie.TopRoot().Hex(), "root hash after insertion")
}

/*
  - Insert EOA A
  - Insert EOA B
    -> check root hash
*/
func TestWorldStateWithTwoAccountMiMC(t *testing.T) {

	accountA := eth.NewEOA(65, big.NewInt(835))
	// This gives a non-zero dummy address to the account
	addressA := DummyAddress(36)

	accountB := eth.NewEOA(42, big.NewInt(354))
	addressB := DummyAddress(41) // must be a different address or the insert will panic

	worldstate := eth.NewWorldState()
	worldstate.AccountTrie.InsertAndProve(addressA, accountA)
	worldstate.AccountTrie.InsertAndProve(addressB, accountB)

	// finally check the root hash after insertion
	assert.Equal(t, "0x020e2a836e973eebd3c6367ef432ff21bb35102bc2ae3258b385e8cfbf4d46d4", worldstate.AccountTrie.TopRoot().Hex())
}

/*
- Insert EOA A
- Insert contract B (empty storage) -> root hash
- Insert slot in storage of B -> root hash
- Delete A -> root hash
- Delete the slot in the storage of B -> root hash
- Rewrite again in the storage of B -> root hash
*/
func TestWorldStateEoaAndContractMiMC(t *testing.T) {

	// dummy EOA
	accountA := eth.NewEOA(65, big.NewInt(835))
	addressA := DummyAddress(36)

	// dummy Account
	accountB := eth.NewContractEmptyStorage(41, big.NewInt(15353), DummyDigest(75), DummyFullByte(15), 7)
	addressB := DummyAddress(47)

	worldstate := eth.NewWorldState()
	worldstate.AccountTrie.InsertAndProve(addressA, accountA)
	worldstate.AccountTrie.InsertAndProve(addressB, accountB)

	// Give a storage trie to B
	worldstate.StorageTries.InsertNew(addressB, eth.NewStorageTrie(addressB))

	// The storage tries map stores pointers to the storage tries so we don't need to
	// update the map after modifying the storage however, we still need to manually
	// set the new storage root hash on the account
	storageB := worldstate.StorageTries.MustGet(addressB)

	// finally check the root hash after insertion
	{
		// Cgeck the root hash after inserting A and B
		assert.Equal(t, "0x0bc47df364adaecf61a5024f2b39603341077be453d88d21e627aee59ef7a6db", worldstate.AccountTrie.TopRoot().Hex(), "after insertion of B")
	}

	// Write something in the storage of B
	{
		storageB.InsertAndProve(DummyFullByte(14), DummyFullByte(18))
		newAccountB := accountB
		newAccountB.StorageRoot = storageB.TopRoot()
		worldstate.AccountTrie.UpdateAndProve(addressB, newAccountB)

		// finally check the root hash after insertion
		assert.Equal(t, "0x069b45f6f789581a3402103cd35168bf8d1de77eb5db9f79390ad29472e0846d", worldstate.AccountTrie.TopRoot().Hex(), "after inserting a slot in the storage of B")
	}

	// Delete the account A
	{
		worldstate.AccountTrie.DeleteAndProve(addressA)
		assert.Equal(t, "0x0b6a3290b85cf230ce33cec4438aa907373c8a471c47346ad32fc170b8644ec3", worldstate.AccountTrie.TopRoot().Hex())
	}

	// Remove what we wrote into B
	{
		storageB.DeleteAndProve(DummyFullByte(14))
		newAccountB := accountB
		newAccountB.StorageRoot = storageB.TopRoot()
		worldstate.AccountTrie.UpdateAndProve(addressB, newAccountB)
		assert.Equal(t, "0x0f11405ba708b9aeb8de0a341d80682b3a59c628e0694af97e357e86bb9567cf", worldstate.AccountTrie.TopRoot().Hex())
	}

	// Write again, somewhere else
	{
		storageB.InsertAndProve(DummyFullByte(11), DummyFullByte(78))
		newAccountB := accountB
		newAccountB.StorageRoot = storageB.TopRoot()
		worldstate.AccountTrie.UpdateAndProve(addressB, newAccountB)
		assert.Equal(t, "0x06825644ff9ddf7d87b8a6f5d813254d535eae9d1bc2d2336b27211b1006f58c", worldstate.AccountTrie.TopRoot().Hex())
	}
}

/*
insert EOA A
insert contract B
delete A
insert C
*/
func TestAddAaddBdelAaddCMiMC(t *testing.T) {

	// dummy EOA
	accountA := eth.NewEOA(65, big.NewInt(835))
	addressA := DummyAddress(36)

	// dummy Account
	accountB := eth.NewContractEmptyStorage(41, big.NewInt(15353), DummyDigest(75), DummyFullByte(15), 7)
	addressB := DummyAddress(47)

	// dummy Contract
	accountC := eth.NewContractEmptyStorage(48, big.NewInt(9835), DummyDigest(54), DummyFullByte(85), 19)
	addressC := DummyAddress(120)

	worldstate := eth.NewWorldState()
	worldstate.AccountTrie.InsertAndProve(addressA, accountA)
	worldstate.AccountTrie.InsertAndProve(addressB, accountB)
	worldstate.AccountTrie.DeleteAndProve(addressA)
	worldstate.AccountTrie.InsertAndProve(addressC, accountC)

	assert.Equal(t, "0x00b43fd65348b5a492ebcbd7ce3933fc963809ca4897d4fcd00d8661e45d9d55", worldstate.AccountTrie.TopRoot().Hex())
}

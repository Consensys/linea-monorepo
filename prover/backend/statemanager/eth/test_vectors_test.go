package eth_test

import (
	"io"
	"math/big"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/accumulator"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/stretchr/testify/require"
)

type WorldState struct {
	AccountTrie  *eth.AccountTrie
	StorageTries collection.Mapping[eth.Address, *eth.StorageTrie]
	Config       *smt.Config
}

func NewWorldState(config *smt.Config) *WorldState {
	return &WorldState{
		AccountTrie:  accumulator.InitializeProverState[eth.Address, eth.Account](config, eth.WS_LOCATION),
		StorageTries: collection.NewMapping[eth.Address, *accumulator.ProverState[eth.FullBytes32, eth.FullBytes32]](),
		Config:       config,
	}
}

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

func Hash(t io.WriterTo) hashtypes.Digest {
	hasher := eth.MIMC_CONFIG.HashFunc()
	t.WriteTo(hasher)
	return hashtypes.BytesToDigest(hasher.Sum(nil))
}

/*
Hash of an account with zero in every fields
*/
func TestHashZeroAccountMiMC(t *testing.T) {
	// Test the hash of an emptyAccount with zeroes everywhere
	emptyAccount := eth.Account{}
	emptyAccount.Balance = big.NewInt(0)

	require.Equal(
		t,
		"0x19be98b429f6e00b8eff84a8aa617d2982421d5cde049c3e2a9b5a30a554a307",
		Hash(emptyAccount).Hex(),
	)
}

/*
- Test the empty storage trie hash
*/
func TestEmptyStorageTrieHash(t *testing.T) {
	require.Equal(t, "0x2e7942bb21022172cbad3ffc38d1c59e998f1ab6ab52feb15345d04bbf859f14", eth.MIMC_EMPTY_STORAGE.Hex())
}

/*
- Root hash of the empty world state
*/
func TestEmptyWorldStateMiMC(t *testing.T) {
	worldstate := NewWorldState(eth.MIMC_CONFIG)
	// should be the top root of an empty accumulator
	require.Equal(t, "0x2e7942bb21022172cbad3ffc38d1c59e998f1ab6ab52feb15345d04bbf859f14", worldstate.AccountTrie.TopRoot().Hex())
}

/*
  - Insert EOA A
    -> check root hash
*/
func TestWorldStateWithAnAccountMiMC(t *testing.T) {

	account := eth.NewEOA(eth.MIMC_CONFIG, 65, big.NewInt(835))
	// This gives a non-zero dummy address to the account
	address := DummyAddress(36)

	worldstate := NewWorldState(eth.MIMC_CONFIG)
	worldstate.AccountTrie.InsertAndProve(address, account)

	// check that the hash of the inserted account matches
	require.Equal(t, "0x25ddd6106526ffb2c9b923617cf3bcab669a5d57821d0ec81daa23155c1513ea", Hash(account).Hex())
	// check that the hash of the address matches
	require.Equal(t, "0x198b65f4680ea1536bedf00813a9f6a55c6e6d923cf2c88ba8791698fbc65f8d", Hash(address).Hex())

	// finally check the root hash after insertion
	require.Equal(t, "0x11aed727a707f2f1962e399bd4787153ba0e69b7224e8eecf4d1e4e6a8e8dafd", worldstate.AccountTrie.TopRoot().Hex())
}

/*
  - Insert EOA A
  - Insert EOA B
    -> check root hash
*/
func TestWorldStateWithTwoAccountMiMC(t *testing.T) {

	accountA := eth.NewEOA(eth.MIMC_CONFIG, 65, big.NewInt(835))
	// This gives a non-zero dummy address to the account
	addressA := DummyAddress(36)

	accountB := eth.NewEOA(eth.MIMC_CONFIG, 42, big.NewInt(354))
	addressB := DummyAddress(41) // must be a different address or the insert will panic

	worldstate := NewWorldState(eth.MIMC_CONFIG)
	worldstate.AccountTrie.InsertAndProve(addressA, accountA)
	worldstate.AccountTrie.InsertAndProve(addressB, accountB)

	// finally check the root hash after insertion
	require.Equal(t, "0x1f8b53f5cf08c25611e11f8bfd2ffdbdab2f12f7e0578e54282f31e0e6267ab4", worldstate.AccountTrie.TopRoot().Hex())
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
	accountA := eth.NewEOA(eth.MIMC_CONFIG, 65, big.NewInt(835))
	addressA := DummyAddress(36)

	// dummy Account
	accountB := eth.NewContractEmptyStorage(eth.MIMC_CONFIG, 41, big.NewInt(15353), hashtypes.DummyDigest(75), DummyFullByte(15), 7)
	addressB := DummyAddress(47)

	worldstate := NewWorldState(eth.MIMC_CONFIG)
	worldstate.AccountTrie.InsertAndProve(addressA, accountA)
	worldstate.AccountTrie.InsertAndProve(addressB, accountB)

	// Give a storage trie to B
	worldstate.StorageTries.InsertNew(addressB, eth.NewStorageTrie(eth.MIMC_CONFIG, addressB))

	// The storage tries map stores pointers to the storage tries so we don't need to
	// update the map after modifying the storage however, we still need to manually
	// set the new storage root hash on the account
	storageB := worldstate.StorageTries.MustGet(addressB)

	// finally check the root hash after insertion
	{
		// Cgeck the root hash after inserting A and B
		require.Equal(t, "0x15471b9c6443332dccaef5b1544c5881e2c2a6e4576ad1696cec3d1769061e21", worldstate.AccountTrie.TopRoot().Hex(), "after insertion of B")
	}

	// Write something in the storage of B
	{
		storageB.InsertAndProve(DummyFullByte(14), DummyFullByte(18))
		newAccountB := accountB
		newAccountB.StorageRoot = storageB.TopRoot()
		worldstate.AccountTrie.UpdateAndProve(addressB, newAccountB)

		// finally check the root hash after insertion
		require.Equal(t, "0x16df792bd76d708e98bccd2f037d15e5b8d5fa816febf0bf0a30487b8e0ba117", worldstate.AccountTrie.TopRoot().Hex(), "after inserting a slot in the storage of B")
	}

	// Delete the account A
	{
		worldstate.AccountTrie.DeleteAndProve(addressA)
		require.Equal(t, "0x2e603c5f62481d627428d9efbfccd33fc1474e1d191b9e93cefa337b4a0e67da", worldstate.AccountTrie.TopRoot().Hex())
	}

	// Remove what we wrote into B
	{
		storageB.DeleteAndProve(DummyFullByte(14))
		newAccountB := accountB
		newAccountB.StorageRoot = storageB.TopRoot()
		worldstate.AccountTrie.UpdateAndProve(addressB, newAccountB)
		require.Equal(t, "0x1aee37cbf805a51f827b48eb8fab44a7012575876045dca6ea6faaaa2233b0b5", worldstate.AccountTrie.TopRoot().Hex())
	}

	// Write again, somewhere else
	{
		storageB.InsertAndProve(DummyFullByte(11), DummyFullByte(78))
		newAccountB := accountB
		newAccountB.StorageRoot = storageB.TopRoot()
		worldstate.AccountTrie.UpdateAndProve(addressB, newAccountB)
		require.Equal(t, "0x02b9fb86c95b0e45a3ad401f1267b62f80e1ec16057d1491c2c9b32b36a1478f", worldstate.AccountTrie.TopRoot().Hex())
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
	accountA := eth.NewEOA(eth.MIMC_CONFIG, 65, big.NewInt(835))
	addressA := DummyAddress(36)

	// dummy Account
	accountB := eth.NewContractEmptyStorage(eth.MIMC_CONFIG, 41, big.NewInt(15353), hashtypes.DummyDigest(75), DummyFullByte(15), 7)
	addressB := DummyAddress(47)

	// dummy Contract
	accountC := eth.NewContractEmptyStorage(eth.MIMC_CONFIG, 48, big.NewInt(9835), hashtypes.DummyDigest(54), DummyFullByte(85), 19)
	addressC := DummyAddress(120)

	worldstate := NewWorldState(eth.MIMC_CONFIG)
	worldstate.AccountTrie.InsertAndProve(addressA, accountA)
	worldstate.AccountTrie.InsertAndProve(addressB, accountB)
	worldstate.AccountTrie.DeleteAndProve(addressA)
	worldstate.AccountTrie.InsertAndProve(addressC, accountC)

	require.Equal(t, "0x1cb213eb41f295fded1c6850d570beec729ca15541a33586320e0f097f0ed11b", worldstate.AccountTrie.TopRoot().Hex())
}

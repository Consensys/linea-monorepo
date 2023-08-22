package accumulator_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/accumulator"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/stretchr/testify/require"
)

func newTestAccumulatorMiMC() *accumulator.ProverState[DummyKey, DummyVal] {
	config := &smt.Config{
		HashFunc: hashtypes.MiMC,
		Depth:    40,
	}
	return accumulator.InitializeProverState[DummyKey, DummyVal](config, LOCATION_TESTING)
}

// Test the root hash of an empty accumulator
func TestEmptyAccumulatorMiMC(t *testing.T) {

	acc := newTestAccumulatorMiMC()

	// next free node
	require.Equal(t, int64(2), acc.NextFreeNode)

	// Check the value of the first leaf, it should be "head" with an updated next value
	{
		leafOpening := acc.Data.MustGet(0).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 0, Next: 1, HKey: 0x0000000000000000000000000000000000000000000000000000000000000000, HVal: 0x0000000000000000000000000000000000000000000000000000000000000000}",
			leafOpening.String(),
		)

		// also checks its hash value (i.e, the corresponding leaf in the tree)
		require.Equal(t, "0x276935e06bee60ac996e056c4917ae55afb4d43efd636447f52baf8d174db1f9", leafOpening.Hash(acc.Config()).Hex())
	}

	// check the value of the second leaf opening. It should be tail with an updated prev value
	{
		leafOpening := acc.Data.MustGet(1).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 0, Next: 1, HKey: 0x30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000000, HVal: 0x0000000000000000000000000000000000000000000000000000000000000000}",
			leafOpening.String(),
		)

		// also check its hash value
		require.Equal(t, "0x103cca8163b021373b6f24e3098a091b40ada3564f9b5d40a4f8185c61e49c60", leafOpening.Hash(acc.Config()).Hex())
	}

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x21bf7e28464bf26302be46623b706eacf89b08134665b8f425d437f13218091b",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x2e7942bb21022172cbad3ffc38d1c59e998f1ab6ab52feb15345d04bbf859f14",
		acc.TopRoot().Hex(),
	)
}

// Test the root after inserting a new entry
func TestInsertionRootHashMiMC(t *testing.T) {

	acc := newTestAccumulatorMiMC()

	key := dumkey(58) // "KEY_58"
	val := dumval(42) // "VAL_42"

	_ = acc.InsertAndProve(key, val)

	// We inserted, so the next free node should have been increased
	require.Equal(t, int64(3), acc.NextFreeNode)

	// Check the value of the first leaf, it should be "head" with an updated next value
	{
		leafOpening := acc.Data.MustGet(0).LeafOpening
		require.Equal(t, int64(0), leafOpening.Prev)
		require.Equal(t, int64(2), leafOpening.Next)
		require.Equal(
			t,
			"LeafOpening{Prev: 0, Next: 2, HKey: 0x0000000000000000000000000000000000000000000000000000000000000000, HVal: 0x0000000000000000000000000000000000000000000000000000000000000000}",
			leafOpening.String(),
		)

		// also checks its hash value (i.e, the corresponding leaf in the tree)
		leaf := leafOpening.Hash(acc.Config())
		require.Equal(t, "0x0b070604db69fe26e6fff2c547a04321e96f99557a54427111d344a7f9c4fe21", leaf.Hex())
	}

	// check the value of the second leaf opening. It should be tail with an updated prev value
	{
		leafOpening := acc.Data.MustGet(1).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 2, Next: 1, HKey: 0x30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000000, HVal: 0x0000000000000000000000000000000000000000000000000000000000000000}",
			leafOpening.String(),
		)

		// also check its hash value
		leaf := leafOpening.Hash(acc.Config())
		require.Equal(t, "0x2d87eaae6ee164a80a573baac4c0c747296fe504c130a06a78f63c488ca43cdf", leaf.Hex())
	}

	// check the value of the second leaf opening, corresponds to the inserted entry
	{
		leafOpening := acc.Data.MustGet(2).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 0, Next: 1, HKey: 0x2114af98d4b60e71fe3df51a0eb9e6b1f0625267d0788ec1e38988deb93f54a5, HVal: 0x15cc289ebc18cb3ba9301f46f0619391ee79007ea289fd3d9155d574f121e953}",
			leafOpening.String(),
		)

		// also check its hash value
		leaf := leafOpening.Hash(acc.Config())
		require.Equal(t, "0x065dc977bfb75fc5fb5f09925e42c1d13900b39f7e6b965d6e484595cb1c93a0", leaf.Hex())
	}

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x1b8c9d905e0561b9ae850cf04ba39ff166f7b7306ea348b87f3f682a45cc82c1",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x186cb83c254c5f6985f7929e02cfd5f19e6c953c4f01bc2af9e0276f75184d6b",
		acc.TopRoot().Hex(),
	)
}

// Test the root after inserting a new entry
func TestInsertAndUpdateRootHashMiMC(t *testing.T) {

	acc := newTestAccumulatorMiMC()

	key := dumkey(58)    // "KEY_58"
	val := dumval(41)    // "VAL_41"
	newval := dumval(42) // "VAL_42"

	_ = acc.InsertAndProve(key, val)
	_ = acc.UpdateAndProve(key, newval)

	// Note : the tree should be in exactly the same state as after directly
	// inserting 42

	// We inserted, so the next free node should have been increased
	require.Equal(t, int64(3), acc.NextFreeNode)

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x1b8c9d905e0561b9ae850cf04ba39ff166f7b7306ea348b87f3f682a45cc82c1",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x186cb83c254c5f6985f7929e02cfd5f19e6c953c4f01bc2af9e0276f75184d6b",
		acc.TopRoot().Hex(),
	)
}

// Test the root after inserting a new entry and deleting right away
func TestInsertAndDeleteRootHashMiMC(t *testing.T) {

	acc := newTestAccumulatorMiMC()

	key := dumkey(58) // "KEY_58"
	val := dumval(41) // "VAL_41"

	_ = acc.InsertAndProve(key, val)
	_ = acc.DeleteAndProve(key)

	// Note : the tree should be in exactly the same state as after directly
	// inserting 42

	// We inserted, so the next free node should have been increased
	require.Equal(t, int64(3), acc.NextFreeNode)

	// root of the subtree (e.g. exluding the next free node). It equals the one
	// of the empty tree
	require.Equal(
		t,
		"0x21bf7e28464bf26302be46623b706eacf89b08134665b8f425d437f13218091b",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node). It does not
	// equal the one of the empty subtree because we have bumped next free node when
	// inserting.
	require.Equal(
		t,
		"0x2844b523cac49f6c5205b7b44065a741fd5c4ba8481c0ebee1d06ecf33f5ef40",
		acc.TopRoot().Hex(),
	)
}

//go:build !fuzzlight

package accumulator_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/stretchr/testify/require"
)

func newTestAccumulatorMiMC() *accumulator.ProverState[DummyKey, DummyVal] {
	config := &smt.Config{
		HashFunc: hashtypes.MiMC,
		Depth:    40,
	}
	return accumulator.InitializeProverState[DummyKey, DummyVal](config, locationTesting)
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
		require.Equal(t, "0x0891fa77c3d0c9b745840d71d41dcb58b638d4734bb4f0bba4a3d1a2d847b672", leafOpening.Hash(acc.Config()).Hex())
	}

	// check the value of the second leaf opening. It should be tail with an updated prev value
	{
		leafOpening := acc.Data.MustGet(1).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 0, Next: 1, HKey: 0x12ab655e9a2ca55660b44d1e5c37b00159aa76fed00000010a11800000000000, HVal: 0x0000000000000000000000000000000000000000000000000000000000000000}",
			leafOpening.String(),
		)

		// also check its hash value
		require.Equal(t, "0x10ba2286f648a549b50ea5f1b6e1155d22c31eb4727c241e76c420200cd5dbe0", leafOpening.Hash(acc.Config()).Hex())
	}

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x0951bfcd4ac808d195af8247140b906a4379b3f2d37ec66e34d2f4a5d35fa166",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x07977874126658098c066972282d4c85f230520af3847e297fe7524f976873e5",
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
		require.Equal(t, "0x09dd19003c62ec025c50e9eacb4fd9509bc5a150b43bce9240162eb4b1778e4b", leaf.Hex())
	}

	// check the value of the second leaf opening. It should be tail with an updated prev value
	{
		leafOpening := acc.Data.MustGet(1).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 2, Next: 1, HKey: 0x12ab655e9a2ca55660b44d1e5c37b00159aa76fed00000010a11800000000000, HVal: 0x0000000000000000000000000000000000000000000000000000000000000000}",
			leafOpening.String(),
		)

		// also check its hash value
		leaf := leafOpening.Hash(acc.Config())
		require.Equal(t, "0x0ab3e0de05fd2b67ffa26aadd17e91b632a26013784e757028ae9d851508b055", leaf.Hex())
	}

	// check the value of the second leaf opening, corresponds to the inserted entry
	{
		leafOpening := acc.Data.MustGet(2).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 0, Next: 1, HKey: 0x096ccb5dfe8d38d9e11f9f3ed692d0b377f4f12cb204d15338fe9862a7997743, HVal: 0x05d7b878349fc53b779b4b7e3a303694e88fba9521b6add1ed824a3be74104fd}",
			leafOpening.String(),
		)

		// also check its hash value
		leaf := leafOpening.Hash(acc.Config())
		require.Equal(t, "0x0002a72d394f16c6a063ca6b15a8bf36c6efbfc67831534707c39bee252400df", leaf.Hex())
	}

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x0882afe875656680dceb7b17fcba7c136cec0c32becbe9039546c79f71c56d36",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x0cfdc3990045390093be4e1cc9907b220324cccd1c8ea9ede980c7afa898ef8d",
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
		"0x0882afe875656680dceb7b17fcba7c136cec0c32becbe9039546c79f71c56d36",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x0cfdc3990045390093be4e1cc9907b220324cccd1c8ea9ede980c7afa898ef8d",
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
		"0x0951bfcd4ac808d195af8247140b906a4379b3f2d37ec66e34d2f4a5d35fa166",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node). It does not
	// equal the one of the empty subtree because we have bumped next free node when
	// inserting.
	require.Equal(
		t,
		"0x0bcb88342825fa7a079a5cf5f77d07b1590a140c311a35acd765080eea120329",
		acc.TopRoot().Hex(),
	)
}

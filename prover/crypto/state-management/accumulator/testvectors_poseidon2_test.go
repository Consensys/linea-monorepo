//go:build !fuzzlight

package accumulator_test

import (
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/require"
)

// Test the root hash of an empty accumulator
func TestEmptyAccumulatorPoseidon2(t *testing.T) {

	acc := newTestAccumulatorPoseidon2DummyVal()

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
		require.Equal(t, "0x7097ec8b2b4beecf1986c5c244fe5b7b54e0626741494ced60b68360719a80fc", leafOpening.Hash(acc.Config()).Hex())
	}

	// check the value of the second leaf opening. It should be tail with an updated prev value
	{
		leafOpening := acc.Data.MustGet(1).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 0, Next: 1, HKey: 0x7f0000007f0000007f0000007f0000007f0000007f0000007f0000007f000000, HVal: 0x0000000000000000000000000000000000000000000000000000000000000000}",
			leafOpening.String(),
		)

		// also check its hash value
		require.Equal(t, "0x088e8d326cd388b41347959041c11a9d6082eeb15868a9f62b1bc777123949fa", leafOpening.Hash(acc.Config()).Hex())
	}

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x16cb2617218f334532c8fc4b7a9689166ea7488d5d93457913a0b72216437033",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x5138a2ab3bdc003b3a3664ab02bf93c52c9d9c156dbf2079241e5e3319540985",
		acc.TopRoot().Hex(),
	)
}

// Test the root after inserting a new entry
func TestInsertionRootHashPoseidon2(t *testing.T) {

	acc := newTestAccumulatorPoseidon2DummyVal()

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
		require.Equal(t, "0x22e7a2ea33304d915fd93eac62b418ba7ca915131bea42827670d7f31a3b3445", leaf.Hex())
	}

	// check the value of the second leaf opening. It should be tail with an updated prev value
	{
		leafOpening := acc.Data.MustGet(1).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 2, Next: 1, HKey: 0x7f0000007f0000007f0000007f0000007f0000007f0000007f0000007f000000, HVal: 0x0000000000000000000000000000000000000000000000000000000000000000}",
			leafOpening.String(),
		)

		// also check its hash value
		leaf := leafOpening.Hash(acc.Config())
		require.Equal(t, "0x7172ab730766d9b6367938c031afbd13607d95796a60930945ab6b88434339b2", leaf.Hex())
	}

	// check the value of the third leaf opening, corresponds to the inserted entry
	{
		leafOpening := acc.Data.MustGet(2).LeafOpening
		require.Equal(
			t,
			"LeafOpening{Prev: 0, Next: 1, HKey: 0x27b48fe33995c55d329a805f2c4acdc05c1229592579fcab0a46d23239f19498, HVal: 0x43b6adef3fc8455b573aadd56e3ab14a11084a9d667c2c8449b707c31cb32560}",
			leafOpening.String(),
		)

		// also check its hash value
		leaf := leafOpening.Hash(acc.Config())
		require.Equal(t, "0x42bee0052f812f13511ced262f95e62d5e572dc971f68fa8759914a441332ff1", leaf.Hex())
	}

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x50b4bb9d2b4c917f48ca9465613d7efe092934a95bcdc7a004ef3a6b7900fd5b",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x7a5cf31710a49b59489deb53505a3d5458625ef01b33914d0d784380464177ad",
		acc.TopRoot().Hex(),
	)
}

// Test the root after inserting a new entry, and then updateting it
func TestInsertAndUpdateRootHashPoseidon2(t *testing.T) {

	acc := newTestAccumulatorPoseidon2DummyVal()

	key := dumkey(58)    // "KEY_58"
	val := dumval(41)    // "VAL_41"
	newval := dumval(42) // "VAL_42"

	_ = acc.InsertAndProve(key, val)
	_ = acc.UpdateAndProve(key, newval)

	// Note : the tree should be in exactly the same state as after directly
	// inserting 42,
	// The same roots as in TestInsertionRootHashPoseidon2

	// We inserted, so the next free node should have been increased
	require.Equal(t, int64(3), acc.NextFreeNode)

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x50b4bb9d2b4c917f48ca9465613d7efe092934a95bcdc7a004ef3a6b7900fd5b",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x7a5cf31710a49b59489deb53505a3d5458625ef01b33914d0d784380464177ad",
		acc.TopRoot().Hex(),
	)
}

// Test the root after inserting a new entry and deleting right away
func TestInsertAndDeleteRootHashPoseidon2(t *testing.T) {

	acc := newTestAccumulatorPoseidon2DummyVal()

	key := dumkey(58) // "KEY_58"
	val := dumval(41) // "VAL_41"

	_ = acc.InsertAndProve(key, val)
	_ = acc.DeleteAndProve(key)

	// We inserted, so the next free node should have been increased
	require.Equal(t, int64(3), acc.NextFreeNode)

	// root of the subtree (e.g. exluding the next free node). It equals the one
	// of the empty tree
	// The same SubTreeRoot as in TestEmptyAccumulatorPoseidon2, but the TopRoot are different as the NextFreeNode are distinct
	require.Equal(
		t,
		"0x3a00a8e34a16f8a1225fee734816edb326f783bd6678d793345a28f046586ba6",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node). It does not
	// equal the one of the empty subtree because we have bumped next free node when
	// inserting.
	require.Equal(
		t,
		"0x391010e55fad441a1e3a33b11fb8271f68e86b6c1812ebd402d678d343fded62",
		acc.TopRoot().Hex(),
	)
}

func newTestAccumulatorPoseidon2() *accumulator.ProverState[types.EthAddress, types.Account] {
	config := &smt.Config{
		HashFunc: poseidon2.Poseidon2,
		Depth:    40,
	}
	return accumulator.InitializeProverState[types.EthAddress, types.Account](config, locationTesting)
}

// Test the root after inserting a new entry, without dummy key and value types
func TestRealKeyAndVal(t *testing.T) {

	acc := newTestAccumulatorPoseidon2()

	key, _ := types.AddressFromHex("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	account :=
		types.Account{
			Nonce:             65,
			Balance:           big.NewInt(5690),
			StorageRoot:       types.Bytes32FromHex("0x0b1dfeef3db4956540da8a5f785917ef1ba432e521368da60a0a1ce430425666"),
			Poseidon2CodeHash: types.Bytes32FromHex("0x729aac4455d43f2c69e53bb75f8430193332a4c32cafd9995312fa8346929e73"),
			KeccakCodeHash:    types.FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
			CodeSize:          0,
		}

	_ = acc.InsertAndProve(key, account)

	// We inserted, so the next free node should have been increased
	require.Equal(t, int64(3), acc.NextFreeNode)

	// root of the subtree (e.g. exluding the next free node)
	require.Equal(
		t,
		"0x06d0e7c249cff5d00868a56f2007b81b33a71efa265c5f97694e350923ce19d6",
		acc.SubTreeRoot().Hex(),
	)

	// root of the complete accumulator (e.g including the last node)
	require.Equal(
		t,
		"0x17ae4260246c749f10ef846b704305b35956df8f26efdc2b3634a0ec787ac1f3",
		acc.TopRoot().Hex(),
	)
}

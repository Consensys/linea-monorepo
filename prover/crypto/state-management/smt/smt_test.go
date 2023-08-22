package smt_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/stretchr/testify/require"
)

// Deterministically creates random digests
func RandDigest(pos int) smt.Digest {
	res := smt.Digest{}
	for i := range res {
		res[i] = byte(pos ^ (i + pos*156) ^ (pos + i*256) ^ (i * pos))
	}
	return res
}

// Creates a empty SMT, test its root hash in hard and tests that
// all the leaves are zero.
func TestTreeInitialization(t *testing.T) {

	config := &smt.Config{
		HashFunc: hashtypes.Keccak,
		Depth:    40,
	}

	tree := smt.NewEmptyTree(config)

	// Only contains empty leaves
	for pos := 0; pos < 1000; pos++ {
		x := tree.GetLeaf(pos)
		require.Equal(t, x, smt.EmptyLeaf())
	}
}

func TestTreeUpdateLeaf(t *testing.T) {
	config := &smt.Config{
		HashFunc: hashtypes.Keccak,
		Depth:    40,
	}

	tree := smt.NewEmptyTree(config)

	// Only contains empty leaves
	for pos := 0; pos < 1000; pos++ {
		// Make a valid digest
		newLeaf := RandDigest(pos)
		tree.Update(pos, newLeaf)
		recovered := tree.GetLeaf(pos)
		require.Equal(t, newLeaf, recovered)
	}
}

func TestMerkleProof(t *testing.T) {
	config := &smt.Config{
		HashFunc: hashtypes.Keccak,
		Depth:    40,
	}

	tree := smt.NewEmptyTree(config)

	// Only contains empty leaves
	for pos := 0; pos < 1000; pos++ {
		// Make a valid digest
		oldLeaf := tree.GetLeaf(pos)
		proof := tree.Prove(pos)

		// Directly verify the proof
		valid := proof.Verify(config, oldLeaf, tree.Root)
		require.Truef(t, valid, "pos #%v, proof #%v", pos, proof)
	}
}

func TestMerkleProofWithUpdate(t *testing.T) {
	config := &smt.Config{
		HashFunc: hashtypes.Keccak,
		Depth:    40,
	}

	tree := smt.NewEmptyTree(config)

	// Only contains empty leaves
	for pos := 0; pos < 1000; pos++ {
		proof := tree.Prove(pos)

		// Updat the leaf with a random-looking value before
		// checking the proof
		newLeaf := RandDigest(pos)
		tree.Update(pos, newLeaf)

		// After updating the old proof should still be valid
		// (because we only changed the current leaf)
		valid := proof.Verify(config, newLeaf, tree.Root)
		require.Truef(t, valid, "pos #%v", pos)
	}
}

func TestBuildFromScratch(t *testing.T) {

	config := &smt.Config{
		HashFunc: hashtypes.Keccak,
		Depth:    10,
	}

	// Generate random field elements and cast them into digestes
	leavesFr := vector.Rand(1 << config.Depth)
	leaves := make([]smt.Digest, len(leavesFr))
	for i := range leaves {
		leaves[i] = smt.Digest(leavesFr[i].Bytes())
	}

	// And generate the
	tree := smt.BuildComplete(leaves, config.HashFunc)

	// Test-Merkle tests the merkle proof point by point
	for i := range leaves {
		proof := tree.Prove(i)
		ok := proof.Verify(config, leaves[i], tree.Root)

		if !ok {
			t.Fatalf("failed to verify pos %v", i)
		}
	}

	// Build the same tree by adding the leaves one by one
	oneByoneTree := smt.NewEmptyTree(config)
	for i := range leaves {
		oneByoneTree.Update(i, leaves[i])
	}

	// We should obtain the same roots
	require.Equal(t, oneByoneTree.Root.Hex(), tree.Root.Hex())

}

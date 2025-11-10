package smt_koalabear

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/require"
)

// Creates a empty SMT, test its root hash in hard and tests that
// all the leaves are zero.
func TestTreeInitialization(t *testing.T) {

	config := &Config{
		Depth: 40,
	}

	tree := NewEmptyTree(config)

	// Only contains empty leaves
	for pos := 0; pos < 1000; pos++ {
		x, _ := tree.GetLeaf(pos)
		require.Equal(t, x, EmptyLeaf())
	}
}

func TestTreeUpdateLeaf(t *testing.T) {
	config := &Config{
		Depth: 40,
	}

	tree := NewEmptyTree(config)

	// Only contains empty leaves
	for pos := 0; pos < 1000; pos++ {
		var newLeaf field.Octuplet
		for i := range newLeaf {
			newLeaf[i] = field.RandomElement()
		}
		tree.Update(pos, newLeaf)
		recovered, _ := tree.GetLeaf(pos)
		require.Equal(t, newLeaf, recovered)
	}
}
func TestMerkleProofNative(t *testing.T) {
	config := &Config{
		Depth: 40,
	}

	tree := NewEmptyTree(config)

	// Only contains empty leaves
	for pos := 0; pos < 1000; pos++ {
		oldLeaf, _ := tree.GetLeaf(pos)
		proof, _ := tree.Prove(pos)

		// Directly verify the proof
		valid := proof.Verify(config, oldLeaf, tree.Root)
		require.Truef(t, valid, "pos #%v, proof #%v", pos, proof)
	}
}

func TestMerkleProofWithUpdate(t *testing.T) {
	config := &Config{
		Depth: 40,
	}

	tree := NewEmptyTree(config)

	// Only contains empty leaves
	for pos := 0; pos < 1000; pos++ {
		proof, _ := tree.Prove(pos)

		// Updat the leaf with a random-looking value before
		// checking the proof
		var newLeaf field.Octuplet
		for i := range newLeaf {
			newLeaf[i] = field.RandomElement()
		}
		tree.Update(pos, newLeaf)

		// After updating the old proof should still be valid
		// (because we only changed the current leaf)
		valid := proof.Verify(config, newLeaf, tree.Root)
		require.Truef(t, valid, "pos #%v", pos)
	}
}

func TestBuildFromScratch(t *testing.T) {

	config := &Config{
		Depth: 8,
	}

	// Generate random field elements and cast them into field.Octuplet
	// every 8 field elements forms a leaf
	leavesFr := make([]field.Element, 8*(1<<config.Depth))
	for i := 0; i < len(leavesFr); i++ {
		leavesFr[i] = field.RandomElement()
	}

	leaves := make([]field.Octuplet, len(leavesFr)/8)
	for i := range leaves {
		var arr [8]field.Element
		copy(arr[:], leavesFr[8*i:8*i+8])
		leaves[i] = arr
	}

	// And generate the
	tree := BuildComplete(leaves)

	// Test-Merkle tests the merkle proof point by point
	for i := range leaves {
		proof, _ := tree.Prove(i)
		ok := proof.Verify(config, leaves[i], tree.Root)

		if !ok {
			t.Fatalf("failed to verify pos %v", i)
		}
	}

	// Build the same tree by adding the leaves one by one
	oneByoneTree := NewEmptyTree(config)
	for i := range leaves {
		oneByoneTree.Update(i, leaves[i])
	}

	// We should obtain the same roots
	require.Equal(t, oneByoneTree.Root, tree.Root)

}

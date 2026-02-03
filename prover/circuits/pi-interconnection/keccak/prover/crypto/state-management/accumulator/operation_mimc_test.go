//go:build !fuzzlight

package accumulator_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/accumulator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializationMiMC(t *testing.T) {
	// Just check that the code returns
	acc := newTestAccumulatorMiMC()
	ver := acc.VerifierState()

	// The next free nodes are well initialized
	assert.Equal(t, int64(2), acc.NextFreeNode, "bad next free node for the prover state")
	assert.Equal(t, int64(2), ver.NextFreeNode, "bad next free node for the verifier state")

	// The roots are consistent
	assert.Equal(t, acc.SubTreeRoot(), ver.SubTreeRoot, "inconsistent roots")

	headHash := accumulator.Head().Hash(acc.Config())
	tailHash := accumulator.Tail(acc.Config()).Hash(acc.Config())

	// First leaf is head
	assert.Equal(t, acc.Tree.MustGetLeaf(0), accumulator.Head().Hash(acc.Config()))
	assert.Equal(t, acc.Tree.MustGetLeaf(1), accumulator.Tail(acc.Config()).Hash(acc.Config()))

	// Can we prover membership of the leaf
	proofHead := acc.Tree.MustProve(0)
	proofHead.Verify(acc.Config(), headHash, acc.SubTreeRoot())

	proofTail := acc.Tree.MustProve(1)
	proofTail.Verify(acc.Config(), tailHash, acc.SubTreeRoot())
}

func TestInsertionMiMC(t *testing.T) {

	// Performs an insertion
	acc := newTestAccumulatorMiMC()
	ver := acc.VerifierState()

	for i := 0; i < numRepetion; i++ {
		trace := acc.InsertAndProve(dumkey(i), dumval(i))
		err := ver.VerifyInsertion(trace)
		require.NoErrorf(t, err, "check #%v - trace %++v", i, trace)
	}

	// Roots of the verifier should be correct
	assert.Equal(t, acc.NextFreeNode, ver.NextFreeNode)
	assert.Equal(t, acc.SubTreeRoot(), ver.SubTreeRoot)
}

func TestReadZeroMiMC(t *testing.T) {

	// Performs an insertion
	acc := newTestAccumulatorMiMC()
	ver := acc.VerifierState()

	for i := 0; i < numRepetion; i++ {
		key := dumkey(i)
		trace := acc.ReadZeroAndProve(key)
		err := ver.ReadZeroVerify(trace)
		require.NoErrorf(t, err, "check #%v - trace %++v", i, trace)
	}

	// Roots of the verifier should be correct
	assert.Equal(t, acc.NextFreeNode, ver.NextFreeNode)
	assert.Equal(t, acc.SubTreeRoot(), ver.SubTreeRoot)
}

func TestReadNonZeroMiMC(t *testing.T) {

	// Performs an insertion
	acc := newTestAccumulatorMiMC()

	// Fill the tree
	for i := 0; i < numRepetion; i++ {
		_ = acc.InsertAndProve(dumkey(i), dumval(i))
	}

	// Snapshot the verifier after the insertions because of the verifier
	ver := acc.VerifierState()

	for i := 0; i < numRepetion; i++ {
		trace := acc.ReadNonZeroAndProve(dumkey(i))
		err := ver.ReadNonZeroVerify(trace)
		require.NoErrorf(t, err, "check #%v - trace %++v", i, trace)
		require.Equal(t, dumval(i), trace.Value)
	}

	// Roots of the verifier should be correct
	assert.Equal(t, acc.NextFreeNode, ver.NextFreeNode)
	assert.Equal(t, acc.SubTreeRoot(), ver.SubTreeRoot)
}

func TestUpdateMiMC(t *testing.T) {
	// Performs an insertion
	acc := newTestAccumulatorMiMC()

	// Fill the tree
	for i := 0; i < numRepetion; i++ {
		_ = acc.InsertAndProve(dumkey(i), dumval(i))
	}

	// Snapshot the verifier after the insertions because of the verifier
	ver := acc.VerifierState()

	for i := 0; i < numRepetion; i++ {
		trace := acc.UpdateAndProve(dumkey(i), dumval(i+1000))
		err := ver.UpdateVerify(trace)
		require.NoErrorf(t, err, "check #%v - trace %++v", i, trace)
	}

	// Roots of the verifier should be correct
	assert.Equal(t, acc.NextFreeNode, ver.NextFreeNode)
	assert.Equal(t, acc.SubTreeRoot(), ver.SubTreeRoot)
}

func TestDeletionMiMC(t *testing.T) {
	// Performs an insertion
	acc := newTestAccumulatorMiMC()

	// Fill the tree
	for i := 0; i < numRepetion; i++ {
		_ = acc.InsertAndProve(dumkey(i), dumval(i))
	}

	// Snapshot the verifier after the insertions because of the verifier
	ver := acc.VerifierState()

	for i := 0; i < numRepetion; i++ {
		trace := acc.DeleteAndProve(dumkey(i))
		err := ver.VerifyDeletion(trace)
		require.NoErrorf(t, err, "check #%v - trace %++v", i, trace)
	}

	// Roots of the verifier should be correct
	assert.Equal(t, acc.NextFreeNode, ver.NextFreeNode)
	assert.Equal(t, acc.SubTreeRoot(), ver.SubTreeRoot)
}

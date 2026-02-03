package accumulator_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"

	. "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Number of repetition steps
const numRepetion = 255

// Dummy hashable type that we can use for the accumulator
type DummyKey = Bytes32
type DummyVal = Bytes32

const locationTesting = "location"

func dumkey(i int) DummyKey {
	return DummyBytes32(i)
}

func dumval(i int) DummyVal {
	return DummyBytes32(i)
}

func newTestAccumulatorKeccak() *accumulator.ProverState[DummyKey, DummyVal] {
	config := &smt.Config{
		HashFunc: hashtypes.Keccak,
		Depth:    40,
	}
	return accumulator.InitializeProverState[DummyKey, DummyVal](config, locationTesting)
}

func TestInitialization(t *testing.T) {
	// Just check that the code returns
	acc := newTestAccumulatorKeccak()
	ver := acc.VerifierState()

	// The next free nodes are well initialized
	assert.Equal(t, int64(2), acc.NextFreeNode, "bad next free node for the prover state")
	assert.Equal(t, int64(2), ver.NextFreeNode, "bad next free node for the verifier state")

	// The roots are consistent
	assert.Equal(t, acc.SubTreeRoot(), ver.SubTreeRoot, "inconsistent roots")

	headHash := accumulator.Head().Hash(acc.Config())
	tailHash := accumulator.Head().Hash(acc.Config())

	// First leaf is head
	assert.Equal(t, acc.Tree.MustGetLeaf(0), accumulator.Head().Hash(acc.Config()))
	assert.Equal(t, acc.Tree.MustGetLeaf(1), accumulator.Tail(acc.Config()).Hash(acc.Config()))

	// Can we prover membership of the leaf
	proofHead := acc.Tree.MustProve(0)
	proofHead.Verify(acc.Config(), headHash, acc.SubTreeRoot())

	proofTail := acc.Tree.MustProve(1)
	proofTail.Verify(acc.Config(), tailHash, acc.SubTreeRoot())
}

func TestInsertion(t *testing.T) {

	// Performs an insertion
	acc := newTestAccumulatorKeccak()
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

func TestReadZero(t *testing.T) {

	// Performs an insertion
	acc := newTestAccumulatorKeccak()
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

func TestReadNonZero(t *testing.T) {

	// Performs an insertion
	acc := newTestAccumulatorKeccak()

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

func TestUpdate(t *testing.T) {
	// Performs an insertion
	acc := newTestAccumulatorKeccak()

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

func TestDeletion(t *testing.T) {
	// Performs an insertion
	acc := newTestAccumulatorKeccak()

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

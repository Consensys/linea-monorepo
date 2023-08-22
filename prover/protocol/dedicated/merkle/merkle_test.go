package merkle_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/merkle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMerklePow2(t *testing.T) {

	// Generates a list of Merkle proofs for the same tree
	depth := 4
	numProofs := 1 << 2

	leaves := make([]hashtypes.Digest, 1<<depth)
	for i := range leaves {
		// #nosec G404 -- no need for a cryptographically strong PRNG for testing purposes
		var x field.Element
		x.SetRandom()
		leaves[i] = x.Bytes()
	}

	tree := smt.BuildComplete(leaves, hashtypes.MiMC)
	root := tree.Root

	// Generates a vector of merkle proofs
	proofs := make([]smt.Proof, 0, numProofs)
	pos := make([]field.Element, numProofs)
	roots := make([]field.Element, numProofs)
	leavesAssignment := make([]field.Element, numProofs)
	for i := 0; i < numProofs; i++ {
		proof := tree.Prove(i)
		proofs = append(proofs, proof)
		pos[i].SetUint64(uint64(i))
		roots[i].SetBytes(root[:])
		leavesAssignment[i].SetBytes(leaves[i][:])
	}

	define := func(b *wizard.Builder) {
		proofcol := b.RegisterCommit("PROOF", depth*numProofs)
		rootscol := b.RegisterCommit("ROOTS", numProofs)
		leavescol := b.RegisterCommit("LEAVES", numProofs)
		poscol := b.RegisterCommit("POS", numProofs)

		merkle.MerkleProofCheck(b.CompiledIOP, "TEST", depth, numProofs, proofcol, rootscol, leavescol, poscol)
	}

	prove := func(run *wizard.ProverRuntime) {
		proofAssignment := merkle.PackMerkleProofs(proofs)
		run.AssignColumn("PROOF", proofAssignment)
		run.AssignColumn("ROOTS", smartvectors.NewRegular(roots))
		run.AssignColumn("LEAVES", smartvectors.NewRegular(leavesAssignment))
		run.AssignColumn("POS", smartvectors.NewRegular(pos))
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prove)
	err := wizard.Verify(comp, proof)

	require.NoError(t, err)

}

func TestMerkleNotPow2(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)

	// Generates a list of Merkle proofs for the same tree
	depth := 3
	numProofs := 1 << 2
	smallSize := utils.NextPowerOfTwo(numProofs)
	largeSize := utils.NextPowerOfTwo(numProofs * depth)

	leaves := make([]hashtypes.Digest, 1<<depth)
	for i := range leaves {
		// #nosec G404 -- no need for a cryptographically strong PRNG for testing purposes
		var x field.Element
		x.SetRandom()
		leaves[i] = x.Bytes()
	}

	tree := smt.BuildComplete(leaves, hashtypes.MiMC)
	root := tree.Root

	// Generates a vector of merkle proofs
	proofs := make([]smt.Proof, 0, numProofs)
	pos := make([]field.Element, numProofs)
	roots := make([]field.Element, numProofs)
	leavesAssignment := make([]field.Element, numProofs)
	for i := 0; i < numProofs; i++ {
		proof := tree.Prove(i)
		proofs = append(proofs, proof)
		pos[i].SetUint64(uint64(i))
		roots[i].SetBytes(root[:])
		leavesAssignment[i].SetBytes(leaves[i][:])
	}

	define := func(b *wizard.Builder) {
		proofcol := b.RegisterCommit("PROOF", largeSize)
		rootscol := b.RegisterCommit("ROOTS", smallSize)
		leavescol := b.RegisterCommit("LEAVES", smallSize)
		poscol := b.RegisterCommit("POS", smallSize)

		merkle.MerkleProofCheck(b.CompiledIOP, "TEST", depth, numProofs, proofcol, rootscol, leavescol, poscol)
	}

	prove := func(run *wizard.ProverRuntime) {
		proofAssignment := merkle.PackMerkleProofs(proofs)
		run.AssignColumn("PROOF", proofAssignment)
		run.AssignColumn("ROOTS", padWithLast(roots))
		run.AssignColumn("LEAVES", padWithLast(leavesAssignment))
		run.AssignColumn("POS", padWithLast(pos))
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prove)
	err := wizard.Verify(comp, proof)

	require.NoError(t, err)

}

func TestMerkleManySizes(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	dimensions := []map[string]int{
		{"depth": 4, "numProofs": 16},
		{"depth": 8, "numProofs": 16},
		{"depth": 4, "numProofs": 14},
		{"depth": 5, "numProofs": 14},
		{"depth": 6, "numProofs": 14},
		{"depth": 6, "numProofs": 48},
	}

	runTest := func(dims map[string]int) {

		// Generates a list of Merkle proofs for the same tree
		depth := dims["depth"]
		numProofs := dims["numProofs"]

		smallSize := utils.NextPowerOfTwo(numProofs)
		largeSize := utils.NextPowerOfTwo(numProofs * depth)

		leaves := make([]hashtypes.Digest, 1<<depth)
		for i := range leaves {
			// #nosec G404 -- no need for a cryptographically strong PRNG for testing purposes
			var x field.Element
			x.SetRandom()
			leaves[i] = x.Bytes()
		}

		tree := smt.BuildComplete(leaves, hashtypes.MiMC)
		root := tree.Root

		// Generates a vector of merkle proofs
		proofs := make([]smt.Proof, 0, numProofs)
		pos := make([]field.Element, numProofs)
		roots := make([]field.Element, numProofs)
		leavesAssignment := make([]field.Element, numProofs)
		for i := 0; i < numProofs; i++ {
			proof := tree.Prove(i)
			proofs = append(proofs, proof)
			pos[i].SetUint64(uint64(i))
			roots[i].SetBytes(root[:])
			leavesAssignment[i].SetBytes(leaves[i][:])
		}

		define := func(b *wizard.Builder) {
			proofcol := b.RegisterCommit("PROOF", largeSize)
			rootscol := b.RegisterCommit("ROOTS", smallSize)
			leavescol := b.RegisterCommit("LEAVES", smallSize)
			poscol := b.RegisterCommit("POS", smallSize)

			merkle.MerkleProofCheck(b.CompiledIOP, "TEST", depth, numProofs, proofcol, rootscol, leavescol, poscol)
		}

		prove := func(run *wizard.ProverRuntime) {
			proofAssignment := merkle.PackMerkleProofs(proofs)
			run.AssignColumn("PROOF", proofAssignment)
			run.AssignColumn("ROOTS", smartvectors.RightZeroPadded(roots, smallSize))
			run.AssignColumn("LEAVES", smartvectors.RightZeroPadded(leavesAssignment, smallSize))
			run.AssignColumn("POS", smartvectors.RightZeroPadded(pos, smallSize))
		}

		comp := wizard.Compile(define, dummy.Compile)
		proof := wizard.Prove(comp, prove)
		err := wizard.Verify(comp, proof)
		require.NoError(t, err)
	}

	for i := range dimensions {
		t.Logf("run test case #%v", i)
		runTest(dimensions[i])
	}

}

func padWithLast(v []field.Element) smartvectors.SmartVector {
	return smartvectors.RightPadded(v, v[len(v)-1], utils.NextPowerOfTwo(len(v)))
}

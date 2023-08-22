package statemanager

import (
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/accumulator"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/merkle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/sirupsen/logrus"
)

const (
	// Dimensions of the Merkle-proof verification module
	maxNumProofs    int = 1 << 13
	merkleTreeDepth int = 40

	// Column names
	STATEMANAGER_MERKLE_PROOFS             ifaces.ColID = "STATEMANAGER_MERKLE_PROOFS"
	STATEMANAGER_MERKLE_ROOTS              ifaces.ColID = "STATEMANAGER_MERKLE_ROOTS"
	STATEMANAGER_MERKLE_POSITIONS          ifaces.ColID = "STATEMANAGER_MERKLE_POSITIONS"
	STATEMANAGER_MERKLE_LEAVES             ifaces.ColID = "STATEMANAGER_MERKLE_LEAVES"
	STATEMANAGER_MERKLE_PROOF_VERIFICATION string       = "STATEMANAGER_MERKLE_PROOF_VERIFICATION"
)

// Config of the Merkle-tree
var mtConfig = &smt.Config{HashFunc: hashtypes.MiMC, Depth: merkleTreeDepth}

// RegisterStateManagerMerkleProof registers the state manager
// merkle-proof verification in the builder
func RegisterStateManagerMerkleProof(
	comp *wizard.CompiledIOP,
	round int,
) {

	// computes the size of the modules
	leaveSize := utils.NextPowerOfTwo(maxNumProofs)
	proofSize := utils.NextPowerOfTwo(maxNumProofs * merkleTreeDepth)

	// initializes the columns
	leaves := comp.InsertCommit(round, STATEMANAGER_MERKLE_LEAVES, leaveSize)
	roots := comp.InsertCommit(round, STATEMANAGER_MERKLE_ROOTS, leaveSize)
	positions := comp.InsertCommit(round, STATEMANAGER_MERKLE_POSITIONS, leaveSize)
	proofs := comp.InsertCommit(round, STATEMANAGER_MERKLE_PROOFS, proofSize)

	// and constrain the proofs to be consistents with the claims
	merkle.MerkleProofCheck(
		comp,
		"STATEMANAGER_MERKLE_PROOFS",
		merkleTreeDepth, maxNumProofs,
		proofs, roots, leaves, positions,
	)
}

// AssignStateManagerMerkleProof assigns from a list of merkle proofs
func AssignStateManagerMerkleProof(
	run *wizard.ProverRuntime,
	// The traces parsed for the state-manager inspection process
	traces [][]any,
) {

	// Counts the number of merkle-proofs check needed to assess the
	// correctness of the storage accesses.
	provedClaims := make([]smt.ProvedClaim, 0, maxNumProofs)

	// Accumulates the Merkle proof claims
	for _, blockTrace := range traces {
		for _, trace := range blockTrace {
			switch t := trace.(type) {
			case accumulator.DeferableCheck:
				provedClaims = t.DeferMerkleChecks(mtConfig, provedClaims)
			default:
				utils.Panic("unexpected type : %T", t)
			}

		}
	}

	logrus.Debugf("parsed %v merkle proofs checks in the traces", len(provedClaims))

	// Log if there are too many claims
	if len(provedClaims) > maxNumProofs {
		logrus.Errorf("got more proofs than what we can prove: %v -> truncating to %v", len(provedClaims), maxNumProofs)
		provedClaims = provedClaims[:maxNumProofs]
	}

	// Pad with dummy proof so that we reach the target number of proofs
	if len(provedClaims) < maxNumProofs {
		provedClaims = padClaimsWithDummy(provedClaims, maxNumProofs)
	}

	// Allocate the Merkle proofs modules with
	numProofs := len(provedClaims)
	paddedSize := utils.NextPowerOfTwo(numProofs)
	leaves := make([]field.Element, numProofs)
	positions := make([]field.Element, numProofs)
	roots := make([]field.Element, numProofs)
	proofs := make([]smt.Proof, numProofs)

	for i := range provedClaims {
		leaves[i].SetBytes(provedClaims[i].Leaf[:])
		roots[i].SetBytes(provedClaims[i].Root[:])
		positions[i].SetInt64(int64(provedClaims[i].Proof.Path))
		proofs[i] = provedClaims[i].Proof
	}

	run.AssignColumn(STATEMANAGER_MERKLE_PROOFS, merkle.PackMerkleProofs(proofs))
	run.AssignColumn(STATEMANAGER_MERKLE_ROOTS, smartvectors.RightZeroPadded(roots, paddedSize))
	run.AssignColumn(STATEMANAGER_MERKLE_POSITIONS, smartvectors.RightZeroPadded(positions, paddedSize))
	run.AssignColumn(STATEMANAGER_MERKLE_LEAVES, smartvectors.RightZeroPadded(leaves, paddedSize))
}

func padClaimsWithDummy(claims []smt.ProvedClaim, targetNumClaims int) []smt.ProvedClaim {

	logrus.Infof("(state-manager) padding the claims from %v to %v", len(claims), targetNumClaims)

	// sanity-check to avoid an infinite-loop
	if len(claims) >= targetNumClaims {
		return claims
	}

	tree := smt.NewEmptyTree(mtConfig)
	root := tree.Root
	proof := tree.Prove(0)
	leaf := tree.GetLeaf(0)

	// sanity-check
	if !proof.Verify(mtConfig, leaf, root) {
		panic("unexpected failure")
	}

	for i := len(claims); i < targetNumClaims; i++ {
		claims = append(claims, smt.ProvedClaim{
			Proof: proof, Root: root, Leaf: leaf,
		})
	}

	return claims
}

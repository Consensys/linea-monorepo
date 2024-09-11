package statemanager

import (
	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/merkle"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// SettingsLegacy specifies the parameters with which to instantiate the statemanager
// module.
type SettingsLegacy struct {
	// In production mode, the statemanager is not optional but for the partial
	// prover and the checker it is not activated.
	Enabled        bool
	MaxMerkleProof int
}

const (
	// Dimensions of the Merkle-proof verification module
	merkleTreeDepth int = 40

	// Column names
	MERKLE_PROOFS_NAME             ifaces.ColID = "STATEMANAGER_MERKLE_PROOFS"
	MERKLE_ROOTS_NAME              ifaces.ColID = "STATEMANAGER_MERKLE_ROOTS"
	MERKLE_POSITIONS_NAME          ifaces.ColID = "STATEMANAGER_MERKLE_POSITIONS"
	MERKLE_LEAVES_NAME             ifaces.ColID = "STATEMANAGER_MERKLE_LEAVES"
	MERKLE_PROOF_VERIFICATION_NAME string       = "STATEMANAGER_MERKLE_PROOF_VERIFICATION"
)

// Config of the Merkle-tree
var mtConfig = &smt.Config{HashFunc: hashtypes.MiMC, Depth: merkleTreeDepth}

// State manager module
type StateManagerLegacy struct {
	Settings  *SettingsLegacy
	Leaves    ifaces.Column
	Roots     ifaces.Column
	Positions ifaces.Column
	Proofs    ifaces.Column
}

// Define registers the state manager merkle-proof verification in the builder
func (sm *StateManagerLegacy) Define(comp *wizard.CompiledIOP) {

	// All the columns and queries from the state-manager are for the round 0
	round := 0
	maxNumProofs := sm.Settings.MaxMerkleProof

	// Computes the size of the modules
	leaveSize := utils.NextPowerOfTwo(maxNumProofs)
	proofSize := utils.NextPowerOfTwo(maxNumProofs * merkleTreeDepth)

	// Initializes the columns
	sm.Leaves = comp.InsertCommit(round, MERKLE_LEAVES_NAME, leaveSize)
	sm.Roots = comp.InsertCommit(round, MERKLE_ROOTS_NAME, leaveSize)
	sm.Positions = comp.InsertCommit(round, MERKLE_POSITIONS_NAME, leaveSize)
	sm.Proofs = comp.InsertCommit(round, MERKLE_PROOFS_NAME, proofSize)

	// And constrain the proofs to be consistents with the claims
	merkle.MerkleProofCheck(
		comp,
		"STATEMANAGER_MERKLE_PROOFS",
		merkleTreeDepth, maxNumProofs,
		sm.Proofs, sm.Roots, sm.Leaves, sm.Positions,
	)
}

// Assign assigns from a list of merkle proofs
func (sm *StateManagerLegacy) Assign(
	run *wizard.ProverRuntime,
	// The traces parsed for the state-manager inspection process
	traces [][]statemanager.DecodedTrace,
) {

	maxNumProofs := sm.Settings.MaxMerkleProof

	// Counts the number of merkle-proofs check needed to assess the
	// correctness of the storage accesses.
	provedClaims := make([]smt.ProvedClaim, 0, maxNumProofs)

	// Accumulates the Merkle proof claims
	for _, blockTrace := range traces {
		for _, trace := range blockTrace {
			switch t := trace.Underlying.(type) {
			case accumulator.Trace:
				provedClaims = t.DeferMerkleChecks(mtConfig, provedClaims)
			default:
				utils.Panic("Unexpected type : %T", t)
			}
		}
	}

	logrus.Debugf("Parsed %v merkle proofs checks in the traces", len(provedClaims))

	// Log if there are too many claims
	if len(provedClaims) > maxNumProofs {
		logrus.Warnf(
			"Got more proofs than what we can prove: %v -> truncating to %v",
			len(provedClaims), maxNumProofs,
		)
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

	run.AssignColumn(sm.Proofs.GetColID(), merkle.PackMerkleProofs(proofs))
	run.AssignColumn(sm.Roots.GetColID(), smartvectors.RightZeroPadded(roots, paddedSize))
	run.AssignColumn(sm.Positions.GetColID(), smartvectors.RightZeroPadded(positions, paddedSize))
	run.AssignColumn(sm.Positions.GetColID(), smartvectors.RightZeroPadded(leaves, paddedSize))
}

func padClaimsWithDummy(claims []smt.ProvedClaim, targetNumClaims int) []smt.ProvedClaim {

	logrus.Infof("(State-manager) Padding the claims from %v to %v", len(claims), targetNumClaims)

	// sanity-check to avoid an infinite-loop
	if len(claims) >= targetNumClaims {
		return claims
	}

	tree := smt.NewEmptyTree(mtConfig)
	root := tree.Root
	proof := tree.MustProve(0)
	leaf := tree.MustGetLeaf(0)

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

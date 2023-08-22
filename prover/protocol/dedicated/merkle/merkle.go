package merkle

import (
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizardutils"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Wizard gadget allowing to verify a Merkle proof
// See : https://github.com/consensys/zkevm-monorepo/issues/67
func MerkleProofCheck(
	// compiled IOP
	comp *wizard.CompiledIOP,
	// name of the Merkle proof check instance
	name string,
	// depth of the tree
	depth, numProofs int,
	// column representing the proofs. If the number
	// of proof is a non-power of two, roots, leaves and pos
	// should be padded by repeating the last entry.
	proofs, roots, leaves, pos ifaces.Column,
) {

	round := wizardutils.MaxRound(proofs, roots, leaves, pos)

	// define the compute module
	cm := ComputeMod{}
	cm.Cols.Proof = proofs
	cm.Define(comp, round, name, numProofs, depth)

	// define the result module
	rm := ResultMod{}
	rm.Roots = roots
	rm.Leaf = leaves
	rm.Pos = pos
	rm.Define(comp, round, name, numProofs)

	// define the lookup relation
	comp.InsertInclusion(
		round,
		ifaces.QueryIDf("MERKLE_MODULE_LOOKUP_%v", name),
		[]ifaces.Column{cm.Cols.NewProof, cm.Cols.Curr, cm.Cols.PosAcc, cm.Cols.Root},
		[]ifaces.Column{rm.IsActive, rm.Leaf, rm.Pos, rm.Roots},
	)

	// assigns the compute module
	comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
		leaves := leaves.GetColAssignment(run)
		pos := pos.GetColAssignment(run)
		cm.assign(run, leaves, pos)
	})
}

// pack a list of merkle-proofs into a single vector
func PackMerkleProofs(proofs []smt.Proof) smartvectors.SmartVector {

	numProofs := len(proofs)
	depth := len(proofs[0].Siblings)
	numRows := utils.NextPowerOfTwo(numProofs * depth)

	res := make([]field.Element, 0, numProofs*depth)
	for i := range proofs {
		for j := range proofs[i].Siblings {
			// assertion, all proofs have the assumed depth
			if len(proofs[i].Siblings) != depth {
				utils.Panic("expected depth %v, got %v", depth, len(proofs[i].Siblings))
			}
			proofentry := proofs[i].Siblings[depth-j-1]
			var x field.Element
			x.SetBytes(proofentry[:])
			res = append(res, x)
		}
	}

	return smartvectors.RightZeroPadded(res, numRows)
}

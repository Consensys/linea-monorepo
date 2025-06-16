package merkle

import (
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// FlatProof is a collection of columns representing a Merkle proof in a flat
// manner.
type FlatProof struct {
	Nodes [common.NbLimbU256][]ifaces.Column
}

// Depth returns the depth of the tree represented by the proof
func (p *FlatProof) Depth() int {
	return len(p.Nodes[0])
}

// NumRow returns the number of rows of the [FlatProof]
func (p *FlatProof) NumRow() int {
	return p.Nodes[0][0].Size()
}

// NewProof instantiates a new [FlatProof] object in the wizard
func NewProof(comp *wizard.CompiledIOP, round int, name string, depth, numRows int) *FlatProof {
	proof := &FlatProof{}
	for i := 0; i < depth; i++ {
		for j := range proof.Nodes {
			node := comp.InsertCommit(round, ifaces.ColIDf("%v_NODE_%v_%v", name, i, j), numRows)
			proof.Nodes[j] = append(proof.Nodes[j], node)
		}
	}
	return proof
}

// WithStatus changes the status of all the columns of the [FlatProof] to
// "status" and returns a pointer to the receiver.
func (p *FlatProof) WithStatus(comp *wizard.CompiledIOP, status column.Status) *FlatProof {
	for _, nodeLimbs := range p.Nodes {
		for _, limb := range nodeLimbs {
			comp.Columns.SetStatus(limb.GetColID(), status)
		}
	}
	return p
}

// Assign assigns the proof columns to the [FlatProof] object from a list of
// [merkletree.Proof] objects. The columns are zero-padded on the right.
func (p *FlatProof) Assign(run *wizard.ProverRuntime, proofs []smt.Proof) {

	assignment := make([][][]field.Element, p.Depth())

	for i := range proofs {
		for j := range proofs[i].Siblings {
			siblingsLimbsBytes := common.SplitBytes(proofs[i].Siblings[j][:])

			var nodeAsFrLimbs []field.Element
			for _, limbBytes := range siblingsLimbsBytes {
				var limb field.Element
				limb.SetBytes(limbBytes)

				nodeAsFrLimbs = append(nodeAsFrLimbs, limb)
			}

			if len(assignment[j]) == 0 {
				assignment[j] = make([][]field.Element, len(nodeAsFrLimbs))
			}

			for k, limb := range nodeAsFrLimbs {
				assignment[j][k] = append(assignment[j][k], limb)
			}
		}
	}

	for i := range p.Nodes[0] {
		for j := range p.Nodes {
			run.AssignColumn(p.Nodes[j][i].GetColID(),
				smartvectors.RightZeroPadded(assignment[i][j], p.NumRow()))
		}
	}
}

// Unpack reads the assignment of a proof and returns it as a list of
// [smt.Proof] objects. The function assumes that the columns are all
// padded in the same fashion.
//
// The function also takes as additional input a smart-vector containing
// the positions corresponding for each proofs.
func (p *FlatProof) Unpack(run ifaces.Runtime, pos smartvectors.SmartVector) []smt.Proof {

	var (
		proofs = make([]smt.Proof, 0)
		// The assumption here is two-fold: first, this relies on the
		// fact that we know all the columns are structured and padded the same.
		//
		// Since every column of Nodes limbs has the same size, the window range can
		// be retrieved from the first limb column.
		start, stop = smartvectors.CoWindowRange(p.Nodes[0][0].GetColAssignment(run))
	)

	for i := start; i <= stop; i++ {

		newProof := smt.Proof{
			Path:     field.ToInt(pos.GetPtr(i)),
			Siblings: make([]types.Bytes32, len(p.Nodes)),
		}

		for n := range len(p.Nodes[0]) {
			siblingLimbBytes := make([]byte, len(p.Nodes))

			for _, limbCol := range p.Nodes {
				element := limbCol[n].GetColAssignmentAt(run, i)
				elementBytes := element.Bytes()
				siblingLimbBytes = append(siblingLimbBytes, elementBytes[field.Bytes-common.LimbBytes:]...)
			}

			copy(newProof.Siblings[n][:], siblingLimbBytes)
		}

		proofs = append(proofs, newProof)
	}

	return proofs
}

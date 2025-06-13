package merkle

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/bits"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// FlatProofVerificationInputs collects parameters helpful for defining a Merkle-tree verification
// wizard.
type FlatProofVerificationInputs struct {
	// Name is the name of the Merkle tree verification instance
	Name string
	// Proof are the columns reserved for storing the Merkle proofs
	Proof FlatProof
	// Leaf contains the alleged leaves
	Leaf [common.NbLimbU256]ifaces.Column
	// Roots contains the Merkle roots
	Roots [common.NbLimbU256]ifaces.Column
	// Position contains the positions of the alleged leaves
	Position [common.NbLimbU64]ifaces.Column
	// Use for looking up and selecting only the the columns containing the
	// root in the ComputeMod.
	IsActive ifaces.Column
}

// FlatMerkleProofVerification represents a Merkle-tree verification wizard
// instance. It contains all the elements
type FlatMerkleProofVerification struct {
	// FlatProofVerificationInputs provides all the parameters needed to define the Merkle tree
	FlatProofVerificationInputs
	// PosBits contains the bit decomposition of the position
	PosBits *bits.BitDecomposed
	// Lefts and Rights respectively represent the left and the right
	// children of the current node being hashed. They are constructed
	// as ternaries
	Lefts, Rights []*dedicated.TernaryCtx
	// Node contains the hash of nodes, obtained by hashing the left and
	// right.
	Nodes []*mimc.HashingCtx
}

// CheckFlatMerkleProofs checks a list of Merkle proof using an horizontal
// wizard arithmetization.
func CheckFlatMerkleProofs(comp *wizard.CompiledIOP, inputs FlatProofVerificationInputs) *FlatMerkleProofVerification {

	if err := checkColumnsAllHaveSameSize(&inputs); err != nil {
		panic(err)
	}

	ctx := &FlatMerkleProofVerification{
		FlatProofVerificationInputs: inputs,
	}

	ctx.PosBits = bits.BitDecompose(comp, inputs.Position[:], len(inputs.Proof.Nodes[0]))

	prevNode := inputs.Leaf

	for j := range len(inputs.Proof.Nodes[0]) {
		var leftLimbs [common.NbLimbU256]ifaces.Column
		var rightLimbs [common.NbLimbU256]ifaces.Column

		for k := range inputs.Proof.Nodes {

			var (
				left  = dedicated.Ternary(comp, ctx.PosBits.Bits[j], inputs.Proof.Nodes[k][j], prevNode[k])
				right = dedicated.Ternary(comp, ctx.PosBits.Bits[j], prevNode[k], inputs.Proof.Nodes[k][j])
			)

			leftLimbs[k] = left.Result
			rightLimbs[k] = right.Result

			ctx.Lefts = append(ctx.Lefts, left)
			ctx.Rights = append(ctx.Rights, right)
		}

		node := mimc.HashOf(comp, [][]ifaces.Column{leftLimbs[:], rightLimbs[:]})
		prevNode = node.Result()

		ctx.Nodes = append(ctx.Nodes, node)
	}

	// This check ensures that the computed and the provided root match. Note
	// that prevNode is the last node computed, hence the root.
	for i := range prevNode {
		comp.InsertGlobal(
			max(prevNode[i].Round(), inputs.Roots[i].Round()),
			ifaces.QueryIDf("%v_ROOT_MATCH_%v", inputs.Name, i),
			symbolic.Mul(inputs.IsActive, symbolic.Sub(prevNode[i], inputs.Roots[i])),
		)

		break
	}

	return ctx
}

// AddProofReuseConstraint adds a constraint on the reuse of the same proof
// across different rows of the Merkle proof verification module.
func (ctx *FlatMerkleProofVerification) AddProofReuseConstraint(comp *wizard.CompiledIOP, mustReuseForNext ifaces.Column) {

	for i := range ctx.Proof.Nodes {
		for j := range ctx.Proof.Nodes[i] {
			comp.InsertGlobal(
				max(mustReuseForNext.Round(), ctx.Proof.Nodes[0][0].Round()),
				ifaces.QueryIDf("%v_PROOF_REUSE_%v_%v", ctx.FlatProofVerificationInputs.Name, i, j),
				symbolic.Mul(
					mustReuseForNext,
					symbolic.Sub(
						ctx.Proof.Nodes[i][j],
						column.Shift(ctx.Proof.Nodes[i][j], 1),
					),
				),
			)
		}
	}
}

// Run implements the [wizard.ProverAction] interface and implements all the
// steps of the Merkle tree verification. It assigns node after node, all the
// columns needed to complete the Merkle tree verification from the leaf up
// to the root.
func (ctx *FlatMerkleProofVerification) Run(run *wizard.ProverRuntime) {

	ctx.PosBits.Run(run)

	for i := range ctx.Lefts {
		ctx.Lefts[i].Run(run)
		ctx.Rights[i].Run(run)

		if (i+1)%len(ctx.Leaf) == 0 {
			ctx.Nodes[i/len(ctx.Leaf)].Run(run)
		}
	}
}

// checkColumnsAllHaveSameSize checks if all the columns have the same size
// and returns an error if they don't.
func checkColumnsAllHaveSameSize(inp *FlatProofVerificationInputs) error {

	size := inp.Roots[0].Size()
	for i := range inp.Roots {
		for _, node := range inp.Proof.Nodes[i] {
			if node.Size() != size {
				return fmt.Errorf("all nodes must have the same size: root=%v proof=%v", size, node.Size())
			}
		}

		if inp.Leaf[i].Size() != size {
			return fmt.Errorf("all nodes must have the same size: root=%v leaf=%v", size, inp.Leaf[i].Size())
		}
	}

	for i := range inp.Position {
		if inp.Position[i].Size() != size {
			return fmt.Errorf("all nodes must have the same size: root=%v position=%v", size, inp.Position[i].Size())
		}
	}

	if inp.IsActive.Size() != size {
		return fmt.Errorf("all nodes must have the same size: root=%v isActive=%v", size, inp.IsActive.Size())
	}

	return nil
}

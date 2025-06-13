package merkle

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const (
	merkleTestDepth  = 16
	merkleTestNumRow = 16
)

// testcases is a list of test-cases scenarios for the merkle tree
var testcases = [][]merkleTestCaseInstance{
	{
		{
			IsWrite: true,
			Pos:     0,
			Leaf:    types.Bytes32{1, 2, 3, 4},
		},
		{
			Pos: 1,
		},
	},
}

func TestMerkleTreeFlat(t *testing.T) {

	for i, testcase := range testcases {
		t.Run(fmt.Sprintf("testcase-%v", i), func(t *testing.T) {

			tree := newMerkleTestBuilder(merkleTestDepth)

			for _, instance := range testcase {
				if instance.IsWrite {
					tree.AddWrite(instance.Pos, instance.Leaf)
				} else {
					tree.AddRead(instance.Pos)
				}
			}

			var (
				tr   = &merkleTestRunnerFlat{}
				comp = wizard.Compile(tr.Define, dummy.CompileAtProverLvl())
				_    = wizard.Prove(comp, func(run *wizard.ProverRuntime) { tr.Assign(run, tree) })
			)

		})
	}
}

// merkleTestRunnerFlat is a runner for the merkle tree test-cases
type merkleTestRunnerFlat struct {
	ctx *FlatMerkleProofVerification
}

// Define implements the [wizard.DefineFunc] interface
func (ctx *merkleTestRunnerFlat) Define(b *wizard.Builder) {

	mpvInputs := FlatProofVerificationInputs{
		Name:     "test",
		Proof:    *NewProof(b.CompiledIOP, 0, "test", merkleTestDepth, merkleTestNumRow),
		IsActive: b.RegisterCommit("ACTIVE", merkleTestNumRow),
	}

	for i := range mpvInputs.Leaf {
		mpvInputs.Leaf[i] = b.RegisterCommit(ifaces.ColIDf("LEAF_%d", i), merkleTestNumRow)
		mpvInputs.Roots[i] = b.RegisterCommit(ifaces.ColIDf("ROOTS_%d", i), merkleTestNumRow)
	}

	for i := range mpvInputs.Position {
		mpvInputs.Position[i] = b.RegisterCommit(ifaces.ColIDf("POS_%d", i), merkleTestNumRow)
	}

	ctx.ctx = CheckFlatMerkleProofs(b.CompiledIOP, mpvInputs)
}

// Assign assigns the merkle tree test-cases at runtime
func (ctx *merkleTestRunnerFlat) Assign(run *wizard.ProverRuntime, data *merkleTestBuilder) {

	for i := range data.leaves {
		run.AssignColumn(ifaces.ColIDf("LEAF_%d", i), smartvectors.RightZeroPadded(data.leaves[i], merkleTestNumRow))

		run.AssignColumn(ifaces.ColIDf("ROOTS_%d", i), smartvectors.RightZeroPadded(data.roots[i], merkleTestNumRow))
	}

	for i := range data.pos {
		run.AssignColumn(ifaces.ColIDf("POS_%d", i), smartvectors.RightZeroPadded(data.pos[i], merkleTestNumRow))
	}

	run.AssignColumn("ACTIVE", smartvectors.RightZeroPadded(data.isActive, merkleTestNumRow))

	ctx.ctx.Proof.Assign(run, data.proofs)
	ctx.ctx.Run(run)

	for i := 0; i < merkleTestNumRow; i++ {
		for l := 0; l < merkleTestDepth; l++ {

			var (
				left  = ctx.ctx.Lefts[l].Result.GetColAssignmentAt(run, i)
				right = ctx.ctx.Rights[l].Result.GetColAssignmentAt(run, i)
			)

			for j := range data.leaves {
				node := ctx.ctx.Nodes[l].Result()[j].GetColAssignmentAt(run, i)

				fmt.Printf("proof=%v level=%v left=%v right=%v node=%v\n", i, l, left.Text(16), right.Text(16), node.Text(16))
			}
		}
	}
}

// merkleTestCaseInstance represents either a read or a write operation to add to
// the test-cases.
type merkleTestCaseInstance struct {
	IsWrite bool
	Pos     int
	// Leaf is only taken into consideration if Write is true
	Leaf types.Bytes32
}

// merkleTestBuilder is used to build the assignment of merkle proofs
// and is implemented like a writer.
type merkleTestBuilder struct {
	proofs             []smt.Proof
	pos                [common.NbLimbU64][]field.Element
	roots              [common.NbLimbU256][]field.Element
	leaves             [common.NbLimbU256][]field.Element
	useNextMerkleProof []field.Element
	isActive           []field.Element
	counter            [common.NbLimbU64][]field.Element
	tree               smt.Tree
}

// merkleTestBuilderRow is a pure-data structure specifying a row in the test builder
type merkleTestBuilderRow struct {
	proof              smt.Proof
	pos                int
	leaf               types.Bytes32
	root               types.Bytes32
	useNextMerkleProof bool
}

func newMerkleTestBuilder(depth int) *merkleTestBuilder {
	return &merkleTestBuilder{
		tree: *smt.BuildComplete(make([]types.Bytes32, 1<<depth), hashtypes.MiMC),
	}
}

func (mt *merkleTestBuilder) AddRead(pos int) {

	var (
		proof              = mt.tree.MustProve(pos)
		leaf, _            = mt.tree.GetLeaf(pos)
		root               = mt.tree.Root
		useNextMerkleProof = false
	)

	mt.pushRow(merkleTestBuilderRow{
		proof:              proof,
		pos:                pos,
		leaf:               leaf,
		root:               root,
		useNextMerkleProof: useNextMerkleProof,
	})
}

func (mt *merkleTestBuilder) AddWrite(pos int, newLeaf types.Bytes32) {

	proof := mt.tree.MustProve(pos)

	mt.pushRow(merkleTestBuilderRow{
		proof:              proof,
		pos:                pos,
		leaf:               mt.tree.MustGetLeaf(pos),
		root:               mt.tree.Root,
		useNextMerkleProof: true,
	})

	mt.tree.Update(pos, newLeaf)

	mt.pushRow(merkleTestBuilderRow{
		proof: proof,
		pos:   pos,
		leaf:  newLeaf,
		root:  mt.tree.Root,
	})
}

func valueToLimbs(limbsNb int, value uint64) (res []field.Element) {
	res = make([]field.Element, limbsNb)
	for i := range limbsNb - 1 {
		res[i] = field.Zero()
	}

	res[limbsNb-1] = field.NewElement(value)
	return res
}

// pushRow adds a row to the builder
func (mt *merkleTestBuilder) pushRow(row merkleTestBuilderRow) {
	counterLimbs := valueToLimbs(4, uint64(len(mt.counter)))
	posLimbs := valueToLimbs(4, uint64(row.pos))
	for i := range counterLimbs {
		mt.counter[i] = append(mt.counter[i], counterLimbs[i])
		mt.pos[i] = append(mt.pos[i], posLimbs[i])
	}

	mt.proofs = append(mt.proofs, row.proof)

	leafs := common.SplitElement(row.leaf.ToField())
	roots := common.SplitElement(row.root.ToField())
	for i := range leafs {
		mt.leaves[i] = append(mt.leaves[i], leafs[i])
		mt.roots[i] = append(mt.roots[i], roots[i])
	}

	mt.useNextMerkleProof = append(mt.useNextMerkleProof, field.FromBool(row.useNextMerkleProof))
	mt.isActive = append(mt.isActive, field.One())
}

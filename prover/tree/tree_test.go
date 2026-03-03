package tree

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/localcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mpts"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/splitextension"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		panic(err)
	}
}

// leafCompilationSuite returns a minimal compilation suite for leaf wizard IOPs.
// This is kept small for fast test execution.
func leafCompilationSuite() []func(*wizard.CompiledIOP) {
	return []func(*wizard.CompiledIOP){
		logderivativesum.CompileLookups,
		localcs.Compile,
		globalcs.Compile,
		univariates.Naturalize,
		mpts.Compile(),
		splitextension.CompileSplitExtToBase,
		vortex.Compile(
			2,
			false,
			vortex.ForceNumOpenedColumns(4),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.PremarkAsSelfRecursed(),
			vortex.WithOptionalSISHashingThreshold(64),
		),
	}
}

// testTreeCompilationSuite returns a minimal compilation suite for tree nodes.
// Uses smaller parameters than production to keep test time reasonable.
func testTreeCompilationSuite() CompilationSuite {
	return CompilationSuite{
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1 << 13),
		),
		vortex.Compile(
			8,
			false,
			vortex.ForceNumOpenedColumns(32),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.WithOptionalSISHashingThreshold(64),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1 << 13),
		),
		vortex.Compile(
			8,
			true, // IsLastRound for the root level
			vortex.ForceNumOpenedColumns(32),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.WithOptionalSISHashingThreshold(64),
		),
	}
}

// defineLeafFunc creates a define function for a simple wizard IOP with a
// lookup constraint. The prefix ensures unique column names per test.
func defineLeafFunc(prefix string) func(bui *wizard.Builder) {
	return func(bui *wizard.Builder) {
		a := bui.RegisterCommit(ifaces.ColID(prefix+"_A"), 1024)
		b := bui.RegisterCommit(ifaces.ColID(prefix+"_B"), 1024)
		bui.Inclusion(ifaces.QueryID(prefix+"_Q"), []ifaces.Column{a}, []ifaces.Column{b})
	}
}

// proveLeafFunc creates a prover function for the leaf wizard IOP.
func proveLeafFunc(prefix string) func(run *wizard.ProverRuntime) {
	return func(run *wizard.ProverRuntime) {
		run.AssignColumn(ifaces.ColID(prefix+"_A"), smartvectors.NewConstant(field.Zero(), 1024))
		run.AssignColumn(ifaces.ColID(prefix+"_B"), smartvectors.NewConstant(field.Zero(), 1024))
	}
}

// TestCompileLevel verifies that a single tree aggregation level can be
// compiled successfully with MaxNumProof=2.
func TestCompileLevel(t *testing.T) {
	leafComp := wizard.Compile(defineLeafFunc("compile"), leafCompilationSuite()...)
	require.NotNil(t, leafComp)

	node := compileLevelWithSuite(leafComp, 0, "compile-level", testTreeCompilationSuite())
	require.NotNil(t, node)
	require.NotNil(t, node.Recursion)
	require.NotNil(t, node.CompiledIOP)
	assert.Equal(t, leafComp, node.ChildComp)

	t.Logf("Tree level 0 compiled: %d rounds", node.CompiledIOP.NumRounds())
}

// compileLevelWithSuite is a test helper that compiles a single tree level
// with a custom compilation suite (to keep tests fast).
func compileLevelWithSuite(childComp *wizard.CompiledIOP, levelIdx int, name string, suite CompilationSuite) *AggregationNode {
	node := &AggregationNode{ChildComp: childComp}
	define := func(b *wizard.Builder) {
		node.Recursion = recursion.DefineRecursionOf(
			b.CompiledIOP,
			childComp,
			recursion.Parameters{
				Name:        name,
				MaxNumProof: 2,
			},
		)
	}
	node.CompiledIOP = wizard.Compile(define, suite...)
	return node
}

// TestProvePair verifies that two leaf witnesses can be aggregated by a
// single tree node and the resulting proof verifies.
func TestProvePair(t *testing.T) {
	// Compile leaf IOP
	leafComp := wizard.Compile(defineLeafFunc("pair"), leafCompilationSuite()...)

	// Compile tree node
	node := compileLevelWithSuite(leafComp, 0, "pair-level", testTreeCompilationSuite())

	// Generate two leaf witnesses
	stoppingRound := recursion.VortexQueryRound(leafComp) + 1

	run1 := wizard.RunProverUntilRound(leafComp, proveLeafFunc("pair"), stoppingRound)
	wit1 := recursion.ExtractWitness(run1)

	run2 := wizard.RunProverUntilRound(leafComp, proveLeafFunc("pair"), stoppingRound)
	wit2 := recursion.ExtractWitness(run2)

	// Prove the root node (using ProveRoot since it's a single-level tree)
	rootProof := node.ProveRoot(wit1, wit2)

	// Verify the proof
	err := wizard.Verify(node.CompiledIOP, rootProof)
	assert.NoError(t, err, "tree root proof should verify")
}

// TestTreeAggregationStructure verifies the TreeAggregation data structure.
func TestTreeAggregationStructure(t *testing.T) {
	tree := &TreeAggregation{
		LeafCompiledIOP: &wizard.CompiledIOP{},
		Levels:          make([]*AggregationNode, 0),
	}

	assert.Equal(t, 0, tree.Depth())
	assert.NotNil(t, tree.RootCompiledIOP(), "root should fallback to leaf when no levels")

	// Add levels
	node1 := &AggregationNode{CompiledIOP: &wizard.CompiledIOP{}}
	node2 := &AggregationNode{CompiledIOP: &wizard.CompiledIOP{}}
	tree.Levels = append(tree.Levels, node1, node2)

	assert.Equal(t, 2, tree.Depth())
	assert.Equal(t, node2.CompiledIOP, tree.RootCompiledIOP())
}

// TestProveTreeTwoLeaves tests end-to-end tree aggregation with 2 leaves
// and a single tree level.
func TestProveTreeTwoLeaves(t *testing.T) {
	// Compile leaf IOP
	leafComp := wizard.Compile(defineLeafFunc("two"), leafCompilationSuite()...)

	// Build a 1-level tree manually (using test suite for speed)
	node := compileLevelWithSuite(leafComp, 0, "two-level", testTreeCompilationSuite())
	tree := &TreeAggregation{
		LeafCompiledIOP: leafComp,
		Levels:          []*AggregationNode{node},
	}

	// Generate two leaf witnesses
	stoppingRound := recursion.VortexQueryRound(leafComp) + 1

	run1 := wizard.RunProverUntilRound(leafComp, proveLeafFunc("two"), stoppingRound)
	wit1 := recursion.ExtractWitness(run1)

	run2 := wizard.RunProverUntilRound(leafComp, proveLeafFunc("two"), stoppingRound)
	wit2 := recursion.ExtractWitness(run2)

	// Prove the tree
	rootProof, err := tree.ProveTree([]recursion.Witness{wit1, wit2})
	require.NoError(t, err, "tree proving should succeed")

	// Verify the root proof
	err = wizard.Verify(tree.RootCompiledIOP(), rootProof)
	assert.NoError(t, err, "root proof should verify")
}

// TestProveTreeSingleLeaf tests tree aggregation with a single leaf
// (which gets duplicated internally).
func TestProveTreeSingleLeaf(t *testing.T) {
	// Compile leaf IOP
	leafComp := wizard.Compile(defineLeafFunc("single"), leafCompilationSuite()...)

	// Build a 1-level tree
	node := compileLevelWithSuite(leafComp, 0, "single-level", testTreeCompilationSuite())
	tree := &TreeAggregation{
		LeafCompiledIOP: leafComp,
		Levels:          []*AggregationNode{node},
	}

	// Generate only one leaf witness
	stoppingRound := recursion.VortexQueryRound(leafComp) + 1
	run1 := wizard.RunProverUntilRound(leafComp, proveLeafFunc("single"), stoppingRound)
	wit1 := recursion.ExtractWitness(run1)

	// Prove with single leaf (should duplicate internally)
	rootProof, err := tree.ProveTree([]recursion.Witness{wit1})
	require.NoError(t, err, "tree proving with single leaf should succeed")

	// Verify the root proof
	err = wizard.Verify(tree.RootCompiledIOP(), rootProof)
	assert.NoError(t, err, "root proof should verify")
}

// TestProveTreeErrors tests error conditions.
func TestProveTreeErrors(t *testing.T) {
	tree := &TreeAggregation{
		LeafCompiledIOP: &wizard.CompiledIOP{},
		Levels:          []*AggregationNode{},
	}

	// Empty witnesses
	_, err := tree.ProveTree(nil)
	assert.Error(t, err, "should error on nil witnesses")

	_, err = tree.ProveTree([]recursion.Witness{})
	assert.Error(t, err, "should error on empty witnesses")

	// No levels
	_, err = tree.ProveTree([]recursion.Witness{{}})
	assert.Error(t, err, "should error with no levels")
}

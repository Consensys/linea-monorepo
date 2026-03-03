package tree

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// CompileLevel compiles a single aggregation level that verifies 2 child
// proofs using recursion.DefineRecursionOf. The child proofs come from
// childComp (which may be the leaf CompiledIOP or a previous level's
// CompiledIOP).
//
// The isFinal parameter controls whether the last Vortex round uses
// PremarkAsSelfRecursed (for intermediate levels that will be further
// recursed) or IsLastRound=true (for the root level that will be
// wrapped in BN254).
func CompileLevel(childComp *wizard.CompiledIOP, levelIdx int, isFinal bool) *AggregationNode {
	node := &AggregationNode{ChildComp: childComp}

	define := func(b *wizard.Builder) {
		node.Recursion = recursion.DefineRecursionOf(
			b.CompiledIOP,
			childComp,
			recursion.Parameters{
				Name:                   fmt.Sprintf("tree-agg-level-%d", levelIdx),
				MaxNumProof:            2,
				WithExternalHasherOpts: true,
				SkipRecursionPrefix:    false,
				Subscript:              fmt.Sprintf("tree-level-%d", levelIdx),
				// Expose all public inputs from both children.
				// nil means "expose all".
				RestrictPublicInputs: nil,
			},
		)
	}

	var suite CompilationSuite
	if isFinal {
		suite = TreeAggregationFinalCompilationSuite()
	} else {
		suite = TreeAggregationCompilationSuite()
	}

	logrus.Infof("Compiling tree aggregation level %d (isFinal=%v)", levelIdx, isFinal)
	node.CompiledIOP = wizard.Compile(define, suite...)
	logrus.Infof("Compiled tree aggregation level %d: %d rounds", levelIdx, node.CompiledIOP.NumRounds())

	return node
}

// CompileTreeAggregation compiles the full binary tree aggregation structure.
// Starting from the leaf CompiledIOP (typically RecursionCompiledIOP from the
// execution prover), it builds maxDepth levels of pairwise aggregation.
//
// Each level verifies 2 proofs from the level below. The final (root) level
// uses a different compilation suite since its output will be wrapped in BN254
// rather than further recursed.
//
// For a tree supporting up to N leaf proofs:
//   - Level 0 aggregates leaves pairwise: N -> N/2
//   - Level 1 aggregates level 0 outputs: N/2 -> N/4
//   - Level k (root): 2 -> 1
//   - Total levels: ceil(log2(N))
func CompileTreeAggregation(leafComp *wizard.CompiledIOP, maxDepth int) *TreeAggregation {
	if maxDepth < 1 {
		panic("tree aggregation requires at least 1 level")
	}

	tree := &TreeAggregation{
		LeafCompiledIOP: leafComp,
		Levels:          make([]*AggregationNode, 0, maxDepth),
	}

	currentComp := leafComp
	for i := 0; i < maxDepth; i++ {
		isFinal := (i == maxDepth-1)
		node := CompileLevel(currentComp, i, isFinal)
		tree.Levels = append(tree.Levels, node)
		currentComp = node.CompiledIOP
	}

	return tree
}

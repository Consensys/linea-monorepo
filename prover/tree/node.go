// Package tree implements a binary tree aggregation system for KoalaBear
// wizard proofs. It aggregates multiple inner proofs (execution, data
// availability) into a single proof through a binary tree of recursion
// nodes, each verifying 2 child proofs using Vortex PCS on KoalaBear
// with Ring-SIS + Poseidon2.
package tree

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// AggregationNode represents a single level in the binary tree aggregation.
// Each node verifies exactly 2 child proofs from the level below (or leaf
// proofs at level 0) using recursion.DefineRecursionOf with MaxNumProof=2.
type AggregationNode struct {
	// Recursion is the recursion context for this node, configured to
	// verify 2 child proofs.
	Recursion *recursion.Recursion

	// CompiledIOP is the compiled IOP for this aggregation level,
	// compiled with Vortex PCS and self-recursion on KoalaBear.
	CompiledIOP *wizard.CompiledIOP

	// ChildComp is the compiled IOP of the child proofs being verified
	// at this level.
	ChildComp *wizard.CompiledIOP
}

// TreeAggregation represents the complete binary tree aggregation structure.
// Given N leaf proofs, it produces a single root proof through ceil(log2(N))
// levels of pairwise aggregation.
type TreeAggregation struct {
	// Levels stores the AggregationNode for each level of the tree.
	// Level 0 aggregates leaf proofs (execution/DA wizard proofs).
	// Level i aggregates proofs produced by level i-1.
	// All nodes at the same level share the same CompiledIOP since they
	// verify the same type of child proof.
	Levels []*AggregationNode

	// LeafCompiledIOP is the CompiledIOP of the leaf proofs
	// (e.g., execution wizard proofs from RecursionCompiledIOP).
	LeafCompiledIOP *wizard.CompiledIOP
}

// RootCompiledIOP returns the CompiledIOP of the root (topmost) level.
// This is the proof that will be wrapped in the BN254 final wrap circuit.
func (t *TreeAggregation) RootCompiledIOP() *wizard.CompiledIOP {
	if len(t.Levels) == 0 {
		return t.LeafCompiledIOP
	}
	return t.Levels[len(t.Levels)-1].CompiledIOP
}

// Depth returns the number of aggregation levels.
func (t *TreeAggregation) Depth() int {
	return len(t.Levels)
}

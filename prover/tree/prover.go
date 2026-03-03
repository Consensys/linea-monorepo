package tree

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// ProvePair proves a single aggregation node by verifying two child witnesses.
// It runs the wizard prover up to the vortex query round and extracts the
// recursion witness for the next level.
//
// Returns the wizard proof and the extracted recursion witness.
func (node *AggregationNode) ProvePair(left, right recursion.Witness) (wizard.Proof, recursion.Witness) {
	stoppingRound := recursion.VortexQueryRound(node.CompiledIOP) + 1

	run := wizard.RunProverUntilRound(
		node.CompiledIOP,
		node.Recursion.GetMainProverStep(
			[]recursion.Witness{left, right},
			&left, // padding: use left witness as filler
		),
		stoppingRound,
	)

	proof := run.ExtractProof()

	// Sanity-check the partial proof
	if err := wizard.VerifyUntilRound(node.CompiledIOP, proof, stoppingRound); err != nil {
		utils.Panic("tree aggregation: generated proof does not pass verifier at stopping round; %v", err.Error())
	}

	wit := recursion.ExtractWitness(run)
	return proof, wit
}

// ProveRoot proves the root (final) aggregation node. Unlike intermediate
// nodes, the root produces a complete wizard proof (not stopped at the
// vortex query round) since it will be wrapped in BN254 rather than
// further recursed.
func (node *AggregationNode) ProveRoot(left, right recursion.Witness) wizard.Proof {
	proof := wizard.Prove(
		node.CompiledIOP,
		node.Recursion.GetMainProverStep(
			[]recursion.Witness{left, right},
			&left,
		),
	)

	// Sanity-check the full proof
	if err := wizard.Verify(node.CompiledIOP, proof); err != nil {
		utils.Panic("tree aggregation root: generated proof does not pass verifier; %v", err.Error())
	}

	return proof
}

// ProveTree proves the entire binary tree aggregation from leaf witnesses
// to a single root proof.
//
// The leaf witnesses are the recursion.Witness values extracted from the
// execution prover (after running ProveInner on the initial wizard IOP
// and extracting via recursion.ExtractWitness).
//
// Returns the final root proof from the topmost aggregation level.
func (t *TreeAggregation) ProveTree(leafWitnesses []recursion.Witness) (wizard.Proof, error) {
	if len(leafWitnesses) == 0 {
		return wizard.Proof{}, fmt.Errorf("no leaf witnesses provided")
	}

	if len(t.Levels) == 0 {
		return wizard.Proof{}, fmt.Errorf("tree has no aggregation levels")
	}

	// Single leaf: just pass through to root
	if len(leafWitnesses) == 1 {
		logrus.Info("Single leaf witness, duplicating for root aggregation")
		leafWitnesses = append(leafWitnesses, leafWitnesses[0])
	}

	currentWitnesses := leafWitnesses

	// Process all levels except the last (which produces a full proof)
	for levelIdx := 0; levelIdx < len(t.Levels)-1; levelIdx++ {
		node := t.Levels[levelIdx]
		currentWitnesses = proveLevel(node, currentWitnesses, levelIdx)
	}

	// Final level: produce a complete proof
	rootNode := t.Levels[len(t.Levels)-1]

	// Pad to 2 if needed
	if len(currentWitnesses) == 1 {
		currentWitnesses = append(currentWitnesses, currentWitnesses[0])
	}
	if len(currentWitnesses) > 2 {
		return wizard.Proof{}, fmt.Errorf(
			"expected at most 2 witnesses at root level, got %d (increase tree depth)",
			len(currentWitnesses),
		)
	}

	logrus.Infof("Proving root level with %d witnesses", len(currentWitnesses))
	rootProof := rootNode.ProveRoot(currentWitnesses[0], currentWitnesses[1])
	return rootProof, nil
}

// proveLevel processes a single intermediate aggregation level, pairing up
// witnesses and producing output witnesses for the next level.
func proveLevel(node *AggregationNode, witnesses []recursion.Witness, levelIdx int) []recursion.Witness {
	// Pad to even count by duplicating the last witness
	if len(witnesses)%2 != 0 {
		logrus.Infof("Level %d: padding %d witnesses to %d", levelIdx, len(witnesses), len(witnesses)+1)
		witnesses = append(witnesses, witnesses[len(witnesses)-1])
	}

	numPairs := len(witnesses) / 2
	logrus.Infof("Proving tree level %d: %d pairs", levelIdx, numPairs)

	nextWitnesses := make([]recursion.Witness, 0, numPairs)
	for i := 0; i < len(witnesses); i += 2 {
		logrus.Infof("Level %d: proving pair %d/%d", levelIdx, i/2+1, numPairs)
		_, wit := node.ProvePair(witnesses[i], witnesses[i+1])
		nextWitnesses = append(nextWitnesses, wit)
	}

	return nextWitnesses
}

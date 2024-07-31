package aggregation

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark/backend/plonk"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/sirupsen/logrus"
)

// Make proof runs the prover of the aggregation circuit and returns the
// corresponding proof.
func MakeProof(
	setup *circuits.Setup,
	maxNbProof int,
	proofClaims []ProofClaimAssignment,
	publicInput fr.Element,
) (
	plonk.Proof,
	error,
) {

	logrus.Infof("Creating the assignment")
	assignment, err := AssignAggregationCircuit(
		maxNbProof,
		proofClaims,
		publicInput,
	)

	if err != nil {
		return nil, fmt.Errorf("while generating the aggregation circuit assignment: %w", err)
	}

	logrus.Infof("Running the prove-check")
	return circuits.ProveCheck(
		setup,
		assignment,
		emPlonk.GetNativeProverOptions(ecc.BN254.ScalarField(), setup.Circuit.Field()),
		emPlonk.GetNativeVerifierOptions(ecc.BN254.ScalarField(), setup.Circuit.Field()),
	)
}

// Assigns the proof using placeholders
func AssignAggregationCircuit(
	maxNbProof int,
	proofClaims []ProofClaimAssignment,
	publicInput fr.Element,
) (
	c *AggregationCircuit,
	err error,
) {

	c = &AggregationCircuit{
		ProofClaims:      make([]proofClaim, maxNbProof),
		DummyPublicInput: publicInput,
	}

	for i := range c.ProofClaims {
		if i < len(proofClaims) {
			c.ProofClaims[i], err = assignProofClaim(&proofClaims[i])
			if err != nil {
				return nil, fmt.Errorf(
					"while emulating the proof claim #%v (circ ID %v): %w",
					i, proofClaims[i].CircuitID, err,
				)
			}
		} else {
			// If we go over capacity, we should use the
			c.ProofClaims[i] = c.ProofClaims[len(proofClaims)-1]
		}
	}

	return c, nil
}

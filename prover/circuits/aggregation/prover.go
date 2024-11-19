package aggregation

import (
	"fmt"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark/backend/plonk"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Make proof runs the prover of the aggregation circuit and returns the
// corresponding proof.
func MakeProof(
	setup *circuits.Setup,
	maxNbProof int,
	proofClaims []ProofClaimAssignment,
	piInfo PiInfo,
	publicInput fr.Element,
) (
	plonk.Proof,
	error,
) {

	logrus.Infof("Creating the assignment")
	assignment, err := AssignAggregationCircuit(
		maxNbProof,
		proofClaims,
		piInfo,
		publicInput,
	)

	if err != nil {
		return nil, fmt.Errorf("while generating the aggregation circuit assignment: %w", err)
	}

	f, err := os.Open(".tmp/aggregation-allowed-vks.bin")
	var t test_utils.FakeTestingT
	require.NoError(t, err)
	vk := plonk.NewVerifyingKey(ecc.BLS12_377)
	var b [1]byte
	_, err = f.Read(b[:])
	require.NoError(t, err)
	assignment.verifyingKeys = make([]emVkey, b[0])
	for i := range assignment.verifyingKeys {
		_, err = vk.ReadFrom(f)
		require.NoError(t, err)
		assignment.verifyingKeys[i], err = emPlonk.ValueOfVerifyingKey[emFr, emG1, emG2](vk)
		require.NoError(t, err)
	}

	_, err = vk.ReadFrom(f)
	require.NoError(t, err)
	assignment.publicInputVerifyingKey, err = emPlonk.ValueOfVerifyingKey[emFr, emG1, emG2](vk)
	require.NoError(t, err)

	f.Close()

	logrus.Infof("Running the prove-check")
	return circuits.ProveCheck(
		setup,
		assignment,
		emPlonk.GetNativeProverOptions(ecc.BN254.ScalarField(), setup.Circuit.Field()),
		emPlonk.GetNativeVerifierOptions(ecc.BN254.ScalarField(), setup.Circuit.Field()),
	)
}

// Assigns the proof using placeholders
func AssignAggregationCircuit(maxNbProof int, proofClaims []ProofClaimAssignment, piInfo PiInfo, publicInput fr.Element) (c *Circuit, err error) {

	c = &Circuit{
		ProofClaims:                    make([]proofClaim, maxNbProof),
		PublicInputWitnessClaimIndexes: utils.ToVariableSlice(piInfo.ActualIndexes),
		PublicInput:                    publicInput,
	}

	if c.PublicInputProof, c.PublicInputWitness, err = piInfo.claim(); err != nil {
		return nil, fmt.Errorf("while emulating the PI proof claim: %w", err)
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

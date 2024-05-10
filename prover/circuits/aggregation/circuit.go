// The bw6circuit package provides an implementation of the BW6 proof aggregation
// circuit. This circuits can aggregate several PLONK proofs.
package aggregation

import (
	"fmt"

	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/native/sw_bls12377"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
)

// shorthand for the emulated types as this can get verbose very quickly with
// generics. `em` stands for emulated
type (
	emFr       = sw_bls12377.ScalarField
	emG1       = sw_bls12377.G1Affine
	emG2       = sw_bls12377.G2Affine
	emGT       = sw_bls12377.GT
	emProof    = emPlonk.Proof[emFr, emG1, emG2]
	emVkey     = emPlonk.VerifyingKey[emFr, emG1, emG2]
	emCircVKey = emPlonk.CircuitVerifyingKey[emFr, emG1]
	emWitness  = emPlonk.Witness[emFr]
	// emBaseVkey = emPlonk.BaseVerifyingKey[emFr, emG1, emG2]
)

// The AggregationCircuit is used to aggregate multiple execution proofs and
// aggregation proofs together.
type AggregationCircuit struct {

	// The list of claims to be provided to the circuit.
	ProofClaims []proofClaim `gnark:",secret"`
	// List of available verifying keys that are available to the circuit. This
	// is treated as a constant by the circuit.
	verifyingKeys []emVkey `gnark:"-"`

	// Dummy general public input
	DummyPublicInput frontend.Variable `gnark:",public"`
}

func (c *AggregationCircuit) Define(api frontend.API) error {
	// Verify the constraints the execution proofs
	err := verifyClaimBatch(api, c.verifyingKeys, c.ProofClaims)
	if err != nil {
		return fmt.Errorf("processing execution proofs: %w", err)
	}

	return err
}

// Instantiate a new AggregationCircuit from a list of verification keys and
// a maximal number of proofs. The function should only be called with the
// purpose of running `frontend.Compile` over it.
func AllocateAggregationCircuit(
	nbProofs int,
	verifyingKeys []plonk.VerifyingKey,
) (*AggregationCircuit, error) {

	var (
		err           error
		emVKeys       = make([]emVkey, len(verifyingKeys))
		csPlaceHolder = getPlaceHolderCS()
		proofClaims   = make([]proofClaim, nbProofs)
	)

	for i := range verifyingKeys {
		emVKeys[i], err = emPlonk.ValueOfVerifyingKey[emFr, emG1, emG2](verifyingKeys[i])
		if err != nil {
			return nil, fmt.Errorf("while converting the verifying key #%v (execution) into its emulated gnark version: %w", i, err)
		}
	}

	for i := range proofClaims {
		proofClaims[i] = allocatableClaimPlaceHolder(csPlaceHolder)
	}

	return &AggregationCircuit{
		verifyingKeys: emVKeys,
		ProofClaims:   proofClaims,
	}, nil

}

func verifyClaimBatch(api frontend.API, vks []emVkey, claims []proofClaim) error {
	verifier, err := emPlonk.NewVerifier[emFr, emG1, emG2, emGT](api)
	if err != nil {
		return fmt.Errorf("while instantiating the verifier: %w", err)
	}

	var (
		bvk       = vks[0].BaseVerifyingKey
		cvks      = make([]emCircVKey, len(vks))
		switches  = make([]frontend.Variable, len(claims))
		proofs    = make([]emProof, len(claims))
		witnesses = make([]emWitness, len(claims))
	)

	for i := range vks {
		cvks[i] = vks[i].CircuitVerifyingKey
	}

	for i := range claims {
		proofs[i] = claims[i].Proof
		switches[i] = claims[i].CircuitID
		witnesses[i] = claims[i].PublicInput
	}

	err = verifier.AssertDifferentProofs(bvk, cvks, switches, proofs, witnesses, emPlonk.WithCompleteArithmetic())
	if err != nil {
		return fmt.Errorf("AssertDifferentProofs returned an error: %w", err)
	}
	return nil
}

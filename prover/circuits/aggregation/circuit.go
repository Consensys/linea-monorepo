// The bw6circuit package provides an implementation of the BW6 proof aggregation
// circuit. This circuits can aggregate several PLONK proofs.
package aggregation

import (
	"errors"
	"fmt"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/native/sw_bls12377"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	"github.com/consensys/gnark/std/math/emulated"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal"
	"slices"
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

// The Circuit is used to aggregate multiple execution proofs and
// aggregation proofs together.
type Circuit struct {

	// The list of claims to be provided to the circuit.
	ProofClaims []proofClaim `gnark:",secret"`
	// List of available verifying keys that are available to the circuit. This
	// is treated as a constant by the circuit.
	verifyingKeys []emVkey `gnark:"-"`

	publicInputVerifyingKey        emVkey              `gnark:"-"`
	PublicInputProof               emProof             `gnark:",secret"`
	PublicInputWitness             emWitness           `gnark:",secret"` // ordered for the PI circuit
	PublicInputWitnessClaimIndexes []frontend.Variable `gnark:",secret"`

	// general public input
	PublicInput frontend.Variable `gnark:",public"`
}

func (c *Circuit) Define(api frontend.API) error {
	// match the public input with the public input of the PI circuit
	field, err := emulated.NewField[emFr](api)
	if err != nil {
		return err
	}
	internal.AssertSliceEquals(api, api.ToBinary(c.PublicInput, emFr{}.Modulus().BitLen()), field.ToBitsCanonical(&c.PublicInputWitness.Public[0]))

	vks := append(slices.Clone(c.verifyingKeys), c.publicInputVerifyingKey)
	piVkIndex := len(vks) - 1

	for i := range c.ProofClaims {
		api.AssertIsLessOrEqual(c.ProofClaims[i].CircuitID, len(c.verifyingKeys)-1) // make sure the prover can't sneak in an extra PI circuit
	}

	// create a lookup table of actual public inputs
	actualPI := make([]*logderivlookup.Table, (emFr{}).NbLimbs())
	for i := range actualPI {
		actualPI[i] = logderivlookup.New(api)
	}
	for _, claim := range c.ProofClaims {
		if len(claim.PublicInput.Public) != 1 {
			return errors.New("expected 1 public input per decompression/execution circuit")
		}
		pi := claim.PublicInput.Public[0]
		for i := range actualPI {
			actualPI[i].Insert(pi.Limbs[i])
		}
	}

	if len(c.PublicInputWitnessClaimIndexes)+1 != len(c.PublicInputWitness.Public) {
		return errors.New("expected the number of public inputs to match the number of public input witness claim indexes")
	}

	// verify that every valid input to the PI circuit is accounted for
	for i, actualI := range c.PublicInputWitnessClaimIndexes {
		hubPI := &c.PublicInputWitness.Public[i+1]
		isNonZero := api.Sub(1, field.IsZero(hubPI)) // if a PI is zero, due to preimage resistance we can infer that the PI circuit is not using it
		for j := range actualPI {
			internal.AssertEqualIf(api, isNonZero, actualPI[j].Lookup(actualI)[0], hubPI.Limbs[j])
		}
	}

	claims := append(slices.Clone(c.ProofClaims), proofClaim{
		CircuitID:   piVkIndex,
		Proof:       c.PublicInputProof,
		PublicInput: c.PublicInputWitness,
	})

	// Verify the constraints the execution proofs
	if err = verifyClaimBatch(api, vks, claims); err != nil {
		return fmt.Errorf("processing execution proofs: %w", err)
	}

	return err
}

// Instantiate a new Circuit from a list of verification keys and
// a maximal number of proofs. The function should only be called with the
// purpose of running `frontend.Compile` over it.
func AllocateCircuit(nbProofs int, allowedInputs []string, key plonk.VerifyingKey, verifyingKeys []plonk.VerifyingKey) (*Circuit, error) {

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

	return &Circuit{
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

package aggregation

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// Assignment collects all the arguments that are necessary to produce a circuit
// assignment for the BW6 aggregation circuit. As the number of required
// arguments is large, it is more convenient to pack them in a struct instead of
// passing them flat.
type ProofClaimAssignment struct {
	CircuitID          int
	Proof              plonk.Proof
	PublicInput        fr.Element
	VerifyingKeyShasum types.FullBytes32
}

// A proof claim represents a single proof to verify
type proofClaim struct {
	// The circuit ID is small integer summarizing which verification key needs
	// to be used for the verification of the circuit. Not that the value 0 is
	// reserved for the placeholder circuit.
	CircuitID frontend.Variable `gnark:",secret"`
	// The proof to verify
	Proof emProof `gnark:",secret"`
	// The public input to be provided to the proof.
	// @alex: this public input is not meant to become public input of the
	// present proof directly.
	PublicInput emWitness `gnark:",secret"`
}

// Generates a placeholder claim. Note that this placeholder claim is meant only
// for circuit allocation. In particular, they should not be used for assigning
// the proof claims that we use to fill the circuit
func allocatableClaimPlaceHolder(cs constraint.ConstraintSystem) proofClaim {
	return proofClaim{
		Proof:       emPlonk.PlaceholderProof[emFr, emG1, emG2](cs),
		PublicInput: emPlonk.PlaceholderWitness[emFr](cs),
	}
}

// Generates a placeholder claim from a place holder asset. This function is
// meant to be used for assigning the main circuit in the event where the
// not enough non-placeholder proofs have been provided to complete the circuit.
// In that case, we can use the ouput of this function to assign the remaining
// "slots" in the circuit.

// Convert a proof claim into a gnark assignment
func assignProofClaim(a *ProofClaimAssignment) (proofClaim, error) {

	var emptyProofClaim proofClaim

	emPi, err := emPlonk.ValueOfProof[emFr, emG1, emG2](a.Proof)
	if err != nil {
		return emptyProofClaim, fmt.Errorf("while emulating the proof over BLS: %w", err)
	}

	// We use the dummy circuit as a placeholder circuit to generate the witness.
	// It works because all of our circuit have a single public input
	aPlace := dummy.Assign(0, a.PublicInput)

	wit, err := frontend.NewWitness(aPlace, ecc.BLS12_377.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return emptyProofClaim, fmt.Errorf("while initializing the gnark witness: %w", err)
	}

	emWit, err := emPlonk.ValueOfWitness[emFr](wit)
	if err != nil {
		return emptyProofClaim, fmt.Errorf("while emulating the witness over BLS: %w", err)
	}

	return proofClaim{
		PublicInput: emWit,
		Proof:       emPi,
		CircuitID:   a.CircuitID,
	}, nil
}

type PiInfo struct {
	Proof         plonk.Proof
	PublicWitness witness.Witness
	ActualIndexes []int
}

func (i *PiInfo) claim() (proof emProof, witness emWitness, err error) {
	if witness, err = emPlonk.ValueOfWitness[emFr](i.PublicWitness); err != nil {
		return
	}
	proof, err = emPlonk.ValueOfProof[emFr, emG1, emG2](i.Proof)
	return
}

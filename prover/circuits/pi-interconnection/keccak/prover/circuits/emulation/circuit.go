// The bn254 package provides an implementation of a gnark circuit that can
// recursively verify a PLONK proof on the BW6 field with an arithmetization
// over the BN254 curve. This circuit is used as the final step before we
// submit a proof to Ethereum.
package emulation

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	frbw6 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bw6761"
	"github.com/consensys/gnark/std/math/emulated"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/dummy"
)

// shorthand for the emulated types as this can get verbose very quickly with
// generics. `em` stands for emulated
type (
	emFr       = sw_bw6761.ScalarField
	emG1       = sw_bw6761.G1Affine
	emG2       = sw_bw6761.G2Affine
	emGT       = sw_bw6761.GTEl
	emProof    = emPlonk.Proof[emFr, emG1, emG2]
	emCircVkey = emPlonk.CircuitVerifyingKey[emFr, emG1]
	emBaseVKey = emPlonk.BaseVerifyingKey[emFr, emG1, emG2]
	emWitness  = emPlonk.Witness[emFr]
)

// The outer-circuits converts a proof over the BW6-761 field into a proof over
// the BN254 field.
type CircuitEmulation struct {
	CircuitVkeys []emCircVkey      `gnark:"-"`
	BaseVKey     emBaseVKey        `gnark:"-"`
	Proof        emProof           `gnark:",secret:"`
	Witness      emWitness         `gnark:",secret:"`
	PublicInput  frontend.Variable `gnark:",public"`
	CircuitID    frontend.Variable `gnark:",secret:"`
}

func (c *CircuitEmulation) Define(api frontend.API) error {

	verifier, err := emPlonk.NewVerifier[emFr, emG1, emG2, emGT](api)
	if err != nil {
		return fmt.Errorf("while instantiating the verifier: %w", err)
	}

	var (
		proofs    = []emProof{c.Proof}
		witnesses = []emWitness{c.Witness}
		switches  = []frontend.Variable{c.CircuitID}
	)

	err = verifier.AssertDifferentProofs(c.BaseVKey, c.CircuitVkeys, switches, proofs, witnesses, emPlonk.WithCompleteArithmetic())
	if err != nil {
		return fmt.Errorf("while asserting the proof are correct: %w", err)
	}

	// @alex: the intent here is to show that public input and witness
	// represent an equal number but not in the same representation.
	f, err := emulated.NewField[emFr](api)
	if err != nil {
		return err
	}

	var (
		piBits   = f.ToBits(&c.Witness.Public[0])
		piNative = api.FromBinary(piBits[:fr.Bits]...)
	)

	for i := fr.Bits; i < len(piBits); i++ {
		api.AssertIsEqual(piBits[i], 0)
	}

	api.AssertIsEqual(piNative, c.PublicInput)

	return nil
}

// Produces a proof for the outer-proof outside on the BN field
func MakeProof(
	setup *circuits.Setup,
	circuitID int,
	innerProof plonk.Proof,
	publicInput fr.Element,
) (
	proof plonk.Proof,
	err error,
) {

	assignment, err := assignOuterCircuit(
		circuitID,
		innerProof,
		publicInput,
	)

	if err != nil {
		return nil, fmt.Errorf("while generating the aggregation circuit assignment: %w", err)
	}

	return circuits.ProveCheck(setup, assignment)
}

// Allocates a new outer-circuit that can be passed to `frontend.Compile`. The
// inner-cs is only needed to allocate the witness of the proof. So any circuit
// with a single public input will do.
func allocateOuterCircuit(
	innerVkeys []plonk.VerifyingKey,
) (*CircuitEmulation, error) {

	singlePiCs := singleInputCS()

	_innerBaseVKey, err := emPlonk.ValueOfBaseVerifyingKey[emFr, emG1, emG2](innerVkeys[0])
	if err != nil {
		return nil, fmt.Errorf("while emulating the Base VK: %w", err)
	}

	_innerCircVKeys := make([]emCircVkey, len(innerVkeys))
	for i := range innerVkeys {
		_innerCircVKeys[i], err = emPlonk.ValueOfCircuitVerifyingKey[emFr, emG1](innerVkeys[i])
		if err != nil {
			return nil, fmt.Errorf("while emulating the Circuit VK: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("while converting the verifying-key in emulated representation: %w", err)
	}

	return &CircuitEmulation{
		Proof:        emPlonk.PlaceholderProof[emFr, emG1, emG2](singlePiCs),
		BaseVKey:     _innerBaseVKey,
		CircuitVkeys: _innerCircVKeys,
		Witness:      emPlonk.PlaceholderWitness[emFr](singlePiCs),
	}, nil
}

// Produces an assignment for the outer-circuit toward proving satisfaction of
// the constraint system.
func assignOuterCircuit(
	circuitID int,
	innerProof plonk.Proof,
	publicInput fr.Element,
) (*CircuitEmulation, error) {

	// Convert x into a witness object

	var (
		piBytes = publicInput.Bytes()
		innerPI frbw6.Element
	)
	innerPI.SetBytes(piBytes[:])
	innerWitness := singleInputWitness(innerPI)

	// This converts the Plonk proofs and witness into equivalent gnark's
	// emulated variables representations.

	emulatedProof, err := emPlonk.ValueOfProof[emFr, emG1, emG2](innerProof)
	if err != nil {
		return nil, fmt.Errorf("while emulating the inner proof in the outer-circuit: %w", err)
	}

	emulatedWitness, err := emPlonk.ValueOfWitness[emFr](innerWitness)
	if err != nil {
		return nil, fmt.Errorf("while emulating the witness of the inner proof in the outer-circuit: %w", err)
	}

	return &CircuitEmulation{
		Proof:       emulatedProof,
		Witness:     emulatedWitness,
		PublicInput: publicInput,
		CircuitID:   circuitID,
	}, nil
}

// This is just a dummy placeholder circuit with a single public input which is
// convenient for us to provide for allocation since the actual circuit is very
// big and we would rather not keep it around in memory.
func singleInputCS() constraint.ConstraintSystem {
	ccs, err := dummy.MakeCS(0, ecc.BW6_761.ScalarField())
	if err != nil {
		panic(err)
	}
	return ccs
}

// This is a utility function allowing us to constraint a witness.Witness object
func singleInputWitness(x frbw6.Element) witness.Witness {
	a := dummy.Assign(0, x)
	w, _ := frontend.NewWitness(a, ecc.BW6_761.ScalarField(), frontend.PublicOnly())
	return w
}

package execution

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/circuits"
	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput"
	"github.com/sirupsen/logrus"

	"github.com/consensys/gnark/std/hash/mimc"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
)

// CircuitExecution for the outer-proof
type CircuitExecution struct {
	// The wizard verifier circuit
	WizardVerifier wizard.WizardVerifierCircuit
	// The extractor is not part of the circuit per se, but hold informations
	// that is used to extract the public inputs from the the WizardVerifier.
	// The extractor only needs to be provided during the definition of the
	// circuit and is omitted during the assignment of the circuit.
	extractor publicInput.FunctionalInputExtractor
	// The functional public inputs are the "actual" statement made by the
	// circuit. They are not part of the public input of the circuit for
	// a number of reasons involving efficiency and simplicity in the aggregation
	// process. What is the public input is their hash.
	FuncInputs FunctionalPublicInputSnark
	// The public input of the proof
	PublicInput frontend.Variable `gnark:",public"`
}

// Allocates the outer-proof circuit
func Allocate(comp *wizard.CompiledIOP, piExtractor *publicInput.FunctionalInputExtractor) CircuitExecution {
	wverifier, err := wizard.AllocateWizardCircuit(comp)
	if err != nil {
		panic(err)
	}
	return CircuitExecution{
		WizardVerifier: *wverifier,
	}
}

// assign the wizard proof to the outer circuit
func assign(
	comp *wizard.CompiledIOP,
	proof wizard.Proof,
	funcInputs FunctionalPublicInput,
) CircuitExecution {

	wizardVerifier := wizard.GetWizardVerifierCircuitAssignment(comp, proof)
	return CircuitExecution{
		WizardVerifier: *wizardVerifier,
		FuncInputs:     funcInputs.ToSnarkType(),
		PublicInput:    new(big.Int).SetBytes(funcInputs.Sum()),
	}
}

// Define of the wizard circuit
func (c *CircuitExecution) Define(api frontend.API) error {
	c.WizardVerifier.Verify(api)
	checkPublicInputs(
		api,
		&c.WizardVerifier,
		c.FuncInputs,
		c.extractor,
	)

	// Add missing public input check
	mimcHasher, _ := mimc.NewMiMC(api)
	api.AssertIsEqual(c.PublicInput, c.FuncInputs.Sum(api, &mimcHasher))
	return nil
}

func MakeProof(
	setup circuits.Setup,
	comp *wizard.CompiledIOP,
	wproof wizard.Proof,
	funcInputs FunctionalPublicInput,
) string {

	assignment := assign(comp, wproof, funcInputs)
	witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
	if err != nil {
		panic(err)
	}

	proof, err := plonk.Prove(
		setup.Circuit,
		setup.ProvingKey,
		witness,
		backend.WithSolverOptions(gkrmimc.SolverOpts(setup.Circuit)...),
		emPlonk.GetNativeProverOptions(ecc.BW6_761.ScalarField(), setup.Circuit.Field()),
	)
	if err != nil {
		panic(err)
	}

	logrus.Infof("generated outer-circuit proof `%++v` for input `%v`", proof, assignment.PublicInput.(*big.Int).String())

	// Sanity-check : the proof must pass
	{
		pubwitness, err := frontend.NewWitness(
			&assignment,
			ecc.BLS12_377.ScalarField(),
			frontend.PublicOnly(),
		)
		if err != nil {
			panic(err)
		}

		err = plonk.Verify(proof, setup.VerifyingKey, pubwitness, emPlonk.GetNativeVerifierOptions(ecc.BW6_761.ScalarField(), setup.Circuit.Field()))
		if err != nil {
			panic(err)
		}
	}

	// Write the serialized proof
	return circuits.SerializeProofRaw(proof)
}

package execution

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/circuits"
	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"

	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
)

// CircuitExecution for the outer-proof
type CircuitExecution struct {
	// The wizard verifier circuit
	WizardVerifier wizard.WizardVerifierCircuit
	// The public input of the proof
	PublicInput frontend.Variable `gnark:",public"`
	X5          frontend.Variable `gnark:",secret"`
}

// Allocates the outer-proof circuit
func Allocate(comp *wizard.CompiledIOP) CircuitExecution {
	wverifier, err := wizard.AllocateWizardCircuit(comp)
	if err != nil {
		panic(err)
	}
	return CircuitExecution{
		WizardVerifier: *wverifier,
	}
}

// assign the wizard proof to the outer circuit
func assign(comp *wizard.CompiledIOP, proof wizard.Proof, publicInput fr.Element) CircuitExecution {
	var x5 fr.Element
	x5.Exp(publicInput, big.NewInt(5))
	wizardVerifier := wizard.GetWizardVerifierCircuitAssignment(comp, proof)
	return CircuitExecution{
		WizardVerifier: *wizardVerifier,
		PublicInput:    publicInput,
		X5:             x5,
	}
}

// Define of the wizard circuit
func (c *CircuitExecution) Define(api frontend.API) error {
	logrus.Infof("defining the outer-proof circuit")
	x5 := api.Mul(c.PublicInput,
		c.PublicInput,
		c.PublicInput,
		c.PublicInput,
		c.PublicInput,
	)
	api.AssertIsEqual(c.X5, x5)
	c.WizardVerifier.Verify(api)
	logrus.Infof("ran successfully")
	return nil
}

func MakeProof(setup circuits.Setup, comp *wizard.CompiledIOP, wproof wizard.Proof, publicInput fr.Element) string {

	assignment := assign(comp, wproof, publicInput)
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

	logrus.Infof("generated outer-circuit proof `%++v` for input `%v`", proof, publicInput.String())

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

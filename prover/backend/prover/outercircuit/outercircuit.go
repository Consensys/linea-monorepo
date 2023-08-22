package outercircuit

import (
	"math/big"
	"path"

	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/plonkutil"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc/gkrmimc"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	"github.com/sirupsen/logrus"
)

// Circuit for the outer-proof
type Circuit struct {
	// The wizard verifier circuit
	WizardVerifier wizard.WizardVerifierCircuit
	// The public input of the proof
	PublicInput frontend.Variable `gnark:",public"`
	X5          frontend.Variable `gnark:",secret"`
}

// Allocates the outer-proof circuit
func Allocate(comp *wizard.CompiledIOP) Circuit {
	wverifier, err := wizard.AllocateWizardCircuit(comp)
	if err != nil {
		panic(err)
	}
	return Circuit{
		WizardVerifier: *wverifier,
	}
}

// assign the wizard proof to the outer circuit
func assign(comp *wizard.CompiledIOP, proof wizard.Proof, publicInput field.Element) Circuit {
	var x5 fr.Element
	x5.Exp(publicInput, big.NewInt(5))
	wizardVerifier := wizard.GetWizardVerifierCircuitAssignment(comp, proof)
	return Circuit{
		WizardVerifier: *wizardVerifier,
		PublicInput:    publicInput,
		X5:             x5,
	}
}

// Define of the wizard circuit
func (c *Circuit) Define(api frontend.API) error {
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

// Generates a SRS from a SRS
func GenSetupFromSRS(comp *wizard.CompiledIOP, srs *kzg.SRS) (pp plonkutil.Setup) {
	scs := buildCircuit(comp)
	return makeSetup(scs, srs)
}

// Generates the setup of the circuit
func GenSetupUnsafe(comp *wizard.CompiledIOP) (pp plonkutil.Setup) {

	scs := buildCircuit(comp)

	// Deterministic (and unsafe) SRS generation. 100M corresponds to the size
	// of Aztec's setup.
	numPointsSrs := 100_000_000

	logrus.Infof("generating the unsafe KZG setup for %v coefficients", numPointsSrs)
	srs, err := kzg.NewSRS(uint64(numPointsSrs), big.NewInt(42))
	if err != nil {
		panic(err)
	}
	logrus.Infof("successfully generated the unsafe KZG setup")

	return makeSetup(scs, srs)
}

// runs the PLONK setup over a circuit
func makeSetup(scs constraint.ConstraintSystem, srs *kzg.SRS) plonkutil.Setup {
	// Runs the deterministic setup of plonk.
	logrus.Infof("running the plonk setup")
	pk, vk, err := plonk.Setup(scs, srs)
	if err != nil {
		panic(err)
	}
	logrus.Info("successfully generated the setup")
	return plonkutil.Setup{PK: pk, VK: vk, SCS: scs}
}

// builds the circuit
func buildCircuit(comp *wizard.CompiledIOP) constraint.ConstraintSystem {
	circuit := Allocate(comp)

	logrus.Infof("compiling the circuit for the outer-proof ...")
	p := profile.Start(profile.WithPath("./profiling/outer-circuit/gnark.pprof"))
	scs, err := frontend.Compile(fr.Modulus(), scs.NewBuilder, &circuit, frontend.WithCapacity(1<<27))
	if err != nil {
		panic(err)
	}
	p.Stop()
	logrus.Infof("successfully compiled the outer-circuit, has %v constraints", scs.GetNbConstraints())
	return scs
}

func MakeProof(pp plonkutil.Setup, comp *wizard.CompiledIOP, wproof wizard.Proof, publicInput fr.Element) string {

	assignment := assign(comp, wproof, publicInput)
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		panic(err)
	}

	proof, err := plonk.Prove(pp.SCS, pp.PK, witness, backend.WithSolverOptions(gkrmimc.SolverOpts(pp.SCS)...))
	if err != nil {
		panic(err)
	}

	// Sanity-check : the proof must pass
	{
		pubwitness, err := frontend.NewWitness(
			&assignment,
			ecc.BN254.ScalarField(),
			frontend.PublicOnly(),
		)
		if err != nil {
			panic(err)
		}

		err = plonk.Verify(proof, pp.VK, pubwitness)
		if err != nil {
			panic(err)
		}
	}

	// Write the serialized proof
	return plonkutil.SerializeProof(proof)
}

// Generate the full setup
func GenerateAndExportSetupUnsafe(comp *wizard.CompiledIOP, folderOut string) {
	pp := GenSetupUnsafe(comp)
	plonkutil.ExportToFile(pp, path.Join(folderOut, "full"))
}

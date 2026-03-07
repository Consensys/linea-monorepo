// Package finalwrap provides a BN254 PLONK circuit that verifies a KoalaBear
// wizard proof using field emulation. This is the final stage of the proof
// pipeline, producing an Ethereum-compatible proof.
//
// The circuit wraps a KoalaBear wizard proof (from the tree aggregation root)
// into a BN254 PLONK proof. KoalaBear field operations are emulated using
// gnark's emulated arithmetic (koalagnark.API) with efficient single-limb
// representation since KoalaBear is a 31-bit field.
package finalwrap

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	gnarkprofile "github.com/consensys/gnark/profile"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// FinalWrapCircuit is a BN254 PLONK circuit that verifies a KoalaBear
// wizard proof. The wizard verifier uses koalagnark.API which auto-detects
// emulated mode (BN254 native field != KoalaBear) and performs all KoalaBear
// arithmetic via gnark's emulated field operations.
type FinalWrapCircuit struct {
	// WizardVerifier is the wizard verifier circuit that replays all
	// verifier steps of the KoalaBear wizard proof.
	WizardVerifier wizard.VerifierCircuit `gnark:",secret"`

	// PublicInput is the BN254-native public input for on-chain verification.
	// It is the hash of the aggregated functional public inputs, fitting
	// within the BN254 scalar field.
	PublicInput frontend.Variable `gnark:",public"`

	// PIDigestProfiler, if non-nil, is called after computePIDigest with
	// the number of BN254 constraints it added. Profiling only.
	PIDigestProfiler func(delta int) `gnark:"-"`
}

// Define implements frontend.Circuit. It verifies the KoalaBear wizard proof
// and asserts the public input consistency.
func (c *FinalWrapCircuit) Define(api frontend.API) error {
	// The wizard verifier auto-detects emulated mode in Verify():
	// - It creates a koalagnark.API which detects BN254 != KoalaBear
	// - It uses GnarkFSKoalagnark for Fiat-Shamir (Poseidon2 on KoalaBear via emulation)
	// - All field operations use emulated.KoalaBear with single-limb (31-bit) representation
	c.WizardVerifier.Verify(api)

	// Extract the wizard's public inputs and verify consistency with
	// the BN254 public input.
	koalaAPI := koalagnark.NewAPI(api)
	var piDigest frontend.Variable
	if c.PIDigestProfiler != nil {
		p := gnarkprofile.Start(gnarkprofile.WithNoOutput())
		piDigest = computePIDigest(api, koalaAPI, &c.WizardVerifier)
		p.Stop()
		c.PIDigestProfiler(p.NbConstraints())
	} else {
		piDigest = computePIDigest(api, koalaAPI, &c.WizardVerifier)
	}
	api.AssertIsEqual(piDigest, c.PublicInput)

	return nil
}

// computePIDigest extracts the wizard verifier's public inputs (KoalaBear
// field elements) and converts them to a single BN254 scalar. Since each
// KoalaBear element is 31 bits, we pack multiple elements into a single
// BN254 scalar (254 bits) using a simple hash-to-field approach.
//
// Public inputs may be base field elements (single KoalaBear) or extension
// field elements (4 KoalaBear coordinates). Both are packed into the digest.
func computePIDigest(api frontend.API, koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit) frontend.Variable {
	// Collect all public inputs from the wizard verifier
	pis := wvc.Spec.PublicInputs
	if len(pis) == 0 {
		return 0
	}

	// Pack the KoalaBear public inputs into BN254 limbs.
	// Each KoalaBear element is 31 bits; we can pack 8 of them into ~248 bits
	// which fits within BN254's 254-bit scalar field.
	//
	// Strategy: treat each KoalaBear PI as a limb and combine using a
	// powers-of-base polynomial evaluation in BN254.
	base := big.NewInt(1 << 31) // slightly larger than KoalaBear modulus
	result := frontend.Variable(0)

	for _, pi := range pis {
		if pi.Acc.IsBase() {
			// Base field: single KoalaBear element
			elem := wvc.GetPublicInput(api, pi.Name)
			nativeVal := koalaAPI.GetFrontendVariable(elem)
			// Horner's method: result = result * base + val
			result = api.Add(api.Mul(result, base), nativeVal)
		} else {
			// Extension field: 4 KoalaBear coordinates (B0.A0, B0.A1, B1.A0, B1.A1)
			ext := wvc.GetPublicInputExt(api, pi.Name)
			result = api.Add(api.Mul(result, base), koalaAPI.GetFrontendVariable(ext.B0.A0))
			result = api.Add(api.Mul(result, base), koalaAPI.GetFrontendVariable(ext.B0.A1))
			result = api.Add(api.Mul(result, base), koalaAPI.GetFrontendVariable(ext.B1.A0))
			result = api.Add(api.Mul(result, base), koalaAPI.GetFrontendVariable(ext.B1.A1))
		}
	}

	return result
}

// Allocate returns a FinalWrapCircuit configured for the given root
// CompiledIOP. The returned circuit is suitable for frontend.Compile.
func Allocate(rootComp *wizard.CompiledIOP) *FinalWrapCircuit {
	wverifier := wizard.AllocateWizardCircuit(rootComp, rootComp.NumRounds())
	return &FinalWrapCircuit{
		WizardVerifier: *wverifier,
	}
}

// MakeCS compiles the FinalWrapCircuit into a BN254 constraint system.
func MakeCS(rootComp *wizard.CompiledIOP) (constraint.ConstraintSystem, error) {
	circuit := Allocate(rootComp)
	logrus.Info("Compiling BN254 final wrap circuit...")
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, circuit)
	if err != nil {
		return nil, fmt.Errorf("compiling final wrap circuit: %w", err)
	}
	logrus.Infof("Final wrap circuit compiled: %d constraints", ccs.GetNbConstraints())
	return ccs, nil
}

// MakeProof generates a BN254 PLONK proof by verifying the root wizard proof
// inside the final wrap circuit.
func MakeProof(
	setup *circuits.Setup,
	rootComp *wizard.CompiledIOP,
	wizardProof wizard.Proof,
	publicInput fr.Element,
) (plonk.Proof, error) {

	assignment := assign(rootComp, wizardProof, publicInput)

	proof, err := circuits.ProveCheck(setup, assignment)
	if err != nil {
		return nil, fmt.Errorf("final wrap proof generation failed: %w", err)
	}

	return proof, nil
}

// assign creates a circuit assignment for the FinalWrapCircuit.
func assign(
	rootComp *wizard.CompiledIOP,
	proof wizard.Proof,
	publicInput fr.Element,
) *FinalWrapCircuit {

	wizardVerifier := wizard.AssignVerifierCircuit(rootComp, proof, rootComp.NumRounds())

	return &FinalWrapCircuit{
		WizardVerifier: *wizardVerifier,
		PublicInput:    new(big.Int).SetBytes(publicInput.Marshal()),
	}
}

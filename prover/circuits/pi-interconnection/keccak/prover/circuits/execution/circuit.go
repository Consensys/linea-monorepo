package execution

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
)

// CircuitExecution for the outer-proof
type CircuitExecution struct {
	// LimitlessMode is set to true if the outer proof is generated for the
	// limitless prover mode.
	LimitlessMode bool `gnark:"-"`
	// CongloVK is used when the [LimitlessMode] is on and is helps checking
	// the validity of the inner-proofs verification-key public input.
	CongloVK [2]field.Element
	// VKMerkleRoot is used when the [LimitlessMode] is on and is helps checking
	// the validity of the inner-proofs verification-key merkle root public
	// input.
	VKMerkleRoot field.Element
	// The wizard verifier circuit
	WizardVerifier wizard.VerifierCircuit `gnark:",secret"`
	// The functional public inputs are the "actual" statement made by the
	// circuit. They are not part of the public input of the circuit for
	// a number of reasons involving efficiency and simplicity in the aggregation
	// process. What is the public input is their hash.
	FuncInputs FunctionalPublicInputSnark `gnark:",secret"`
	// The public input of the proof
	PublicInput frontend.Variable `gnark:",public"`
}

// Define of the wizard circuit
func (c *CircuitExecution) Define(api frontend.API) error {

	c.WizardVerifier.HasherFactory = gkrmimc.NewHasherFactory(api)
	c.WizardVerifier.FS = fiatshamir.NewGnarkFiatShamir(api, c.WizardVerifier.HasherFactory)

	c.WizardVerifier.Verify(api)
	checkPublicInputs(
		api,
		&c.WizardVerifier,
		c.FuncInputs,
		c.LimitlessMode, // limitlessMode = false
	)

	if c.LimitlessMode {
		c.checkLimitlessConglomerationCompletion(api)
	}

	// Add missing public input check
	mimcHasher, _ := mimc.NewMiMC(api)
	api.AssertIsEqual(c.PublicInput, c.FuncInputs.Sum(api, &mimcHasher))
	return nil
}

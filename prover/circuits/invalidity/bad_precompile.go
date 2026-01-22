package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const MAX_L2_LOGS = 16

// BadPrecompileCircuit defines the circuit for the transaction with a bad precompile.
type BadPrecompileCircuit struct {
	// The wizard verifier circuit - this is the main witness
	WizardVerifier wizard.VerifierCircuit `gnark:",secret"`

	// These fields are derived from wizard public inputs during Define()
	// They are excluded from witness generation via gnark:"-" tag
	stateRootHash     frontend.Variable    `gnark:"-"`
	txHash            [2]frontend.Variable `gnark:"-"`
	fromAddress       frontend.Variable    `gnark:"-"`
	hashBadPrecompile frontend.Variable    `gnark:"-"`
	NbL2Logs          frontend.Variable    `gnark:"-"`

	InvalidityType frontend.Variable
}

func (circuit *BadPrecompileCircuit) Allocate(config Config) {
	wverifier := wizard.AllocateWizardCircuit(config.Zkevm.WizardIOP, 0)
	circuit.WizardVerifier = *wverifier
}

func (circuit *BadPrecompileCircuit) Define(api frontend.API) error {
	// set the hasher factory and the fiatshamir scheme
	circuit.WizardVerifier.HasherFactory = gkrmimc.NewHasherFactory(api)
	circuit.WizardVerifier.FS = fiatshamir.NewGnarkFiatShamir(api, circuit.WizardVerifier.HasherFactory)

	circuit.WizardVerifier.Verify(api)

	// Get public inputs from the wizard verifier and store them
	circuit.stateRootHash = circuit.WizardVerifier.GetPublicInput(api, "StateRootHash")
	circuit.txHash[0] = circuit.WizardVerifier.GetPublicInput(api, "TxHash_Hi")
	circuit.txHash[1] = circuit.WizardVerifier.GetPublicInput(api, "TxHash_Lo")
	circuit.fromAddress = circuit.WizardVerifier.GetPublicInput(api, "FromAddress")
	circuit.hashBadPrecompile = circuit.WizardVerifier.GetPublicInput(api, "HashBadPrecompile")
	circuit.NbL2Logs = circuit.WizardVerifier.GetPublicInput(api, "NbL2Logs")

	if circuit.InvalidityType == BadPrecompile {
		// check that hashBadPrecompile is non-zero
		api.AssertIsDifferent(circuit.hashBadPrecompile, 0)
	} else if circuit.InvalidityType == TooManyLogs {
		// check that NbL2Logs is greater than MAX_L2_LOGS
		api.AssertIsLessOrEqual(MAX_L2_LOGS+1, circuit.NbL2Logs)
	}
	return nil
}

func (circuit *BadPrecompileCircuit) Assign(assi AssigningInputs) {
	circuit.WizardVerifier = *wizard.AssignVerifierCircuit(assi.Zkevm.WizardIOP, assi.ZkevmWizardProof, 0)
	circuit.InvalidityType = int(assi.InvalidityType)
}

func (c *BadPrecompileCircuit) FunctionalPublicInputs() FunctionalPublicInputsGnark {
	return FunctionalPublicInputsGnark{
		FromAddress:   c.fromAddress,
		TxHash:        c.txHash,
		StateRootHash: c.stateRootHash,
	}
}

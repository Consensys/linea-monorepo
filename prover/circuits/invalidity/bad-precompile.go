package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// BadPrecompileCircuit defines the circuit for the transaction with a bad precompile.
type BadPrecompileCircuit struct {
	// Indicates if the transaction has a bad precompile.
	HasBadPrecompile frontend.Variable
	//NbL2Logs
	NbL2Logs frontend.Variable
	// The wizard verifier circuit
	WizardVerifier wizard.VerifierCircuit `gnark:",secret"`
}

func (circuit *BadPrecompileCircuit) Allocate(config Config) {
	wverifier := wizard.AllocateWizardCircuit(config.zkevm.WizardIOP, 0)
	circuit.WizardVerifier = *wverifier
}

func (circuit *BadPrecompileCircuit) Define(api frontend.API) error {

	// set the hasher factory and the fiatshamir scheme
	circuit.WizardVerifier.HasherFactory = gkrmimc.NewHasherFactory(api)

	circuit.WizardVerifier.FS = fiatshamir.NewGnarkFiatShamir(api, circuit.WizardVerifier.HasherFactory)

	circuit.WizardVerifier.Verify(api)

	// check that NbL2Logs and HashBadPrecompile are consistent with the wizard verifier circuit
	checkPublicInputs(
		api,
		&circuit.WizardVerifier,
		circuit.FunctionalPublicInputs(),
	)

	return nil
}

func (circuit *BadPrecompileCircuit) Assign(assi AssigningInputs) {
	_ = *wizard.AssignVerifierCircuit(assi.ZkevmWizardIOP, assi.ZkevmWizardProof, 0)
}

func (circuit *BadPrecompileCircuit) FunctionalPublicInputs() FunctionalPublicInputsGnark {
	return FunctionalPublicInputsGnark{
		HasBadPrecompile: circuit.HasBadPrecompile,
	}
}

func checkPublicInputs(
	api frontend.API,
	wvc *wizard.VerifierCircuit,
	gnarkFuncInp FunctionalPublicInputsGnark,
) {

}

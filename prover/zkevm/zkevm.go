package zkevm

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/serialization"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/statemanager"
)

// ZkEvm defines the wizard responsible for proving execution of the zk
type ZkEvm struct {
	// Arithmetization definition function. Generated during the compilation
	// process.
	arithmetization arithmetization.Arithmetization
	// Keccak module in use. Generated during the compilation process.
	keccak keccak.Module
	// State manager module in use. Generated during the compilation process.
	stateManager statemanager.StateManagerLegacy

	// PublicInput gives access to the public inputs of the wizard-IOP and is
	// used to access them to define the outer-circuit.
	PublicInput publicInput.PublicInput

	// Contains the actual wizard-IOP compiled object. This object is called to
	// generate the inner-proof.
	WizardIOP *wizard.CompiledIOP
}

// Instantiate a new ZkEvm instance. The function itself is a noop. Call
// `Compile` to actually instantate the zkEVM proof scheme. It only dispatches
// the configuration across its components.
func NewZkEVM(
	settings Settings, // Settings for the zkEVM
) *ZkEvm {
	return &ZkEvm{
		arithmetization: arithmetization.Arithmetization{
			Settings: &settings.Arithmetization,
		},
		stateManager: statemanager.StateManagerLegacy{
			Settings: &settings.Statemanager,
		},
		keccak: keccak.Module{
			Settings: &settings.Keccak,
		},
	}
}

// Compiles instantiate the prover scheme over the parameterized zkEVM. The
// function can take a bit of time to complete. It will populate the zkEVM
// struct and needs to be called before running the prover of the inner-proof.
// It returns the (populated) receiver.
func (z *ZkEvm) Compile(suite compilationSuite, vm wizard.VersionMetadata) *ZkEvm {
	z.WizardIOP = wizard.Compile(z.define, suite...)
	z.WizardIOP.BootstrapFiatShamir(vm, serialization.SerializeCompiledIOP)
	return z
}

// Prove assigns and runs the inner-prover of the zkEVM and then, it returns the
// inner-proof
func (z *ZkEvm) ProveInner(input *Witness) wizard.Proof {
	return wizard.Prove(z.WizardIOP, z.prove(input))
}

// Verify verifies the inner-proof of the zkEVM
func (z *ZkEvm) VerifyInner(proof wizard.Proof) error {
	return wizard.Verify(z.WizardIOP, proof)
}

// The define function of the zkEVM define module. This function is unexported
// and should not be exported. The user should instead use the "Compile"
// function. This function is meant to be passed as a closure to the
// wizard.Compile function. Thus, this is an internal.
func (z *ZkEvm) define(b *wizard.Builder) {

	// Run the arithmetization function first. Always. Because the other modules
	// are "building" on top of it.
	z.arithmetization.Define(b)

	// Recall, the state-manager is not feature-gated for the full prover but it
	// is disabled for the partial and the checker
	if z.stateManager.Settings.Enabled {
		z.stateManager.Define(b.CompiledIOP)
	}

	// If the keccak module is enabled, set the module.
	if z.keccak.Settings.Enabled {
		var providers []generic.GenericByteModule
		nbKeccakF := z.keccak.Settings.MaxNumKeccakf
		z.keccak.Define(b.CompiledIOP, providers, nbKeccakF)
	}
}

// Returns a prover function for the zkEVM module. The resulting function is
// aimed to be passed to the wizard.Prove function.
func (z *ZkEvm) prove(input *Witness) (prover wizard.ProverStep) {
	return func(run *wizard.ProverRuntime) {

		// Assigns the arithmetization module. From Corset. Must be done first
		// because the following modules use the content of these columns to
		// assign themselves.
		arithmetization.Assign(run, input.ExecTracesFPath)

		// Assign the state-manager module
		if z.stateManager.Settings.Enabled {
			z.stateManager.Assign(run, input.SMTraces)
		}

		// Assign the Keccak module
		if z.keccak.Settings.Enabled {
			z.keccak.AssignKeccak(run)
		}
	}
}

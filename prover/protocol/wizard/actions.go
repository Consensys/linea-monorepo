package wizard

import "github.com/consensys/gnark/frontend"

// ProverAction represents an action to be performed by the prover.
// They have to be registered in the [CompiledIOP] via the
// [CompiledIOP.RegisterProverAction]
type ProverAction interface {
	// Run executes the ProverAction over a [ProverRuntime]
	Run(*ProverRuntime)
}

// VerifierAction represents an action to be performed by the verifier of the
// protocol. Usually, this is used to represent verifier checks. They can be
// registered via [CompiledIOP.RegisterVerifierAction].
type VerifierAction interface {
	// Run executes the VerifierAction over a [VerifierRuntime] it returns an
	// error.
	Run(*VerifierRuntime) error
	// RunGnark is as Run but in a gnark circuit. Instead, of the returning an
	// error the function enforces the passing of the verifier's checks.
	RunGnark(frontend.API, *WizardVerifierCircuit)
}

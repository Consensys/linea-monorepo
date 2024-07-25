package wizard

import "github.com/consensys/gnark/frontend"

type runtimeProverAction struct {
	RuntimeProverAction
	metadata *metadata
}

type runtimeVerifierAction struct {
	RuntimeVerifierAction
	metadata *metadata
}

// RuntimeProverAction represents an action to be performed by the prover.
// They have to be registered in the [CompiledIOP] via the
// [CompiledIOP.RegisterProverAction]
type RuntimeProverAction interface {
	// Run executes the ProverAction over a [ProverRuntime]
	Run(*RuntimeProver)
}

// RuntimeVerifierAction represents an action to be performed by the verifier of the
// protocol. Usually, this is used to represent verifier checks. They can be
// registered via [CompiledIOP.RegisterVerifierAction].
type RuntimeVerifierAction interface {
	// Run executes the VerifierAction over a [VerifierRuntime] it returns an
	// error.
	Run(Runtime) error
	// RunGnark is as Run but in a gnark circuit. Instead, of the returning an
	// error the function enforces the passing of the verifier's checks.
	RunGnark(frontend.API, GnarkRuntime)
}

func (api *API) AddRuntimeProverAction(round int, pa RuntimeProverAction) *runtimeProverAction {

	var (
		registeredPA = runtimeProverAction{
			RuntimeProverAction: pa,
			metadata:            api.newMetadata(),
		}
	)

	api.runtimeProverActions.addToRound(round, registeredPA)
	return &registeredPA
}

func (api *API) AddRuntimeVerifierAction(round int, va RuntimeVerifierAction) *runtimeVerifierAction {

	var (
		registeredVA = runtimeVerifierAction{
			RuntimeVerifierAction: va,
			metadata:              api.newMetadata(),
		}
	)

	api.runtimeVerifierActions.addToRound(round, registeredVA)
	return &registeredVA
}

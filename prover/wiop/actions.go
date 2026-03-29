package wiop

// Action represents an action to be performed during protocol execution,
// by either the prover or the verifier.
type Action interface {
	// Run executes the action against the given [Runtime].
	Run(Runtime)
}

// actionWrapper adapts a plain func(Runtime) to the [Action] interface.
type actionWrapper struct {
	step func(Runtime)
}

// Run implements [Action].
func (w actionWrapper) Run(run Runtime) { w.step(run) }

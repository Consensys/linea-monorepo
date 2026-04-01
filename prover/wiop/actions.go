package wiop

// Action represents an action to be performed during protocol execution,
// by either the prover or the verifier.
type Action interface {
	// Run executes the action against the given [Runtime].
	Run(Runtime)
}

package wiop

// Action represents a prover-side computation to be performed during protocol
// execution.
type Action interface {
	// Run executes the action against the given [Runtime].
	Run(Runtime)
}

// VerifierAction represents a verifier-side check to be performed during
// protocol execution.
type VerifierAction interface {
	// Check executes the verification step against the given [Runtime] and
	// returns an error if the check fails.
	Check(Runtime) error
}

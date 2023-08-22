package column

type Status int

const (
	// Neither of the cases below
	Ignored Status = iota
	// Sent to the oracle, implictly the verifier receive a commitment
	Committed
	// Sent to the verifier as part of the proof
	Proof
	// Part of the setup, evaluated at compilation-time
	Precomputed
	// Public input
	PublicInput
	// Verifying key
	VerifyingKey
	// Defined by the verifier
	VerifierDefined
)

func (s Status) String() string {
	switch s {
	case Committed:
		return "COMMITTED"
	case Ignored:
		return "IGNORED"
	case Proof:
		return "PROOF"
	case Precomputed:
		return "PRECOMPUTED"
	case PublicInput:
		return "PUBLIC_INPUT"
	case VerifyingKey:
		return "VERIFYING_KEY"
	case VerifierDefined:
		return "VERIFIER_DEFINED"
	}
	panic("unknown")
}

// Returns true if the column is visible to the verifier
func (s Status) IsPublic() bool {
	switch s {
	case Proof, PublicInput, VerifyingKey, VerifierDefined:
		return true
	default:
		return false
	}
}

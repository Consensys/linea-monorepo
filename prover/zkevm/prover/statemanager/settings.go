package statemanager

// Settings specifies the parameters with which to instantiate the statemanager
// module.
type Settings struct {
	// In production mode, the statemanager is not optional but for the partial
	// prover and the checker it is not activated.
	Enabled        bool
	MaxMerkleProof int
}

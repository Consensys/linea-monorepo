package invalidity

import (
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

type Response struct {
	// the corresponding request file
	Request
	// Proof in 0x prefixed hexstring format
	Proof string `json:"proof"`

	// The shasum of the verifier key to use to verify the proof. This is used
	// by the aggregation circuit to identify the circuit ID to use in the proof.
	VerifyingKeyShaSum string `json:"verifyingKeyShaSum"`

	// PublicInput is the final value public input of the current proof. This
	// field is used for debugging in case one of the proofs don't pass at the
	// aggregation level.
	PublicInput types.Bytes32 `json:"publicInput"`

	ProverVersion string `json:"proverVersion"`
}

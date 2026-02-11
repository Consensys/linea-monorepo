package invalidity

import (
	"github.com/consensys/linea-monorepo/prover/utils/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type Response struct {
	// the transaction that was attempted to be included in the current aggregation
	Transaction *ethtypes.Transaction
	// signer of the transaction
	Signer types.EthAddress `json:"signer"`
	// hash of the transaction (before signing)
	TxHash string `json:"txHash"`
	// the corresponding request file
	Request
	// Proof in 0x prefixed hexstring format
	Proof string `json:"proof"`

	// The shasum of the verifier key to use to verify the proof. This is used
	// by the aggregation circuit to identify the circuit ID to use in the proof.
	VerifyingKeyShaSum string `json:"verifyingKeyShaSum"`

	// PublicInput is the final value public input of the current proof. This

	PublicInput types.Bls12377Fr `json:"publicInput"`

	ProverVersion string `json:"proverVersion"`

	FtxRollingHash types.Bls12377Fr `json:"ftxRollingHash"`
}

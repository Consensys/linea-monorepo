package bridge

import (
	"github.com/ethereum/go-ethereum/crypto"
)

// Returns the selector for the msg reception receipt
func MsgConfirmSelector() []byte {
	const signature string = "addL1L2MessageHashes(bytes32[])"
	// Encode in utf8, hash it, keep the 4 leftmost bytes
	return crypto.Keccak256([]byte(signature))[:4]
}

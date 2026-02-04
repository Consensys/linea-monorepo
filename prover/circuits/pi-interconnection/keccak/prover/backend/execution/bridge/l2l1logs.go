package bridge

import (
	"github.com/ethereum/go-ethereum/crypto"
)

// Returns the L2L1LogHashes from a list of logs and abi.encode them

// Returns true iff the log is an l2l1 event

// ABI encode an L2L1 log

// Returns the selector for the L2 L1 logs
func L2L1Topic0() (res [32]byte) {
	// signature of the event
	const signature string = "MessageSent(address,address,uint256,uint256,uint256,bytes,bytes32)"
	hashed := crypto.Keccak256([]byte(signature))
	copy(res[:], hashed)
	return res
}

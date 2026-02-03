package bridge

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

// Batch reception index
func BatchReceptionIndex(logs []types.Log, l2BridgeAddress common.Address) []uint16 {
	res := []uint16{}

	for i, log := range logs {
		// logging just in case it can help
		logrus.Tracef("Parsing logs #%v : %v\n", i, logs)

		if !IsL1L2MessageHashesAddedToInbox(log, l2BridgeAddress) {
			continue
		}

		// Push the txIndex
		res = append(res, utils.ToUint16(log.TxIndex))
	}
	return res
}

func IsL1L2MessageHashesAddedToInbox(log types.Log, l2BridgeAddress common.Address) bool {

	// Check the address
	if log.Address != l2BridgeAddress {
		return false
	}

	// Check the topic0
	if len(log.Topics) == 0 || log.Topics[0] != L1L2MessageHashesAddedToInboxTopic0() {
		return false
	}

	return true
}

// Is the L1L2MessageHashes event topic 0
func L1L2MessageHashesAddedToInboxTopic0() common.Hash {
	const signature string = "L1L2MessageHashesAddedToInbox(bytes32[])"
	// Encode in utf8, hash it keep the 4 leftmost bytes
	return crypto.Keccak256Hash([]byte(signature))
}

// Returns the selector for the msg reception receipt
func MsgConfirmSelector() []byte {
	const signature string = "addL1L2MessageHashes(bytes32[])"
	// Encode in utf8, hash it, keep the 4 leftmost bytes
	return crypto.Keccak256([]byte(signature))[:4]
}

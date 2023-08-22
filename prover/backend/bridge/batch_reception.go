package bridge

import (
	"github.com/consensys/accelerated-crypto-monorepo/utils"
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
		logrus.Tracef("parsing logs #%v : %v\n", i, logs)
		// Check the address
		if log.Address != l2BridgeAddress {
			continue
		}

		// Check the topic0
		if len(log.Topics) == 0 || log.Topics[0] != L1L2MessageHashesAddedToInboxTopic0() {
			continue
		}

		// Push the txIndex
		res = append(res, uint16(log.TxIndex))
	}
	return res
}

// return true if the transactions can potentially be a batch reception
// index this check cannot check that the transaction passed
func MustLookLikeABatchReceptionTx(tx types.Transaction, l2BridgeAddress common.Address) {

	if tx.To() == nil || len(tx.Data()) < 4 {
		utils.Panic("could not parse the txData or the to address of the transaction")
	}

	to := *tx.To()
	var selector = [4]byte{}
	copy(selector[:], tx.Data())

	if to != l2BridgeAddress {
		utils.Panic("the recipient of the alleged bridge transaction must be the bridge address : %v,found %v", l2BridgeAddress.Hex(), to.Hex())
	}

	expectedSelector := [4]byte{}
	copy(expectedSelector[:], MsgConfirmSelector())
	if selector != expectedSelector {
		utils.Panic("the selector must be the expected one : %x found %v", expectedSelector, selector)
	}

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

package bridge

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

const rollingHashEventSignature = "RollingHashUpdated(uint256,bytes32)"

var rollingHashUpdateTopic0 = GetRollingHashUpdateTopic0()

// Bridge-event emitted post-compression release to notify the prover that a
// message has been received on L2 from L1.
type RollingHashUpdated struct {
	MessageNumber int64             `json:"messageNumber"`
	RollingHash   types.FullBytes32 `json:"rollingHash"`
}

// Scan the list of logs and returns the parsed  `RollingHashUpdated` events
// that originated from the l2BridgeAddress.
func ExtractRollingHashUpdated(logs []ethtypes.Log, l2BridgeAddress common.Address) []RollingHashUpdated {

	logrus.Tracef("Filtering the following logs for rolling hash updated: %++v", logs)
	res := []RollingHashUpdated{}
	for _, log := range logs {

		if !IsRollingHashUpdated(log, l2BridgeAddress) {
			continue
		}

		res = append(res, RollingHashUpdated{
			MessageNumber: new(big.Int).SetBytes(log.Topics[1][:]).Int64(),
			RollingHash:   types.FullBytes32(log.Topics[2]),
		})
	}

	return res
}

func IsRollingHashUpdated(log ethtypes.Log, l2BridgeAddress common.Address) bool {

	if len(log.Topics) == 0 || log.Topics[0] != rollingHashUpdateTopic0 {
		return false
	}

	if log.Address != l2BridgeAddress {
		return false
	}

	return true
}

func (l *RollingHashUpdated) AsTypesLog(l2BridgeAddress common.Address) ethtypes.Log {
	return ethtypes.Log{
		Address: l2BridgeAddress,
		Topics: []common.Hash{
			rollingHashUpdateTopic0,
			common.Hash(utils.AsBigEndian32Bytes(int(l.MessageNumber))),
			common.Hash(l.RollingHash),
		},
	}
}

// Returns the selector for RollingHashUpdated event.
func GetRollingHashUpdateTopic0() (res [32]byte) {
	// signature of the event
	hashed := crypto.Keccak256([]byte(rollingHashEventSignature))
	copy(res[:], hashed)
	return res
}

package public_input

import (
	"errors"

	"github.com/consensys/linea-monorepo/prover/utils"
)

type Execution struct {
	L2MsgHashes            [][32]byte
	FinalStateRootHash     [32]byte
	FinalBlockNumber       uint64
	FinalBlockTimestamp    uint64
	FinalRollingHash       [32]byte
	FinalRollingHashNumber uint64
}

type ExecutionSerializable struct {
	L2MsgHashes            []string `json:"l2MsgHashes"`
	FinalStateRootHash     string   `json:"finalStateRootHash"`
	FinalBlockNumber       uint64   `json:"finalBlockNumber"`
	FinalBlockTimestamp    uint64   `json:"finalBlockTimestamp"`
	FinalRollingHash       string   `json:"finalRollingHash"`
	FinalRollingHashNumber uint64   `json:"finalRollingHashNumber"`
}

func (e ExecutionSerializable) Decode() (decoded Execution, err error) {
	decoded = Execution{
		L2MsgHashes:            make([][32]byte, len(e.L2MsgHashes)),
		FinalBlockNumber:       e.FinalBlockNumber,
		FinalBlockTimestamp:    e.FinalBlockTimestamp,
		FinalRollingHashNumber: e.FinalRollingHashNumber,
	}

	fillWithHex := func(dst []byte, src string) {
		var d []byte
		if d, err = utils.HexDecodeString(src); err != nil {
			return
		}
		if len(d) > len(dst) {
			err = errors.New("decoded bytes too long")
		}
		n := len(dst) - len(d)
		copy(dst[n:], d)
		for n > 0 {
			n--
			dst[n] = 0
		}
	}

	for i, hex := range e.L2MsgHashes {
		if fillWithHex(decoded.L2MsgHashes[i][:], hex); err != nil {
			return
		}
	}
	if fillWithHex(decoded.FinalStateRootHash[:], e.FinalStateRootHash); err != nil {
		return
	}

	fillWithHex(decoded.FinalRollingHash[:], e.FinalRollingHash)
	return
}

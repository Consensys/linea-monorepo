package public_input

import (
	"encoding/binary"
	"errors"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"hash"

	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

type Execution struct {
	L2MsgHashes                 [][32]byte
	L2MessageServiceAddr        types.EthAddress
	ChainID                     uint64
	InitialBlockTimestamp       uint64
	FinalStateRootHash          [32]byte
	FinalBlockNumber            uint64
	FinalBlockTimestamp         uint64
	FinalRollingHashUpdate      [32]byte
	FinalRollingHashMsgNumber   uint64
	InitialRollingHashUpdate    [32]byte
	InitialRollingHashMsgNumber uint64
	DataChecksum                [32]byte
	L2MessageHashes             [][32]byte
	InitialStateRootHash        [32]byte
	InitialBlockNumber          uint64
}

type ExecutionSerializable struct {
	ChainID                     uint64           `json:"chainID"`
	L2MessageServiceAddr        types.EthAddress `json:"l2MessageServiceAddr"`
	InitialBlockTimestamp       uint64           `json:"initialBlockTimestamp"`
	L2MsgHashes                 []string         `json:"l2MsgHashes"`
	FinalStateRootHash          string           `json:"finalStateRootHash"`
	FinalBlockNumber            uint64           `json:"finalBlockNumber"`
	FinalBlockTimestamp         uint64           `json:"finalBlockTimestamp"`
	InitialRollingHashUpdate    string           `json:"initialRollingHash"`
	InitialRollingHashMsgNumber uint64           `json:"initialRollingHashNumber"`
	FinalRollingHashUpdate      string           `json:"finalRollingHash"`
	FinalRollingHashMsgNumber   uint64           `json:"finalRollingHashNumber"`
}

func (e ExecutionSerializable) Decode() (decoded Execution, err error) {
	decoded = Execution{
		InitialBlockTimestamp:       e.InitialBlockTimestamp,
		L2MsgHashes:                 make([][32]byte, len(e.L2MsgHashes)),
		FinalBlockNumber:            e.FinalBlockNumber,
		FinalBlockTimestamp:         e.FinalBlockTimestamp,
		FinalRollingHashMsgNumber:   e.FinalRollingHashMsgNumber,
		InitialRollingHashMsgNumber: e.InitialRollingHashMsgNumber,
		L2MessageServiceAddr:        e.L2MessageServiceAddr,
		ChainID:                     uint64(e.ChainID),
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

	if fillWithHex(decoded.InitialRollingHashUpdate[:], e.InitialRollingHashUpdate); err != nil {
		return
	}

	fillWithHex(decoded.FinalRollingHashUpdate[:], e.FinalRollingHashUpdate)
	return
}

func (pi *Execution) Sum(hsh hash.Hash) []byte {
	if hsh == nil {
		hsh = mimc.NewMiMC()
	}

	hsh.Reset()
	for i := range pi.L2MessageHashes {
		hsh.Write(pi.L2MessageHashes[i][:16])
		hsh.Write(pi.L2MessageHashes[i][16:])
	}
	l2MessagesSum := hsh.Sum(nil)

	hsh.Reset()

	hsh.Write(pi.DataChecksum[:])
	hsh.Write(l2MessagesSum)
	hsh.Write(pi.FinalStateRootHash[:])

	writeNum(hsh, pi.FinalBlockNumber)
	writeNum(hsh, pi.FinalBlockTimestamp)
	hsh.Write(pi.FinalRollingHashUpdate[:16])
	hsh.Write(pi.FinalRollingHashUpdate[16:])
	writeNum(hsh, pi.FinalRollingHashMsgNumber)
	hsh.Write(pi.InitialStateRootHash[:])
	writeNum(hsh, pi.InitialBlockNumber)
	writeNum(hsh, pi.InitialBlockTimestamp)
	hsh.Write(pi.InitialRollingHashUpdate[:16])
	hsh.Write(pi.InitialRollingHashUpdate[16:])
	writeNum(hsh, pi.InitialRollingHashMsgNumber)
	writeNum(hsh, pi.ChainID)
	hsh.Write(pi.L2MessageServiceAddr[:])

	return hsh.Sum(nil)

}

func (pi *Execution) SumAsField() field.Element {

	var (
		sumBytes = pi.Sum(nil)
		sum      = new(field.Element).SetBytes(sumBytes)
	)

	return *sum
}

func writeNum(hsh hash.Hash, n uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], n)
	hsh.Write(b[:])
}

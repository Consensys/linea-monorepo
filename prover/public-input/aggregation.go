package public_input

import (
	"hash"

	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/linea-monorepo/prover/utils"
	"golang.org/x/crypto/sha3"
)

// Aggregation collects all the field that are used to construct the public
// input of the finalization proof.
type Aggregation struct {
	FinalShnarf                             string
	ParentAggregationFinalShnarf            string
	ParentStateRootHash                     string
	ParentAggregationLastBlockTimestamp     uint
	FinalTimestamp                          uint
	LastFinalizedBlockNumber                uint
	FinalBlockNumber                        uint
	LastFinalizedL1RollingHash              string
	L1RollingHash                           string
	LastFinalizedL1RollingHashMessageNumber uint
	L1RollingHashMessageNumber              uint
	L2MsgRootHashes                         []string
	L2MsgMerkleTreeDepth                    int
}

func (p Aggregation) Sum(hsh hash.Hash) []byte {
	if hsh == nil {
		hsh = sha3.NewLegacyKeccak256()
	}

	writeHex := func(hex string) {
		b, err := utils.HexDecodeString(hex)
		if err != nil {
			panic(err)
		}
		hsh.Write(b)
	}

	writeInt := func(i int) {
		b := utils.FmtInt32Bytes(i)
		hsh.Write(b[:])
	}

	hsh.Reset()

	for _, hex := range p.L2MsgRootHashes {
		writeHex(hex)
	}

	l2Msgs := hsh.Sum(nil)

	hsh.Reset()
	writeHex(p.ParentAggregationFinalShnarf)
	writeHex(p.FinalShnarf)
	writeInt(int(p.ParentAggregationLastBlockTimestamp))
	writeInt(int(p.FinalTimestamp))
	writeInt(int(p.LastFinalizedBlockNumber))
	writeInt(int(p.FinalBlockNumber))
	writeHex(p.LastFinalizedL1RollingHash)
	writeHex(p.L1RollingHash)
	writeInt(int(p.LastFinalizedL1RollingHashMessageNumber))
	writeInt(int(p.L1RollingHashMessageNumber))
	writeInt(p.L2MsgMerkleTreeDepth)
	hsh.Write(l2Msgs)

	// represent canonically as a bn254 scalar
	var x bn254fr.Element
	x.SetBytes(hsh.Sum(nil))

	res := x.Bytes()

	return res[:]

}

// GetPublicInputHex computes the public input of the finalization proof
func (p Aggregation) GetPublicInputHex() string {
	return utils.HexEncodeToString(p.Sum(sha3.NewLegacyKeccak256()))
}

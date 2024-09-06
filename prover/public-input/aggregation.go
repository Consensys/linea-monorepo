package public_input

import (
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/rangecheck"
	"github.com/consensys/zkevm-monorepo/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
	"golang.org/x/crypto/sha3"
	"hash"
	"slices"
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

// AggregationFPI holds the same info as public_input.Aggregation, except in parsed form
type AggregationFPI struct {
	ParentShnarf             [32]byte
	NbDecompression          uint64
	InitialStateRootHash     [32]byte
	InitialBlockNumber       uint64
	InitialBlockTimestamp    uint64
	InitialRollingHash       [32]byte
	InitialRollingHashNumber uint64
	ChainID                  uint64 // for now we're forcing all executions to have the same chain ID
	L2MessageServiceAddr     types.EthAddress
	NbL2Messages             uint64 // TODO not used in hash. delete if not necessary
	L2MsgMerkleTreeRoots     [][32]byte
	//FinalStateRootHash       [32]byte		redundant: incorporated into shnarf
	FinalBlockNumber       uint64
	FinalBlockTimestamp    uint64
	FinalRollingHash       [32]byte
	FinalRollingHashNumber uint64
	FinalShnarf            [32]byte
	L2MsgMerkleTreeDepth   int
}

func (pi *AggregationFPI) ToSnarkType() AggregationFPISnark {
	s := AggregationFPISnark{
		AggregationFPIQSnark: AggregationFPIQSnark{
			InitialBlockNumber:       pi.InitialBlockNumber,
			InitialBlockTimestamp:    pi.InitialBlockTimestamp,
			InitialRollingHash:       [32]frontend.Variable{},
			InitialRollingHashNumber: pi.InitialRollingHashNumber,
			InitialStateRootHash:     pi.InitialStateRootHash[:],

			NbDecompression:      pi.NbDecompression,
			ChainID:              pi.ChainID,
			L2MessageServiceAddr: pi.L2MessageServiceAddr[:],
		},
		L2MsgMerkleTreeRoots:   make([][32]frontend.Variable, len(pi.L2MsgMerkleTreeRoots)),
		FinalBlockNumber:       pi.FinalBlockNumber,
		FinalBlockTimestamp:    pi.FinalBlockTimestamp,
		FinalRollingHashNumber: pi.FinalRollingHashNumber,
		L2MsgMerkleTreeDepth:   pi.L2MsgMerkleTreeDepth,
	}

	utils.Copy(s.FinalRollingHash[:], pi.FinalRollingHash[:])
	utils.Copy(s.InitialRollingHash[:], pi.InitialRollingHash[:])
	utils.Copy(s.ParentShnarf[:], pi.ParentShnarf[:])
	utils.Copy(s.FinalShnarf[:], pi.FinalShnarf[:])

	for i := range s.L2MsgMerkleTreeRoots {
		utils.Copy(s.L2MsgMerkleTreeRoots[i][:], pi.L2MsgMerkleTreeRoots[i][:])
	}

	return s
}

type AggregationFPIQSnark struct {
	ParentShnarf             [32]frontend.Variable
	NbDecompression          frontend.Variable
	InitialStateRootHash     frontend.Variable
	InitialBlockNumber       frontend.Variable
	InitialBlockTimestamp    frontend.Variable
	InitialRollingHash       [32]frontend.Variable
	InitialRollingHashNumber frontend.Variable
	ChainID                  frontend.Variable // for now we're forcing all executions to have the same chain ID
	L2MessageServiceAddr     frontend.Variable // 20 bytes
}

type AggregationFPISnark struct {
	AggregationFPIQSnark
	NbL2Messages         frontend.Variable // TODO not used in hash. delete if not necessary
	L2MsgMerkleTreeRoots [][32]frontend.Variable
	// FinalStateRootHash     frontend.Variable redundant: incorporated into final shnarf
	FinalBlockNumber       frontend.Variable
	FinalBlockTimestamp    frontend.Variable
	FinalRollingHash       [32]frontend.Variable
	FinalRollingHashNumber frontend.Variable
	FinalShnarf            [32]frontend.Variable
	L2MsgMerkleTreeDepth   int
}

// NewAggregationFPI does NOT set all fields, only the ones covered in public_input.Aggregation
func NewAggregationFPI(fpi *Aggregation) (s *AggregationFPI, err error) {
	s = &AggregationFPI{
		InitialBlockNumber:       uint64(fpi.LastFinalizedBlockNumber),
		InitialBlockTimestamp:    uint64(fpi.ParentAggregationLastBlockTimestamp),
		InitialRollingHashNumber: uint64(fpi.LastFinalizedL1RollingHashMessageNumber),
		L2MsgMerkleTreeRoots:     make([][32]byte, len(fpi.L2MsgRootHashes)),
		FinalBlockNumber:         uint64(fpi.FinalBlockNumber),
		FinalBlockTimestamp:      uint64(fpi.FinalTimestamp),
		FinalRollingHashNumber:   uint64(fpi.L1RollingHashMessageNumber),
		L2MsgMerkleTreeDepth:     fpi.L2MsgMerkleTreeDepth,
	}

	if err = copyFromHex(s.InitialStateRootHash[:], fpi.ParentStateRootHash); err != nil {
		return
	}
	if err = copyFromHex(s.FinalRollingHash[:], fpi.L1RollingHash); err != nil {
		return
	}
	if err = copyFromHex(s.InitialRollingHash[:], fpi.LastFinalizedL1RollingHash); err != nil {
		return
	}
	if err = copyFromHex(s.ParentShnarf[:], fpi.ParentAggregationFinalShnarf); err != nil {
		return
	}
	if err = copyFromHex(s.FinalShnarf[:], fpi.FinalShnarf); err != nil {
		return
	}

	for i := range s.L2MsgMerkleTreeRoots {
		if err = copyFromHex(s.L2MsgMerkleTreeRoots[i][:], fpi.L2MsgRootHashes[i]); err != nil {
			return
		}
	}
	return
}

func (pi *AggregationFPISnark) Sum(api frontend.API, hash keccak.BlockHasher) [32]frontend.Variable {
	// number of hashes: 12
	sum := hash.Sum(nil,
		pi.ParentShnarf,
		pi.FinalShnarf,
		utils.ToBytes(api, pi.InitialBlockTimestamp),
		utils.ToBytes(api, pi.FinalBlockTimestamp),
		utils.ToBytes(api, pi.InitialBlockNumber),
		utils.ToBytes(api, pi.FinalBlockNumber),
		pi.InitialRollingHash,
		pi.FinalRollingHash,
		utils.ToBytes(api, pi.InitialRollingHashNumber),
		utils.ToBytes(api, pi.FinalRollingHashNumber),
		utils.ToBytes(api, pi.L2MsgMerkleTreeDepth),
		hash.Sum(nil, pi.L2MsgMerkleTreeRoots...),
	)

	// turn the hash into a bn254 element
	var res [32]frontend.Variable
	copy(res[:], utils.ReduceBytes[emulated.BN254Fr](api, sum[:]))
	return res
}

func (pi *AggregationFPIQSnark) RangeCheck(api frontend.API) {
	rc := rangecheck.New(api)
	for _, v := range append(slices.Clone(pi.InitialRollingHash[:]), pi.ParentShnarf[:]...) {
		rc.Check(v, 8)
	}
	// not checking L2MsgServiceAddr as its range is never assumed in the pi circuit
	// not checking NbDecompressions as the NewRange in the pi circuit range checks it; TODO do it here instead
}

func copyFromHex(dst []byte, src string) error {
	b, err := utils.HexDecodeString(src)
	if err != nil {
		return err
	}
	copy(dst[len(dst)-len(b):], b) // panics if src is too long
	return nil
}

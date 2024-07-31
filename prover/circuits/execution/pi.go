package execution

import (
	"encoding/binary"
	"github.com/consensys/gnark/frontend"
	gnarkHash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/rangecheck"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal"
	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
	"hash"
	"slices"
)

// FunctionalPublicInputQSnark the information on this execution that cannot be extracted from other input in the same aggregation batch
type FunctionalPublicInputQSnark struct {
	DataChecksum           frontend.Variable
	L2MessageHashes        internal.Var32Slice // TODO range check
	FinalStateRootHash     frontend.Variable
	FinalBlockNumber       frontend.Variable
	FinalBlockTimestamp    frontend.Variable
	FinalRollingHash       [32]frontend.Variable // TODO range check
	FinalRollingHashNumber frontend.Variable
}

type FunctionalPublicInputSnark struct {
	FunctionalPublicInputQSnark
	InitialStateRootHash     frontend.Variable
	InitialBlockNumber       frontend.Variable
	InitialBlockTimestamp    frontend.Variable
	InitialRollingHash       [32]frontend.Variable
	InitialRollingHashNumber frontend.Variable
	ChainID                  frontend.Variable
	L2MessageServiceAddr     frontend.Variable
}

type FunctionalPublicInput struct {
	DataChecksum             [32]byte
	L2MessageHashes          [][32]byte
	MaxNbL2MessageHashes     int
	FinalStateRootHash       [32]byte
	FinalBlockNumber         uint64
	FinalBlockTimestamp      uint64
	FinalRollingHash         [32]byte
	FinalRollingHashNumber   uint64
	InitialStateRootHash     [32]byte
	InitialBlockNumber       uint64
	InitialBlockTimestamp    uint64
	InitialRollingHash       [32]byte
	InitialRollingHashNumber uint64
	ChainID                  uint64
	L2MessageServiceAddr     types.EthAddress
}

// RangeCheck checks that values are within range
func (pi *FunctionalPublicInputQSnark) RangeCheck(api frontend.API) {
	// the length of the l2msg slice is range checked in Concat; no need to do it here; TODO do it here instead
	rc := rangecheck.New(api)
	for _, v := range pi.L2MessageHashes.Values {
		for i := range v {
			rc.Check(v[i], 8)
		}
	}
	for i := range pi.FinalRollingHash {
		rc.Check(pi.FinalRollingHash[i], 8)
	}
}

func (pi *FunctionalPublicInputSnark) Sum(api frontend.API, hsh gnarkHash.FieldHasher) frontend.Variable {
	finalRollingHash := internal.CombineBytesIntoElements(api, pi.FinalRollingHash)
	initialRollingHash := internal.CombineBytesIntoElements(api, pi.InitialRollingHash)

	hsh.Reset()
	for _, v := range pi.L2MessageHashes.Values { // it has to be zero padded
		vc := internal.CombineBytesIntoElements(api, v)
		hsh.Write(vc[0], vc[1])
	}
	l2MessagesSum := hsh.Sum()

	hsh.Reset()
	hsh.Write(pi.DataChecksum, l2MessagesSum,
		pi.FinalStateRootHash, pi.FinalBlockNumber, pi.FinalBlockTimestamp, finalRollingHash[0], finalRollingHash[1], pi.FinalRollingHashNumber,
		pi.InitialStateRootHash, pi.InitialBlockNumber, pi.InitialBlockTimestamp, initialRollingHash[0], initialRollingHash[1], pi.InitialRollingHashNumber,
		pi.ChainID, pi.L2MessageServiceAddr)

	return hsh.Sum()
}

func (pi *FunctionalPublicInput) ToSnarkType() FunctionalPublicInputSnark {
	res := FunctionalPublicInputSnark{
		FunctionalPublicInputQSnark: FunctionalPublicInputQSnark{
			DataChecksum:           slices.Clone(pi.DataChecksum[:]),
			L2MessageHashes:        internal.NewSliceOf32Array(pi.L2MessageHashes, pi.MaxNbL2MessageHashes),
			FinalStateRootHash:     slices.Clone(pi.FinalStateRootHash[:]),
			FinalBlockNumber:       pi.FinalBlockNumber,
			FinalBlockTimestamp:    pi.FinalBlockTimestamp,
			FinalRollingHashNumber: pi.FinalRollingHashNumber,
		},
		InitialStateRootHash:     slices.Clone(pi.InitialStateRootHash[:]),
		InitialBlockNumber:       pi.InitialBlockNumber,
		InitialBlockTimestamp:    pi.InitialBlockTimestamp,
		InitialRollingHashNumber: pi.InitialRollingHashNumber,
		ChainID:                  pi.ChainID,
		L2MessageServiceAddr:     slices.Clone(pi.L2MessageServiceAddr[:]),
	}
	internal.Copy(res.FinalRollingHash[:], pi.FinalRollingHash[:])
	internal.Copy(res.InitialRollingHash[:], pi.InitialRollingHash[:])

	return res
}

func (pi *FunctionalPublicInput) Sum() []byte { // all mimc; no need to provide a keccak hasher
	hsh := mimc.NewMiMC()
	// TODO incorporate length too? Technically not necessary

	var zero [1]byte
	for i := range pi.L2MessageHashes {
		hsh.Write(pi.L2MessageHashes[i][:16])
		hsh.Write(pi.L2MessageHashes[i][16:])
	}
	nbZeros := pi.MaxNbL2MessageHashes - len(pi.L2MessageHashes)
	if nbZeros < 0 {
		panic("too many L2 messages")
	}
	for i := 0; i < nbZeros; i++ {
		hsh.Write(zero[:])
		hsh.Write(zero[:])
	}
	l2MessagesSum := hsh.Sum(nil)

	hsh.Reset()

	hsh.Write(pi.DataChecksum[:])
	hsh.Write(l2MessagesSum)
	hsh.Write(pi.FinalStateRootHash[:])

	writeNum(hsh, pi.FinalBlockNumber)
	writeNum(hsh, pi.FinalBlockTimestamp)
	hsh.Write(pi.FinalRollingHash[:16])
	hsh.Write(pi.FinalRollingHash[16:])
	writeNum(hsh, pi.FinalRollingHashNumber)
	hsh.Write(pi.InitialStateRootHash[:])
	writeNum(hsh, pi.InitialBlockNumber)
	writeNum(hsh, pi.InitialBlockTimestamp)
	hsh.Write(pi.InitialRollingHash[:16])
	hsh.Write(pi.InitialRollingHash[16:])
	writeNum(hsh, pi.InitialRollingHashNumber)
	writeNum(hsh, pi.ChainID)
	hsh.Write(pi.L2MessageServiceAddr[:])

	return hsh.Sum(nil)

}

func writeNum(hsh hash.Hash, n uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], n)
	hsh.Write(b[:])
}

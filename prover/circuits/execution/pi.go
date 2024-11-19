package execution

import (
	"encoding/binary"
	"fmt"
	"hash"
	"slices"

	"github.com/consensys/gnark/frontend"
	gnarkHash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/rangecheck"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// FunctionalPublicInputQSnark the information on this execution that cannot be
// extracted from other input in the same aggregation batch
type FunctionalPublicInputQSnark struct {
	DataChecksum           frontend.Variable
	L2MessageHashes        L2MessageHashes
	InitialBlockTimestamp  frontend.Variable
	FinalStateRootHash     frontend.Variable
	FinalBlockNumber       frontend.Variable
	FinalBlockTimestamp    frontend.Variable
	FinalRollingHash       [32]frontend.Variable
	FinalRollingHashNumber frontend.Variable
}

// L2MessageHashes is a wrapper for [Var32Slice] it is use to instantiate the
// sequence of L2MessageHash that we extract from the arithmetization. The
// reason we need a wrapper here is because we hash the L2MessageHashes in a
// specific way.
type L2MessageHashes internal.Var32Slice

// NewL2MessageHashes constructs a new var slice
func NewL2MessageHashes(v [][32]frontend.Variable, max int) L2MessageHashes {
	return L2MessageHashes(internal.NewSliceOf32Array(v, max))
}

// CheckSum returns the hash of the [L2MessageHashes]. The encoding is done as
// follows:
//
//   - each L2 hash is decomposed in a hi and lo part: each over 16 bytes
//   - they are sequentially hashed in the following order: (hi_0, lo_0, hi_1, lo_1 ...)
//
// The function also performs a consistency check to ensure that the length of
// the slice if consistent with the number of non-zero elements. And the function
// also ensures that the non-zero elements are all packed at the beginning of the
// struct. The function returns zero if the slice encodes zero message hashes
// (this is what happens if no L2 message events are emitted during the present
// execution frame).
//
// @alex: it would be nice to make that function compatible with the GKR hasher
// factory though in practice this function will only create 32 calls to the
// MiMC permutation which makes it a non-issue.
func (l *L2MessageHashes) CheckSumMiMC(api frontend.API) frontend.Variable {

	var (
		// sumIsUsed is used to count the number of non-zero hashes that we
		// found in l. It is to be tested against l.Length.
		sumIsUsed = frontend.Variable(0)
		res       = frontend.Variable(0)
	)

	for i := range l.Values {
		var (
			hi     = internal.Pack(api, l.Values[i][:16], 128, 8)[0]
			lo     = internal.Pack(api, l.Values[i][16:], 128, 8)[0]
			isUsed = api.Sub(
				1,
				api.Mul(
					api.IsZero(hi),
					api.IsZero(lo),
				),
			)
		)

		tmpRes := mimc.GnarkBlockCompression(api, res, hi)
		tmpRes = mimc.GnarkBlockCompression(api, tmpRes, lo)

		res = api.Select(isUsed, tmpRes, res)
		sumIsUsed = api.Add(sumIsUsed, isUsed)
	}

	api.AssertIsEqual(sumIsUsed, l.Length)
	return res
}

type FunctionalPublicInputSnark struct {
	FunctionalPublicInputQSnark
	InitialStateRootHash     frontend.Variable
	InitialBlockNumber       frontend.Variable
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

	var (
		finalRollingHash   = internal.CombineBytesIntoElements(api, pi.FinalRollingHash)
		initialRollingHash = internal.CombineBytesIntoElements(api, pi.InitialRollingHash)
		l2MessagesSum      = pi.L2MessageHashes.CheckSumMiMC(api)
	)

	hsh.Reset()
	hsh.Write(pi.DataChecksum, l2MessagesSum,
		pi.FinalStateRootHash, pi.FinalBlockNumber, pi.FinalBlockTimestamp, finalRollingHash[0], finalRollingHash[1], pi.FinalRollingHashNumber,
		pi.InitialStateRootHash, pi.InitialBlockNumber, pi.InitialBlockTimestamp, initialRollingHash[0], initialRollingHash[1], pi.InitialRollingHashNumber,
		pi.ChainID, pi.L2MessageServiceAddr)

	return hsh.Sum()
}

func (pi *FunctionalPublicInput) ToSnarkType() (FunctionalPublicInputSnark, error) {
	res := FunctionalPublicInputSnark{
		FunctionalPublicInputQSnark: FunctionalPublicInputQSnark{
			DataChecksum:           slices.Clone(pi.DataChecksum[:]),
			L2MessageHashes:        L2MessageHashes(internal.NewSliceOf32Array(pi.L2MessageHashes, pi.MaxNbL2MessageHashes)),
			InitialBlockTimestamp:  pi.InitialBlockTimestamp,
			FinalStateRootHash:     slices.Clone(pi.FinalStateRootHash[:]),
			FinalBlockNumber:       pi.FinalBlockNumber,
			FinalBlockTimestamp:    pi.FinalBlockTimestamp,
			FinalRollingHashNumber: pi.FinalRollingHashNumber,
		},
		InitialStateRootHash:     slices.Clone(pi.InitialStateRootHash[:]),
		InitialBlockNumber:       pi.InitialBlockNumber,
		InitialRollingHashNumber: pi.InitialRollingHashNumber,
		ChainID:                  pi.ChainID,
		L2MessageServiceAddr:     slices.Clone(pi.L2MessageServiceAddr[:]),
	}
	utils.Copy(res.FinalRollingHash[:], pi.FinalRollingHash[:])
	utils.Copy(res.InitialRollingHash[:], pi.InitialRollingHash[:])

	var err error
	if nbMsg := len(pi.L2MessageHashes); nbMsg > pi.MaxNbL2MessageHashes {
		err = fmt.Errorf("has %d L2 message hashes but a maximum of %d is allowed", nbMsg, pi.MaxNbL2MessageHashes)
	}

	return res, err
}

func (pi *FunctionalPublicInput) Sum(hsh hash.Hash) []byte {
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

func (pi *FunctionalPublicInput) SumAsField() field.Element {

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

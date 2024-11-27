package execution

import (
	"fmt"
	"github.com/consensys/gnark/frontend"
	gnarkHash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/rangecheck"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// FunctionalPublicInputQSnark the information on this execution that cannot be
// extracted from other input in the same aggregation batch
type FunctionalPublicInputQSnark struct {
	DataChecksum                frontend.Variable
	L2MessageHashes             L2MessageHashes
	InitialBlockTimestamp       frontend.Variable
	FinalStateRootHash          frontend.Variable
	FinalBlockNumber            frontend.Variable
	FinalBlockTimestamp         frontend.Variable
	InitialRollingHashUpdate    [32]frontend.Variable
	InitialRollingHashMsgNumber frontend.Variable
	FinalRollingHashUpdate      [32]frontend.Variable
	FinalRollingHashMsgNumber   frontend.Variable
}

// L2MessageHashes is a wrapper for [Var32Slice] it is use to instantiate the
// sequence of L2MessageHash that we extract from the arithmetization. The
// reason we need a wrapper here is that we hash the L2MessageHashes in a
// specific way.
type L2MessageHashes internal.Var32Slice

func (s *L2MessageHashes) Assign(values [][32]frontend.Variable) error {
	if len(values) > len(s.Values) {
		return fmt.Errorf("%d values cannot fit in %d-long slice", len(values), len(s.Values))
	}
	for i := range values {
		utils.Copy(s.Values[i][:], values[i][:])
	}
	var zeros [32]byte
	for i := len(values); i < len(s.Values); i++ {
		utils.Copy(s.Values[i][:], zeros[:])
	}
	s.Length = len(values)
	return nil
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
	InitialStateRootHash frontend.Variable
	InitialBlockNumber   frontend.Variable
	ChainID              frontend.Variable
	L2MessageServiceAddr frontend.Variable
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
	for i := range pi.FinalRollingHashUpdate {
		rc.Check(pi.FinalRollingHashUpdate[i], 8)
	}
	rc.Check(pi.FinalBlockNumber, 64)
	rc.Check(pi.FinalBlockTimestamp, 64)
	rc.Check(pi.InitialBlockTimestamp, 64)
	rc.Check(pi.InitialRollingHashMsgNumber, 64)
	rc.Check(pi.FinalRollingHashMsgNumber, 64)

}

func (pi *FunctionalPublicInputSnark) Sum(api frontend.API, hsh gnarkHash.FieldHasher) frontend.Variable {

	var (
		finalRollingHash   = internal.CombineBytesIntoElements(api, pi.FinalRollingHashUpdate)
		initialRollingHash = internal.CombineBytesIntoElements(api, pi.InitialRollingHashUpdate)
		l2MessagesSum      = pi.L2MessageHashes.CheckSumMiMC(api)
	)

	hsh.Reset()
	hsh.Write(pi.DataChecksum, l2MessagesSum,
		pi.FinalStateRootHash, pi.FinalBlockNumber, pi.FinalBlockTimestamp, finalRollingHash[0], finalRollingHash[1], pi.FinalRollingHashMsgNumber,
		pi.InitialStateRootHash, pi.InitialBlockNumber, pi.InitialBlockTimestamp, initialRollingHash[0], initialRollingHash[1], pi.InitialRollingHashMsgNumber,
		pi.ChainID, pi.L2MessageServiceAddr)

	return hsh.Sum()
}

func (spi *FunctionalPublicInputSnark) Assign(pi *public_input.Execution) error {

	spi.FunctionalPublicInputQSnark.Assign(pi)

	spi.InitialStateRootHash = pi.InitialStateRootHash
	spi.InitialBlockNumber = pi.InitialBlockNumber
	spi.ChainID = pi.ChainID
	spi.L2MessageServiceAddr = pi.L2MessageServiceAddr

	utils.Copy(spi.FinalRollingHashUpdate[:], pi.FinalRollingHashUpdate[:])
	utils.Copy(spi.InitialRollingHashUpdate[:], pi.InitialRollingHashUpdate[:])

	var err error
	if nbMsg := len(pi.L2MessageHashes); nbMsg > pi.MaxNbL2MessageHashes {
		err = fmt.Errorf("has %d L2 message hashes but a maximum of %d is allowed", nbMsg, pi.MaxNbL2MessageHashes)
	}

	return err

}

func (spiq *FunctionalPublicInputQSnark) Assign(pi *public_input.Execution) error {

	spiq.DataChecksum = pi.DataChecksum
	spiq.InitialBlockTimestamp = pi.InitialBlockTimestamp
	spiq.FinalStateRootHash = pi.FinalStateRootHash
	spiq.FinalBlockNumber = pi.FinalBlockNumber
	spiq.FinalBlockTimestamp = pi.FinalBlockTimestamp
	spiq.InitialRollingHashMsgNumber = pi.InitialRollingHashMsgNumber
	spiq.FinalRollingHashMsgNumber = pi.FinalRollingHashMsgNumber

	utils.Copy(spiq.FinalRollingHashUpdate[:], pi.FinalRollingHashUpdate[:])
	utils.Copy(spiq.InitialRollingHashUpdate[:], pi.InitialRollingHashUpdate[:])

	if len(pi.L2MessageHashes) > len(spiq.L2MessageHashes.Values) {

	}

	return internal.Var32Slice(spiq.L2MessageHashes).Assign(pi.L2MessageHashes)
}

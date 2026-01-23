package public_input

import (
	"bytes"
	"encoding/binary"
	"hash"
	"math/big"
	"math/bits"

	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	gcposeidon2permutation "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
	gchash "github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/compress"
	ghash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// ExecDataChecksum consists of information enabling the computation
// of a checksum for execution data.
type ExecDataChecksum struct {

	// Length of the execution data in bytes
	Length uint64

	// PartialHash consists of hₙ₋₁ where hₙ₋₁ = H(hₙ₋₂, cₙ₋₁), ..., h₁ = H(h₀, c₁), h₀ = c₀ and
	// the cᵢ are the data packed into 31-byte BLS12-377 scalars in Big-Endian fashion.
	// H here is the Poseidon2 compression function for BLS12-377 scalars.
	// The last element is padded with less significant zeros as needed.
	// This sum is not collision resistant. To become so it needs the data length
	// incorporated into it as well.
	PartialHash types.Bls12377Fr

	// Hash is a modified Poseidon2 hash of the execution data over
	// the BLS12-377 scalar field. It is computed as H(PartialHash, Length).
	Hash types.Bls12377Fr
}

// @gusiri
// TODO: make sure we bubble up everything for the dynamic chain configuration
type Execution struct {
	InitialBlockTimestamp        uint64
	FinalStateRootHash           [32]byte
	FinalBlockNumber             uint64
	FinalBlockTimestamp          uint64
	LastRollingHashUpdate        [32]byte
	LastRollingHashUpdateNumber  uint64
	InitialRollingHashUpdate     [32]byte
	FirstRollingHashUpdateNumber uint64
	DataChecksum                 ExecDataChecksum
	L2MessageHashes              [][32]byte
	InitialStateRootHash         [32]byte
	InitialBlockNumber           uint64

	// Dynamic chain configuration
	ChainID              uint64
	BaseFee              uint64
	CoinBase             types.EthAddress
	L2MessageServiceAddr types.EthAddress
}

func (pi *Execution) Sum() []byte {
	hsh := gchash.POSEIDON2_BLS12_377.New()

	for i := range pi.L2MessageHashes {
		hsh.Write(pi.L2MessageHashes[i][:16])
		hsh.Write(pi.L2MessageHashes[i][16:])
	}
	l2MessagesSum := hsh.Sum(nil)

	hsh.Reset()

	hsh.Write(pi.DataChecksum.Hash[:])

	hsh.Write(l2MessagesSum)
	hsh.Write(pi.FinalStateRootHash[:])

	writeNum(hsh, pi.FinalBlockNumber)
	writeNum(hsh, pi.FinalBlockTimestamp)
	hsh.Write(pi.LastRollingHashUpdate[:16])
	hsh.Write(pi.LastRollingHashUpdate[16:])
	writeNum(hsh, pi.LastRollingHashUpdateNumber)
	hsh.Write(pi.InitialStateRootHash[:])
	writeNum(hsh, pi.InitialBlockNumber)
	writeNum(hsh, pi.InitialBlockTimestamp)
	hsh.Write(pi.InitialRollingHashUpdate[:16])
	hsh.Write(pi.InitialRollingHashUpdate[16:])
	writeNum(hsh, pi.FirstRollingHashUpdateNumber)
	// dynamic chain configuration
	writeNum(hsh, pi.ChainID)
	writeNum(hsh, pi.BaseFee)
	hsh.Write(pi.CoinBase[:])
	hsh.Write(pi.L2MessageServiceAddr[:])

	return hsh.Sum(nil)

}

func (pi *Execution) SumAsField() field.Element {

	sumBytes := pi.Sum()
	sum := new(field.Element).SetBytes(sumBytes)

	return *sum
}

func writeNum(hsh hash.Hash, n uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], n)
	hsh.Write(b[:])
}

// compressionChain range-extends a compressor into a non-MerkleDamgard semi-hasher.
// It does not gracefully recover from errors.
type compressionChain struct {
	state      []byte
	buffer     bytes.Buffer
	compressor gchash.Compressor
}

func (c *compressionChain) Write(in []byte) (n int, err error) {
	for len(in)+c.buffer.Len() >= c.compressor.BlockSize() {
		m := c.compressor.BlockSize() - c.buffer.Len()
		c.buffer.Write(in[:m])
		in = in[m:]

		if len(c.state) == 0 {
			c.state = bytes.Clone(c.buffer.Bytes())
		} else if c.state, err = c.compressor.Compress(c.state, c.buffer.Bytes()); err != nil {
			return
		}

		c.buffer.Reset()

		n += m
	}

	c.buffer.Write(in)
	n += len(in)
	return

}

func NewExecDataChecksum(data []byte) (sums ExecDataChecksum, err error) {
	sums.Length = uint64(len(data))

	compressionChain := compressionChain{compressor: gcposeidon2permutation.NewDefaultPermutation()}

	if err = gnarkutil.PackLoose(&compressionChain, data, fr377.Bytes, 1); err != nil {
		return
	}
	copy(sums.PartialHash[:], compressionChain.state)
	var length [fr377.Bytes]byte
	binary.BigEndian.PutUint64(length[len(length)-8:], sums.Length)
	if _, err = compressionChain.Write(length[:]); err != nil {
		return
	}
	copy(sums.Hash[:], compressionChain.state)
	return
}

// ChecksumExecDataSnark computes the checksum of execution data as a BLS12-377 element.
// Caller must ensure that all data values past nbBytes are zero, or the result will be incorrect.
// Each element of data is meant to contain wordNbBits many bits. This claim is not guaranteed to be checked by this function.
func ChecksumExecDataSnark(api frontend.API, data []frontend.Variable, wordNbBits int, nbBytes frontend.Variable, compressor ghash.Compressor) (frontend.Variable, error) {

	if len(data) == 0 {
		return 0, nil
	}

	// turn the data into bytes
	_bytes, err := gnarkutil.ToBytes(api, data, wordNbBits)
	if err != nil {
		return nil, err
	}

	// turn the bytes into blocks
	radix := big.NewInt(256)
	blocks := make([]frontend.Variable, (len(_bytes)+31-1)/31)
	for i := range blocks {
		blocks[i] = compress.ReadNum(api, _bytes[i*31:min(i*31+31, len(_bytes))], radix)
	}

	// pad the last block with zeros
	for range len(blocks)*31 - len(_bytes) {
		blocks[len(blocks)-1] = api.Mul(blocks[len(blocks)-1], radix)
	}

	// chain of hashes
	partials := logderivlookup.New(api)
	state := blocks[0]
	for i := 1; i < len(blocks); i++ {
		partials.Insert(state)
		state = compressor.Compress(state, blocks[i])
	}
	partials.Insert(state)

	// find the partial checksum to use
	// number of used blocks is ⌈nbBytes / 31⌉
	blockI, _, err := gnarkutil.DivBy31(api, api.Add(nbBytes, 30), 1+bits.Len(uint(len(_bytes))))
	if err != nil {
		return nil, err
	}
	blockI = api.Sub(blockI, 1)
	partial := partials.Lookup(blockI)[0]

	return compressor.Compress(partial, nbBytes), nil
}

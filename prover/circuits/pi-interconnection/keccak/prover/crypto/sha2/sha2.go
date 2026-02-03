package sha2

import (
	"bytes"
	"encoding/binary"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
)

const (
	BlockSizeByte       = 64
	DigestSizeByte      = 32
	blockSizeUint32     = BlockSizeByte / 4
	digestSizeUint32    = DigestSizeByte / 4
	domainSeparatorByte = byte(0x80)
)

type (
	blockUint32  [blockSizeUint32]uint32
	digestUint32 [digestSizeUint32]uint32
	Block        = [BlockSizeByte]byte
	Digest       = [DigestSizeByte]byte
)

// iv is the initialization vector of sha2 and is the initial state of the hasher
// when the hashing starts
var iv = digestUint32{
	0x6A09E667,
	0xBB67AE85,
	0x3C6EF372,
	0xA54FF53A,
	0x510E527F,
	0x9B05688C,
	0x1F83D9AB,
	0x5BE0CD19,
}

// HashTraces represents the traces of the sha2 happening when
// hashing a long string
type HashTraces struct {
	// blocks of the message, including the padding
	Blocks         []Block
	BlockOldStates []Digest
	BlockNewStates []Digest
	// indicates whether the current block is  the first block of a hash.
	IsNewHash []bool
}

// PadStream returns the stream, padded following Sha2's specification
func PadStream(stream []byte) []byte {
	return paddedBuffer(stream).Bytes()
}

// Compress runs the compression function of Sha2 over a block and an initial
// hasher state and returns the resulting state.
func Compress(oldState Digest, block Block) (newState Digest) {

	var (
		oldStateUint32 = new(digestUint32).fromDigest(oldState)
		blockUint32    = new(blockUint32).fromBlock(block)
	)

	permSha2(oldStateUint32, *blockUint32)
	return oldStateUint32.intoDigest()
}

// Hash computes the sha2 hash of a slice of bytes. The user can optionally
// pass a HashTraces to which the function will append the generated
// traces throughout the hashing process.
func Hash(stream []byte, optTracer *HashTraces) Digest {

	var (
		currState = iv
		blocks    = splitInBlocksUint32(stream)
	)

	for i, block := range blocks {

		if optTracer != nil {
			optTracer.BlockOldStates = append(optTracer.BlockOldStates, currState.intoDigest())
		}

		permSha2(&currState, block)

		if optTracer != nil {
			optTracer.Blocks = append(optTracer.Blocks, block.intoBlock())
			optTracer.IsNewHash = append(optTracer.IsNewHash, i == 0)
			optTracer.BlockNewStates = append(optTracer.BlockNewStates, currState.intoDigest())
		}
	}

	return currState.intoDigest()
}

// splitInBlocks applies the Sha2 padding to the stream and returns the
// list of blocks to feed to the compression function.
func splitInBlocksUint32(stream []byte) []blockUint32 {

	var (
		paddedBuffer  = paddedBuffer(stream)
		paddedByteLen = paddedBuffer.Len()
		numBlocks     = paddedByteLen / BlockSizeByte
		blocks        = make([]blockUint32, numBlocks)
		tmp           Block
	)

	for i := range blocks {
		paddedBuffer.Read(tmp[:])
		blocks[i].fromBlock(tmp)
	}

	return blocks
}

// paddedBuffer returns a [byte.Buffer] storing the padded input stream
func paddedBuffer(stream []byte) *bytes.Buffer {

	var (
		buf               = &bytes.Buffer{}
		streamByteLen, _  = buf.Write(stream) // can't err
		streamBitLen      = streamByteLen << 3
		numZeroBytesToPad = BlockSizeByte - ((streamByteLen + 9) % BlockSizeByte)
	)

	if numZeroBytesToPad == BlockSizeByte {
		numZeroBytesToPad = 0
	}

	buf.WriteByte(domainSeparatorByte)
	buf.Write(make([]byte, numZeroBytesToPad))
	binary.Write(buf, binary.BigEndian, uint64(streamBitLen))

	var (
		paddedByteLen  = buf.Len()
		paddingByteLen = paddedByteLen - streamByteLen
	)

	if paddingByteLen < 9 || paddingByteLen > 72 {
		utils.Panic("invalid padding size: %v", paddingByteLen)
	}

	if paddedByteLen%BlockSizeByte != 0 {
		utils.Panic("invalid padded size: %v", paddedByteLen)
	}

	return buf
}

// fromDigest sets the value of 'd' from a Digest in byte form.
func (d *digestUint32) fromDigest(dBytes Digest) *digestUint32 {
	for i := 0; i < len(d); i++ {
		d[i] = binary.BigEndian.Uint32(dBytes[4*i : 4*i+4])
	}

	return d
}

// intoDigest recovers a digest in byte form that can be exported by the package
// API
func (d digestUint32) intoDigest() (res Digest) {
	for i := 0; i < len(d); i++ {
		binary.BigEndian.PutUint32(res[4*i:4*i+4], d[i])
	}

	return res
}

// fromBlocks sets the value of 'b' from a Block in byte form.
func (b *blockUint32) fromBlock(bBytes Block) *blockUint32 {
	for i := 0; i < len(b); i++ {
		b[i] = binary.BigEndian.Uint32(bBytes[4*i : 4*i+4])
	}
	return b
}

// intoBlock recovers the block in bytes form that can be exported by the package
// API
func (b blockUint32) intoBlock() (res Block) {
	for i := 0; i < len(b); i++ {
		binary.BigEndian.PutUint32(res[4*i:4*i+4], b[i])
	}

	return res
}

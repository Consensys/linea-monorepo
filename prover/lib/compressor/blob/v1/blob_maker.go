package v1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/sirupsen/logrus"

	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/icza/bitio"

	"github.com/consensys/compress/lzss"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	maxOutputSize        = 1 << 20 // 1MB
	PackingSizeU256      = fr381.Bits - 1
	ByteLenEncodingBytes = 3
	NbElemsEncodingBytes = 2

	// These also impact the circuit constraints (compile / setup time)
	MaxUncompressedBytes = 797.25 * 1024 // defines the max size we can handle for a blob (uncompressed) input
	MaxUsableBytes       = 32 * 4096     // defines the number of bytes available in a blob
)

// BlobMaker is a bm for RLP encoded blocks (see EIP-4844).
// It takes a batch of blocks as input (see StartNewBatch and Write).
// And it compresses them into a "blob" (see Bytes).
type BlobMaker struct {
	limit      int              // maximum size of the compressed data
	compressor *lzss.Compressor // compressor used to compress the blob body
	dict       []byte           // dictionary used for compression

	header Header

	// contains currentBlob data from latest **valid** call to Write
	// that is the header (uncompressed) and the body (compressed)
	// byte aligned to match fr.Element boundary.
	currentBlob       [maxOutputSize]byte
	currentBlobLength int

	// some buffers to avoid repeated allocations
	buf        bytes.Buffer
	packBuffer bytes.Buffer
}

// NewBlobMaker returns a new bm.
func NewBlobMaker(dataLimit int, dictPath string) (*BlobMaker, error) {
	blobMaker := BlobMaker{
		limit: dataLimit,
	}
	blobMaker.buf.Grow(1 << 17)

	// initialize compressor
	dict, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, err
	}
	dict = lzss.AugmentDict(dict)
	blobMaker.dict = dict

	dictChecksum, err := MiMCChecksumPackedData(dict, 8)
	if err != nil {
		return nil, err
	}
	copy(blobMaker.header.DictChecksum[:], dictChecksum)

	blobMaker.compressor, err = lzss.NewCompressor(dict)
	if err != nil {
		return nil, err
	}

	// initialize state
	blobMaker.StartNewBatch()

	return &blobMaker, nil
}

// StartNewBatch starts a new batch of blocks.
func (bm *BlobMaker) StartNewBatch() {
	bm.header.sealBatch()
}

// Reset resets the bm to its initial state.
func (bm *BlobMaker) Reset() {
	bm.header.resetTable()
	bm.currentBlobLength = 0
	bm.buf.Reset()
	bm.packBuffer.Reset()
	bm.compressor.Reset()

	bm.StartNewBatch()
}

// Len returns the length of the compressed data, which includes the header.
func (bm *BlobMaker) Len() int {
	return bm.currentBlobLength
}

func (bm *BlobMaker) Written() int {
	return bm.compressor.Written()
}

// Bytes returns the compressed data. Note that it returns a slice of the internal buffer,
// it is the caller's responsibility to copy the data if needed.
func (bm *BlobMaker) Bytes() []byte {
	if bm.currentBlobLength > 0 {
		// sanity check that we can always decompress.
		header, rawBlocks, _, err := DecompressBlob(bm.currentBlob[:bm.currentBlobLength], bm.dict)
		if err != nil {
			var sbb strings.Builder
			fmt.Fprintf(&sbb, "invalid blob: %v\n", err)
			fmt.Fprintf(&sbb, "header: %v\n", bm.header)
			fmt.Fprintf(&sbb, "bm.currentBlobLength: %v\n", bm.currentBlobLength)
			fmt.Fprintf(&sbb, "bm.currentBlob: %x\n", bm.currentBlob[:bm.currentBlobLength])

			panic(sbb.String())
		}
		// compare the header
		if !header.Equals(&bm.header) {
			panic("invalid blob: header mismatch")
		}
		if !bytes.Equal(rawBlocks, bm.compressor.WrittenBytes()) {
			panic(fmt.Sprintf("invalid blob: body mismatch expected %x, got %x", rawBlocks, bm.compressor.WrittenBytes()))
		}
	}
	return bm.currentBlob[:bm.currentBlobLength]
}

// Write attempts to append the RLP block to the current batch.
// if forceReset is set; this will NOT append the bytes but still returns true if the chunk could have been appended
func (bm *BlobMaker) Write(rlpBlock []byte, forceReset bool) (ok bool, err error) {

	// decode the RLP block.
	var block types.Block
	if err = rlp.Decode(bytes.NewReader(rlpBlock), &block); err != nil {
		return false, fmt.Errorf("when decoding input RLP block: %w", err)
	}

	// re-encode it for compression
	bm.buf.Reset()
	if err = EncodeBlockForCompression(&block, &bm.buf); err != nil {
		return false, fmt.Errorf("when re-encoding block for compression: %w", err)
	}
	blockLen := bm.buf.Len()

	if blockLen > bm.limit {
		// we should panic but logging / alerting is handled by the caller.
		// see https://github.com/Consensys/zkevm-monorepo/issues/2326#issuecomment-1923573005
		logrus.Warn("block size is larger than the blob limit. This should be checked by the coordinator, keeping the log for sanity", "block size", blockLen, "limit", bm.limit)
	}

	// write the block to the bm
	if _, err = bm.compressor.Write(bm.buf.Bytes()); err != nil {
		// The 2 possibles errors are:
		// 1. underlying writer error (shouldn't happen we use a simple in memory buffer)
		// 2. we exceed the maximum input size of 2Mb (shouldn't happen either)
		// In both cases, we can't do anything, so we reset the state.
		return false, fmt.Errorf("when writing block to compressor: %w", err)
	}

	// increment length of the current batch
	bm.header.addBlock(blockLen)
	// write the header to get its length.
	bm.buf.Reset()
	if _, err = bm.header.WriteTo(&bm.buf); err != nil {
		// only possible error is an underlying writer error (shouldn't happen we use a simple in-memory buffer)
		bm.header.removeLastBlock()
		return false, fmt.Errorf("when writing header to buffer: %w", err)
	}

	// check that the header + the uncompressed data is "decompressable" in the circuit
	if uint64(bm.compressor.Written()+bm.buf.Len()) > MaxUncompressedBytes {
		// it means we are not exploiting the full blob capacity; our compression ratio is "too good"
		// and our decompression circuit is not able to handle the uncompressed data.
		// we should reset the state.
		if err := bm.compressor.Revert(); err != nil {
			return false, fmt.Errorf("when reverting compressor because uncompressed blob is > maxUncompressedSize: %w", err)
		}
		bm.header.removeLastBlock()
		return false, nil
	}

	// check that the header + the compressed data fits in the blob
	fitsInBlob := PackAlignSize(bm.buf.Len()+bm.compressor.Len(), fr381.Bits-1) <= bm.limit
	if !fitsInBlob {
		// first thing to check is if we bypass compression, would that fit?
		if bm.compressor.ConsiderBypassing() {
			// we can bypass compression and get a better ratio.
			// let's check if now we fit in the blob.
			if PackAlignSize(bm.buf.Len()+bm.compressor.Len(), fr381.Bits-1) <= bm.limit {
				goto bypass
			}
		}

		// discard.
		if err = bm.compressor.Revert(); err != nil {
			return false, fmt.Errorf("when reverting compressor because blob is full: %w", err)
		}
		bm.header.removeLastBlock()
		return false, nil
	}
bypass:
	if forceReset {
		// we don't want to append the data, but we could have.
		if err = bm.compressor.Revert(); err != nil {
			return false, fmt.Errorf("when reverting compressor (blob is not full but forceReset == true): %w", err)
		}
		bm.header.removeLastBlock()
		return true, nil
	}

	// copy the compressed data to the blob
	bm.packBuffer.Reset()
	n2, err := PackAlign(&bm.packBuffer, bm.buf.Bytes(), fr381.Bits-1, WithAdditionalInput(bm.compressor.Bytes()))
	if err != nil {
		bm.compressor.Revert()
		bm.header.removeLastBlock()
		return false, fmt.Errorf("when packing blob: %w", err)
	}
	bm.currentBlobLength = int(n2)
	copy(bm.currentBlob[:bm.currentBlobLength], bm.packBuffer.Bytes())

	return true, nil
}

// Clone returns a (almost) deep copy of the bm -- this is used for test purposes.
func (bm *BlobMaker) Clone() *BlobMaker {
	deepCopy := *bm
	deepCopy.header.BatchSizes = make([]int, len(bm.header.BatchSizes))

	copy(deepCopy.header.BatchSizes, bm.header.BatchSizes)

	return &deepCopy
}

// Equals returns true if the two compressors are ~equal -- this is used for test purposes.
func (bm *BlobMaker) Equals(other *BlobMaker) bool {
	if bm.limit != other.limit {
		return false
	}
	if bm.currentBlobLength != other.currentBlobLength {
		return false
	}
	if !bytes.Equal(bm.currentBlob[:bm.currentBlobLength], other.currentBlob[:other.currentBlobLength]) {
		return false
	}
	if len(bm.header.BatchSizes) != len(other.header.BatchSizes) {
		return false
	}
	if !slices.Equal(bm.header.BatchSizes, other.header.BatchSizes) {
		return false
	}
	return true
}

// DecompressBlob decompresses a blob and returns the header and the blocks as they were compressed.
func DecompressBlob(b, dict []byte) (blobHeader *Header, rawPayload []byte, blocks [][]byte, err error) {
	// UnpackAlign the blob
	b, err = UnpackAlign(b, fr381.Bits-1, false)
	if err != nil {
		return nil, nil, nil, err
	}

	// read the header
	blobHeader = new(Header)
	read, err := blobHeader.ReadFrom(bytes.NewReader(b))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read blob header: %w", err)
	}
	// ensure the dict hash matches
	{
		expectedDictChecksum, err := MiMCChecksumPackedData(dict, 8)
		if err != nil {
			return nil, nil, nil, err
		}
		if !bytes.Equal(expectedDictChecksum, blobHeader.DictChecksum[:]) {
			return nil, nil, nil, errors.New("invalid dict hash")
		}
	}

	b = b[read:]

	// decompress the data
	rawPayload, err = lzss.Decompress(b, dict)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decompress blob body: %w", err)
	}

	offset := 0
	for _, batchLen := range blobHeader.BatchSizes {

		batchOffset := offset
		for offset < batchOffset+batchLen {
			if blockLen, err := nextRlpListNbBytes(rawPayload[offset:]); err != nil {
				return nil, nil, nil, err
			} else {
				blocks = append(blocks, rawPayload[offset:offset+blockLen])
				offset += blockLen
			}
		}

		if offset != batchOffset+batchLen {
			return nil, nil, nil, errors.New("incorrect batch length")
		}
	}

	return blobHeader, rawPayload, blocks, nil
}

// nextRlpListNbBytes interprets a prefix of b as a list object, returning how many bytes that list contains.
func nextRlpListNbBytes(b []byte) (n int, err error) { // TODO Review VERY CAREFULLY
	if len(b) == 0 {
		return 0, errors.New("empty input")
	}
	if b[0] < 0xc0 {
		return 0, errors.New("not a list")
	}
	if b[0] <= 0xf7 {
		n = int(b[0]) - 0xc0 + 1 // 1-byte prefix
	} else {
		l := int(b[0]) - 0xf7
		var buf [8]byte
		copy(buf[8-l:], b[1:])
		payloadLen := binary.BigEndian.Uint64(buf[:])
		n = 1 + l + int(payloadLen)
	}
	if n > len(b) {
		return 0, errors.New("incomplete input")
	}
	return n, nil
}

// PackAlignSize returns the size of the data when packed with PackAlign.
func PackAlignSize(length0, packingSize int, options ...packAlignOption) (n int) {
	var s packAlignSettings
	s.initialize(length0, options...)

	// we may need to add some bits to a and b to ensure we can process some blocks of 248 bits
	extraBits := (packingSize - s.dataNbBits%packingSize) % packingSize
	nbBits := s.dataNbBits + extraBits

	return (nbBits / packingSize) * ((packingSize + 7) / 8)
}

type packAlignSettings struct {
	dataNbBits           int
	lastByteNbUnusedBits uint8
	noTerminalSymbol     bool
	additionalInput      [][]byte
}

type packAlignOption func(*packAlignSettings)

func NoTerminalSymbol() packAlignOption {
	return func(o *packAlignSettings) {
		o.noTerminalSymbol = true
	}
}

func WithAdditionalInput(data ...[]byte) packAlignOption {
	return func(o *packAlignSettings) {
		o.additionalInput = append(o.additionalInput, data...)
	}
}

func WithLastByteNbUnusedBits(n uint8) packAlignOption {
	if n > 7 {
		panic("only 8 bits to a byte")
	}
	return func(o *packAlignSettings) {
		o.lastByteNbUnusedBits = n
	}
}

func (s *packAlignSettings) initialize(length int, options ...packAlignOption) {

	for _, opt := range options {
		opt(s)
	}

	nbBytes := length
	for _, data := range s.additionalInput {
		nbBytes += len(data)
	}

	if !s.noTerminalSymbol {
		nbBytes++
	}

	s.dataNbBits = nbBytes*8 - int(s.lastByteNbUnusedBits)
}

// PackAlign writes a and b to w, aligned to fr.Element (bls12-377) boundary.
// It returns the length of the data written to w.
func PackAlign(w io.Writer, a []byte, packingSize int, options ...packAlignOption) (n int64, err error) {

	var s packAlignSettings
	s.initialize(len(a), options...)
	if !s.noTerminalSymbol && s.lastByteNbUnusedBits != 0 {
		return 0, errors.New("terminal symbols with byte aligned input not yet supported")
	}

	// we may need to add some bits to a and b to ensure we can process some blocks of packingSize bits
	nbBits := (s.dataNbBits + (packingSize - 1)) / packingSize * packingSize
	extraBits := nbBits - s.dataNbBits

	// padding will always be less than bytesPerElem bytes
	bytesPerElem := (packingSize + 7) / 8
	packingSizeLastU64 := uint8(packingSize % 64)
	if packingSizeLastU64 == 0 {
		packingSizeLastU64 = 64
	}
	bytePadding := (extraBits + 7) / 8
	buf := make([]byte, bytesPerElem, bytesPerElem+1)

	// the last nonzero byte is 0xff
	if !s.noTerminalSymbol {
		buf = append(buf, 0)
		buf[0] = 0xff
	}

	inReaders := make([]io.Reader, 2+len(s.additionalInput))
	inReaders[0] = bytes.NewReader(a)
	for i, data := range s.additionalInput {
		inReaders[i+1] = bytes.NewReader(data)
	}
	inReaders[len(inReaders)-1] = bytes.NewReader(buf[:bytePadding+1])

	r := bitio.NewReader(io.MultiReader(inReaders...))

	var tryWriteErr error
	tryWrite := func(v uint64) {
		if tryWriteErr == nil {
			tryWriteErr = binary.Write(w, binary.BigEndian, v)
		}
	}

	for i := 0; i < nbBits/packingSize; i++ {
		tryWrite(r.TryReadBits(packingSizeLastU64))
		for j := int(packingSizeLastU64); j < packingSize; j += 64 {
			tryWrite(r.TryReadBits(64))
		}
	}

	if tryWriteErr != nil {
		return 0, fmt.Errorf("when writing to w: %w", tryWriteErr)
	}

	if r.TryError != nil {
		return 0, fmt.Errorf("when reading from multi-reader: %w", r.TryError)
	}

	n1 := (nbBits / packingSize) * bytesPerElem
	if n1 != PackAlignSize(len(a), packingSize, options...) {
		return 0, errors.New("inconsistent PackAlignSize")
	}
	return int64(n1), nil
}

// UnpackAlign unpacks r (packed with PackAlign) and returns the unpacked data.
func UnpackAlign(r []byte, packingSize int, noTerminalSymbol bool) ([]byte, error) {
	bytesPerElem := (packingSize + 7) / 8
	packingSizeLastU64 := uint8(packingSize % 64)
	if packingSizeLastU64 == 0 {
		packingSizeLastU64 = 64
	}

	n := len(r) / bytesPerElem
	if n*bytesPerElem != len(r) {
		return nil, fmt.Errorf("invalid data length; expected multiple of %d", bytesPerElem)
	}

	var out bytes.Buffer
	w := bitio.NewWriter(&out)
	for i := 0; i < n; i++ {
		// read bytes
		element := r[bytesPerElem*i : bytesPerElem*(i+1)]
		// write bits
		w.TryWriteBits(binary.BigEndian.Uint64(element[0:8]), packingSizeLastU64)
		for j := 8; j < bytesPerElem; j += 8 {
			w.TryWriteBits(binary.BigEndian.Uint64(element[j:j+8]), 64)
		}
	}
	if w.TryError != nil {
		return nil, fmt.Errorf("when writing to bitio.Writer: %w", w.TryError)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("when closing bitio.Writer: %w", err)
	}

	if !noTerminalSymbol {
		// the last nonzero byte should be 0xff
		outLen := out.Len() - 1
		for out.Bytes()[outLen] == 0 {
			outLen--
		}
		if out.Bytes()[outLen] != 0xff {
			return nil, errors.New("invalid terminal symbol")
		}
		out.Truncate(outLen)
	}

	return out.Bytes(), nil
}

// MiMCChecksumPackedData re-packs the data tightly into bls12-377 elements and computes the MiMC checksum.
// only supporting packing without a terminal symbol. Input with a terminal symbol will be interpreted in full padded length.
func MiMCChecksumPackedData(data []byte, inputPackingSize int, hashPackingOptions ...packAlignOption) ([]byte, error) {
	dataNbBits := len(data) * 8
	if inputPackingSize%8 != 0 {
		inputBytesPerElem := (inputPackingSize + 7) / 8
		dataNbBits = dataNbBits / inputBytesPerElem * inputPackingSize
		var err error
		if data, err = UnpackAlign(data, inputPackingSize, true); err != nil {
			return nil, err
		}
	}

	lastByteNbUnusedBits := 8 - dataNbBits%8
	if lastByteNbUnusedBits == 8 {
		lastByteNbUnusedBits = 0
	}

	var bb bytes.Buffer
	packingOptions := make([]packAlignOption, len(hashPackingOptions)+1)
	copy(packingOptions, hashPackingOptions)
	packingOptions[len(packingOptions)-1] = WithLastByteNbUnusedBits(uint8(lastByteNbUnusedBits))
	if _, err := PackAlign(&bb, data, fr377.Bits-1, packingOptions...); err != nil {
		return nil, err
	}

	hsh := hash.MIMC_BLS12_377.New()
	hsh.Write(bb.Bytes())
	return hsh.Sum(nil), nil
}

// WorstCompressedBlockSize returns the size of the given block, as compressed by an "empty" blob maker.
// That is, with more context, blob maker could compress the block further, but this function
// returns the maximum size that can be achieved.
//
// The input is a RLP encoded block.
// Returns the length of the compressed data, or -1 if an error occurred.
//
// This function is thread-safe. Concurrent calls are allowed,
// but the other functions may not be thread-safe.
func (bm *BlobMaker) WorstCompressedBlockSize(rlpBlock []byte) (bool, int, error) {
	// decode the RLP block.
	var block types.Block
	if err := rlp.Decode(bytes.NewReader(rlpBlock), &block); err != nil {
		return false, -1, fmt.Errorf("failed to decode RLP block: %w", err)
	}

	// encode the block in Linea format.
	var buf bytes.Buffer
	if err := EncodeBlockForCompression(&block, &buf); err != nil {
		return false, -1, fmt.Errorf("failed to encode block: %w", err)
	}

	inputSlice := buf.Bytes()
	n, err := bm.compressor.CompressedSize256k(inputSlice)
	if err != nil {
		return false, -1, err
	}
	expandingBlock := n > len(inputSlice)
	if expandingBlock {
		// this simulates the fallback to "no compression"
		// this case may happen if the input is not compressible
		// in which case the compressed size is the input size + the header size
		n = len(inputSlice) + lzss.HeaderSize
	}

	// account for the padding
	n = PackAlignSize(n, fr381.Bits-1, NoTerminalSymbol())

	return expandingBlock, n, nil
}

// WorstCompressedTxSize returns the size of the given transaction, as compressed by an "empty" blob maker.
// That is, with more context, blob maker could compress the transaction further, but this function
// returns the maximum size that can be achieved.
//
// The input is a RLP encoded transaction.
// Returns the length of the compressed data, or -1 if an error occurred.
//
// This function is thread-safe. Concurrent calls are allowed,
// but the other functions may not be thread-safe.
func (bm *BlobMaker) WorstCompressedTxSize(rlpTx []byte) (int, error) {

	// decode the RLP transaction.
	var tx types.Transaction
	if err := rlp.Decode(bytes.NewReader(rlpTx), &tx); err != nil {
		return -1, err
	}

	// encode the transaction in Linea format.
	var buf bytes.Buffer
	txTrimmed := trimTxForCompression(&tx)
	if err := rlp.Encode(&buf, txTrimmed); err != nil {
		return -1, fmt.Errorf("failed to encode transaction: %w", err)
	}

	inputSlice := buf.Bytes()
	return bm.compressor.CompressedSize256k(inputSlice)
}

// RawCompressedSize compresses the (raw) input and returns the length of the compressed data.
// The returned length account for the "padding" used by the blob maker to
// fit the data in field elements.
// Input size must be less than 256kB.
// If an error occurred, returns -1.
//
// This function is thread-safe. Concurrent calls are allowed,
// but the other functions are not thread-safe.
func (bm *BlobMaker) RawCompressedSize(data []byte) (int, error) {
	n, err := bm.compressor.CompressedSize256k(data)
	if err != nil {
		return -1, err
	}
	if n > len(data) {
		// this simulates the fallback to "no compression"
		// this case may happen if the input is not compressible
		// in which case the compressed size is the input size + the header size
		n = len(data) + lzss.HeaderSize
	}

	// account for the padding
	n = PackAlignSize(n, fr381.Bits-1, NoTerminalSymbol())

	return n, nil
}

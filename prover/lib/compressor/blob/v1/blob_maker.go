package v1

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	"os"
	"slices"
	"strings"

	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/sirupsen/logrus"

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
	MaxUncompressedBytes = 740 * 1024 // defines the max size we can handle for a blob (uncompressed) input
	MaxUsableBytes       = 32 * 4096  // defines the number of bytes available in a blob
)

// BlobMaker is a bm for RLP encoded blocks (see EIP-4844).
// It takes a batch of blocks as input (see StartNewBatch and Write).
// And it compresses them into a "blob" (see Bytes).
type BlobMaker struct {
	Limit      int              // maximum size of the compressed data
	compressor *lzss.Compressor // compressor used to compress the blob body
	dict       []byte           // dictionary used for compression
	dictStore  dictionary.Store // dictionary store comprising only dict, used for decompression sanity checks

	Header Header

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
		Limit: dataLimit,
	}
	blobMaker.buf.Grow(1 << 17)

	// initialize compressor
	dict, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, err
	}
	dict = lzss.AugmentDict(dict)
	blobMaker.dict = dict
	if blobMaker.dictStore, err = dictionary.SingletonStore(dict, 1); err != nil {
		return nil, err
	}

	dictChecksum, err := encode.MiMCChecksumPackedData(dict, 8)
	if err != nil {
		return nil, err
	}
	copy(blobMaker.Header.DictChecksum[:], dictChecksum)

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
	bm.Header.sealBatch()
}

// Reset resets the bm to its initial state.
func (bm *BlobMaker) Reset() {
	bm.Header.resetTable()
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
		header, rawBlocks, _, err := DecompressBlob(bm.currentBlob[:bm.currentBlobLength], bm.dictStore)
		if err != nil {
			var sbb strings.Builder
			fmt.Fprintf(&sbb, "invalid blob: %v\n", err)
			fmt.Fprintf(&sbb, "header: %v\n", bm.Header)
			fmt.Fprintf(&sbb, "bm.currentBlobLength: %v\n", bm.currentBlobLength)
			fmt.Fprintf(&sbb, "bm.currentBlob: %x\n", bm.currentBlob[:bm.currentBlobLength])

			panic(sbb.String())
		}
		// compare the header
		if !header.Equals(&bm.Header) {
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

	if blockLen > bm.Limit {
		// we should panic but logging / alerting is handled by the caller.
		logrus.Warn("block size is larger than the blob Limit. This should be checked by the coordinator, keeping the log for sanity", "block size", blockLen, "Limit", bm.Limit)
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
	bm.Header.addBlock(blockLen)
	// write the header to get its length.
	bm.buf.Reset()
	if _, err = bm.Header.WriteTo(&bm.buf); err != nil {
		// only possible error is an underlying writer error (shouldn't happen we use a simple in-memory buffer)
		bm.Header.removeLastBlock()
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
		bm.Header.removeLastBlock()
		return false, nil
	}

	// check that the header + the compressed data fits in the blob
	fitsInBlob := encode.PackAlignSize(bm.buf.Len()+bm.compressor.Len(), fr381.Bits-1) <= bm.Limit
	if !fitsInBlob {
		// first thing to check is if we bypass compression, would that fit?
		if bm.compressor.ConsiderBypassing() {
			// we can bypass compression and get a better ratio.
			// let's check if now we fit in the blob.
			if encode.PackAlignSize(bm.buf.Len()+bm.compressor.Len(), fr381.Bits-1) <= bm.Limit {
				goto bypass
			}
		}

		// discard.
		if err = bm.compressor.Revert(); err != nil {
			return false, fmt.Errorf("when reverting compressor because blob is full: %w", err)
		}
		bm.Header.removeLastBlock()
		return false, nil
	}
bypass:
	if forceReset {
		// we don't want to append the data, but we could have.
		if err = bm.compressor.Revert(); err != nil {
			return false, fmt.Errorf("when reverting compressor (blob is not full but forceReset == true): %w", err)
		}
		bm.Header.removeLastBlock()
		return true, nil
	}

	// copy the compressed data to the blob
	bm.packBuffer.Reset()
	n2, err := encode.PackAlign(&bm.packBuffer, bm.buf.Bytes(), fr381.Bits-1, encode.WithAdditionalInput(bm.compressor.Bytes()))
	if err != nil {
		bm.compressor.Revert()
		bm.Header.removeLastBlock()
		return false, fmt.Errorf("when packing blob: %w", err)
	}
	bm.currentBlobLength = int(n2)
	copy(bm.currentBlob[:bm.currentBlobLength], bm.packBuffer.Bytes())

	return true, nil
}

// Clone returns a (almost) deep copy of the bm -- this is used for test purposes.
func (bm *BlobMaker) Clone() *BlobMaker {
	deepCopy := *bm
	deepCopy.Header.BatchSizes = make([]int, len(bm.Header.BatchSizes))

	copy(deepCopy.Header.BatchSizes, bm.Header.BatchSizes)

	return &deepCopy
}

// Equals returns true if the two compressors are ~equal -- this is used for test purposes.
func (bm *BlobMaker) Equals(other *BlobMaker) bool {
	if bm.Limit != other.Limit {
		return false
	}
	if bm.currentBlobLength != other.currentBlobLength {
		return false
	}
	if !bytes.Equal(bm.currentBlob[:bm.currentBlobLength], other.currentBlob[:other.currentBlobLength]) {
		return false
	}
	if len(bm.Header.BatchSizes) != len(other.Header.BatchSizes) {
		return false
	}
	if !slices.Equal(bm.Header.BatchSizes, other.Header.BatchSizes) {
		return false
	}
	return true
}

// DecompressBlob decompresses a blob and returns the header and the blocks as they were compressed.
func DecompressBlob(b []byte, dictStore dictionary.Store) (blobHeader *Header, rawPayload []byte, blocks [][]byte, err error) {
	// UnpackAlign the blob
	b, err = encode.UnpackAlign(b, fr381.Bits-1, false)
	if err != nil {
		return nil, nil, nil, err
	}

	// read the header
	blobHeader = new(Header)
	read, err := blobHeader.ReadFrom(bytes.NewReader(b))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read blob header: %w", err)
	}
	// retrieve dict
	dict, err := dictStore.Get(blobHeader.DictChecksum[:], 1)
	if err != nil {
		return nil, nil, nil, err
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
			if blockLen, err := ScanBlockByteLen(rawPayload[offset:]); err != nil {
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
	n = encode.PackAlignSize(n, fr381.Bits-1, encode.NoTerminalSymbol())

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
	if err := encodeTxForCompression(&tx, &buf); err != nil {
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
	n = encode.PackAlignSize(n, fr381.Bits-1, encode.NoTerminalSymbol())

	return n, nil
}

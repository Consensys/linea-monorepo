package v1

import (
	"bytes"
	"fmt"
	"os"

	"github.com/consensys/compress/lzss"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// TxCompressor compresses RLP-encoded transactions additively, maintaining compression
// context across transactions for better compression ratios. This is designed for
// sequencer block building where transactions are added one by one until the
// compressed size threshold is reached.
//
// Unlike BlobMaker which operates on RLP-encoded blocks, TxCompressor operates on
// individual RLP-encoded transactions. The caller is responsible for accounting for
// blob overhead (header + block metadata, approximately 500 bytes) when setting the limit.
type TxCompressor struct {
	Limit      int              // maximum size of the compressed data
	compressor *lzss.Compressor // compressor used to compress transactions
	dict       []byte           // dictionary used for compression
	buf        bytes.Buffer     // reusable buffer for encoding
}

// NewTxCompressor returns a new transaction compressor.
// The dataLimit argument is the maximum size of the compressed data.
// The caller should account for blob overhead (~500 bytes) when setting this limit.
func NewTxCompressor(dataLimit int, dictPath string) (*TxCompressor, error) {
	tc := &TxCompressor{
		Limit: dataLimit,
	}
	tc.buf.Grow(1 << 17)

	// initialize compressor with dictionary
	dict, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dictionary: %w", err)
	}
	dict = lzss.AugmentDict(dict)
	tc.dict = dict

	tc.compressor, err = lzss.NewCompressor(dict)
	if err != nil {
		return nil, fmt.Errorf("failed to create compressor: %w", err)
	}

	return tc, nil
}

// Reset resets the compressor to its initial state.
func (tc *TxCompressor) Reset() {
	tc.buf.Reset()
	tc.compressor.Reset()
}

// Len returns the current length of the compressed data.
func (tc *TxCompressor) Len() int {
	return tc.compressor.Len()
}

// Written returns the number of uncompressed bytes written to the compressor.
func (tc *TxCompressor) Written() int {
	return tc.compressor.Written()
}

// Bytes returns the compressed data.
// Note: this returns a slice of the internal buffer; the caller should copy if needed.
func (tc *TxCompressor) Bytes() []byte {
	return tc.compressor.Bytes()
}

// fitsInLimit checks if the current compressed size (with packing overhead) fits within the limit.
func (tc *TxCompressor) fitsInLimit() bool {
	return encode.PackAlignSize(tc.compressor.Len(), fr381.Bits-1) <= tc.Limit
}

// Write attempts to append an RLP-encoded transaction to the compressed data.
// Returns true if the transaction was appended, false if it would exceed the limit.
// If forceReset is true, the transaction is not actually appended but the return value
// indicates whether it could have been appended.
func (tc *TxCompressor) Write(rlpTx []byte, forceReset bool) (ok bool, err error) {
	// Snapshot the current state BEFORE any modifications.
	// We need both uncompressed payload (for recompression) and compressed bytes
	// (for exact state restoration if recompression fails).
	prevWritten := tc.compressor.Written()
	prevLen := tc.compressor.Len()
	var snapshotPayload []byte
	var snapshotCompressed []byte
	if prevWritten > 0 {
		snapshotPayload = make([]byte, prevWritten)
		copy(snapshotPayload, tc.compressor.WrittenBytes())
		snapshotCompressed = make([]byte, prevLen)
		copy(snapshotCompressed, tc.compressor.Bytes())
	}

	// decode the RLP transaction
	var tx types.Transaction
	if err = rlp.Decode(bytes.NewReader(rlpTx), &tx); err != nil {
		return false, fmt.Errorf("failed to decode RLP transaction: %w", err)
	}

	// encode the transaction for compression (from address + tx RLP for signing)
	tc.buf.Reset()
	from := ethereum.GetFrom(&tx)
	txRlp := ethereum.EncodeTxForSigning(&tx)
	tc.buf.Write(from[:])
	tc.buf.Write(txRlp)

	// write to compressor
	if _, err = tc.compressor.Write(tc.buf.Bytes()); err != nil {
		if innerErr := tc.compressor.Revert(); innerErr != nil {
			return false, fmt.Errorf("failed to revert after write error: %w (original: %w)", innerErr, err)
		}
		return false, fmt.Errorf("failed to write to compressor: %w", err)
	}

	// check if we fit in the limit
	if tc.fitsInLimit() {
		// Fits without recompression
		if forceReset {
			if err = tc.compressor.Revert(); err != nil {
				return false, fmt.Errorf("failed to revert after forceReset: %w", err)
			}
		}
		return true, nil
	}

	// Doesn't fit with incremental compression - try recompression for better ratio.
	// Capture full payload (including new tx) before reset.
	fullPayload := make([]byte, tc.compressor.Written())
	copy(fullPayload, tc.compressor.WrittenBytes())

	// Recompress everything in one go
	tc.compressor.Reset()
	if _, err = tc.compressor.Write(fullPayload); err != nil {
		// Restore previous state
		tc.restoreSnapshot(snapshotPayload, snapshotCompressed)
		return false, fmt.Errorf("failed to recompress: %w", err)
	}

	if tc.fitsInLimit() {
		if forceReset {
			// Restore previous state for CanWrite
			tc.restoreSnapshot(snapshotPayload, snapshotCompressed)
		}
		return true, nil
	}

	// Try bypassing compression
	if tc.compressor.ConsiderBypassing() {
		if tc.fitsInLimit() {
			if forceReset {
				// Restore previous state for CanWrite
				tc.restoreSnapshot(snapshotPayload, snapshotCompressed)
			}
			return true, nil
		}
	}

	// Doesn't fit even after recompression - restore previous state
	tc.restoreSnapshot(snapshotPayload, snapshotCompressed)
	return false, nil
}

// restoreSnapshot restores the compressor to a previously saved state.
// The semantic content (Written/WrittenBytes) is restored exactly.
// The compressed size (Len) may vary slightly because recompressing the same
// data can produce different output than incremental compression.
// The compressed parameter is kept for potential future use if the lzss library
// adds support for direct state restoration.
func (tc *TxCompressor) restoreSnapshot(payload, _ []byte) {
	tc.compressor.Reset()
	if len(payload) > 0 {
		tc.compressor.Write(payload)
	}
}

// CanWrite checks if an RLP-encoded transaction can be appended without actually appending it.
// This is equivalent to Write(rlpTx, true).
func (tc *TxCompressor) CanWrite(rlpTx []byte) (bool, error) {
	return tc.Write(rlpTx, true)
}

// RawCompressedSize compresses the (raw) input statelessly and returns the length of the compressed data.
// The returned length accounts for the "padding" used to fit the data in field elements.
// If an error occurred, returns -1.
//
// This function is fully stateless and does not modify the compressor's internal state.
// It uses the same compression algorithm and dictionary as the stateful Write method.
// Input size must be less than 256kB (sufficient for any single transaction).
// It is useful for estimating the compressed size of a transaction for profitability calculations.
func (tc *TxCompressor) RawCompressedSize(data []byte) (int, error) {
	n, err := tc.compressor.CompressedSize256k(data)
	if err != nil {
		return -1, err
	}
	if n > len(data) {
		// Fallback to "no compression" if compressed is larger than original
		n = len(data) + lzss.HeaderSize
	}
	return encode.PackAlignSize(n, fr381.Bits-1), nil
}

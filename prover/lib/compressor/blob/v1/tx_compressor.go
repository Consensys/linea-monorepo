package v1

import (
	"fmt"
	"os"

	"github.com/consensys/compress/lzss"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
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
	Limit            int              // maximum size of the compressed data
	compressor       *lzss.Compressor // compressor used to compress transactions
	dict             []byte           // dictionary used for compression
	enableRecompress bool             // whether to attempt recompression when near limit

	// Reusable buffers to avoid allocations
	snapshotPayload    []byte // reusable buffer for snapshot payload
	snapshotCompressed []byte // reusable buffer for snapshot compressed data
	fullPayload        []byte // reusable buffer for full payload during recompression
}

// NewTxCompressor returns a new transaction compressor.
// The dataLimit argument is the maximum size of the compressed data.
// The caller should account for blob overhead (~500 bytes) when setting this limit.
// enableRecompress controls whether the compressor attempts recompression when
// incremental compression exceeds the limit. Recompression can achieve better
// compression ratios but is expensive. Set to false for faster operation.
func NewTxCompressor(dataLimit int, dictPath string, enableRecompress bool) (*TxCompressor, error) {
	tc := &TxCompressor{
		Limit:            dataLimit,
		enableRecompress: enableRecompress,
	}

	// Pre-allocate reusable buffers for snapshots.
	// These hold UNCOMPRESSED data which can be larger than the compressed limit.
	// Initial capacity of 128KB covers most cases; buffers grow dynamically if needed.
	tc.snapshotPayload = make([]byte, 0, 1<<17)
	tc.snapshotCompressed = make([]byte, 0, 1<<17)
	tc.fullPayload = make([]byte, 0, 1<<17)

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
	tc.compressor.Reset()
	// Reset reusable buffers (keep capacity)
	tc.snapshotPayload = tc.snapshotPayload[:0]
	tc.snapshotCompressed = tc.snapshotCompressed[:0]
	tc.fullPayload = tc.fullPayload[:0]
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

// ensureCapacity returns a slice with at least the given capacity.
// If the existing slice has enough capacity, it returns a slice of that.
// Otherwise, it allocates a new slice.
func ensureCapacity(buf []byte, size int) []byte {
	if cap(buf) >= size {
		return buf[:size]
	}
	// Grow to at least 2x the required size to reduce future allocations
	newCap := size * 2
	return make([]byte, size, newCap)
}

// WriteRaw attempts to append pre-encoded transaction data to the compressed data.
// The txData should be: from address (20 bytes) + RLP-encoded transaction for signing.
// This is the fast path that avoids RLP decoding and signature recovery.
// Returns true if the transaction was appended, false if it would exceed the limit.
// If forceReset is true, the transaction is not actually appended but the return value
// indicates whether it could have been appended.
func (tc *TxCompressor) WriteRaw(txData []byte, forceReset bool) (ok bool, err error) {
	// Snapshot the current state BEFORE any modifications (only if recompression is enabled)
	prevWritten := tc.compressor.Written()
	prevLen := tc.compressor.Len()

	if tc.enableRecompress && prevWritten > 0 {
		// Reuse snapshot buffers, growing if necessary
		tc.snapshotPayload = ensureCapacity(tc.snapshotPayload, prevWritten)
		copy(tc.snapshotPayload, tc.compressor.WrittenBytes())
		tc.snapshotCompressed = ensureCapacity(tc.snapshotCompressed, prevLen)
		copy(tc.snapshotCompressed, tc.compressor.Bytes())
	}

	// write to compressor directly
	if _, err = tc.compressor.Write(txData); err != nil {
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

	// Doesn't fit with incremental compression
	if !tc.enableRecompress {
		// Recompression disabled - just revert and return false
		if err = tc.compressor.Revert(); err != nil {
			return false, fmt.Errorf("failed to revert: %w", err)
		}
		return false, nil
	}

	// Try recompression for better ratio.
	// Capture full payload (including new tx) before reset.
	fullWritten := tc.compressor.Written()
	tc.fullPayload = ensureCapacity(tc.fullPayload, fullWritten)
	copy(tc.fullPayload, tc.compressor.WrittenBytes())

	// Recompress everything in one go
	tc.compressor.Reset()
	if _, err = tc.compressor.Write(tc.fullPayload); err != nil {
		// Restore previous state
		tc.restoreSnapshot(tc.snapshotPayload)
		return false, fmt.Errorf("failed to recompress: %w", err)
	}

	if tc.fitsInLimit() {
		if forceReset {
			// Restore previous state for CanWrite
			tc.restoreSnapshot(tc.snapshotPayload)
		}
		return true, nil
	}

	// Try bypassing compression
	if tc.compressor.ConsiderBypassing() {
		if tc.fitsInLimit() {
			if forceReset {
				// Restore previous state for CanWrite
				tc.restoreSnapshot(tc.snapshotPayload)
			}
			return true, nil
		}
	}

	// Doesn't fit even after recompression - restore previous state
	tc.restoreSnapshot(tc.snapshotPayload)
	return false, nil
}

// restoreSnapshot restores the compressor to a previously saved state.
func (tc *TxCompressor) restoreSnapshot(payload []byte) {
	tc.compressor.Reset()
	if len(payload) > 0 {
		tc.compressor.Write(payload)
	}
}

// CanWriteRaw checks if pre-encoded transaction data can be appended without actually appending it.
// This is equivalent to WriteRaw(txData, true).
func (tc *TxCompressor) CanWriteRaw(txData []byte) (bool, error) {
	return tc.WriteRaw(txData, true)
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

// StartRawWrite begins a raw write operation by writing the from address.
// This is an optimization for callers who want to stream the transaction data.
// After calling this, use ContinueRawWrite to append the RLP data, then FinishRawWrite.
// This avoids the need to concatenate from + rlp in the caller.
func (tc *TxCompressor) StartRawWrite(from []byte) error {
	if len(from) != 20 {
		return fmt.Errorf("from address must be 20 bytes, got %d", len(from))
	}
	// This is a placeholder for potential future streaming optimization
	// For now, callers should use WriteRaw with concatenated data
	return nil
}

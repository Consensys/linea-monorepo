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
// blob overhead (header + block metadata, approximately 100 bytes) when setting the limit.
type TxCompressor struct {
	Limit      int              // maximum size of the compressed data
	compressor *lzss.Compressor // compressor used to compress transactions
	dict       []byte           // dictionary used for compression
	buf        bytes.Buffer     // reusable buffer for encoding
}

// NewTxCompressor returns a new transaction compressor.
// The dataLimit argument is the maximum size of the compressed data.
// The caller should account for blob overhead (~100 bytes) when setting this limit.
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
	// Save the current written length before any modifications (for reverting)
	prevWrittenLen := tc.compressor.Written()

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

	// Capture payload before any recompression (needed for proper revert)
	payload := tc.compressor.WrittenBytes()
	recompressionAttempted := false

	// Helper to revert to previous state
	revert := func() error {
		if !recompressionAttempted {
			// Fast path: compressor's own Revert works
			return tc.compressor.Revert()
		}
		// After recompression, we can't use Revert - must manually restore
		tc.compressor.Reset()
		if prevWrittenLen > 0 {
			if _, err := tc.compressor.Write(payload[:prevWrittenLen]); err != nil {
				return fmt.Errorf("failed to restore previous state: %w", err)
			}
		}
		return nil
	}

	// check if we fit in the limit
	if !tc.fitsInLimit() {
		recompressionAttempted = true

		// try to recompress everything in one go for better ratio
		tc.compressor.Reset()
		if _, err = tc.compressor.Write(payload); err != nil {
			// revert to previous state
			if revertErr := revert(); revertErr != nil {
				return false, fmt.Errorf("failed to revert after recompression error: %w (original: %w)", revertErr, err)
			}
			return false, fmt.Errorf("failed to recompress: %w", err)
		}

		if !tc.fitsInLimit() {
			// try bypassing compression
			if tc.compressor.ConsiderBypassing() {
				if tc.fitsInLimit() {
					goto success
				}
			}

			// doesn't fit, revert to previous state
			if err := revert(); err != nil {
				return false, err
			}
			return false, nil
		}
	}

success:
	if forceReset {
		// we don't want to append the data, but we could have
		if err = revert(); err != nil {
			return false, fmt.Errorf("failed to revert after forceReset: %w", err)
		}
		return true, nil
	}

	return true, nil
}

// CanWrite checks if an RLP-encoded transaction can be appended without actually appending it.
// This is equivalent to Write(rlpTx, true).
func (tc *TxCompressor) CanWrite(rlpTx []byte) (bool, error) {
	return tc.Write(rlpTx, true)
}

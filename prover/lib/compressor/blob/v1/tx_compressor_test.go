//go:build !fuzzlight

package v1_test

import (
	"bytes"
	cRand "crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	v1Testing "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encodeTxForCompressor encodes a transaction in the format expected by TxCompressor:
// from address (20 bytes) + RLP-encoded transaction for signing
func encodeTxForCompressor(tx *types.Transaction) []byte {
	from := ethereum.GetFrom(tx)
	txRlp := ethereum.EncodeTxForSigning(tx)
	result := make([]byte, len(from)+len(txRlp))
	copy(result[:20], from[:])
	copy(result[20:], txRlp)
	return result
}

func TestTxCompressorBasic(t *testing.T) {
	tc, err := v1.NewTxCompressor(64*1024, testDictPath, true)
	assert.NoError(t, err, "init should succeed")

	// Create a simple transaction
	tx := makeTestTx(100)
	txData := encodeTxForCompressor(tx)

	// WriteRaw should succeed
	ok, err := tc.WriteRaw(txData, false)
	assert.NoError(t, err)
	assert.True(t, ok, "transaction should be appended")
	assert.Greater(t, tc.Len(), 0, "compressed size should be > 0")
	assert.Greater(t, tc.Written(), 0, "written bytes should be > 0")
}

func TestTxCompressorCanWriteRaw(t *testing.T) {
	tc, err := v1.NewTxCompressor(64*1024, testDictPath, true)
	assert.NoError(t, err, "init should succeed")

	tx := makeTestTx(100)
	txData := encodeTxForCompressor(tx)

	// CanWriteRaw should not mutate state
	initialLen := tc.Len()
	initialWritten := tc.Written()

	canWrite, err := tc.CanWriteRaw(txData)
	assert.NoError(t, err)
	assert.True(t, canWrite)

	assert.Equal(t, initialLen, tc.Len(), "CanWriteRaw should not change Len")
	assert.Equal(t, initialWritten, tc.Written(), "CanWriteRaw should not change Written")
}

func TestTxCompressorWithRecompressionDisabled(t *testing.T) {
	// Test with recompression disabled - should be faster but may fit fewer txs
	tc, err := v1.NewTxCompressor(8*1024, testDictPath, false)
	require.NoError(t, err)

	var txCount int
	for i := 0; i < 200; i++ {
		tx := makeRandomizedTransferTx(uint64(i))
		txData := encodeTxForCompressor(tx)

		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err)
		if !ok {
			break
		}
		txCount++
	}

	t.Logf("With recompression disabled: accepted %d transactions, size: %d bytes", txCount, tc.Len())
	require.Greater(t, txCount, 0, "should accept at least one transaction")
}

func TestTxCompressorWithRecompressionEnabled(t *testing.T) {
	// Test with recompression enabled - should fit more txs
	tc, err := v1.NewTxCompressor(8*1024, testDictPath, true)
	require.NoError(t, err)

	var txCount int
	for i := 0; i < 200; i++ {
		tx := makeRandomizedTransferTx(uint64(i))
		txData := encodeTxForCompressor(tx)

		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err)
		if !ok {
			break
		}
		txCount++
	}

	t.Logf("With recompression enabled: accepted %d transactions, size: %d bytes", txCount, tc.Len())
	require.Greater(t, txCount, 0, "should accept at least one transaction")
}

// TestTxCompressorCanWriteRawDoesNotCorruptStateNearLimit tests that CanWriteRaw
// does not corrupt compressor state when called near the limit where
// recompression is triggered. This is a regression test for a bug where
// CanWriteRaw would corrupt state after recompression was attempted.
func TestTxCompressorCanWriteRawDoesNotCorruptStateNearLimit(t *testing.T) {
	// Use a small limit to trigger recompression quickly
	const limit = 8 * 1024
	tc, err := v1.NewTxCompressor(limit, testDictPath, true)
	require.NoError(t, err)

	// Generate transactions
	var txDataList [][]byte
	for i := 0; i < 200; i++ {
		tx := makeRandomizedTransferTx(uint64(i))
		txDataList = append(txDataList, encodeTxForCompressor(tx))
	}

	// Fill the compressor, checking CanWriteRaw before each WriteRaw
	var prevLen int
	for i, txData := range txDataList {
		lenBefore := tc.Len()
		writtenBefore := tc.Written()

		// Call CanWriteRaw - this should NOT change semantic state (Written)
		// Compressed size (Len) may vary slightly due to recompression attempt
		canWrite, err := tc.CanWriteRaw(txData)
		require.NoError(t, err, "CanWriteRaw should not error on tx %d", i)

		// Verify CanWriteRaw did not change semantic state
		require.Equal(t, writtenBefore, tc.Written(),
			"CanWriteRaw should not change Written (tx %d, canWrite=%v)", i, canWrite)
		// Len may vary slightly due to recompression, but should be close
		lenAfterCanWrite := tc.Len()
		require.InDelta(t, lenBefore, lenAfterCanWrite, float64(lenBefore)*0.05,
			"CanWriteRaw should not significantly change Len (tx %d, canWrite=%v): before=%d, after=%d",
			i, canWrite, lenBefore, lenAfterCanWrite)

		// Now actually write
		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err, "WriteRaw should not error on tx %d", i)

		// Verify consistency between CanWriteRaw and WriteRaw
		require.Equal(t, canWrite, ok,
			"CanWriteRaw and WriteRaw should return same result (tx %d)", i)

		if !ok {
			// Transaction didn't fit - semantic content (Written) must be preserved exactly.
			// Compressed size (Len) may vary slightly after recompression attempt,
			// but should not decrease dramatically (which would indicate corruption).
			require.Equal(t, writtenBefore, tc.Written(),
				"Written should be unchanged when tx doesn't fit (tx %d)", i)
			lenAfter := tc.Len()
			// Allow small variation in Len due to recompression, but catch major corruption
			require.InDelta(t, lenBefore, lenAfter, float64(lenBefore)*0.05,
				"Len should be approximately unchanged when tx doesn't fit (tx %d): before=%d, after=%d", i, lenBefore, lenAfter)
			t.Logf("Compressor full after %d transactions, final size: %d bytes", i, tc.Len())
			break
		}

		// Transaction was written - size should have increased (or stayed same due to recompression)
		currentLen := tc.Len()
		currentWritten := tc.Written()

		// Detect unexpected size decrease (the bug we're testing for)
		if i > 0 && currentLen < prevLen/2 {
			t.Fatalf("Unexpected size decrease at tx %d: %d -> %d (this indicates CanWriteRaw corrupted state)",
				i, prevLen, currentLen)
		}

		// Written should always increase when a tx is added
		require.Greater(t, currentWritten, writtenBefore,
			"Written should increase when tx is added (tx %d)", i)

		prevLen = currentLen
	}

	// Verify we actually filled the compressor (test is meaningful)
	require.Greater(t, tc.Len(), limit/2, "Compressor should be at least half full")
}

// TestTxCompressorCanWriteRawRepeatedCallsNearLimit tests that repeated CanWriteRaw
// calls near the limit don't accumulate state corruption.
func TestTxCompressorCanWriteRawRepeatedCallsNearLimit(t *testing.T) {
	const limit = 8 * 1024
	tc, err := v1.NewTxCompressor(limit, testDictPath, true)
	require.NoError(t, err)

	// Fill compressor to near capacity
	var lastAcceptedTxData []byte
	for i := 0; i < 100; i++ {
		tx := makeRandomizedTransferTx(uint64(i))
		txData := encodeTxForCompressor(tx)

		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err)
		if !ok {
			break
		}
		lastAcceptedTxData = txData
	}

	// Record state after filling
	lenAfterFill := tc.Len()
	writtenAfterFill := tc.Written()
	t.Logf("After filling: Len=%d, Written=%d", lenAfterFill, writtenAfterFill)

	// Create a transaction that won't fit
	largeTx := makeTestTx(5000) // large tx unlikely to fit
	largeTxData := encodeTxForCompressor(largeTx)

	// Call CanWriteRaw multiple times with the large tx - should all return false
	// and should NOT change state
	for i := 0; i < 10; i++ {
		canWrite, err := tc.CanWriteRaw(largeTxData)
		require.NoError(t, err, "CanWriteRaw iteration %d", i)
		require.False(t, canWrite, "Large tx should not fit, iteration %d", i)

		// State should be unchanged
		require.Equal(t, lenAfterFill, tc.Len(),
			"Len should be unchanged after CanWriteRaw iteration %d", i)
		require.Equal(t, writtenAfterFill, tc.Written(),
			"Written should be unchanged after CanWriteRaw iteration %d", i)
	}

	// Also test with a small tx that might trigger recompression
	if lastAcceptedTxData != nil {
		for i := 0; i < 10; i++ {
			lenBefore := tc.Len()
			writtenBefore := tc.Written()

			_, err := tc.CanWriteRaw(lastAcceptedTxData)
			require.NoError(t, err, "CanWriteRaw with small tx iteration %d", i)

			// State should be unchanged regardless of result
			require.Equal(t, lenBefore, tc.Len(),
				"Len should be unchanged after CanWriteRaw with small tx iteration %d", i)
			require.Equal(t, writtenBefore, tc.Written(),
				"Written should be unchanged after CanWriteRaw with small tx iteration %d", i)
		}
	}
}

func TestTxCompressorReset(t *testing.T) {
	tc, err := v1.NewTxCompressor(64*1024, testDictPath, true)
	assert.NoError(t, err)

	tx := makeTestTx(100)
	txData := encodeTxForCompressor(tx)

	// Get baseline after init (compressor has a small header)
	baselineLen := tc.Len()
	baselineWritten := tc.Written()

	// Write a transaction
	ok, err := tc.WriteRaw(txData, false)
	assert.NoError(t, err)
	assert.True(t, ok)

	lenAfterWrite := tc.Len()
	writtenAfterWrite := tc.Written()
	assert.Greater(t, lenAfterWrite, baselineLen)
	assert.Greater(t, writtenAfterWrite, baselineWritten)

	// Reset
	tc.Reset()

	// After reset, should be back to baseline
	assert.Equal(t, baselineLen, tc.Len(), "Len should be back to baseline after reset")
	assert.Equal(t, baselineWritten, tc.Written(), "Written should be back to baseline after reset")

	// Write same transaction again, should get same compressed size
	ok, err = tc.WriteRaw(txData, false)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, lenAfterWrite, tc.Len(), "compressed size should be same after reset")
}

func TestTxCompressorLimitExceeded(t *testing.T) {
	// Use a very small limit
	tc, err := v1.NewTxCompressor(100, testDictPath, true)
	assert.NoError(t, err)

	// Get baseline after init (compressor has a small header)
	baselineLen := tc.Len()

	// Create a large transaction that won't fit
	tx := makeTestTx(10000)
	txData := encodeTxForCompressor(tx)

	// WriteRaw should fail (return false) but not error
	ok, err := tc.WriteRaw(txData, false)
	assert.NoError(t, err)
	assert.False(t, ok, "large transaction should not fit in small limit")
	assert.Equal(t, baselineLen, tc.Len(), "Len should remain at baseline when limit exceeded")
}

func TestTxCompressorMultipleTransactions(t *testing.T) {
	tc, err := v1.NewTxCompressor(64*1024, testDictPath, true)
	assert.NoError(t, err)

	var txDataList [][]byte

	// Create multiple transactions
	for i := 0; i < 10; i++ {
		tx := makeTestTx(100 + i*50)
		txDataList = append(txDataList, encodeTxForCompressor(tx))
	}

	// Write all transactions
	prevLen := 0
	for i, txData := range txDataList {
		ok, err := tc.WriteRaw(txData, false)
		assert.NoError(t, err)
		assert.True(t, ok, "transaction %d should be appended", i)
		assert.Greater(t, tc.Len(), prevLen, "compressed size should increase")
		prevLen = tc.Len()
	}
}

func TestTxCompressorCompressionContextBenefit(t *testing.T) {
	// This test verifies that maintaining compression context across transactions
	// results in better compression than compressing each transaction individually

	tc, err := v1.NewTxCompressor(128*1024, testDictPath, true)
	assert.NoError(t, err)

	// Create transactions with similar data (should compress well with context)
	var txDataList [][]byte
	for i := 0; i < 20; i++ {
		tx := makeTestTx(500) // same size, similar structure
		txDataList = append(txDataList, encodeTxForCompressor(tx))
	}

	// Measure individual compression sizes
	individualTotal := 0
	for _, txData := range txDataList {
		tc.Reset()
		ok, err := tc.WriteRaw(txData, false)
		assert.NoError(t, err)
		assert.True(t, ok)
		individualTotal += tc.Len()
	}

	// Measure additive compression size
	tc.Reset()
	for _, txData := range txDataList {
		ok, err := tc.WriteRaw(txData, false)
		assert.NoError(t, err)
		assert.True(t, ok)
	}
	additiveTotal := tc.Len()

	// Additive compression should be smaller due to shared context
	t.Logf("Individual total: %d bytes, Additive total: %d bytes, Savings: %.1f%%",
		individualTotal, additiveTotal, 100.0*(1.0-float64(additiveTotal)/float64(individualTotal)))
	assert.Less(t, additiveTotal, individualTotal, "additive compression should be smaller than individual")
}

// TestTxCompressorCompatibilityWithBlobMaker is the critical compatibility test.
// It proves that transactions compressed with TxCompressor will fit into BlobMaker
// when assembled into a block.
func TestTxCompressorCompatibilityWithBlobMaker(t *testing.T) {
	const blobLimit = 128 * 1024
	const overhead = 100 // conservative blob overhead estimate

	// Create TxCompressor with limit accounting for overhead
	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath, true)
	require.NoError(t, err)

	// Create BlobMaker with full blob limit
	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Generate test transactions
	var txs []*types.Transaction
	var txDataList [][]byte
	for i := 0; i < 100; i++ {
		tx := makeTestTx(200 + i*10)
		txs = append(txs, tx)
		txDataList = append(txDataList, encodeTxForCompressor(tx))
	}

	// Add transactions to TxCompressor until it says "full"
	var acceptedTxs []*types.Transaction
	for i, txData := range txDataList {
		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err)
		if !ok {
			t.Logf("TxCompressor full after %d transactions", i)
			break
		}
		acceptedTxs = append(acceptedTxs, txs[i])
	}

	require.Greater(t, len(acceptedTxs), 0, "should accept at least one transaction")
	t.Logf("TxCompressor accepted %d transactions, compressed size: %d bytes", len(acceptedTxs), tc.Len())

	// Build a block with the accepted transactions
	block := types.NewBlock(
		&types.Header{Time: 12345},
		&types.Body{Transactions: acceptedTxs},
		nil,
		trie.NewStackTrie(nil),
	)

	// RLP encode the block
	var blockBuf bytes.Buffer
	err = rlp.Encode(&blockBuf, block)
	require.NoError(t, err)

	// Verify BlobMaker accepts the block
	ok, err := bm.Write(blockBuf.Bytes(), false)
	require.NoError(t, err)
	require.True(t, ok, "Block built with TxCompressor should fit in BlobMaker")

	t.Logf("BlobMaker accepted block, blob size: %d bytes", bm.Len())

	// Verify we can decompress the blob
	dict, err := os.ReadFile(testDictPath)
	require.NoError(t, err)
	dictStore, err := dictionary.SingletonStore(dict, 1)
	require.NoError(t, err)

	resp, err := v1.DecompressBlob(bm.Bytes(), dictStore)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Blocks), "should have 1 block")
}

// TestTxCompressorWorstCaseManyRandomizedTxs is the critical worst-case test.
// Randomized transactions prevent compression from being too efficient,
// ensuring we test the scenario where TxCompressor context sharing provides
// minimal benefit. This test ensures that blocks built with TxCompressor
// still fit in BlobMaker even in this worst-case scenario.
func TestTxCompressorWorstCaseManyRandomizedTxs(t *testing.T) {
	const blobLimit = 128 * 1024
	const overhead = 100

	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath, true)
	require.NoError(t, err)

	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Generate many randomized transfer transactions (worst case for compression)
	// Randomized fields prevent compression from being too efficient
	var txs []*types.Transaction
	var txDataList [][]byte
	var totalPlainTxSize int
	for i := 0; i < 5000; i++ {
		tx := makeRandomizedTransferTx(uint64(i))
		txs = append(txs, tx)
		txData := encodeTxForCompressor(tx)
		txDataList = append(txDataList, txData)
		totalPlainTxSize += len(txData)
	}

	// Add transactions to TxCompressor until it says "full"
	var acceptedTxs []*types.Transaction
	var acceptedPlainSize int
	for i, txData := range txDataList {
		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err)
		if !ok {
			t.Logf("TxCompressor full after %d transactions", i)
			break
		}
		acceptedTxs = append(acceptedTxs, txs[i])
		acceptedPlainSize += len(txData)
	}

	require.Greater(t, len(acceptedTxs), 0, "should accept at least one transaction")
	txCompressorSize := tc.Len()

	// Build a block with the accepted transactions
	block := types.NewBlock(
		&types.Header{Time: 12345},
		&types.Body{Transactions: acceptedTxs},
		nil,
		trie.NewStackTrie(nil),
	)

	var blockBuf bytes.Buffer
	err = rlp.Encode(&blockBuf, block)
	require.NoError(t, err)
	blockRlpSize := blockBuf.Len()

	// Verify BlobMaker accepts the block
	ok, err := bm.Write(blockBuf.Bytes(), false)
	require.NoError(t, err)
	require.True(t, ok, "Block built with TxCompressor (worst case: many small plain txs) should fit in BlobMaker")

	blobMakerSize := bm.Len()

	// Log comprehensive comparison
	t.Logf("=== Compression Comparison (Worst Case: %d small plain txs) ===", len(acceptedTxs))
	t.Logf("Plain transaction data size:     %d bytes", acceptedPlainSize)
	t.Logf("Block RLP size (with header):    %d bytes", blockRlpSize)
	t.Logf("TxCompressor compressed size:    %d bytes (ratio: %.2f%%)", txCompressorSize, 100.0*float64(txCompressorSize)/float64(acceptedPlainSize))
	t.Logf("BlobMaker compressed size:       %d bytes (ratio: %.2f%%)", blobMakerSize, 100.0*float64(blobMakerSize)/float64(blockRlpSize))
	t.Logf("TxCompressor vs BlobMaker diff:  %d bytes (%.2f%%)", blobMakerSize-txCompressorSize, 100.0*float64(blobMakerSize-txCompressorSize)/float64(txCompressorSize))
	t.Logf("Blob limit:                      %d bytes", blobLimit)
	t.Logf("Headroom remaining:              %d bytes (%.2f%%)", blobLimit-blobMakerSize, 100.0*float64(blobLimit-blobMakerSize)/float64(blobLimit))

	// Verify we can decompress the blob
	dict, err := os.ReadFile(testDictPath)
	require.NoError(t, err)
	dictStore, err := dictionary.SingletonStore(dict, 1)
	require.NoError(t, err)

	resp, err := v1.DecompressBlob(bm.Bytes(), dictStore)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Blocks), "should have 1 block")

	// Decode the block to verify transaction count
	decodedBlock, err := v1.DecodeBlockFromUncompressed(bytes.NewReader(resp.Blocks[0]))
	require.NoError(t, err)
	require.Equal(t, len(acceptedTxs), len(decodedBlock.Txs), "should have same number of transactions")
}

// TestTxCompressorCompatibilityWithVariousTxTypes tests compatibility with different transaction types
func TestTxCompressorCompatibilityWithVariousTxTypes(t *testing.T) {
	const blobLimit = 128 * 1024
	const overhead = 100

	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath, true)
	require.NoError(t, err)

	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Create various transaction types
	var txs []*types.Transaction
	var txDataList [][]byte

	// Legacy transactions
	for i := 0; i < 10; i++ {
		tx := makeLegacyTx(100 + i*20)
		txs = append(txs, tx)
		txDataList = append(txDataList, encodeTxForCompressor(tx))
	}

	// EIP-1559 transactions
	for i := 0; i < 10; i++ {
		tx := makeEIP1559Tx(100 + i*20)
		txs = append(txs, tx)
		txDataList = append(txDataList, encodeTxForCompressor(tx))
	}

	// Add transactions to TxCompressor
	var acceptedTxs []*types.Transaction
	for i, txData := range txDataList {
		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err)
		if !ok {
			break
		}
		acceptedTxs = append(acceptedTxs, txs[i])
	}

	require.Greater(t, len(acceptedTxs), 0)

	// Build block and verify BlobMaker accepts it
	block := types.NewBlock(
		&types.Header{Time: 12345},
		&types.Body{Transactions: acceptedTxs},
		nil,
		trie.NewStackTrie(nil),
	)

	var blockBuf bytes.Buffer
	rlp.Encode(&blockBuf, block)

	ok, err := bm.Write(blockBuf.Bytes(), false)
	require.NoError(t, err)
	require.True(t, ok, "Block with various tx types should fit in BlobMaker")
}

// TestTxCompressorCompatibilityWithHighEntropyData tests worst-case compression scenario
func TestTxCompressorCompatibilityWithHighEntropyData(t *testing.T) {
	const blobLimit = 128 * 1024
	const overhead = 100

	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath, true)
	require.NoError(t, err)

	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Create transactions with random (high entropy) data
	var txs []*types.Transaction
	var txDataList [][]byte

	for i := 0; i < 50; i++ {
		tx := makeHighEntropyTx(500)
		txs = append(txs, tx)
		txDataList = append(txDataList, encodeTxForCompressor(tx))
	}

	// Add transactions to TxCompressor
	var acceptedTxs []*types.Transaction
	for i, txData := range txDataList {
		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err)
		if !ok {
			break
		}
		acceptedTxs = append(acceptedTxs, txs[i])
	}

	if len(acceptedTxs) == 0 {
		t.Skip("No transactions fit - high entropy data expands too much")
	}

	// Build block and verify BlobMaker accepts it
	block := types.NewBlock(
		&types.Header{Time: 12345},
		&types.Body{Transactions: acceptedTxs},
		nil,
		trie.NewStackTrie(nil),
	)

	var blockBuf bytes.Buffer
	rlp.Encode(&blockBuf, block)

	ok, err := bm.Write(blockBuf.Bytes(), false)
	require.NoError(t, err)
	require.True(t, ok, "Block with high entropy data should fit in BlobMaker")
}

// TestTxCompressorCompatibilityWithRealTestData uses real test blocks
func TestTxCompressorCompatibilityWithRealTestData(t *testing.T) {
	const blobLimit = 128 * 1024
	const overhead = 100

	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath, true)
	require.NoError(t, err)

	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Load real test blocks
	repoRoot, err := test_utils.GetRepoRootPath()
	require.NoError(t, err)

	testBlocks, err := v1Testing.LoadTestBlocks(filepath.Join(repoRoot, "testdata/prover-v2/prover-execution/requests"))
	require.NoError(t, err)
	require.Greater(t, len(testBlocks), 0, "should have test blocks")

	// Extract transactions from test blocks
	var allTxs []*types.Transaction
	var allTxDataList [][]byte

	for _, blockRlp := range testBlocks {
		var block types.Block
		err := rlp.Decode(bytes.NewReader(blockRlp), &block)
		require.NoError(t, err)

		for _, tx := range block.Transactions() {
			allTxs = append(allTxs, tx)
			allTxDataList = append(allTxDataList, encodeTxForCompressor(tx))
		}
	}

	t.Logf("Loaded %d transactions from %d test blocks", len(allTxs), len(testBlocks))

	// Add transactions to TxCompressor
	var acceptedTxs []*types.Transaction
	for i, txData := range allTxDataList {
		ok, err := tc.WriteRaw(txData, false)
		require.NoError(t, err)
		if !ok {
			t.Logf("TxCompressor full after %d transactions", i)
			break
		}
		acceptedTxs = append(acceptedTxs, allTxs[i])
	}

	require.Greater(t, len(acceptedTxs), 0, "should accept at least one transaction")
	t.Logf("TxCompressor accepted %d transactions, compressed size: %d bytes", len(acceptedTxs), tc.Len())

	// Build a block with the accepted transactions
	block := types.NewBlock(
		&types.Header{Time: 12345},
		&types.Body{Transactions: acceptedTxs},
		nil,
		trie.NewStackTrie(nil),
	)

	var blockBuf bytes.Buffer
	err = rlp.Encode(&blockBuf, block)
	require.NoError(t, err)

	// Verify BlobMaker accepts the block
	ok, err := bm.Write(blockBuf.Bytes(), false)
	require.NoError(t, err)
	require.True(t, ok, "Block built with TxCompressor from real test data should fit in BlobMaker")

	t.Logf("BlobMaker accepted block with %d transactions, blob size: %d bytes", len(acceptedTxs), bm.Len())
}

// TestTxCompressorRawCompressedSizeDoesNotAffectState verifies that calling
// RawCompressedSize does not affect subsequent WriteRaw operations
func TestTxCompressorRawCompressedSizeDoesNotAffectState(t *testing.T) {
	tc1, err := v1.NewTxCompressor(64*1024, testDictPath, true)
	require.NoError(t, err)

	tc2, err := v1.NewTxCompressor(64*1024, testDictPath, true)
	require.NoError(t, err)

	// Create test transactions
	var txDataList [][]byte
	for i := 0; i < 10; i++ {
		tx := makeTestTx(100 + i*20)
		txDataList = append(txDataList, encodeTxForCompressor(tx))
	}

	// tc1: call RawCompressedSize before each WriteRaw
	for _, txData := range txDataList {
		_, err := tc1.RawCompressedSize(txData)
		require.NoError(t, err)

		ok, err := tc1.WriteRaw(txData, false)
		require.NoError(t, err)
		require.True(t, ok)
	}

	// tc2: just WriteRaw without RawCompressedSize
	for _, txData := range txDataList {
		ok, err := tc2.WriteRaw(txData, false)
		require.NoError(t, err)
		require.True(t, ok)
	}

	// Both compressors should have the same state
	require.Equal(t, tc1.Len(), tc2.Len(), "Len should be same regardless of RawCompressedSize calls")
	require.Equal(t, tc1.Written(), tc2.Written(), "Written should be same regardless of RawCompressedSize calls")
}

// Helper functions

func makeTestTx(dataSize int) *types.Transaction {
	return makeLegacyTx(dataSize)
}

// makeRandomizedTransferTx creates a transfer transaction with randomized fields.
// Randomizing all fields prevents compression from being too efficient,
// which is important for testing the worst case.
func makeRandomizedTransferTx(nonce uint64) *types.Transaction {
	// Random recipient address
	addrBytes := make([]byte, 20)
	cRand.Read(addrBytes)
	address := common.BytesToAddress(addrBytes)

	// Random gas price (1-100 gwei range)
	gasPrice := big.NewInt(1000000000 + int64(nonce%99)*1000000000)

	// Random gas limit (21000-100000 range)
	gasLimit := uint64(21000 + nonce%79000)

	// Random value (0-10 ETH range)
	value := big.NewInt(int64(nonce%10) * 1000000000000000000)

	// Random small payload (0-100 bytes)
	payloadSize := int(nonce % 100)
	var data []byte
	if payloadSize > 0 {
		data = make([]byte, payloadSize)
		cRand.Read(data)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		To:       &address,
		Value:    value,
		Data:     data,
	})

	return signTx(tx)
}

func makeLegacyTx(dataSize int) *types.Transaction {
	address := common.HexToAddress("0x000042")
	data := make([]byte, dataSize)
	// Fill with some pattern for better compression
	for i := range data {
		data[i] = byte(i % 256)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    42,
		Gas:      21000,
		GasPrice: big.NewInt(1000000000),
		To:       &address,
		Value:    big.NewInt(0),
		Data:     data,
	})

	return signTx(tx)
}

func makeEIP1559Tx(dataSize int) *types.Transaction {
	address := common.HexToAddress("0x000042")
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	chainID := big.NewInt(1)
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     42,
		Gas:       21000,
		GasFeeCap: big.NewInt(2000000000),
		GasTipCap: big.NewInt(1000000000),
		To:        &address,
		Value:     big.NewInt(0),
		Data:      data,
	})

	return signTxWithSigner(tx, types.NewLondonSigner(chainID))
}

func makeHighEntropyTx(dataSize int) *types.Transaction {
	address := common.HexToAddress("0x000042")
	data := make([]byte, dataSize)
	// Use crypto/rand for high entropy data
	cRand.Read(data)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    42,
		Gas:      21000,
		GasPrice: big.NewInt(1000000000),
		To:       &address,
		Value:    big.NewInt(0),
		Data:     data,
	})

	return signTx(tx)
}

func signTx(tx *types.Transaction) *types.Transaction {
	return signTxWithSigner(tx, types.NewEIP155Signer(big.NewInt(1)))
}

func signTxWithSigner(tx *types.Transaction, signer types.Signer) *types.Transaction {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		panic(err)
	}
	return signedTx
}

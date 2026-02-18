//go:build !fuzzlight

package v1_test

import (
	"bytes"
	cRand "crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"testing"

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

func TestTxCompressorBasic(t *testing.T) {
	tc, err := v1.NewTxCompressor(64*1024, testDictPath)
	assert.NoError(t, err, "init should succeed")

	// Create a simple transaction
	tx := makeTestTx(100)
	var buf bytes.Buffer
	rlp.Encode(&buf, tx)
	rlpTx := buf.Bytes()

	// Write should succeed
	ok, err := tc.Write(rlpTx, false)
	assert.NoError(t, err)
	assert.True(t, ok, "transaction should be appended")
	assert.Greater(t, tc.Len(), 0, "compressed size should be > 0")
	assert.Greater(t, tc.Written(), 0, "written bytes should be > 0")
}

func TestTxCompressorCanWrite(t *testing.T) {
	tc, err := v1.NewTxCompressor(64*1024, testDictPath)
	assert.NoError(t, err, "init should succeed")

	tx := makeTestTx(100)
	var buf bytes.Buffer
	rlp.Encode(&buf, tx)
	rlpTx := buf.Bytes()

	// CanWrite should not mutate state
	initialLen := tc.Len()
	initialWritten := tc.Written()

	canWrite, err := tc.CanWrite(rlpTx)
	assert.NoError(t, err)
	assert.True(t, canWrite)

	assert.Equal(t, initialLen, tc.Len(), "CanWrite should not change Len")
	assert.Equal(t, initialWritten, tc.Written(), "CanWrite should not change Written")
}

// TestTxCompressorCanWriteDoesNotCorruptStateNearLimit tests that CanWrite
// does not corrupt compressor state when called near the limit where
// recompression is triggered. This is a regression test for a bug where
// CanWrite would corrupt state after recompression was attempted.
func TestTxCompressorCanWriteDoesNotCorruptStateNearLimit(t *testing.T) {
	// Use a small limit to trigger recompression quickly
	const limit = 8 * 1024
	tc, err := v1.NewTxCompressor(limit, testDictPath)
	require.NoError(t, err)

	// Generate transactions
	var txs []*types.Transaction
	var rlpTxs [][]byte
	for i := 0; i < 200; i++ {
		tx := makeRandomizedTransferTx(uint64(i))
		txs = append(txs, tx)
		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTxs = append(rlpTxs, buf.Bytes())
	}

	// Fill the compressor, checking CanWrite before each Write
	var prevLen, prevWritten int
	for i, rlpTx := range rlpTxs {
		lenBefore := tc.Len()
		writtenBefore := tc.Written()

		// Call CanWrite - this should NOT change state
		canWrite, err := tc.CanWrite(rlpTx)
		require.NoError(t, err, "CanWrite should not error on tx %d", i)

		// Verify CanWrite did not change state
		require.Equal(t, lenBefore, tc.Len(),
			"CanWrite should not change Len (tx %d, canWrite=%v)", i, canWrite)
		require.Equal(t, writtenBefore, tc.Written(),
			"CanWrite should not change Written (tx %d, canWrite=%v)", i, canWrite)

		// Now actually write
		ok, err := tc.Write(rlpTx, false)
		require.NoError(t, err, "Write should not error on tx %d", i)

		// Verify consistency between CanWrite and Write
		require.Equal(t, canWrite, ok,
			"CanWrite and Write should return same result (tx %d)", i)

		if !ok {
			// Transaction didn't fit - compressor should be unchanged
			require.Equal(t, lenBefore, tc.Len(),
				"Len should be unchanged when tx doesn't fit (tx %d)", i)
			require.Equal(t, writtenBefore, tc.Written(),
				"Written should be unchanged when tx doesn't fit (tx %d)", i)
			t.Logf("Compressor full after %d transactions, final size: %d bytes", i, tc.Len())
			break
		}

		// Transaction was written - size should have increased (or stayed same due to recompression)
		currentLen := tc.Len()
		currentWritten := tc.Written()

		// Detect unexpected size decrease (the bug we're testing for)
		if i > 0 && currentLen < prevLen/2 {
			t.Fatalf("Unexpected size decrease at tx %d: %d -> %d (this indicates CanWrite corrupted state)",
				i, prevLen, currentLen)
		}

		// Written should always increase when a tx is added
		require.Greater(t, currentWritten, writtenBefore,
			"Written should increase when tx is added (tx %d)", i)

		prevLen = currentLen
		prevWritten = currentWritten
	}

	// Verify we actually filled the compressor (test is meaningful)
	require.Greater(t, tc.Len(), limit/2, "Compressor should be at least half full")
}

// TestTxCompressorCanWriteRepeatedCallsNearLimit tests that repeated CanWrite
// calls near the limit don't accumulate state corruption.
func TestTxCompressorCanWriteRepeatedCallsNearLimit(t *testing.T) {
	const limit = 8 * 1024
	tc, err := v1.NewTxCompressor(limit, testDictPath)
	require.NoError(t, err)

	// Fill compressor to near capacity
	var lastAcceptedTx []byte
	for i := 0; i < 100; i++ {
		tx := makeRandomizedTransferTx(uint64(i))
		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTx := buf.Bytes()

		ok, err := tc.Write(rlpTx, false)
		require.NoError(t, err)
		if !ok {
			break
		}
		lastAcceptedTx = rlpTx
	}

	// Record state after filling
	lenAfterFill := tc.Len()
	writtenAfterFill := tc.Written()
	t.Logf("After filling: Len=%d, Written=%d", lenAfterFill, writtenAfterFill)

	// Create a transaction that won't fit
	largeTx := makeTestTx(5000) // large tx unlikely to fit
	var largeBuf bytes.Buffer
	rlp.Encode(&largeBuf, largeTx)
	largeTxRlp := largeBuf.Bytes()

	// Call CanWrite multiple times with the large tx - should all return false
	// and should NOT change state
	for i := 0; i < 10; i++ {
		canWrite, err := tc.CanWrite(largeTxRlp)
		require.NoError(t, err, "CanWrite iteration %d", i)
		require.False(t, canWrite, "Large tx should not fit, iteration %d", i)

		// State should be unchanged
		require.Equal(t, lenAfterFill, tc.Len(),
			"Len should be unchanged after CanWrite iteration %d", i)
		require.Equal(t, writtenAfterFill, tc.Written(),
			"Written should be unchanged after CanWrite iteration %d", i)
	}

	// Also test with a small tx that might trigger recompression
	if lastAcceptedTx != nil {
		for i := 0; i < 10; i++ {
			lenBefore := tc.Len()
			writtenBefore := tc.Written()

			_, err := tc.CanWrite(lastAcceptedTx)
			require.NoError(t, err, "CanWrite with small tx iteration %d", i)

			// State should be unchanged regardless of result
			require.Equal(t, lenBefore, tc.Len(),
				"Len should be unchanged after CanWrite with small tx iteration %d", i)
			require.Equal(t, writtenBefore, tc.Written(),
				"Written should be unchanged after CanWrite with small tx iteration %d", i)
		}
	}
}

func TestTxCompressorReset(t *testing.T) {
	tc, err := v1.NewTxCompressor(64*1024, testDictPath)
	assert.NoError(t, err)

	tx := makeTestTx(100)
	var buf bytes.Buffer
	rlp.Encode(&buf, tx)
	rlpTx := buf.Bytes()

	// Get baseline after init (compressor has a small header)
	baselineLen := tc.Len()
	baselineWritten := tc.Written()

	// Write a transaction
	ok, err := tc.Write(rlpTx, false)
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
	ok, err = tc.Write(rlpTx, false)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, lenAfterWrite, tc.Len(), "compressed size should be same after reset")
}

func TestTxCompressorLimitExceeded(t *testing.T) {
	// Use a very small limit
	tc, err := v1.NewTxCompressor(100, testDictPath)
	assert.NoError(t, err)

	// Get baseline after init (compressor has a small header)
	baselineLen := tc.Len()

	// Create a large transaction that won't fit
	tx := makeTestTx(10000)
	var buf bytes.Buffer
	rlp.Encode(&buf, tx)
	rlpTx := buf.Bytes()

	// Write should fail (return false) but not error
	ok, err := tc.Write(rlpTx, false)
	assert.NoError(t, err)
	assert.False(t, ok, "large transaction should not fit in small limit")
	assert.Equal(t, baselineLen, tc.Len(), "Len should remain at baseline when limit exceeded")
}

func TestTxCompressorMultipleTransactions(t *testing.T) {
	tc, err := v1.NewTxCompressor(64*1024, testDictPath)
	assert.NoError(t, err)

	var txs []*types.Transaction
	var rlpTxs [][]byte

	// Create multiple transactions
	for i := 0; i < 10; i++ {
		tx := makeTestTx(100 + i*50)
		txs = append(txs, tx)

		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTxs = append(rlpTxs, buf.Bytes())
	}

	// Write all transactions
	prevLen := 0
	for i, rlpTx := range rlpTxs {
		ok, err := tc.Write(rlpTx, false)
		assert.NoError(t, err)
		assert.True(t, ok, "transaction %d should be appended", i)
		assert.Greater(t, tc.Len(), prevLen, "compressed size should increase")
		prevLen = tc.Len()
	}
}

func TestTxCompressorCompressionContextBenefit(t *testing.T) {
	// This test verifies that maintaining compression context across transactions
	// results in better compression than compressing each transaction individually

	tc, err := v1.NewTxCompressor(128*1024, testDictPath)
	assert.NoError(t, err)

	// Create transactions with similar data (should compress well with context)
	var rlpTxs [][]byte
	for i := 0; i < 20; i++ {
		tx := makeTestTx(500) // same size, similar structure
		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTxs = append(rlpTxs, buf.Bytes())
	}

	// Measure individual compression sizes
	individualTotal := 0
	for _, rlpTx := range rlpTxs {
		tc.Reset()
		ok, err := tc.Write(rlpTx, false)
		assert.NoError(t, err)
		assert.True(t, ok)
		individualTotal += tc.Len()
	}

	// Measure additive compression size
	tc.Reset()
	for _, rlpTx := range rlpTxs {
		ok, err := tc.Write(rlpTx, false)
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
	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath)
	require.NoError(t, err)

	// Create BlobMaker with full blob limit
	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Generate test transactions
	var txs []*types.Transaction
	var rlpTxs [][]byte
	for i := 0; i < 100; i++ {
		tx := makeTestTx(200 + i*10)
		txs = append(txs, tx)

		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTxs = append(rlpTxs, buf.Bytes())
	}

	// Add transactions to TxCompressor until it says "full"
	var acceptedTxs []*types.Transaction
	for i, rlpTx := range rlpTxs {
		ok, err := tc.Write(rlpTx, false)
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

	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath)
	require.NoError(t, err)

	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Generate many randomized transfer transactions (worst case for compression)
	// Randomized fields prevent compression from being too efficient
	var txs []*types.Transaction
	var rlpTxs [][]byte
	var totalPlainTxSize int
	for i := 0; i < 5000; i++ {
		tx := makeRandomizedTransferTx(uint64(i))
		txs = append(txs, tx)

		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTxs = append(rlpTxs, buf.Bytes())
		totalPlainTxSize += buf.Len()
	}

	// Add transactions to TxCompressor until it says "full"
	var acceptedTxs []*types.Transaction
	var acceptedPlainSize int
	for i, rlpTx := range rlpTxs {
		ok, err := tc.Write(rlpTx, false)
		require.NoError(t, err)
		if !ok {
			t.Logf("TxCompressor full after %d transactions", i)
			break
		}
		acceptedTxs = append(acceptedTxs, txs[i])
		acceptedPlainSize += len(rlpTx)
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
	require.Equal(t, len(acceptedTxs), len(resp.Blocks[0].Txs), "should have same number of transactions")
}

// TestTxCompressorCompatibilityWithVariousTxTypes tests compatibility with different transaction types
func TestTxCompressorCompatibilityWithVariousTxTypes(t *testing.T) {
	const blobLimit = 128 * 1024
	const overhead = 100

	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath)
	require.NoError(t, err)

	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Create various transaction types
	var txs []*types.Transaction
	var rlpTxs [][]byte

	// Legacy transactions
	for i := 0; i < 10; i++ {
		tx := makeLegacyTx(100 + i*20)
		txs = append(txs, tx)
		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTxs = append(rlpTxs, buf.Bytes())
	}

	// EIP-1559 transactions
	for i := 0; i < 10; i++ {
		tx := makeEIP1559Tx(100 + i*20)
		txs = append(txs, tx)
		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTxs = append(rlpTxs, buf.Bytes())
	}

	// Add transactions to TxCompressor
	var acceptedTxs []*types.Transaction
	for i, rlpTx := range rlpTxs {
		ok, err := tc.Write(rlpTx, false)
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

	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath)
	require.NoError(t, err)

	bm, err := v1.NewBlobMaker(blobLimit, testDictPath)
	require.NoError(t, err)

	// Create transactions with random (high entropy) data
	var txs []*types.Transaction
	var rlpTxs [][]byte

	for i := 0; i < 50; i++ {
		tx := makeHighEntropyTx(500)
		txs = append(txs, tx)
		var buf bytes.Buffer
		rlp.Encode(&buf, tx)
		rlpTxs = append(rlpTxs, buf.Bytes())
	}

	// Add transactions to TxCompressor
	var acceptedTxs []*types.Transaction
	for i, rlpTx := range rlpTxs {
		ok, err := tc.Write(rlpTx, false)
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

	tc, err := v1.NewTxCompressor(blobLimit-overhead, testDictPath)
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
	var allRlpTxs [][]byte

	for _, blockRlp := range testBlocks {
		var block types.Block
		err := rlp.Decode(bytes.NewReader(blockRlp), &block)
		require.NoError(t, err)

		for _, tx := range block.Transactions() {
			allTxs = append(allTxs, tx)
			var buf bytes.Buffer
			rlp.Encode(&buf, tx)
			allRlpTxs = append(allRlpTxs, buf.Bytes())
		}
	}

	t.Logf("Loaded %d transactions from %d test blocks", len(allTxs), len(testBlocks))

	// Add transactions to TxCompressor
	var acceptedTxs []*types.Transaction
	for i, rlpTx := range allRlpTxs {
		ok, err := tc.Write(rlpTx, false)
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

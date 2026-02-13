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

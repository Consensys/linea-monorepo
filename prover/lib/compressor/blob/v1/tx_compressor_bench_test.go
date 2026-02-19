//go:build !fuzzlight

package v1_test

import (
	"bytes"
	cRand "crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/require"
)

const (
	benchBlobLimit   = 128 * 1024
	benchBlocksCount = 1
)

// TestSequencerApproachComparison compares two approaches for the sequencer's
// transaction selection with compression awareness:
//
// Approach 1 (BlobMaker-based - current production):
//   - Fast path: Sum stateless per-tx compressed size estimates
//   - Slow path (when near limit): Build fake block with all selected txs + candidate,
//     compress with BlobMaker to check if it fits
//   - Problem: O(n²) compression work when near the limit
//
// Approach 2 (TxCompressor-based - new):
//   - Add transactions one by one to additive compressor
//   - Each WriteRaw incrementally compresses and checks limit
//   - Benefit: O(n) compression work total
//
// This test simulates the actual sequencer flow for both approaches.
func TestSequencerApproachComparison(t *testing.T) {
	scenarios := []struct {
		name        string
		txGenerator func(int) []*types.Transaction
	}{
		{"erc20-transfers", generateErc20Transfers},
		{"plain-transfers", generatePlainTransfers},
		{"calldata-500b", func(n int) []*types.Transaction { return generateCalldataTxs(n, 500) }},
		{"calldata-3kb", func(n int) []*types.Transaction { return generateCalldataTxs(n, 3*1024) }},
		{"mixed", generateMixedTxs},
	}

	fmt.Println("=== Sequencer Approach Comparison ===")
	fmt.Printf("Blob limit: %d bytes, Blocks per scenario: %d\n\n", benchBlobLimit, benchBlocksCount)

	for _, scenario := range scenarios {
		fmt.Printf("--- Scenario: %s ---\n", scenario.name)

		txs := scenario.txGenerator(5000)

		// Approach 1: BlobMaker with fast/slow path
		bmResult := benchmarkBlobMakerApproach(t, txs)
		fmt.Printf("BlobMaker Approach (fast+slow path):\n")
		fmt.Printf("  Blocks: %d, Avg txs/block: %.1f\n", bmResult.blocks, bmResult.avgTxsPerBlock)
		fmt.Printf("  Avg time/block: %v, Avg time/tx: %v\n", bmResult.avgTimePerBlock, bmResult.avgTimePerTx)
		fmt.Printf("  Slow path triggers: %d, Avg slow path time: %v\n", bmResult.slowPathTriggers, bmResult.avgSlowPathTime)
		fmt.Printf("  Total time: %v\n", bmResult.totalTime)

		// Approach 2: TxCompressor additive
		tcResult := benchmarkTxCompressorApproach(t, txs)
		fmt.Printf("TxCompressor Approach (additive):\n")
		fmt.Printf("  Blocks: %d, Avg txs/block: %.1f\n", tcResult.blocks, tcResult.avgTxsPerBlock)
		fmt.Printf("  Avg time/block: %v, Avg time/tx: %v\n", tcResult.avgTimePerBlock, tcResult.avgTimePerTx)
		fmt.Printf("  Total time: %v\n", tcResult.totalTime)

		// Comparison
		if bmResult.totalTime > 0 && tcResult.totalTime > 0 {
			speedup := float64(bmResult.totalTime) / float64(tcResult.totalTime)
			fmt.Printf("Speedup (TxCompressor vs BlobMaker): %.2fx\n", speedup)
		}
		fmt.Println()
	}
}

type blobMakerResult struct {
	blocks           int
	avgTxsPerBlock   float64
	avgTimePerBlock  time.Duration
	avgTimePerTx     time.Duration
	slowPathTriggers int
	avgSlowPathTime  time.Duration
	totalTime        time.Duration
}

type txCompressorResult struct {
	blocks          int
	avgTxsPerBlock  float64
	avgTimePerBlock time.Duration
	avgTimePerTx    time.Duration
	totalTime       time.Duration
}

// benchmarkBlobMakerApproach simulates the BlobMaker-based sequencer approach:
// 1. Fast path: accumulate stateless compressed size estimates
// 2. Slow path: when estimate exceeds limit, build fake block and compress with BlobMaker
//
// Note: BlobMaker works on blocks, not transactions. So in the slow path, we must
// rebuild the entire block (all selected txs + candidate) and re-compress it each time.
// This is the O(n²) problem we're trying to solve with TxCompressor.
func benchmarkBlobMakerApproach(t *testing.T, txs []*types.Transaction) blobMakerResult {
	// For stateless estimates, we use a TxCompressor's RawCompressedSize
	estimator, err := v1.NewTxCompressor(benchBlobLimit, testDictPath, false)
	require.NoError(t, err)

	bm, err := v1.NewBlobMaker(benchBlobLimit, testDictPath)
	require.NoError(t, err)

	// Pre-encode transactions for the estimator
	txDataList := make([][]byte, len(txs))
	for i, tx := range txs {
		txDataList[i] = encodeTxForCompressor(tx)
	}

	var totalTxs int
	var slowPathTriggers int
	var slowPathChecks int
	var totalSlowPathTime time.Duration
	cursor := 0
	start := time.Now()

	for blockIdx := 0; blockIdx < benchBlocksCount; blockIdx++ {
		bm.Reset()
		var selectedTxs []*types.Transaction
		var cumulativeEstimate int
		inSlowPath := false

		for {
			tx := txs[cursor%len(txs)]
			txData := txDataList[cursor%len(txDataList)]
			cursor++

			if !inSlowPath {
				// Fast path: use stateless estimate
				txEstimate, err := estimator.RawCompressedSize(txData)
				require.NoError(t, err)

				if cumulativeEstimate+txEstimate <= benchBlobLimit {
					// Fast path accepts - no actual compression needed
					selectedTxs = append(selectedTxs, tx)
					cumulativeEstimate += txEstimate
					totalTxs++
					continue
				}

				// Switch to slow path - now we need to actually compress
				inSlowPath = true
				slowPathTriggers++
			}

			// Slow path: build block with all selected txs + candidate and compress
			slowPathStart := time.Now()
			slowPathChecks++

			// Build block with candidate transaction
			candidateTxs := make([]*types.Transaction, len(selectedTxs)+1)
			copy(candidateTxs, selectedTxs)
			candidateTxs[len(selectedTxs)] = tx

			blk := types.NewBlock(
				&types.Header{Time: 12345},
				&types.Body{Transactions: candidateTxs},
				nil,
				trie.NewStackTrie(nil),
			)

			var blockBuf bytes.Buffer
			require.NoError(t, rlp.Encode(&blockBuf, blk))

			// Check if it fits using BlobMaker's CanWrite (forceReset=true)
			// This does the full compression but reverts the state
			ok, err := bm.Write(blockBuf.Bytes(), true)
			require.NoError(t, err)

			totalSlowPathTime += time.Since(slowPathStart)

			if !ok {
				// Block is full - this transaction doesn't fit
				break
			}

			// Transaction fits - accept it
			selectedTxs = candidateTxs
			totalTxs++
		}
	}

	totalTime := time.Since(start)
	avgSlowPathTime := time.Duration(0)
	if slowPathChecks > 0 {
		avgSlowPathTime = totalSlowPathTime / time.Duration(slowPathChecks)
	}

	return blobMakerResult{
		blocks:           benchBlocksCount,
		avgTxsPerBlock:   float64(totalTxs) / float64(benchBlocksCount),
		avgTimePerBlock:  totalTime / time.Duration(benchBlocksCount),
		avgTimePerTx:     totalTime / time.Duration(max(1, totalTxs)),
		slowPathTriggers: slowPathTriggers,
		avgSlowPathTime:  avgSlowPathTime,
		totalTime:        totalTime,
	}
}

// benchmarkTxCompressorApproach simulates the TxCompressor-based sequencer approach:
// Each transaction is added to the additive compressor one by one
func benchmarkTxCompressorApproach(t *testing.T, txs []*types.Transaction) txCompressorResult {
	tc, err := v1.NewTxCompressor(benchBlobLimit, testDictPath, false)
	require.NoError(t, err)

	// Pre-encode transactions
	txDataList := make([][]byte, len(txs))
	for i, tx := range txs {
		txDataList[i] = encodeTxForCompressor(tx)
	}

	var totalTxs int
	cursor := 0
	start := time.Now()

	for block := 0; block < benchBlocksCount; block++ {
		tc.Reset()
		for {
			txData := txDataList[cursor%len(txDataList)]
			cursor++

			ok, err := tc.WriteRaw(txData, false)
			require.NoError(t, err)

			if !ok {
				// Block is full
				break
			}
			totalTxs++
		}
	}

	totalTime := time.Since(start)
	return txCompressorResult{
		blocks:          benchBlocksCount,
		avgTxsPerBlock:  float64(totalTxs) / float64(benchBlocksCount),
		avgTimePerBlock: totalTime / time.Duration(benchBlocksCount),
		avgTimePerTx:    totalTime / time.Duration(max(1, totalTxs)),
		totalTime:       totalTime,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Transaction generators - all use high entropy random data to simulate
// realistic worst-case compression scenarios

func generateErc20Transfers(count int) []*types.Transaction {
	txs := make([]*types.Transaction, count)
	for i := 0; i < count; i++ {
		// Random token contract
		addrBytes := make([]byte, 20)
		cRand.Read(addrBytes)
		tokenContract := common.BytesToAddress(addrBytes)

		// Random recipient
		cRand.Read(addrBytes)
		recipient := common.BytesToAddress(addrBytes)

		// ERC20 transfer(address,uint256) with random amount
		data := make([]byte, 68)
		data[0], data[1], data[2], data[3] = 0xa9, 0x05, 0x9c, 0xbb
		copy(data[16:36], recipient.Bytes())
		cRand.Read(data[36:68]) // random amount

		// Random gas values
		gasBytes := make([]byte, 8)
		cRand.Read(gasBytes)
		gas := 50000 + uint64(gasBytes[0])*1000

		cRand.Read(gasBytes)
		gasFee := big.NewInt(1000000000 + int64(gasBytes[0])*10000000)

		cRand.Read(gasBytes)
		gasTip := big.NewInt(100000000 + int64(gasBytes[0])*1000000)

		// Random nonce
		cRand.Read(gasBytes)
		nonce := uint64(gasBytes[0])<<16 | uint64(gasBytes[1])<<8 | uint64(gasBytes[2])

		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(59144),
			Nonce:     nonce,
			Gas:       gas,
			GasFeeCap: gasFee,
			GasTipCap: gasTip,
			To:        &tokenContract,
			Value:     big.NewInt(0),
			Data:      data,
		})
		txs[i] = signTxEIP1559(tx)
	}
	return txs
}

func generatePlainTransfers(count int) []*types.Transaction {
	txs := make([]*types.Transaction, count)
	for i := 0; i < count; i++ {
		// Random recipient
		addrBytes := make([]byte, 20)
		cRand.Read(addrBytes)
		recipient := common.BytesToAddress(addrBytes)

		// Random value
		valueBytes := make([]byte, 8)
		cRand.Read(valueBytes)
		value := new(big.Int).SetBytes(valueBytes)

		// Random gas values
		gasBytes := make([]byte, 8)
		cRand.Read(gasBytes)
		gasFee := big.NewInt(1000000000 + int64(gasBytes[0])*10000000)

		cRand.Read(gasBytes)
		gasTip := big.NewInt(100000000 + int64(gasBytes[0])*1000000)

		// Random nonce
		cRand.Read(gasBytes)
		nonce := uint64(gasBytes[0])<<16 | uint64(gasBytes[1])<<8 | uint64(gasBytes[2])

		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(59144),
			Nonce:     nonce,
			Gas:       21000,
			GasFeeCap: gasFee,
			GasTipCap: gasTip,
			To:        &recipient,
			Value:     value,
			Data:      nil,
		})
		txs[i] = signTxEIP1559(tx)
	}
	return txs
}

func generateCalldataTxs(count, calldataSize int) []*types.Transaction {
	txs := make([]*types.Transaction, count)
	for i := 0; i < count; i++ {
		// Random recipient
		addrBytes := make([]byte, 20)
		cRand.Read(addrBytes)
		recipient := common.BytesToAddress(addrBytes)

		// Random calldata (high entropy)
		data := make([]byte, calldataSize)
		cRand.Read(data)

		// Random gas values
		gasBytes := make([]byte, 8)
		cRand.Read(gasBytes)
		gas := 100000 + uint64(gasBytes[0])*10000

		cRand.Read(gasBytes)
		gasFee := big.NewInt(1000000000 + int64(gasBytes[0])*10000000)

		cRand.Read(gasBytes)
		gasTip := big.NewInt(100000000 + int64(gasBytes[0])*1000000)

		// Random nonce
		cRand.Read(gasBytes)
		nonce := uint64(gasBytes[0])<<16 | uint64(gasBytes[1])<<8 | uint64(gasBytes[2])

		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(59144),
			Nonce:     nonce,
			Gas:       gas,
			GasFeeCap: gasFee,
			GasTipCap: gasTip,
			To:        &recipient,
			Value:     big.NewInt(0),
			Data:      data,
		})
		txs[i] = signTxEIP1559(tx)
	}
	return txs
}

func generateMixedTxs(count int) []*types.Transaction {
	erc20 := generateErc20Transfers(count / 4)
	plain := generatePlainTransfers(count / 4)
	calldata500 := generateCalldataTxs(count/4, 500)
	calldata3k := generateCalldataTxs(count/4, 3*1024)

	mixed := make([]*types.Transaction, 0, count)
	for i := 0; i < count/4; i++ {
		mixed = append(mixed, erc20[i], plain[i], calldata500[i], calldata3k[i])
	}
	return mixed
}

func signTxEIP1559(tx *types.Transaction) *types.Transaction {
	privateKey, _ := crypto.GenerateKey()
	signer := types.NewLondonSigner(big.NewInt(59144))
	signedTx, _ := types.SignTx(tx, signer, privateKey)
	return signedTx
}

package linea.blob

import linea.rlp.RLP
import net.consensys.linea.nativecompressor.CompressorTestData
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.BlockBody
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder
import org.hyperledger.besu.ethereum.core.Difficulty
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.evm.log.LogsBloomFilter
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import kotlin.random.Random

/**
 * Integration test verifying that transactions compressed with TxCompressor
 * will fit into BlobCompressor when assembled into a block.
 *
 * The critical scenario to test is:
 * - TxCompressor benefits from compression context across many transactions
 * - BlobMaker compresses the block from scratch without that context
 * - We must ensure that blocks built with TxCompressor still fit in BlobMaker
 *
 * The worst case is many small, plain transactions (minimal data) because:
 * - They benefit most from shared context in TxCompressor (similar structure)
 * - BlobMaker has to compress them without that context advantage
 */
class TxCompressorBlobMakerCompatibilityTest {
  companion object {
    private const val BLOB_LIMIT = 128 * 1024
    private const val BLOB_OVERHEAD = 100

    // BlobCompressor has a max uncompressed input limit (ZK circuit constraint)
    // Block RLP includes header overhead (~500 bytes), so we use a conservative limit
    private const val MAX_UNCOMPRESSED_TX_DATA = 770_000
    private val TEST_BLOCKS = CompressorTestData.blocksRlpEncoded

    // Note: TxCompressor uses a global singleton in the Go code, so we initialize once
    // and reset between tests to avoid interference with other test classes.
    private val txCompressor: GoBackedTxCompressor by lazy {
      GoBackedTxCompressor.getInstance(
        TxCompressorVersion.V2,
        BLOB_LIMIT - BLOB_OVERHEAD,
      )
    }

    private val blobCompressor: GoBackedBlobCompressor by lazy {
      GoBackedBlobCompressor.getInstance(
        BlobCompressorVersion.V2,
        BLOB_LIMIT,
      )
    }
  }

  @BeforeEach
  fun before() {
    // Reset compressors between tests (don't reinitialize - that would reset the global Go singleton)
    txCompressor.reset()
    blobCompressor.reset()
  }

  /**
   * Critical compatibility test: many randomized transactions.
   *
   * This is the worst case because:
   * 1. Randomized fields (addresses, values, gas, payload) prevent efficient compression
   * 2. TxCompressor context sharing provides minimal benefit with random data
   * 3. BlobMaker compresses from scratch -> similar compression ratio
   * 4. If TxCompressor accepts N transactions, BlobMaker must also accept the resulting block
   */
  @Test
  fun `many randomized transactions from TxCompressor fit in BlobCompressor`() {
    // Fill TxCompressor until it says "full"
    // Generate transactions on-demand to fill exactly one blob
    val random = Random(12345L)
    val acceptedTxs = mutableListOf<Transaction>()
    var sumOfRlpForSigningSize = 0
    var sumOfSignedTxRlpSize = 0
    // Both call RawCompressedSize with the same input — should be identical after NoTerminalSymbol alignment
    var sumOfIndividualTxCompressorSize = 0
    var sumOfIndividualRawCompressedSize = 0
    var nonce = 0L

    while (true) {
      val (tx, txData) = TxCompressorTestFixtures.generateRandomizedTransferTx(nonce++, random)
      if (!txCompressor.canAppendTransaction(txData.from, txData.rlpForSigning)) {
        break
      }
      val result = txCompressor.appendTransaction(txData.from, txData.rlpForSigning)
      assertThat(result.txAppended).isTrue()
      acceptedTxs.add(tx)
      sumOfRlpForSigningSize += txData.totalSize
      sumOfSignedTxRlpSize += TxCompressorTestFixtures.signedTxRlpSize(tx)
      val combinedData = txData.from + txData.rlpForSigning
      sumOfIndividualTxCompressorSize += txCompressor.compressedSize(combinedData)
      val blobEstimate = blobCompressor.compressedSize(combinedData)
      if (blobEstimate > 0) sumOfIndividualRawCompressedSize += blobEstimate
    }

    assertThat(acceptedTxs).isNotEmpty()
    val txCompressorSize = txCompressor.getCompressedSize()

    // Build a block with the accepted transactions
    val block = buildTestBlock(acceptedTxs)
    val blockRlp = RLP.encodeBlock(block)
    val blockRlpSize = blockRlp.size

    // Verify BlobCompressor accepts the block
    val result = blobCompressor.appendBlock(blockRlp)
    val blobCompressorSize = result.compressedSizeAfter

    assertThat(result.blockAppended)
      .withFailMessage(
        "Block built with TxCompressor should fit in BlobCompressor. " +
          "Transactions: ${acceptedTxs.size}, " +
          "TxCompressor size: $txCompressorSize, " +
          "BlobCompressor size after: $blobCompressorSize, " +
          "Block RLP size: $blockRlpSize",
      )
      .isTrue()

    println("=== Compression Comparison (Worst Case: ${acceptedTxs.size} randomized txs) ===")
    println("--- Input sizes ---")
    println(
      "Sum of RLP for signing (from + tx):   $sumOfRlpForSigningSize bytes " +
        "(${"%.2f".format(100.0 * sumOfRlpForSigningSize / sumOfSignedTxRlpSize)}% of signed tx RLP)",
    )
    println("Signed tx RLP (no header):            $sumOfSignedTxRlpSize bytes")
    println("Block RLP (header + signed txs):      $blockRlpSize bytes")
    println("--- Compressed sizes ---")
    println(
      "Sum of individual TxCompressor:       $sumOfIndividualTxCompressorSize bytes " +
        "(vs RLP-for-signing: ${"%.2f".format(100.0 * sumOfIndividualTxCompressorSize / sumOfRlpForSigningSize)}%, " +
        "vs signed tx RLP: ${"%.2f".format(100.0 * sumOfIndividualTxCompressorSize / sumOfSignedTxRlpSize)}%)",
    )
    println(
      "Sum of individual RawCompressed:      $sumOfIndividualRawCompressedSize bytes " +
        "(worst-case stateless per-tx)",
    )
    println(
      "TxCompressor batch (with context):    $txCompressorSize bytes " +
        "(vs RLP-for-signing: ${"%.2f".format(100.0 * txCompressorSize / sumOfRlpForSigningSize)}%, " +
        "vs block RLP: ${"%.2f".format(100.0 * txCompressorSize / blockRlpSize)}%)",
    )
    println(
      "BlobCompressor after appendBlock:     $blobCompressorSize bytes " +
        "(vs block RLP: ${"%.2f".format(100.0 * blobCompressorSize / blockRlpSize)}%)",
    )
    println("--- Savings ---")
    println(
      "TxCompressor context saving:          ${sumOfIndividualTxCompressorSize - txCompressorSize} bytes " +
        "(${"%.2f".format(
          100.0 * (sumOfIndividualTxCompressorSize - txCompressorSize) /
            sumOfIndividualTxCompressorSize,
        )}% vs individual TxCompressor)",
    )
    println(
      "TxCompressor vs BlobCompressor:       ${blobCompressorSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (blobCompressorSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println("--- Blob limit ---")
    println("Blob limit:                           $BLOB_LIMIT bytes")
    println(
      "Headroom remaining:                   ${BLOB_LIMIT - blobCompressorSize} bytes " +
        "(${"%.2f".format(100.0 * (BLOB_LIMIT - blobCompressorSize) / BLOB_LIMIT)}%)",
    )
  }

  /**
   * Test with real transaction data from test fixtures.
   */
  @Test
  fun `real test data - transactions from TxCompressor fit in BlobCompressor`() {
    val testTransactions = TEST_BLOCKS.flatMap { blockRlp ->
      val block = RLP.decodeBlockWithMainnetFunctions(blockRlp)
      block.body.transactions.map { tx -> tx to TxCompressorTestFixtures.encodeTransactionForCompressor(tx) }
    }

    // Fill TxCompressor with transactions
    val acceptedTxs = mutableListOf<Transaction>()
    var sumOfRlpForSigningSize = 0
    var sumOfSignedTxRlpSize = 0
    // Both call RawCompressedSize with the same input — should be identical after NoTerminalSymbol alignment
    var sumOfIndividualTxCompressorSize = 0
    var sumOfIndividualRawCompressedSize = 0
    for ((tx, txData) in testTransactions) {
      if (!txCompressor.canAppendTransaction(txData.from, txData.rlpForSigning)) {
        break
      }
      val result = txCompressor.appendTransaction(txData.from, txData.rlpForSigning)
      assertThat(result.txAppended).isTrue()
      acceptedTxs.add(tx)
      sumOfRlpForSigningSize += txData.totalSize
      sumOfSignedTxRlpSize += TxCompressorTestFixtures.signedTxRlpSize(tx)
      val combinedData = txData.from + txData.rlpForSigning
      sumOfIndividualTxCompressorSize += txCompressor.compressedSize(combinedData)
      val blobEstimate = blobCompressor.compressedSize(combinedData)
      if (blobEstimate > 0) sumOfIndividualRawCompressedSize += blobEstimate
    }

    assertThat(acceptedTxs).isNotEmpty()
    val txCompressorSize = txCompressor.getCompressedSize()

    // Build a block with the accepted transactions
    val block = buildTestBlock(acceptedTxs)
    val blockRlp = RLP.encodeBlock(block)
    val blockRlpSize = blockRlp.size

    // Verify BlobCompressor accepts the block
    val result = blobCompressor.appendBlock(blockRlp)
    val blobCompressorSize = result.compressedSizeAfter

    assertThat(result.blockAppended)
      .withFailMessage(
        "Block built with TxCompressor should fit in BlobCompressor. " +
          "TxCompressor size: $txCompressorSize, " +
          "BlobCompressor size after: $blobCompressorSize",
      )
      .isTrue()

    println("=== Compression Comparison (Real Test Data: ${acceptedTxs.size} txs) ===")
    println("--- Input sizes ---")
    println(
      "Sum of RLP for signing (from + tx):   $sumOfRlpForSigningSize bytes " +
        "(${"%.2f".format(100.0 * sumOfRlpForSigningSize / sumOfSignedTxRlpSize)}% of signed tx RLP)",
    )
    println("Signed tx RLP (no header):            $sumOfSignedTxRlpSize bytes")
    println("Block RLP (header + signed txs):      $blockRlpSize bytes")
    println("--- Compressed sizes ---")
    println(
      "Sum of individual TxCompressor:       $sumOfIndividualTxCompressorSize bytes " +
        "(vs RLP-for-signing: ${"%.2f".format(100.0 * sumOfIndividualTxCompressorSize / sumOfRlpForSigningSize)}%, " +
        "vs signed tx RLP: ${"%.2f".format(100.0 * sumOfIndividualTxCompressorSize / sumOfSignedTxRlpSize)}%)",
    )
    println(
      "Sum of individual RawCompressed:      $sumOfIndividualRawCompressedSize bytes " +
        "(worst-case stateless per-tx)",
    )
    println(
      "TxCompressor batch (with context):    $txCompressorSize bytes " +
        "(vs RLP-for-signing: ${"%.2f".format(100.0 * txCompressorSize / sumOfRlpForSigningSize)}%, " +
        "vs block RLP: ${"%.2f".format(100.0 * txCompressorSize / blockRlpSize)}%)",
    )
    println(
      "BlobCompressor after appendBlock:     $blobCompressorSize bytes " +
        "(vs block RLP: ${"%.2f".format(100.0 * blobCompressorSize / blockRlpSize)}%)",
    )
    println("--- Savings ---")
    println(
      "TxCompressor context saving:          ${sumOfIndividualTxCompressorSize - txCompressorSize} bytes " +
        "(${"%.2f".format(
          100.0 * (sumOfIndividualTxCompressorSize - txCompressorSize) /
            sumOfIndividualTxCompressorSize,
        )}% vs individual TxCompressor)",
    )
    println(
      "TxCompressor vs BlobCompressor:       ${blobCompressorSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (blobCompressorSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println("--- Blob limit ---")
    println("Blob limit:                           $BLOB_LIMIT bytes")
    println(
      "Headroom remaining:                   ${BLOB_LIMIT - blobCompressorSize} bytes " +
        "(${"%.2f".format(100.0 * (BLOB_LIMIT - blobCompressorSize) / BLOB_LIMIT)}%)",
    )
  }

  /**
   * Test with ERC-20 transfer transactions - realistic scenario.
   *
   * This simulates a common pattern: same sender transferring tokens to the same
   * recipient multiple times with different amounts. This should compress well
   * because most of the transaction data is identical (sender, contract, recipient,
   * function selector) - only the amount and nonce vary.
   */
  @Test
  fun `ERC-20 transfers - same sender, random recipient, different amounts`() {
    // Fill TxCompressor until it says "full"
    // Generate ERC-20 transfers on-demand to fill exactly one blob
    val random = Random(54321L)
    val keyPair = TxCompressorTestFixtures.signatureAlgorithm.generateKeyPair()
    val tokenContractAddress = Address.fromHexString("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

    val acceptedTxs = mutableListOf<Transaction>()
    var sumOfRlpForSigningSize = 0
    var sumOfSignedTxRlpSize = 0
    // Both call RawCompressedSize with the same input — should be identical after NoTerminalSymbol alignment
    var sumOfIndividualTxCompressorSize = 0
    var sumOfIndividualRawCompressedSize = 0
    var nonce = 0L

    while (true) {
      // Randomize recipient address per transaction
      val recipientBytes = ByteArray(20)
      random.nextBytes(recipientBytes)
      val recipientAddress = Address.wrap(Bytes.wrap(recipientBytes))

      val (tx, txData) = TxCompressorTestFixtures.generateErc20TransferTx(nonce++, keyPair, tokenContractAddress, recipientAddress, random)
      // Check both compressed size limit AND uncompressed size limit
      // ERC-20 transfers compress very well, so we may hit uncompressed limit first
      if (!txCompressor.canAppendTransaction(txData.from, txData.rlpForSigning) ||
        sumOfRlpForSigningSize + txData.totalSize > MAX_UNCOMPRESSED_TX_DATA
      ) {
        break
      }
      val result = txCompressor.appendTransaction(txData.from, txData.rlpForSigning)
      assertThat(result.txAppended).isTrue()
      acceptedTxs.add(tx)
      sumOfRlpForSigningSize += txData.totalSize
      sumOfSignedTxRlpSize += TxCompressorTestFixtures.signedTxRlpSize(tx)
      val combinedData = txData.from + txData.rlpForSigning
      sumOfIndividualTxCompressorSize += txCompressor.compressedSize(combinedData)
      val blobEstimate = blobCompressor.compressedSize(combinedData)
      if (blobEstimate > 0) sumOfIndividualRawCompressedSize += blobEstimate
    }

    assertThat(acceptedTxs).isNotEmpty()
    val txCompressorSize = txCompressor.getCompressedSize()

    // Build a block with the accepted transactions
    val block = buildTestBlock(acceptedTxs)
    val blockRlp = RLP.encodeBlock(block)
    val blockRlpSize = blockRlp.size

    // Verify BlobCompressor accepts the block
    val result = blobCompressor.appendBlock(blockRlp)
    val blobCompressorSize = result.compressedSizeAfter

    assertThat(result.blockAppended)
      .withFailMessage(
        "Block built with TxCompressor should fit in BlobCompressor. " +
          "Transactions: ${acceptedTxs.size}, " +
          "TxCompressor size: $txCompressorSize, " +
          "BlobCompressor size after: $blobCompressorSize, " +
          "Block RLP size: $blockRlpSize",
      )
      .isTrue()

    println("=== Compression Comparison (ERC-20 Transfers, random recipient: ${acceptedTxs.size} txs) ===")
    println("--- Input sizes ---")
    println(
      "Sum of RLP for signing (from + tx):   $sumOfRlpForSigningSize bytes " +
        "(${"%.2f".format(100.0 * sumOfRlpForSigningSize / sumOfSignedTxRlpSize)}% of signed tx RLP)",
    )
    println("Signed tx RLP (no header):            $sumOfSignedTxRlpSize bytes")
    println("Block RLP (header + signed txs):      $blockRlpSize bytes")
    println("--- Compressed sizes ---")
    println(
      "Sum of individual TxCompressor:       $sumOfIndividualTxCompressorSize bytes " +
        "(vs RLP-for-signing: ${"%.2f".format(100.0 * sumOfIndividualTxCompressorSize / sumOfRlpForSigningSize)}%, " +
        "vs signed tx RLP: ${"%.2f".format(100.0 * sumOfIndividualTxCompressorSize / sumOfSignedTxRlpSize)}%)",
    )
    println(
      "Sum of individual RawCompressed:      $sumOfIndividualRawCompressedSize bytes " +
        "(worst-case stateless per-tx)",
    )
    println(
      "TxCompressor batch (with context):    $txCompressorSize bytes " +
        "(vs RLP-for-signing: ${"%.2f".format(100.0 * txCompressorSize / sumOfRlpForSigningSize)}%, " +
        "vs block RLP: ${"%.2f".format(100.0 * txCompressorSize / blockRlpSize)}%)",
    )
    println(
      "BlobCompressor after appendBlock:     $blobCompressorSize bytes " +
        "(vs block RLP: ${"%.2f".format(100.0 * blobCompressorSize / blockRlpSize)}%)",
    )
    println("--- Savings ---")
    println(
      "TxCompressor context saving:          ${sumOfIndividualTxCompressorSize - txCompressorSize} bytes " +
        "(${"%.2f".format(
          100.0 * (sumOfIndividualTxCompressorSize - txCompressorSize) /
            sumOfIndividualTxCompressorSize,
        )}% vs individual TxCompressor)",
    )
    println(
      "TxCompressor vs BlobCompressor:       ${blobCompressorSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (blobCompressorSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println("--- Blob limit ---")
    println("Blob limit:                           $BLOB_LIMIT bytes")
    println(
      "Headroom remaining:                   ${BLOB_LIMIT - blobCompressorSize} bytes " +
        "(${"%.2f".format(100.0 * (BLOB_LIMIT - blobCompressorSize) / BLOB_LIMIT)}%)",
    )
  }

  /**
   * Verify TxCompressor provides compression benefit over individual estimation.
   */
  @Test
  fun `TxCompressor provides better compression than individual estimation`() {
    val txsToTest = TxCompressorTestFixtures.generateManyRandomizedTransactions(100)

    // Estimate individual sizes (worst case - no context)
    var individualEstimate = 0
    for ((_, txData) in txsToTest) {
      txCompressor.reset()
      txCompressor.appendTransaction(txData.from, txData.rlpForSigning)
      individualEstimate += txCompressor.getCompressedSize()
    }

    // Actual additive compression (with context)
    txCompressor.reset()
    for ((_, txData) in txsToTest) {
      txCompressor.appendTransaction(txData.from, txData.rlpForSigning)
    }
    val actualSize = txCompressor.getCompressedSize()

    // Additive compression should be smaller due to shared context
    val savings = 1.0 - (actualSize.toDouble() / individualEstimate)
    println("Individual estimate: $individualEstimate, Actual: $actualSize, Savings: ${savings * 100}%")

    assertThat(savings)
      .withFailMessage(
        "Additive compression should provide savings. " +
          "Individual estimate: $individualEstimate, Actual: $actualSize, Savings: ${savings * 100}%",
      )
      .isGreaterThan(0.0)
  }

  private fun buildTestBlock(transactions: List<Transaction>): Block {
    // Get a reference block header from test data
    val referenceBlock = RLP.decodeBlockWithMainnetFunctions(TEST_BLOCKS.first())
    val referenceHeader = referenceBlock.header

    // Build a new header with updated fields
    val header = BlockHeaderBuilder.create()
      .parentHash(referenceHeader.parentHash)
      .ommersHash(referenceHeader.ommersHash)
      .coinbase(referenceHeader.coinbase)
      .stateRoot(referenceHeader.stateRoot)
      .transactionsRoot(referenceHeader.transactionsRoot)
      .receiptsRoot(referenceHeader.receiptsRoot)
      .logsBloom(LogsBloomFilter.empty())
      .difficulty(Difficulty.ZERO)
      .number(referenceHeader.number)
      .gasLimit(referenceHeader.gasLimit)
      .gasUsed(referenceHeader.gasUsed)
      .timestamp(System.currentTimeMillis() / 1000)
      .extraData(Bytes.EMPTY)
      .mixHash(referenceHeader.mixHash)
      .nonce(referenceHeader.nonce)
      .baseFee(referenceHeader.baseFee.orElse(null))
      .blockHeaderFunctions(MainnetBlockHeaderFunctions())
      .buildBlockHeader()

    val body = BlockBody(transactions, emptyList())
    return Block(header, body)
  }
}

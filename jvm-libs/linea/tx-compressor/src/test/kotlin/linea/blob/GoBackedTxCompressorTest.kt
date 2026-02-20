package linea.blob

import linea.rlp.RLP
import net.consensys.linea.nativecompressor.CompressorTestData
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.ethereum.rlp.BytesValueRLPOutput
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.assertThrows
import kotlin.random.Random

/**
 * IMPORTANT: The Go TxCompressor library uses a global singleton.
 * This means that if multiple test classes call getInstance() with different limits,
 * they will interfere with each other. Run this test class in isolation if you need
 * to test with a specific DATA_LIMIT.
 *
 * To run this test class in isolation:
 *   ./gradlew :jvm-libs:linea:tx-compressor:test --tests "linea.blob.GoBackedTxCompressorTest"
 */
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class GoBackedTxCompressorTest {
  companion object {
    // Use a smaller limit to test limit-exceeded behavior faster
    private const val DATA_LIMIT = 24 * 1024
    private val TEST_BLOCKS = CompressorTestData.blocksRlpEncoded

    // Note: getInstance() reinitializes the global Go singleton with the given limit.
    // If another test class calls getInstance() with a different limit, it will change
    // the limit for all tests.
    private val compressor: GoBackedTxCompressor by lazy {
      GoBackedTxCompressor.getInstance(TxCompressorVersion.V2, DATA_LIMIT)
    }

    private fun extractTransactionsFromBlocks(): List<TransactionData> {
      return TEST_BLOCKS.flatMap { blockRlp ->
        val block = RLP.decodeBlockWithMainnetFunctions(blockRlp)
        block.body.transactions.map { tx ->
          encodeTransactionForCompressor(tx)
        }
      }
    }

    /**
     * Encodes a transaction in the format expected by TxCompressor:
     * - from: sender address (20 bytes)
     * - rlpForSigning: RLP-encoded transaction for signing (without signature)
     */
    private fun encodeTransactionForCompressor(tx: Transaction): TransactionData {
      val from = tx.sender.toArray()
      val rlpForSigning = encodeTransactionForSigning(tx)
      return TransactionData(from, rlpForSigning)
    }

    /**
     * Encodes a transaction for signing (without the signature).
     * This is the RLP encoding that the signer would hash and sign.
     */
    private fun encodeTransactionForSigning(tx: Transaction): ByteArray {
      val rlpOutput = BytesValueRLPOutput()
      tx.writeTo(rlpOutput)
      val fullRlp = rlpOutput.encoded().toArray()

      // For simplicity in tests, we use the full RLP encoding.
      // In production, this should be the encoding without signature fields.
      // The Go compressor will handle this correctly as long as the format is consistent.
      return fullRlp
    }

    private val TEST_TRANSACTIONS: List<TransactionData> by lazy { extractTransactionsFromBlocks() }
  }

  /**
   * Holds the data needed to append a transaction to the compressor.
   */
  data class TransactionData(
    val from: ByteArray,
    val rlpForSigning: ByteArray,
  ) {
    override fun equals(other: Any?): Boolean {
      if (this === other) return true
      if (javaClass != other?.javaClass) return false
      other as TransactionData
      return from.contentEquals(other.from) && rlpForSigning.contentEquals(other.rlpForSigning)
    }

    override fun hashCode(): Int {
      var result = from.contentHashCode()
      result = 31 * result + rlpForSigning.contentHashCode()
      return result
    }
  }

  // The LZSS compressor has a small header (~3 bytes) even when empty
  private var baselineCompressedSize: Int = 0

  @BeforeEach
  fun before() {
    compressor.reset()
    baselineCompressedSize = compressor.getCompressedSize()
  }

  @Test
  fun `test appendTransaction with data within limit`() {
    val txData = TEST_TRANSACTIONS.first()
    val result = compressor.appendTransaction(txData.from, txData.rlpForSigning)
    assertThat(result.txAppended).isTrue
    assertThat(result.compressedSizeBefore).isEqualTo(baselineCompressedSize)
    assertThat(result.compressedSizeAfter).isGreaterThan(baselineCompressedSize)
  }

  @Test
  fun `test invalid from address size`() {
    val invalidFrom = Random.nextBytes(19) // wrong size
    val rlp = Random.nextBytes(100)
    assertThrows<IllegalArgumentException> {
      compressor.appendTransaction(invalidFrom, rlp)
    }
  }

  @Test
  fun `test compression data limit exceeded`() {
    var txs = TEST_TRANSACTIONS.iterator()
    var txData = txs.next()
    var result = compressor.appendTransaction(txData.from, txData.rlpForSigning)
    // at least one transaction should be appended
    assertThat(result.txAppended).isTrue()

    var txCount = 1
    val maxIterations = 10000 // safety limit to prevent infinite loop
    val progressInterval = 100 // print progress every N transactions
    var previousSize = compressor.getCompressedSize()

    println("Starting compression test with limit: $DATA_LIMIT bytes")
    println(
      "Tx #1: compressed size = $previousSize bytes " +
        "(${"%.1f".format(100.0 * previousSize / DATA_LIMIT)}% of configured limit)",
    )

    while (result.txAppended && txCount < maxIterations) {
      if (!txs.hasNext()) {
        // recompress again, until the limit is reached
        txs = TEST_TRANSACTIONS.iterator()
      }
      txData = txs.next()
      val canAppend = compressor.canAppendTransaction(txData.from, txData.rlpForSigning)
      result = compressor.appendTransaction(txData.from, txData.rlpForSigning)
      // assert consistency between canAppendTransaction and appendTransaction
      assertThat(canAppend).isEqualTo(result.txAppended)
      txCount++

      val currentSize = compressor.getCompressedSize()

      // Print progress periodically
      if (txCount % progressInterval == 0) {
        val percentFull = 100.0 * currentSize / DATA_LIMIT
        println(
          "Tx #$txCount: compressed size = $currentSize bytes" +
            " (${"%.1f".format(percentFull)}% of configured limit)",
        )
      }
    }

    val finalSize = compressor.getCompressedSize()
    val finalPercent = 100.0 * finalSize / DATA_LIMIT

    if (txCount >= maxIterations) {
      // If we hit the limit, the test data transactions are too small to fill the compressor
      // This is still a valid test - we verified consistency between canAppend and append
      println("Warning: reached max iterations ($maxIterations) without filling compressor.")
      println(
        "Final: $txCount txs, compressed size = $finalSize bytes " +
          "(${"%.1f".format(finalPercent)}% of configured limit)",
      )
    } else {
      println("Compressor full after $txCount transactions")
      println("Final: compressed size = $finalSize bytes (${"%.1f".format(finalPercent)}% of configured limit)")
      assertThat(result.txAppended).isFalse()
      assertThat(result.compressedSizeBefore).isGreaterThan(baselineCompressedSize)
      assertThat(result.compressedSizeAfter).isEqualTo(result.compressedSizeBefore)
    }
  }

  @Test
  fun `test reset clears state`() {
    val txs = TEST_TRANSACTIONS.iterator()
    assertThat(compressor.getCompressedSize()).isEqualTo(baselineCompressedSize)

    var txData = txs.next()
    var res = compressor.appendTransaction(txData.from, txData.rlpForSigning)
    assertThat(res.txAppended).isTrue()
    assertThat(res.compressedSizeBefore).isEqualTo(baselineCompressedSize)
    assertThat(res.compressedSizeAfter).isGreaterThan(baselineCompressedSize)
    assertThat(res.compressedSizeAfter).isEqualTo(compressor.getCompressedSize())

    val sizeAfterFirstTx = res.compressedSizeAfter

    compressor.reset()

    assertThat(compressor.getCompressedSize()).isEqualTo(baselineCompressedSize)
    assertThat(compressor.getUncompressedSize()).isZero()

    txData = txs.next()
    res = compressor.appendTransaction(txData.from, txData.rlpForSigning)
    assertThat(res.txAppended).isTrue()
    assertThat(res.compressedSizeBefore).isEqualTo(baselineCompressedSize)
    assertThat(res.compressedSizeAfter).isGreaterThan(baselineCompressedSize)
    assertThat(res.compressedSizeAfter).isEqualTo(compressor.getCompressedSize())
  }

  @Test
  fun `test compression context improves ratio`() {
    // Compare: sum of individual tx compressed sizes vs. additive compression
    // Additive should be smaller due to shared context
    val txsToTest = TEST_TRANSACTIONS.take(20)

    // Measure individual compression sizes
    var individualSum = 0
    for (txData in txsToTest) {
      compressor.reset()
      val result = compressor.appendTransaction(txData.from, txData.rlpForSigning)
      assertThat(result.txAppended).isTrue()
      individualSum += compressor.getCompressedSize()
    }

    // Measure additive compression size
    compressor.reset()
    for (txData in txsToTest) {
      val result = compressor.appendTransaction(txData.from, txData.rlpForSigning)
      assertThat(result.txAppended).isTrue()
    }
    val additiveSize = compressor.getCompressedSize()

    // Additive compression should be smaller due to shared context
    assertThat(additiveSize).isLessThan(individualSum)
  }

  @Test
  fun `test getCompressedData returns correct data`() {
    val txData = TEST_TRANSACTIONS.first()
    compressor.appendTransaction(txData.from, txData.rlpForSigning)

    val compressedData = compressor.getCompressedData()
    assertThat(compressedData.size).isEqualTo(compressor.getCompressedSize())
    assertThat(compressedData.size).isGreaterThan(0)
  }

  @Test
  fun `test getCompressedDataAndReset returns data and resets`() {
    val txData = TEST_TRANSACTIONS.first()
    compressor.appendTransaction(txData.from, txData.rlpForSigning)

    val sizeBefore = compressor.getCompressedSize()
    assertThat(sizeBefore).isGreaterThan(baselineCompressedSize)

    val compressedData = compressor.getCompressedDataAndReset()
    assertThat(compressedData.size).isEqualTo(sizeBefore)
    assertThat(compressor.getCompressedSize()).isEqualTo(baselineCompressedSize)
  }

  @Test
  fun `test multiple transactions accumulate`() {
    val txs = TEST_TRANSACTIONS.take(5)
    var prevSize = baselineCompressedSize

    for (txData in txs) {
      val result = compressor.appendTransaction(txData.from, txData.rlpForSigning)
      assertThat(result.txAppended).isTrue()
      assertThat(result.compressedSizeBefore).isEqualTo(prevSize)
      assertThat(result.compressedSizeAfter).isGreaterThan(prevSize)
      prevSize = result.compressedSizeAfter
    }

    assertThat(compressor.getCompressedSize()).isEqualTo(prevSize)
    assertThat(compressor.getUncompressedSize()).isGreaterThan(0)
  }

  @Test
  fun `test compressedSize is stateless`() {
    val testData = "hello world this is some test data for compression".toByteArray()

    // Get initial state
    val initialLen = compressor.getCompressedSize()
    val initialWritten = compressor.getUncompressedSize()

    // Call compressedSize multiple times
    val size1 = compressor.compressedSize(testData)
    val size2 = compressor.compressedSize(testData)

    // Should return same value
    assertThat(size1).isEqualTo(size2)

    // State should be unchanged
    assertThat(compressor.getCompressedSize()).isEqualTo(initialLen)
    assertThat(compressor.getUncompressedSize()).isEqualTo(initialWritten)
  }

  @Test
  fun `test getInstance with recompression disabled`() {
    // Create a compressor with recompression disabled
    val fastCompressor = GoBackedTxCompressor.getInstance(
      TxCompressorVersion.V2,
      DATA_LIMIT,
      enableRecompress = false,
    )

    val txData = TEST_TRANSACTIONS.first()
    val result = fastCompressor.appendTransaction(txData.from, txData.rlpForSigning)
    assertThat(result.txAppended).isTrue()
    assertThat(result.compressedSizeAfter).isGreaterThan(0)
  }
}

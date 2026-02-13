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

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class GoBackedTxCompressorTest {
  companion object {
    private const val DATA_LIMIT = 24 * 1024
    private val TEST_BLOCKS = CompressorTestData.blocksRlpEncoded
    private val compressor = GoBackedTxCompressor.getInstance(TxCompressorVersion.V1, DATA_LIMIT)

    private fun extractTransactionsFromBlocks(): List<ByteArray> {
      return TEST_BLOCKS.flatMap { blockRlp ->
        val block = RLP.decodeBlockWithMainnetFunctions(blockRlp)
        block.body.transactions.map { tx ->
          encodeTransaction(tx)
        }
      }
    }

    private fun encodeTransaction(tx: Transaction): ByteArray {
      val rlpOutput = BytesValueRLPOutput()
      tx.writeTo(rlpOutput)
      return rlpOutput.encoded().toArray()
    }

    private val TEST_TRANSACTIONS: List<ByteArray> by lazy { extractTransactionsFromBlocks() }
  }

  @BeforeEach
  fun before() {
    compressor.reset()
  }

  @Test
  fun `test appendTransaction with data within limit`() {
    val tx = TEST_TRANSACTIONS.first()
    val result = compressor.appendTransaction(tx)
    assertThat(result.txAppended).isTrue
    assertThat(result.compressedSizeBefore).isZero()
    assertThat(result.compressedSizeAfter).isGreaterThan(0)
  }

  @Test
  fun `test invalid rlp transaction`() {
    val invalidTx = Random.nextBytes(100)
    assertThrows<TxCompressionException> {
      compressor.appendTransaction(invalidTx)
    }
  }

  @Test
  fun `test compression data limit exceeded`() {
    var txs = TEST_TRANSACTIONS.iterator()
    var result = compressor.appendTransaction(txs.next())
    // at least one transaction should be appended
    assertThat(result.txAppended).isTrue()
    while (result.txAppended) {
      if (!txs.hasNext()) {
        // recompress again, until the limit is reached
        txs = TEST_TRANSACTIONS.iterator()
      }
      val txRlp = txs.next()
      val canAppend = compressor.canAppendTransaction(txRlp)
      result = compressor.appendTransaction(txRlp)
      // assert consistency between canAppendTransaction and appendTransaction
      assertThat(canAppend).isEqualTo(result.txAppended)
    }
    assertThat(result.txAppended).isFalse()
    assertThat(result.compressedSizeBefore).isGreaterThan(0)
    assertThat(result.compressedSizeAfter).isEqualTo(result.compressedSizeBefore)
  }

  @Test
  fun `test reset clears state`() {
    val txs = TEST_TRANSACTIONS.iterator()
    assertThat(compressor.getCompressedSize()).isZero()

    var res = compressor.appendTransaction(txs.next())
    assertThat(res.txAppended).isTrue()
    assertThat(res.compressedSizeBefore).isZero()
    assertThat(res.compressedSizeAfter).isGreaterThan(0)
    assertThat(res.compressedSizeAfter).isEqualTo(compressor.getCompressedSize())

    compressor.reset()

    assertThat(compressor.getCompressedSize()).isZero()
    assertThat(compressor.getUncompressedSize()).isZero()

    res = compressor.appendTransaction(txs.next())
    assertThat(res.txAppended).isTrue()
    assertThat(res.compressedSizeBefore).isZero()
    assertThat(res.compressedSizeAfter).isGreaterThan(0)
    assertThat(res.compressedSizeAfter).isEqualTo(compressor.getCompressedSize())
  }

  @Test
  fun `test compression context improves ratio`() {
    // Compare: sum of individual tx compressed sizes vs. additive compression
    // Additive should be smaller due to shared context
    val txsToTest = TEST_TRANSACTIONS.take(20)

    // Measure individual compression sizes
    var individualSum = 0
    for (tx in txsToTest) {
      compressor.reset()
      val result = compressor.appendTransaction(tx)
      assertThat(result.txAppended).isTrue()
      individualSum += compressor.getCompressedSize()
    }

    // Measure additive compression size
    compressor.reset()
    for (tx in txsToTest) {
      val result = compressor.appendTransaction(tx)
      assertThat(result.txAppended).isTrue()
    }
    val additiveSize = compressor.getCompressedSize()

    // Additive compression should be smaller due to shared context
    assertThat(additiveSize).isLessThan(individualSum)
  }

  @Test
  fun `test getCompressedData returns correct data`() {
    val tx = TEST_TRANSACTIONS.first()
    compressor.appendTransaction(tx)

    val compressedData = compressor.getCompressedData()
    assertThat(compressedData.size).isEqualTo(compressor.getCompressedSize())
    assertThat(compressedData.size).isGreaterThan(0)
  }

  @Test
  fun `test getCompressedDataAndReset returns data and resets`() {
    val tx = TEST_TRANSACTIONS.first()
    compressor.appendTransaction(tx)

    val sizeBefore = compressor.getCompressedSize()
    assertThat(sizeBefore).isGreaterThan(0)

    val compressedData = compressor.getCompressedDataAndReset()
    assertThat(compressedData.size).isEqualTo(sizeBefore)
    assertThat(compressor.getCompressedSize()).isZero()
  }

  @Test
  fun `test multiple transactions accumulate`() {
    val txs = TEST_TRANSACTIONS.take(5)
    var prevSize = 0

    for (tx in txs) {
      val result = compressor.appendTransaction(tx)
      assertThat(result.txAppended).isTrue()
      assertThat(result.compressedSizeBefore).isEqualTo(prevSize)
      assertThat(result.compressedSizeAfter).isGreaterThan(prevSize)
      prevSize = result.compressedSizeAfter
    }

    assertThat(compressor.getCompressedSize()).isEqualTo(prevSize)
    assertThat(compressor.getUncompressedSize()).isGreaterThan(0)
  }
}

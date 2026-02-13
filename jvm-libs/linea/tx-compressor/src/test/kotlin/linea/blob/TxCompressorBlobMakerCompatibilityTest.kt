package linea.blob

import linea.rlp.RLP
import net.consensys.linea.nativecompressor.CompressorTestData
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.BlockBody
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder
import org.hyperledger.besu.ethereum.core.Difficulty
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.ethereum.rlp.BytesValueRLPOutput
import org.hyperledger.besu.evm.log.LogsBloomFilter
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance

/**
 * Integration test verifying that transactions compressed with TxCompressor
 * will fit into BlobCompressor when assembled into a block.
 *
 * This is a critical compatibility test ensuring the sequencer can safely use
 * TxCompressor for block building and then pass the resulting blocks to the
 * coordinator's BlobMaker.
 */
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class TxCompressorBlobMakerCompatibilityTest {
  companion object {
    private const val BLOB_LIMIT = 128 * 1024
    private const val BLOB_OVERHEAD = 100 // Conservative overhead estimate
    private val TEST_BLOCKS = CompressorTestData.blocksRlpEncoded

    private fun extractTransactionsFromBlocks(): List<Pair<Transaction, ByteArray>> {
      return TEST_BLOCKS.flatMap { blockRlp ->
        val block = RLP.decodeBlockWithMainnetFunctions(blockRlp)
        block.body.transactions.map { tx ->
          tx to encodeTransaction(tx)
        }
      }
    }

    private fun encodeTransaction(tx: Transaction): ByteArray {
      val rlpOutput = BytesValueRLPOutput()
      tx.writeTo(rlpOutput)
      return rlpOutput.encoded().toArray()
    }

    private val TEST_TRANSACTIONS: List<Pair<Transaction, ByteArray>> by lazy {
      extractTransactionsFromBlocks()
    }
  }

  private lateinit var txCompressor: GoBackedTxCompressor
  private lateinit var blobCompressor: GoBackedBlobCompressor

  @BeforeEach
  fun before() {
    // Create TxCompressor with limit accounting for blob overhead
    txCompressor = GoBackedTxCompressor.getInstance(
      TxCompressorVersion.V1,
      BLOB_LIMIT - BLOB_OVERHEAD,
    )

    // Create BlobCompressor with full blob limit
    blobCompressor = GoBackedBlobCompressor.getInstance(
      BlobCompressorVersion.V2,
      BLOB_LIMIT,
    )
  }

  @Test
  fun `transactions from TxCompressor fit in BlobCompressor`() {
    // Fill TxCompressor with transactions
    val acceptedTxs = mutableListOf<Transaction>()
    for ((tx, rlpTx) in TEST_TRANSACTIONS) {
      if (!txCompressor.canAppendTransaction(rlpTx)) {
        break
      }
      val result = txCompressor.appendTransaction(rlpTx)
      assertThat(result.txAppended).isTrue()
      acceptedTxs.add(tx)
    }

    assertThat(acceptedTxs).isNotEmpty()

    // Build a block with the accepted transactions
    val block = buildTestBlock(acceptedTxs)
    val blockRlp = RLP.encodeBlock(block)

    // Verify BlobCompressor accepts the block
    val result = blobCompressor.appendBlock(blockRlp)
    assertThat(result.blockAppended)
      .withFailMessage(
        "Block built with TxCompressor should fit in BlobCompressor. " +
          "TxCompressor size: ${txCompressor.getCompressedSize()}, " +
          "BlobCompressor size after: ${result.compressedSizeAfter}",
      )
      .isTrue()
  }

  @Test
  fun `multiple blocks from TxCompressor fit in single blob`() {
    val allAcceptedTxs = mutableListOf<List<Transaction>>()
    var txIterator = TEST_TRANSACTIONS.iterator()

    // Build multiple blocks
    repeat(3) {
      txCompressor.reset()
      val blockTxs = mutableListOf<Transaction>()

      while (txIterator.hasNext()) {
        val (tx, rlpTx) = txIterator.next()
        if (!txCompressor.canAppendTransaction(rlpTx)) {
          break
        }
        txCompressor.appendTransaction(rlpTx)
        blockTxs.add(tx)
      }

      if (blockTxs.isNotEmpty()) {
        allAcceptedTxs.add(blockTxs)
      }

      if (!txIterator.hasNext()) {
        txIterator = TEST_TRANSACTIONS.iterator()
      }
    }

    // Add all blocks to BlobCompressor
    for (blockTxs in allAcceptedTxs) {
      val block = buildTestBlock(blockTxs)
      val blockRlp = RLP.encodeBlock(block)
      val result = blobCompressor.appendBlock(blockRlp)
      assertThat(result.blockAppended)
        .withFailMessage("All blocks built with TxCompressor should fit in BlobCompressor")
        .isTrue()
    }

    assertThat(blobCompressor.getCompressedData().size).isGreaterThan(0)
  }

  @Test
  fun `TxCompressor provides better compression than individual estimation`() {
    // This test verifies the key benefit of TxCompressor: maintaining compression
    // context across transactions results in better compression than estimating
    // each transaction individually.

    val txsToTest = TEST_TRANSACTIONS.take(50)

    // Estimate individual sizes (worst case)
    var individualEstimate = 0
    for ((_, rlpTx) in txsToTest) {
      txCompressor.reset()
      txCompressor.appendTransaction(rlpTx)
      individualEstimate += txCompressor.getCompressedSize()
    }

    // Actual additive compression
    txCompressor.reset()
    for ((_, rlpTx) in txsToTest) {
      txCompressor.appendTransaction(rlpTx)
    }
    val actualSize = txCompressor.getCompressedSize()

    // Additive compression should be significantly smaller
    val savings = 1.0 - (actualSize.toDouble() / individualEstimate)
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

package linea.blob

import linea.rlp.RLP
import net.consensys.linea.nativecompressor.CompressorTestData
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Wei
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
import java.math.BigInteger
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
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class TxCompressorBlobMakerCompatibilityTest {
  companion object {
    private const val BLOB_LIMIT = 128 * 1024
    private const val BLOB_OVERHEAD = 100 // Conservative overhead estimate
    private val TEST_BLOCKS = CompressorTestData.blocksRlpEncoded

    private val signatureAlgorithm = SignatureAlgorithmFactory.getInstance()

    private fun encodeTransaction(tx: Transaction): ByteArray {
      val rlpOutput = BytesValueRLPOutput()
      tx.writeTo(rlpOutput)
      return rlpOutput.encoded().toArray()
    }

    /**
     * Generate a transfer transaction with randomized fields.
     * Randomizing all fields prevents compression from being too efficient,
     * which is important for testing the worst case where TxCompressor
     * context sharing provides minimal benefit.
     */
    private fun generateRandomizedTransferTx(nonce: Long, random: Random): Pair<Transaction, ByteArray> {
      // Generate a new key pair for each transaction (different sender)
      val keyPair = signatureAlgorithm.generateKeyPair()

      // Random recipient address
      val toAddressBytes = ByteArray(20)
      random.nextBytes(toAddressBytes)
      val toAddress = Address.wrap(Bytes.wrap(toAddressBytes))

      // Random gas price (1-100 gwei range)
      val gasPrice = Wei.of(random.nextLong(1_000_000_000L, 100_000_000_000L))

      // Random gas limit (21000-100000 range)
      val gasLimit = random.nextLong(21000, 100000)

      // Random value (0-1 ETH range in wei, using smaller range to fit in Long)
      val value = Wei.of(random.nextLong(0, 1_000_000_000_000_000_000L))

      // Random small payload (0-100 bytes of random data)
      val payloadSize = random.nextInt(0, 100)
      val payload = if (payloadSize > 0) {
        val payloadBytes = ByteArray(payloadSize)
        random.nextBytes(payloadBytes)
        Bytes.wrap(payloadBytes)
      } else {
        Bytes.EMPTY
      }

      val tx = Transaction.builder()
        .type(org.hyperledger.besu.datatypes.TransactionType.FRONTIER)
        .nonce(nonce)
        .gasPrice(gasPrice)
        .gasLimit(gasLimit)
        .to(toAddress)
        .value(value)
        .payload(payload)
        .chainId(BigInteger.ONE)
        .signAndBuild(keyPair)

      return tx to encodeTransaction(tx)
    }

    /**
     * Generate many randomized transactions - worst case for compression.
     * Uses a seeded random for reproducibility.
     */
    fun generateManyRandomizedTransactions(count: Int, seed: Long = 12345L): List<Pair<Transaction, ByteArray>> {
      val random = Random(seed)
      return (0 until count).map { generateRandomizedTransferTx(it.toLong(), random) }
    }

    // ERC-20 transfer function selector: transfer(address,uint256)
    // keccak256("transfer(address,uint256)") = 0xa9059cbb...
    private val ERC20_TRANSFER_SELECTOR = Bytes.fromHexString("0xa9059cbb")

    /**
     * Generate an ERC-20 transfer transaction.
     * - Same sender (keyPair)
     * - Same token contract address
     * - Same recipient address
     * - Random transfer amount
     *
     * This simulates realistic ERC-20 transfer patterns where the same user
     * transfers tokens to the same recipient multiple times with different amounts.
     */
    private fun generateErc20TransferTx(
      nonce: Long,
      keyPair: org.hyperledger.besu.crypto.KeyPair,
      tokenContractAddress: Address,
      recipientAddress: Address,
      random: Random,
    ): Pair<Transaction, ByteArray> {
      // Random transfer amount (1 to 1000 tokens with 18 decimals)
      val tokenAmount = BigInteger.valueOf(random.nextLong(1, 1000))
        .multiply(BigInteger.TEN.pow(18))

      // ERC-20 transfer(address to, uint256 amount) calldata
      // 4 bytes selector + 32 bytes address (padded) + 32 bytes amount
      val recipientPadded = Bytes.concatenate(
        Bytes.wrap(ByteArray(12)), // 12 zero bytes padding
        recipientAddress,
      )
      val amountBytes = Bytes.wrap(tokenAmount.toByteArray())
      val amountPadded = Bytes.concatenate(
        Bytes.wrap(ByteArray(32 - amountBytes.size())), // pad to 32 bytes
        amountBytes,
      )
      val payload = Bytes.concatenate(ERC20_TRANSFER_SELECTOR, recipientPadded, amountPadded)

      // Random gas price (1-50 gwei range)
      val gasPrice = Wei.of(random.nextLong(1_000_000_000L, 50_000_000_000L))

      val tx = Transaction.builder()
        .type(org.hyperledger.besu.datatypes.TransactionType.FRONTIER)
        .nonce(nonce)
        .gasPrice(gasPrice)
        .gasLimit(100000) // ERC-20 transfers typically need ~65k gas
        .to(tokenContractAddress)
        .value(Wei.ZERO) // ERC-20 transfers don't send ETH
        .payload(payload)
        .chainId(BigInteger.ONE)
        .signAndBuild(keyPair)

      return tx to encodeTransaction(tx)
    }

    /**
     * Generate many ERC-20 transfer transactions from the same sender to the same recipient.
     * This tests a realistic scenario where compression can benefit from repeated patterns
     * (same sender, same contract, same recipient, only amount varies).
     */
    fun generateManyErc20Transfers(count: Int, seed: Long = 54321L): List<Pair<Transaction, ByteArray>> {
      val random = Random(seed)

      // Fixed sender
      val keyPair = signatureAlgorithm.generateKeyPair()

      // Fixed token contract address (e.g., USDC-like)
      val tokenContractAddress = Address.fromHexString("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

      // Fixed recipient address
      val recipientAddress = Address.fromHexString("0x1234567890123456789012345678901234567890")

      return (0 until count).map { i ->
        generateErc20TransferTx(i.toLong(), keyPair, tokenContractAddress, recipientAddress, random)
      }
    }

    // Note: TxCompressor uses a global singleton in the Go code, so we initialize once
    // and reset between tests to avoid interference with other test classes.
    private val txCompressor: GoBackedTxCompressor by lazy {
      GoBackedTxCompressor.getInstance(
        TxCompressorVersion.V1,
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
    // Generate many randomized transfer transactions
    val allTxs = generateManyRandomizedTransactions(5000)

    // Fill TxCompressor until it says "full"
    val acceptedTxs = mutableListOf<Transaction>()
    var acceptedPlainSize = 0
    for ((tx, rlpTx) in allTxs) {
      if (!txCompressor.canAppendTransaction(rlpTx)) {
        break
      }
      val result = txCompressor.appendTransaction(rlpTx)
      assertThat(result.txAppended).isTrue()
      acceptedTxs.add(tx)
      acceptedPlainSize += rlpTx.size
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

    // Get RawCompressedSize for comparison (stateless compression of block RLP)
    val rawCompressedSize = blobCompressor.compressedSize(blockRlp)

    // Log comprehensive comparison
    println("=== Compression Comparison (Worst Case: ${acceptedTxs.size} randomized txs) ===")
    println("Plain transaction data size:     $acceptedPlainSize bytes")
    println("Block RLP size (with header):    $blockRlpSize bytes")
    println(
      "TxCompressor compressed size:    $txCompressorSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * txCompressorSize / acceptedPlainSize)}%)",
    )
    println(
      "BlobCompressor compressed size:  $blobCompressorSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * blobCompressorSize / blockRlpSize)}%)",
    )
    println(
      "RawCompressedSize (stateless):   $rawCompressedSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * rawCompressedSize / blockRlpSize)}%)",
    )
    println(
      "TxCompressor vs BlobCompressor:  ${blobCompressorSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (blobCompressorSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println(
      "TxCompressor vs RawCompressed:   ${rawCompressedSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (rawCompressedSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println("Blob limit:                      $BLOB_LIMIT bytes")
    println(
      "Headroom remaining:              ${BLOB_LIMIT - blobCompressorSize} bytes " +
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
      block.body.transactions.map { tx -> tx to encodeTransaction(tx) }
    }

    // Fill TxCompressor with transactions
    val acceptedTxs = mutableListOf<Transaction>()
    var acceptedPlainSize = 0
    for ((tx, rlpTx) in testTransactions) {
      if (!txCompressor.canAppendTransaction(rlpTx)) {
        break
      }
      val result = txCompressor.appendTransaction(rlpTx)
      assertThat(result.txAppended).isTrue()
      acceptedTxs.add(tx)
      acceptedPlainSize += rlpTx.size
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

    // Get RawCompressedSize for comparison (stateless compression of block RLP)
    val rawCompressedSize = blobCompressor.compressedSize(blockRlp)

    // Log comprehensive comparison
    println("=== Compression Comparison (Real Test Data: ${acceptedTxs.size} txs) ===")
    println("Plain transaction data size:     $acceptedPlainSize bytes")
    println("Block RLP size (with header):    $blockRlpSize bytes")
    println(
      "TxCompressor compressed size:    $txCompressorSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * txCompressorSize / acceptedPlainSize)}%)",
    )
    println(
      "BlobCompressor compressed size:  $blobCompressorSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * blobCompressorSize / blockRlpSize)}%)",
    )
    println(
      "RawCompressedSize (stateless):   $rawCompressedSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * rawCompressedSize / blockRlpSize)}%)",
    )
    println(
      "TxCompressor vs BlobCompressor:  ${blobCompressorSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (blobCompressorSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println(
      "TxCompressor vs RawCompressed:   ${rawCompressedSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (rawCompressedSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println("Blob limit:                      $BLOB_LIMIT bytes")
    println(
      "Headroom remaining:              ${BLOB_LIMIT - blobCompressorSize} bytes " +
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
  fun `ERC-20 transfers - same sender and recipient with different amounts`() {
    // Generate many ERC-20 transfer transactions
    val allTxs = generateManyErc20Transfers(5000)

    // Fill TxCompressor until it says "full"
    val acceptedTxs = mutableListOf<Transaction>()
    var acceptedPlainSize = 0
    for ((tx, rlpTx) in allTxs) {
      if (!txCompressor.canAppendTransaction(rlpTx)) {
        break
      }
      val result = txCompressor.appendTransaction(rlpTx)
      assertThat(result.txAppended).isTrue()
      acceptedTxs.add(tx)
      acceptedPlainSize += rlpTx.size
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

    // Get RawCompressedSize for comparison
    val rawCompressedSize = blobCompressor.compressedSize(blockRlp)

    // Log comprehensive comparison
    println("=== Compression Comparison (ERC-20 Transfers: ${acceptedTxs.size} txs) ===")
    println("Plain transaction data size:     $acceptedPlainSize bytes")
    println("Block RLP size (with header):    $blockRlpSize bytes")
    println(
      "TxCompressor compressed size:    $txCompressorSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * txCompressorSize / acceptedPlainSize)}%)",
    )
    println(
      "BlobCompressor compressed size:  $blobCompressorSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * blobCompressorSize / blockRlpSize)}%)",
    )
    println(
      "RawCompressedSize (stateless):   $rawCompressedSize bytes " +
        "(ratio: ${"%.2f".format(100.0 * rawCompressedSize / blockRlpSize)}%)",
    )
    println(
      "TxCompressor vs BlobCompressor:  ${blobCompressorSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (blobCompressorSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println(
      "TxCompressor vs RawCompressed:   ${rawCompressedSize - txCompressorSize} bytes diff " +
        "(${"%.2f".format(100.0 * (rawCompressedSize - txCompressorSize) / txCompressorSize)}%)",
    )
    println("Blob limit:                      $BLOB_LIMIT bytes")
    println(
      "Headroom remaining:              ${BLOB_LIMIT - blobCompressorSize} bytes " +
        "(${"%.2f".format(100.0 * (BLOB_LIMIT - blobCompressorSize) / BLOB_LIMIT)}%)",
    )
  }

  /**
   * Verify TxCompressor provides compression benefit over individual estimation.
   */
  @Test
  fun `TxCompressor provides better compression than individual estimation`() {
    val txsToTest = generateManyRandomizedTransactions(100)

    // Estimate individual sizes (worst case - no context)
    var individualEstimate = 0
    for ((_, rlpTx) in txsToTest) {
      txCompressor.reset()
      txCompressor.appendTransaction(rlpTx)
      individualEstimate += txCompressor.getCompressedSize()
    }

    // Actual additive compression (with context)
    txCompressor.reset()
    for ((_, rlpTx) in txsToTest) {
      txCompressor.appendTransaction(rlpTx)
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

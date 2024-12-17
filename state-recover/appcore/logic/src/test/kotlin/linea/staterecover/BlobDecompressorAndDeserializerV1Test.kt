package linea.staterecover

import build.linea.staterecover.BlockL1RecoveredData
import build.linea.staterecover.TransactionL1RecoveredData
import io.vertx.core.Vertx
import kotlinx.datetime.Instant
import linea.blob.BlobCompressor
import linea.blob.GoBackedBlobCompressor
import linea.rlp.RLP
import net.consensys.encodeHex
import net.consensys.linea.blob.BlobCompressorVersion
import net.consensys.linea.blob.BlobDecompressorVersion
import net.consensys.linea.blob.GoNativeBlobDecompressorFactory
import net.consensys.linea.nativecompressor.CompressorTestData
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.Transaction
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import kotlin.jvm.optionals.getOrNull

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class BlobDecompressorAndDeserializerV1Test {
  private lateinit var compressor: BlobCompressor
  private val blockStaticFields = BlockHeaderStaticFields(
    coinbase = Address.ZERO.toArray(),
    gasLimit = 30_000_000UL,
    difficulty = 0UL
  )
  private lateinit var decompressorToDomain: BlobDecompressorAndDeserializer
  private lateinit var vertx: Vertx

  @BeforeEach
  fun setUp() {
    vertx = Vertx.vertx()
    compressor = GoBackedBlobCompressor.getInstance(
      compressorVersion = BlobCompressorVersion.V1_0_1,
      dataLimit = 124u * 1024u
    )
    val decompressor = GoNativeBlobDecompressorFactory.getInstance(BlobDecompressorVersion.V1_1_0)
    decompressorToDomain = BlobDecompressorToDomainV1(decompressor, blockStaticFields, vertx)
  }

  @AfterEach
  fun afterEach() {
    vertx.close()
  }

  @Test
  fun `should decompress block and transactions`() {
    val blocksRLP = CompressorTestData.blocksRlpEncoded.toList()
    assertBlockCompressionAndDecompression(blocksRLP)
  }

//  @Test
//  fun `should decompress block and transactions - tx with contract deployment`() {
//    assertBlockCompressionAndDecompression(CompressorTestData.blocksRlpEncodedV2)
//  }

  private fun assertBlockCompressionAndDecompression(
    blocksRLP: List<ByteArray>
  ) {
    val blocks = blocksRLP.map(RLP::decodeBlockWithMainnetFunctions)
    val startingBlockNumber = blocks[0].header.number.toULong()
    println("starting block number: $startingBlockNumber")

    val blob1 = compress(blocksRLP.slice(0..0))
    val blob2 = compress(blocksRLP.slice(3..3))

    val recoveredBlocks = decompressorToDomain.decompress(
      startBlockNumber = startingBlockNumber,
      blobs = listOf(blob1, blob2)
    ).get()
    assertThat(recoveredBlocks[0].blockNumber).isEqualTo(startingBlockNumber)

    recoveredBlocks.zip(blocks) { recoveredBlock, originalBlock ->
      assertBlockData(recoveredBlock, originalBlock)
    }
  }

  private fun assertBlockData(
    uncompressed: BlockL1RecoveredData,
    original: Block
  ) {
    println("asserting block: ${original.header.number} ${original.header}")
    assertThat(uncompressed.blockNumber).isEqualTo(original.header.number.toULong())
    assertThat(uncompressed.blockHash.encodeHex()).isEqualTo(original.header.hash.toArray().encodeHex())
    assertThat(uncompressed.coinbase).isEqualTo(blockStaticFields.coinbase)
    assertThat(uncompressed.blockTimestamp).isEqualTo(Instant.fromEpochSeconds(original.header.timestamp))
    assertThat(uncompressed.gasLimit).isEqualTo(blockStaticFields.gasLimit)
    assertThat(uncompressed.difficulty).isEqualTo(0UL)
    uncompressed.transactions.zip(original.body.transactions) { uncompressedTransaction, originalTransaction ->
      assertTransactionData(uncompressedTransaction, originalTransaction)
    }
  }

  private fun assertTransactionData(
    uncompressed: TransactionL1RecoveredData,
    original: Transaction
  ) {
    assertThat(uncompressed.type).isEqualTo(original.type.serializedType.toUByte())
    assertThat(uncompressed.from).isEqualTo(original.sender.toArray())
    assertThat(uncompressed.nonce).isEqualTo(original.nonce.toULong())
    assertThat(uncompressed.to).isEqualTo(original.to.getOrNull()?.toArray())
    assertThat(uncompressed.gasLimit).isEqualTo(original.gasLimit.toULong())
    assertThat(uncompressed.maxFeePerGas).isEqualTo(original.maxFeePerGas.getOrNull()?.asBigInteger)
    assertThat(uncompressed.maxPriorityFeePerGas).isEqualTo(original.maxPriorityFeePerGas.getOrNull()?.asBigInteger)
    assertThat(uncompressed.gasPrice).isEqualTo(original.gasPrice.getOrNull()?.asBigInteger)
    assertThat(uncompressed.value).isEqualTo(original.value.asBigInteger)
    assertThat(uncompressed.data?.encodeHex()).isEqualTo(original.data.getOrNull()?.toArray()?.encodeHex())
    if (uncompressed.accessList.isNullOrEmpty() != original.accessList.getOrNull().isNullOrEmpty()) {
      assertThat(uncompressed.accessList).isEqualTo(original.accessList.getOrNull())
    } else {
      uncompressed.accessList?.zip(original.accessList.getOrNull()!!) { a, b ->
        assertThat(a.address).isEqualTo(b.address.toArray())
        assertThat(a.storageKeys).isEqualTo(b.storageKeys.map { it.toArray() })
      }
    }
  }

  private fun compressBlocks(blocks: List<Block>): ByteArray {
    return compress(blocks.map { block -> block.toRlp().toArray() })
  }

  private fun compress(blocks: List<ByteArray>): ByteArray {
    blocks.forEach { blockRlp ->
      compressor.appendBlock(blockRlp)
    }

    return compressor.getCompressedDataAndReset()
  }
}

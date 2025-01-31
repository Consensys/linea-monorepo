package net.consensys.linea.blob

import kotlinx.datetime.Instant
import linea.blob.BlobCompressor
import linea.blob.GoBackedBlobCompressor
import linea.domain.AccessListEntry
import linea.domain.TransactionFactory
import linea.domain.createBlock
import linea.domain.toBesu
import linea.rlp.BesuRlpBlobDecoder
import linea.rlp.RLP
import net.consensys.decodeHex
import net.consensys.eth
import net.consensys.linea.nativecompressor.CompressorTestData
import net.consensys.toBigInteger
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import kotlin.jvm.optionals.getOrNull

class GoNativeBlobDecompressorTest {
  private val blobCompressedLimit = 30 * 1024
  private lateinit var compressor: BlobCompressor
  private lateinit var decompressor: BlobDecompressor

  @BeforeEach
  fun beforeEach() {
    compressor = GoBackedBlobCompressor
      .getInstance(BlobCompressorVersion.V1_0_1, blobCompressedLimit.toUInt())
    decompressor = GoNativeBlobDecompressorFactory.getInstance(BlobDecompressorVersion.V1_1_0)
  }

  @Test
  fun `when blocks are compressed with compressor shall decompress them back`() {
    val blocks = CompressorTestData.blocksRlpEncoded
    compressor.appendBlock(blocks[0])
    compressor.appendBlock(blocks[1])

    val compressedData = compressor.getCompressedData()

    val decompressedBlob = decompressor.decompress(compressedData)
    assertThat(decompressedBlob.size).isGreaterThan(compressedData.size)
    val decompressedBlocks: List<ByteArray> = rlpDecodeAsListOfBytes(decompressedBlob)
    assertThat(decompressedBlocks).hasSize(2)
  }

  @Test
  fun `should decompress original data`() {
    val tx0 = TransactionFactory.createTransactionFrontier(
      nonce = 1uL,
      gasLimit = 22_0000uL,
      to = null,
      value = 1uL.eth.toBigInteger(),
      input = byteArrayOf()
    )
    val tx1 = TransactionFactory.createTransactionEip1559(
      nonce = 2uL,
      gasLimit = 23_0000uL,
      to = null,
      value = 2uL.eth.toBigInteger(),
      input = "0x1234".toByteArray(),
      accessList = listOf(
        AccessListEntry(
          address = "0x00000000000000000000000000000000000000ff".decodeHex(),
          storageKeys = listOf(
            "0x0000000000000000000000000000000000000000000000000000000000000001".decodeHex(),
            "0x0000000000000000000000000000000000000000000000000000000000000002".decodeHex()
          )
        )
      )
    )
    val originalBesuBlock = createBlock(
      number = 123uL,
      timestamp = Instant.parse("2025-01-02T12:23:45Z"),
      transactions = listOf(tx0, tx1)
    ).toBesu()

    compressor.appendBlock(RLP.encodeBlock(originalBesuBlock))
    val decompressedData = decompressor.decompress(compressor.getCompressedData())
    val decompressedBlocks: List<ByteArray> = rlpDecodeAsListOfBytes(decompressedData)
    assertThat(decompressedBlocks).hasSize(1)
    val decompressedBlock = decompressedBlocks[0]
    val decodedBlock = BesuRlpBlobDecoder.decode(decompressedBlock)

    // Only BlockHash and Timestamp are compressed to the Blob
    assertThat(decodedBlock.header.hash.toArray()).isEqualTo(originalBesuBlock.header.hash)
    assertThat(decodedBlock.header.timestamp).isEqualTo(Instant.parse("2025-01-02T12:23:45Z").epochSeconds)

    assertThat(decodedBlock.body.transactions).hasSize(2)
    val decompressedTx0 = decodedBlock.body.transactions[0]
    val decompressedTx1 = decodedBlock.body.transactions[0]

    assertThat(decompressedTx0.type).isEqualTo(tx0.type.toBesu())
    assertThat(decompressedTx0.sender.toArray()).isEqualTo(tx0.toBesu().sender.toArray())
    assertThat(decompressedTx0.nonce.toULong()).isEqualTo(tx0.nonce)
    assertThat(decompressedTx0.gasLimit.toULong()).isEqualTo(tx0.gasLimit)
    assertThat(decompressedTx0.maxFeePerGas.getOrNull()?.asBigInteger).isEqualTo(tx0.maxFeePerGas)
    assertThat(decompressedTx0.maxPriorityFeePerGas.getOrNull()?.asBigInteger).isEqualTo(tx0.maxPriorityFeePerGas)
    assertThat(decompressedTx0.gasPrice.getOrNull()?.asBigInteger).isEqualTo(tx0.gasPrice)
    assertThat(decompressedTx0.to.getOrNull()?.toArray()).isEqualTo(tx0.to)
    assertThat(decompressedTx0.value.asBigInteger).isEqualTo(tx0.value)
    assertThat(decompressedTx0.payload.toArray()).isEqualTo(tx0.input)
    assertThat(decompressedTx0.accessList).isNull()

    assertThat(decompressedTx1.type.serializedType.toUByte()).isEqualTo(tx1.type)
    assertThat(decompressedTx1.sender.toArray()).isEqualTo(tx1.toBesu().sender.toArray())
    assertThat(decompressedTx1.nonce.toULong()).isEqualTo(tx1.nonce)
    assertThat(decompressedTx1.gasLimit.toULong()).isEqualTo(tx1.gasLimit)
    assertThat(decompressedTx1.maxFeePerGas.getOrNull()?.asBigInteger).isEqualTo(tx1.maxFeePerGas)
    assertThat(decompressedTx1.maxPriorityFeePerGas.getOrNull()?.asBigInteger).isEqualTo(tx1.maxPriorityFeePerGas)
    assertThat(decompressedTx1.gasPrice.getOrNull()?.asBigInteger).isEqualTo(tx1.gasPrice)
    assertThat(decompressedTx1.to.getOrNull()?.toArray()).isEqualTo(tx1.to)
    assertThat(decompressedTx1.value.asBigInteger).isEqualTo(tx1.value)
    assertThat(decompressedTx1.payload.toArray()).isEqualTo(tx1.input)
    assertThat(decompressedTx1.accessList).isNull()
  }
}

package net.consensys.linea.blob

import net.consensys.linea.nativecompressor.CompressorTestData
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class GoNativeBlobDecompressorTest {
  private val blobCompressedLimit = 10 * 1024
  private lateinit var compressor: GoNativeBlobCompressor
  private lateinit var decompressor: BlobDecompressor

  @BeforeEach
  fun beforeEach() {
    compressor = GoNativeBlobCompressorFactory
      .getInstance(BlobCompressorVersion.V1_0_1)
      .apply {
        Init(
          dataLimit = blobCompressedLimit,
          dictPath = GoNativeBlobCompressorFactory.dictionaryPath.toAbsolutePath().toString()
        )
        Reset()
      }
    decompressor = GoNativeBlobDecompressorFactory.getInstance(BlobDecompressorVersion.V1_1_0)
  }

  @Test
  fun `when blocks are compressed with compressor shall decompress them back`() {
    val blocks = CompressorTestData.blocksRlpEncoded.slice(0..1)
    assertTrue(compressor.Write(blocks[0], blocks[0].size))
    assertTrue(compressor.Write(blocks[1], blocks[1].size))

    val compressedData = ByteArray(compressor.Len())
    compressor.Bytes(compressedData)

    val decompressedDataBuffer = ByteArray(blobCompressedLimit * 20)
    val decompressedBlobSize = decompressor.decompress(compressedData, decompressedDataBuffer)
    assertThat(decompressedBlobSize).isGreaterThan(0)
    // TODO: assert decompressedDataBuffer content
  }
}

import net.consensys.linea.blob.BlobCompressorVersion
import net.consensys.linea.blob.BlobDecompressor
import net.consensys.linea.blob.BlobDecompressorVersion
import net.consensys.linea.blob.GoNativeBlobCompressor
import net.consensys.linea.blob.GoNativeBlobCompressorFactory
import net.consensys.linea.blob.GoNativeBlobDecompressorFactory
import net.consensys.linea.nativecompressor.CompressorTestData
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Assertions.assertDoesNotThrow
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.parallel.Execution
import org.junit.jupiter.api.parallel.ExecutionMode

@Execution(ExecutionMode.SAME_THREAD)
class BlobDecompressorToDomainImplTest {

  private val blobCompressedLimit = 10 * 1024
  private lateinit var compressor: GoNativeBlobCompressor
  private lateinit var decompressor: BlobDecompressor
  private lateinit var blobToBlockDomainMapper: BlobDecompressorToDomainImpl

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
    blobToBlockDomainMapper = BlobDecompressorToDomainImpl()
  }

  @Test
  fun `when blocks are compressed with compressor shall decompress them back`() {
    val blocks = CompressorTestData.blocksRlpEncoded.slice(0..1)
    assertTrue(compressor.Write(blocks[0], blocks[0].size))
    assertTrue(compressor.Write(blocks[1], blocks[1].size))

    val compressedData = ByteArray(compressor.Len())
    compressor.Bytes(compressedData)

    val decompressedBlob = decompressor.decompress(compressedData)
    Assertions.assertThat(decompressedBlob.size).isGreaterThan(compressedData.size)

    val decodedBlocks = blobToBlockDomainMapper.decompress(decompressedBlob)
    Assertions.assertThat(decodedBlocks).hasSameSizeAs(blocks)
  }

  @Test
  fun `verify rlp encoded blocks are decoded`() {
    val encodedBlocks = CompressorTestData.blocksRlpEncoded.slice(0..1)
    encodedBlocks.forEach {
      assertDoesNotThrow {
        blobToBlockDomainMapper.decodeRLPBlock(it)
      }
    }
  }
}

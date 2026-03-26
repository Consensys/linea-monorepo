package linea.blob

import com.sun.jna.ptr.PointerByReference
import net.consensys.linea.nativecompressor.CompressorTestData
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.random.Random

class GoBackedBlobCompressorTest {
  companion object {
    private const val DATA_LIMIT = 24 * 1024
    private val TEST_DATA = CompressorTestData.blocksRlpEncoded
  }

  private fun newCompressor() = GoBackedBlobCompressor.getInstance(BlobCompressorVersion.V4, DATA_LIMIT)

  @Test
  fun `test appendBlock with data within limit`() {
    newCompressor().use { compressor ->
      val result = compressor.appendBlock(TEST_DATA.first())
      assertThat(result.blockAppended).isTrue
      assertThat(result.compressedSizeBefore).isZero()
      assertThat(result.compressedSizeAfter).isGreaterThan(0)
    }
  }

  @Test
  fun `test invalid rlp block`() {
    newCompressor().use { compressor ->
      assertThrows<BlobCompressionException>("rlp: expected input list for types.extblock") {
        compressor.appendBlock(Random.nextBytes(100))
      }
    }
  }

  @Test
  fun `test reset`() {
    newCompressor().use { compressor ->
      val blocks = TEST_DATA.iterator()
      assertThat(compressor.getCompressedData()).isEmpty()
      var res = compressor.appendBlock(blocks.next())
      assertThat(res.blockAppended).isTrue()
      assertThat(res.compressedSizeAfter).isGreaterThan(0)
      assertThat(res.compressedSizeAfter).isEqualTo(compressor.getCompressedData().size)

      compressor.reset()

      assertThat(compressor.getCompressedData()).isEmpty()
      res = compressor.appendBlock(blocks.next())
      assertThat(res.blockAppended).isTrue()
      assertThat(res.compressedSizeAfter).isEqualTo(compressor.getCompressedData().size)
    }
  }

  @Test
  fun `multiple instances have independent state`() {
    newCompressor().use { c1 ->
      newCompressor().use { c2 ->
        // append one block to c1 only
        val res = c1.appendBlock(TEST_DATA.first())
        assertThat(res.blockAppended).isTrue()

        // c2 must still be empty
        assertThat(c2.getCompressedData()).isEmpty()

        // append a different block to c2
        val res2 = c2.appendBlock(TEST_DATA[1])
        assertThat(res2.blockAppended).isTrue()

        // c1 must be unaffected by c2's write
        assertThat(c1.getCompressedData().size).isEqualTo(res.compressedSizeAfter)
      }
    }
  }

  @Test
  fun `init with bad dictionary path returns native error message`() {
    val lib = GoNativeBlobCompressorFactory.getInstance(BlobCompressorVersion.V4)
    val errOut = PointerByReference()
    val handle = lib.Init(DATA_LIMIT, "/nonexistent/dictionary.bin", errOut)
    assertThat(handle).isEqualTo(-1)
    assertThat(errOut.value?.getString(0)).isNotBlank()
  }

  @Test
  fun `close releases instance and a new one can be created`() {
    val c1 = newCompressor()
    c1.appendBlock(TEST_DATA.first())
    c1.close()

    // must be able to create a fresh instance after closing
    newCompressor().use { c2 ->
      assertThat(c2.getCompressedData()).isEmpty()
    }
  }
}

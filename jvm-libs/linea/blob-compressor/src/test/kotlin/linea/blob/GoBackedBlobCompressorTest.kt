package linea.blob

import com.sun.jna.ptr.PointerByReference
import linea.blob.BlobCompressorSelectorTest.Companion.compressBlocks
import net.consensys.linea.nativecompressor.CompressorTestData
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import java.util.concurrent.CopyOnWriteArrayList
import java.util.concurrent.CountDownLatch
import java.util.concurrent.CyclicBarrier
import java.util.concurrent.TimeUnit
import kotlin.concurrent.atomics.AtomicReference
import kotlin.concurrent.atomics.ExperimentalAtomicApi
import kotlin.random.Random

class GoBackedBlobCompressorTest {
  companion object {
    private const val DATA_LIMIT = 24 * 1024
    private val TEST_DATA = CompressorTestData.blocksRlpEncoded
  }

  private fun newCompressor() = BlobCompressorFactory.getInstance(BlobCompressorVersion.V4, DATA_LIMIT)

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
  fun `test compression data limit exceeded`() {
    newCompressor().use { compressor ->
      var blocks = TEST_DATA.iterator()
      var result = compressor.appendBlock(blocks.next())
      assertThat(result.blockAppended).isTrue()
      while (result.blockAppended) {
        val blockRlp = blocks.next()
        val canAppend = compressor.canAppendBlock(blockRlp)
        result = compressor.appendBlock(blockRlp)
        assertThat(canAppend).isEqualTo(result.blockAppended)
        if (!blocks.hasNext()) {
          blocks = TEST_DATA.iterator()
        }
      }
      assertThat(result.blockAppended).isFalse()
      assertThat(result.compressedSizeBefore).isGreaterThan(0)
      assertThat(result.compressedSizeAfter).isEqualTo(result.compressedSizeBefore)
    }
  }

  @Test
  fun `test batches`() {
    newCompressor().use { compressor ->
      val blocks = TEST_DATA.iterator()
      var res = compressor.appendBlock(blocks.next())
      assertThat(res.blockAppended).isTrue()

      compressor.startNewBatch()

      res = compressor.appendBlock(blocks.next())
      assertThat(res.blockAppended).isTrue()
      assertThat(compressor.getCompressedData().size).isGreaterThan(0)
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

  @OptIn(ExperimentalAtomicApi::class)
  @Test
  fun `compresses multiple blocks concurrently with multiple v4 instances without exception`() {
    val sampleBlocks = CompressorTestData.blocksRlpEncoded
    assertThat(sampleBlocks).isNotEmpty

    newCompressor().use { compressor1 ->
      newCompressor().use { compressor2 ->
        assertTrue(compressor1 != compressor2)

        val compressed1 = compressBlocks(compressor1, sampleBlocks)
        val compressed2 = compressBlocks(compressor2, sampleBlocks)

        assertThat(compressed1).isEqualTo(compressBlocks(compressor1, sampleBlocks))
        assertThat(compressed2).isEqualTo(compressBlocks(compressor2, sampleBlocks))

        val compressed1Parallel: AtomicReference<ByteArray> = AtomicReference(ByteArray(0))
        val compressed2Parallel: AtomicReference<ByteArray> = AtomicReference(ByteArray(0))

        val errors = CopyOnWriteArrayList<Throwable>()
        val startBarrier = CyclicBarrier(3)
        val doneLatch = CountDownLatch(2)

        val thread1 = Thread {
          try {
            startBarrier.await()
            compressed1Parallel.store(compressBlocks(compressor1, sampleBlocks))
          } catch (t: Throwable) {
            errors.add(t)
          } finally {
            doneLatch.countDown()
          }
        }

        val thread2 = Thread {
          try {
            startBarrier.await()
            compressed2Parallel.store(compressBlocks(compressor2, sampleBlocks))
          } catch (t: Throwable) {
            errors.add(t)
          } finally {
            doneLatch.countDown()
          }
        }

        thread1.start()
        thread2.start()
        startBarrier.await()
        assertTrue(doneLatch.await(10, TimeUnit.SECONDS), "compression threads did not complete in time")

        assertThat(compressed1Parallel.load()).isEqualTo(compressed1)
        assertThat(compressed2Parallel.load()).isEqualTo(compressed2)
        assertThat(errors).isEmpty()
      }
    }
  }
}

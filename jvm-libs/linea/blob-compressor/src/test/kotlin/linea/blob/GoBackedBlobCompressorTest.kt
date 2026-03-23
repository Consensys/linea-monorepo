package linea.blob

import net.consensys.linea.nativecompressor.CompressorTestData
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.assertThrows
import java.util.concurrent.CopyOnWriteArrayList
import java.util.concurrent.CountDownLatch
import java.util.concurrent.CyclicBarrier
import java.util.concurrent.TimeUnit
import kotlin.concurrent.atomics.AtomicReference
import kotlin.concurrent.atomics.ExperimentalAtomicApi
import kotlin.random.Random

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class GoBackedBlobCompressorTest {
  companion object {
    private const val DATA_LIMIT = 24 * 1024
    private val TEST_DATA = CompressorTestData.blocksRlpEncoded
    private val compressor = GoBackedBlobCompressor.getInstance(BlobCompressorVersion.V3, DATA_LIMIT)
  }

  @BeforeEach
  fun before() {
    compressor.reset()
  }

  @Test
  fun `test appendBlock with data within limit`() {
    val blocks = TEST_DATA
    val result = compressor.appendBlock(blocks.first())
    assertThat(result.blockAppended).isTrue
    assertThat(result.compressedSizeBefore).isZero()
    assertThat(result.compressedSizeAfter).isGreaterThan(0)
  }

  @Test
  fun `test invalid rlp block`() {
    val block = Random.nextBytes(100)
    assertThrows<BlobCompressionException>("rlp: expected input list for types.extblock") {
      compressor.appendBlock(block)
    }
  }

  @Test
  fun `test compression data limit exceeded`() {
    var blocks = TEST_DATA.iterator()
    var result = compressor.appendBlock(blocks.next())
    // at least one block should be appended
    assertThat(result.blockAppended).isTrue()
    while (result.blockAppended) {
      val blockRlp = blocks.next()
      val canAppend = compressor.canAppendBlock(blockRlp)
      result = compressor.appendBlock(blockRlp)
      // assert consistency between canAppendBlock and appendBlock
      assertThat(canAppend).isEqualTo(result.blockAppended)
      if (!blocks.hasNext()) {
        // recompress again, until the limit is reached
        blocks = TEST_DATA.iterator()
      }
    }
    assertThat(result.blockAppended).isFalse()
    assertThat(result.compressedSizeBefore).isGreaterThan(0)
    assertThat(result.compressedSizeAfter).isEqualTo(result.compressedSizeBefore)
  }

  @Test
  fun `test reset`() {
    val blocks = TEST_DATA.iterator()
    assertThat(compressor.goNativeBlobCompressor.Len()).isZero()
    var res = compressor.appendBlock(blocks.next())
    assertThat(res.blockAppended).isTrue()
    assertThat(res.compressedSizeBefore).isZero()
    assertThat(res.compressedSizeAfter).isGreaterThan(0)
    assertThat(res.compressedSizeAfter).isEqualTo(compressor.goNativeBlobCompressor.Len())

    compressor.reset()

    assertThat(compressor.goNativeBlobCompressor.Len()).isZero()
    res = compressor.appendBlock(blocks.next())
    assertThat(res.blockAppended).isTrue()
    assertThat(res.compressedSizeBefore).isZero()
    assertThat(res.compressedSizeAfter).isGreaterThan(0)
    assertThat(res.compressedSizeAfter).isEqualTo(compressor.goNativeBlobCompressor.Len())
  }

  @Test
  fun `test batches`() {
    val blocks = TEST_DATA.iterator()
    var res = compressor.appendBlock(blocks.next())
    assertThat(res.blockAppended).isTrue()

    compressor.startNewBatch()

    res = compressor.appendBlock(blocks.next())
    assertThat(res.blockAppended).isTrue()
    assertThat(compressor.getCompressedData().size).isGreaterThan(0)
  }

  @Test
  fun `should calculate the compression size of raw data`() {
    val data = TEST_DATA.first()
    val compressedSize = compressor.compressedSize(data)
    assertThat(compressedSize).isBetween(1, data.size - 1)
  }

  @OptIn(ExperimentalAtomicApi::class)
  @Test
  fun `compresses multiple blocks concurrently with v2 and v3 without exception`() {
    val v2 = BlobCompressorVersion.V2
    val v3 = BlobCompressorVersion.V3
    val compressorV2 = GoBackedBlobCompressor.getInstance(v2, DATA_LIMIT)
    val compressorV3 = GoBackedBlobCompressor.getInstance(v3, DATA_LIMIT)
    assertTrue(compressorV2 != compressorV3)

    val sampleBlocks = CompressorTestData.blocksRlpEncoded
    assertThat(sampleBlocks).isNotEmpty
    val compressedV2 = compressBlocks(compressorV2, sampleBlocks)
    val compressedV3 = compressBlocks(compressorV3, sampleBlocks)

    assertThat(compressedV2).isEqualTo(compressBlocks(compressorV2, sampleBlocks))
    assertThat(compressedV3).isEqualTo(compressBlocks(compressorV3, sampleBlocks))

    val compressedV2Parallel: AtomicReference<ByteArray> = AtomicReference(ByteArray(0))
    val compressedV3Parallel: AtomicReference<ByteArray> = AtomicReference(ByteArray(0))

    val errors = CopyOnWriteArrayList<Throwable>()
    val startBarrier = CyclicBarrier(3)
    val doneLatch = CountDownLatch(2)

    val v2Thread = Thread {
      try {
        startBarrier.await()
        compressedV2Parallel.store(compressBlocks(compressorV2, sampleBlocks))
      } catch (t: Throwable) {
        errors.add(t)
      } finally {
        doneLatch.countDown()
      }
    }

    val v3Thread = Thread {
      try {
        startBarrier.await()
        compressedV3Parallel.store(compressBlocks(compressorV3, sampleBlocks))
      } catch (t: Throwable) {
        errors.add(t)
      } finally {
        doneLatch.countDown()
      }
    }

    v2Thread.start()
    v3Thread.start()
    startBarrier.await()
    assertTrue(doneLatch.await(10, TimeUnit.SECONDS), "compression threads did not complete in time")

    assertThat(compressedV2Parallel.load()).isEqualTo(compressedV2)
    assertThat(compressedV3Parallel.load()).isEqualTo(compressedV3)
    assertThat(errors).isEmpty()
  }

  @OptIn(ExperimentalAtomicApi::class)
  @Test
  fun `compresses multiple blocks concurrently with two v2 process isolated instances without exception`() {
    val v2 = BlobCompressorVersion.V2
    val compressorV2 = GoBackedBlobCompressor.getInstance(v2, DATA_LIMIT)
    GoBackedBlobCompressor.getProcessIsolatedInstance(
      v2,
      DATA_LIMIT,
    ).use { compressorV2A ->
      GoBackedBlobCompressor.getProcessIsolatedInstance(
        v2,
        DATA_LIMIT,
      ).use { compressorV2B ->
        assertTrue(compressorV2A != compressorV2B)

        val sampleBlocks = CompressorTestData.blocksRlpEncoded
        assertThat(sampleBlocks).isNotEmpty
        val compressedV2 = compressBlocks(compressorV2, sampleBlocks)

        assertThat(compressedV2).isEqualTo(compressBlocks(compressorV2A, sampleBlocks))
        assertThat(compressedV2).isEqualTo(compressBlocks(compressorV2B, sampleBlocks))

        val compressedV2AParallel: AtomicReference<ByteArray> = AtomicReference(ByteArray(0))
        val compressedV2BParallel: AtomicReference<ByteArray> = AtomicReference(ByteArray(0))

        val errors = CopyOnWriteArrayList<Throwable>()
        val startBarrier = CyclicBarrier(3)
        val doneLatch = CountDownLatch(2)

        val v2AThread = Thread {
          try {
            startBarrier.await()
            compressedV2AParallel.store(compressBlocks(compressorV2A, sampleBlocks))
          } catch (t: Throwable) {
            errors.add(t)
          } finally {
            doneLatch.countDown()
          }
        }

        val v2BThread = Thread {
          try {
            startBarrier.await()
            compressedV2BParallel.store(compressBlocks(compressorV2B, sampleBlocks))
          } catch (t: Throwable) {
            errors.add(t)
          } finally {
            doneLatch.countDown()
          }
        }

        v2AThread.start()
        v2BThread.start()
        startBarrier.await()
        assertTrue(doneLatch.await(10, TimeUnit.SECONDS), "compression threads did not complete in time")

        assertThat(compressedV2AParallel.load()).isEqualTo(compressedV2)
        assertThat(compressedV2BParallel.load()).isEqualTo(compressedV2)
        assertThat(errors).isEmpty()
      }
    }
  }

  @OptIn(ExperimentalAtomicApi::class)
  @Test
  @Disabled(
    "This test currently fails as jna backed compressors are not thread safe. " +
      "This test should be enabled once the underlying compressor is made thread safe.",
  )
  fun `compresses multiple blocks concurrently with two v2 jna backed instances without exception`() {
    val v2 = BlobCompressorVersion.V2
    val compressorV2 = GoBackedBlobCompressor.getInstance(v2, DATA_LIMIT)
    val compressorV2A = GoBackedBlobCompressor.getInstance(
      v2,
      DATA_LIMIT,
    )
    val compressorV2B = GoBackedBlobCompressor.getInstance(
      v2,
      DATA_LIMIT,
    )
    assertTrue(compressorV2A != compressorV2B)

    val sampleBlocks = CompressorTestData.blocksRlpEncoded
    assertThat(sampleBlocks).isNotEmpty
    val compressedV2 = compressBlocks(compressorV2, sampleBlocks)

    assertThat(compressedV2).isEqualTo(compressBlocks(compressorV2A, sampleBlocks))
    assertThat(compressedV2).isEqualTo(compressBlocks(compressorV2B, sampleBlocks))

    val compressedV2AParallel: AtomicReference<ByteArray> = AtomicReference(ByteArray(0))
    val compressedV2BParallel: AtomicReference<ByteArray> = AtomicReference(ByteArray(0))

    val errors = CopyOnWriteArrayList<Throwable>()
    val startBarrier = CyclicBarrier(3)
    val doneLatch = CountDownLatch(2)

    val v2AThread = Thread {
      try {
        startBarrier.await()
        compressedV2AParallel.store(compressBlocks(compressorV2A, sampleBlocks))
      } catch (t: Throwable) {
        errors.add(t)
      } finally {
        doneLatch.countDown()
      }
    }

    val v2BThread = Thread {
      try {
        startBarrier.await()
        compressedV2BParallel.store(compressBlocks(compressorV2B, sampleBlocks))
      } catch (t: Throwable) {
        errors.add(t)
      } finally {
        doneLatch.countDown()
      }
    }

    v2AThread.start()
    v2BThread.start()
    startBarrier.await()
    assertTrue(doneLatch.await(10, TimeUnit.SECONDS), "compression threads did not complete in time")

    assertThat(compressedV2AParallel.load()).isEqualTo(compressedV2)
    assertThat(compressedV2BParallel.load()).isEqualTo(compressedV2)
    assertThat(errors).isEmpty()
  }

  fun compressBlocks(compressor: BlobCompressor, blocks: List<ByteArray>): ByteArray {
    blocks.forEach { block ->
      compressor.startNewBatch()
      compressor.appendBlock(block)
    }
    return compressor.getCompressedDataAndReset()
  }
}

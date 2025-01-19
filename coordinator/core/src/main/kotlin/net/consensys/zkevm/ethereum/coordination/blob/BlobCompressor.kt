package net.consensys.zkevm.ethereum.coordination.blob

import net.consensys.encodeHex
import net.consensys.linea.blob.BlobCompressorVersion
import net.consensys.linea.blob.GoNativeBlobCompressor
import net.consensys.linea.blob.GoNativeBlobCompressorFactory
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.metrics.TimerCapture
import org.apache.logging.log4j.LogManager
import kotlin.random.Random

class BlobCompressionException(message: String) : RuntimeException(message)

interface BlobCompressor {

  /**
   * @Throws(BlobCompressionException::class) when blockRLPEncoded is invalid
   */
  fun canAppendBlock(blockRLPEncoded: ByteArray): Boolean

  /**
   * @Throws(BlobCompressionException::class) when blockRLPEncoded is invalid
   */
  fun appendBlock(blockRLPEncoded: ByteArray): AppendResult

  fun startNewBatch()
  fun getCompressedData(): ByteArray
  fun reset()

  data class AppendResult(
    // returns false if last chunk would go over dataLimit. Does  not append last block.
    val blockAppended: Boolean,
    val compressedSizeBefore: Int,
    // even when block is not appended, compressedSizeAfter should as if it was appended
    val compressedSizeAfter: Int
  )
}

class GoBackedBlobCompressor private constructor(
  internal val goNativeBlobCompressor: GoNativeBlobCompressor,
  private val metricsFacade: MetricsFacade
) : BlobCompressor {

  companion object {
    @Volatile
    private var instance: GoBackedBlobCompressor? = null

    fun getInstance(
      compressorVersion: BlobCompressorVersion = BlobCompressorVersion.V0_1_0,
      dataLimit: UInt,
      metricsFacade: MetricsFacade
    ): GoBackedBlobCompressor {
      if (instance == null) {
        synchronized(this) {
          if (instance == null) {
            val goNativeBlobCompressor = GoNativeBlobCompressorFactory.getInstance(compressorVersion)
            val initialized = goNativeBlobCompressor.Init(
              dataLimit.toInt(),
              GoNativeBlobCompressorFactory.dictionaryPath.toString()
            )
            if (!initialized) {
              throw InstantiationException(goNativeBlobCompressor.Error())
            }
            instance = GoBackedBlobCompressor(goNativeBlobCompressor, metricsFacade)
          } else {
            throw IllegalStateException("Compressor singleton instance already created")
          }
        }
      } else {
        throw IllegalStateException("Compressor singleton instance already created")
      }
      return instance!!
    }
  }

  private val canAppendBlockTimer: TimerCapture<Boolean> = metricsFacade.createSimpleTimer(
    category = LineaMetricsCategory.BLOB,
    name = "go.backed.blob.compressor.can.append.block",
    description = "Time taken to run CanWrite method"
  )
  private val appendBlockTimer: TimerCapture<BlobCompressor.AppendResult> = metricsFacade.createSimpleTimer(
    category = LineaMetricsCategory.BLOB,
    name = "go.backed.blob.compressor.append.block",
    description = "Time taken to run AppendResult method"
  )

  private val log = LogManager.getLogger(GoBackedBlobCompressor::class.java)

  override fun canAppendBlock(blockRLPEncoded: ByteArray): Boolean {
    return canAppendBlockTimer.captureTime {
      goNativeBlobCompressor.CanWrite(blockRLPEncoded, blockRLPEncoded.size)
    }
  }

  override fun appendBlock(blockRLPEncoded: ByteArray): BlobCompressor.AppendResult {
    val compressionSizeBefore = goNativeBlobCompressor.Len()
    val appended = goNativeBlobCompressor.Write(blockRLPEncoded, blockRLPEncoded.size)
    val compressedSizeAfter = goNativeBlobCompressor.Len()
    val compressionRatio = 1.0 - ((compressedSizeAfter - compressionSizeBefore).toDouble() / blockRLPEncoded.size)
    log.trace(
      "block compressed: blockRlpSize={} compressionDataBefore={} compressionDataAfter={} compressionRatio={}",
      blockRLPEncoded.size,
      compressionSizeBefore,
      compressedSizeAfter,
      compressionRatio
    )
    val error = goNativeBlobCompressor.Error()
    if (error != null) {
      log.error("Failure while writing the following RLP encoded block: {}", blockRLPEncoded.encodeHex())
      throw BlobCompressionException(error)
    }
    return appendBlockTimer.captureTime {
      BlobCompressor.AppendResult(appended, compressionSizeBefore, compressedSizeAfter)
    }
  }

  override fun startNewBatch() {
    goNativeBlobCompressor.StartNewBatch()
  }

  override fun getCompressedData(): ByteArray {
    val compressedData = ByteArray(goNativeBlobCompressor.Len())
    goNativeBlobCompressor.Bytes(compressedData)
    return compressedData
  }

  override fun reset() {
    goNativeBlobCompressor.Reset()
  }
}

class FakeBlobCompressor(
  private val dataLimit: Int,
  // 1 means no compression
  // 0.5 means compressed data is 50% smaller than original
  // 1.1 means compressed data is 10% bigger than original
  private val fakeCompressionRatio: Double = 1.0
) : BlobCompressor {
  val log = LogManager.getLogger(FakeBlobCompressor::class.java)

  init {
    require(dataLimit > 0) { "dataLimit must be greater than 0" }
    require(fakeCompressionRatio > 0.0) { "fakeCompressionRatio must be greater than 0" }
  }

  private var compressedData = byteArrayOf()

  override fun canAppendBlock(blockRLPEncoded: ByteArray): Boolean {
    return compressedData.size + blockRLPEncoded.size * fakeCompressionRatio <= dataLimit
  }

  override fun appendBlock(blockRLPEncoded: ByteArray): BlobCompressor.AppendResult {
    return buildAppendResult(blockRLPEncoded)
  }

  private fun buildAppendResult(blockRLPEncoded: ByteArray): BlobCompressor.AppendResult {
    val compressedBlock = Random.nextBytes((blockRLPEncoded.size * fakeCompressionRatio).toInt())
    val compressedSizeBefore = compressedData.size
    val compressedSizeAfter = compressedData.size + compressedBlock.size
    return if (compressedSizeAfter > dataLimit) {
      BlobCompressor.AppendResult(
        false,
        compressedSizeBefore,
        compressedSizeAfter
      )
    } else {
      compressedData += compressedBlock
      BlobCompressor.AppendResult(true, compressedSizeBefore, compressedSizeAfter)
    }
  }

  override fun startNewBatch() {
    // we don't care.
  }

  override fun getCompressedData(): ByteArray {
    return compressedData
  }

  override fun reset() {
    log.trace("resetting to empty state")
    compressedData = byteArrayOf()
  }
}

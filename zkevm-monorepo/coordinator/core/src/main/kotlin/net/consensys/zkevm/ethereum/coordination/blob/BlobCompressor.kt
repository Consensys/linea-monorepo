package net.consensys.zkevm.ethereum.coordination.blob

import net.consensys.encodeHex
import net.consensys.linea.nativecompressor.GoNativeBlobCompressor
import net.consensys.linea.nativecompressor.GoNativeBlobCompressorFactory
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

class GoBackedBlobCompressor private constructor(internal val goNativeBlobCompressor: GoNativeBlobCompressor) :
  BlobCompressor {
  companion object {
    @Volatile
    private var instance: GoBackedBlobCompressor? = null

    fun getInstance(dataLimit: Int): GoBackedBlobCompressor {
      if (instance == null) {
        synchronized(this) {
          if (instance == null) {
            val goNativeBlobCompressor = GoNativeBlobCompressorFactory.getInstance()
            val initialized = goNativeBlobCompressor.Init(dataLimit, GoNativeBlobCompressorFactory.getDictionaryPath())
            if (!initialized) {
              throw InstantiationException(goNativeBlobCompressor.Error())
            }
            instance = GoBackedBlobCompressor(goNativeBlobCompressor)
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

  private val log = LogManager.getLogger(GoBackedBlobCompressor::class.java)

  override fun canAppendBlock(blockRLPEncoded: ByteArray): Boolean {
    return goNativeBlobCompressor.CanWrite(blockRLPEncoded, blockRLPEncoded.size)
  }

  override fun appendBlock(blockRLPEncoded: ByteArray): BlobCompressor.AppendResult {
    val compressionSizeBefore = goNativeBlobCompressor.Len()
    val appended = goNativeBlobCompressor.Write(blockRLPEncoded, blockRLPEncoded.size)
    val compressedSizeAfter = goNativeBlobCompressor.Len()
    log.trace(
      "block compressed: blockRlpSize={} compressionDataBefore={} compressionDataAfter={} compressionRatio={}",
      blockRLPEncoded.size,
      compressionSizeBefore,
      compressedSizeAfter,
      1.0 - ((compressedSizeAfter - compressionSizeBefore).toDouble() / blockRLPEncoded.size)
    )
    val error = goNativeBlobCompressor.Error()
    if (error != null) {
      log.error("Failure while writing the following RLP encoded block: {}", blockRLPEncoded.encodeHex())
      throw BlobCompressionException(error)
    }
    return BlobCompressor.AppendResult(appended, compressionSizeBefore, compressedSizeAfter)
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

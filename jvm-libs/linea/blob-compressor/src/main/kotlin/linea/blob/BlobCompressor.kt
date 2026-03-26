package linea.blob

import linea.kotlin.encodeHex
import org.apache.logging.log4j.LogManager

class BlobCompressionException(message: String) : RuntimeException(message)

interface BlobCompressor {

  val version: BlobCompressorVersion

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

  fun getCompressedDataAndReset(): ByteArray {
    val compressedData = getCompressedData()
    reset()
    return compressedData
  }

  data class AppendResult(
    // returns false if last chunk would go over dataLimit. Does  not append last block.
    val blockAppended: Boolean,
    val compressedSizeBefore: Int,
    // even when block is not appended, compressedSizeAfter should as if it was appended
    val compressedSizeAfter: Int,
  )

  fun compressedSize(data: ByteArray): Int
}

// Private adapter so GoBackedBlobCompressor is insulated from the two native API shapes.
private interface NativeCompressorInstance : AutoCloseable {
  fun reset()
  fun startNewBatch()
  fun write(data: ByteArray, len: Int): Boolean
  fun canWrite(data: ByteArray, len: Int): Boolean
  fun error(): String?
  fun len(): Int
  fun bytes(out: ByteArray)
  fun rawCompressedSize(data: ByteArray, len: Int): Int
}

private class LegacyNativeCompressorInstance(
  private val lib: GoNativeBlobCompressor,
) : NativeCompressorInstance {
  override fun reset() = lib.Reset()
  override fun startNewBatch() = lib.StartNewBatch()
  override fun write(data: ByteArray, len: Int) = lib.Write(data, len)
  override fun canWrite(data: ByteArray, len: Int) = lib.CanWrite(data, len)
  override fun error() = lib.Error()
  override fun len() = lib.Len()
  override fun bytes(out: ByteArray) = lib.Bytes(out)
  override fun rawCompressedSize(data: ByteArray, len: Int) = lib.RawCompressedSize(data, len)
  override fun close() {} // global singleton in the native lib; nothing to free
}

private class HandleNativeCompressorInstance(
  private val lib: GoNativeBlobCompressorJnaLib,
  private val handle: Int,
) : NativeCompressorInstance {
  override fun reset() = lib.Reset(handle)
  override fun startNewBatch() = lib.StartNewBatch(handle)
  override fun write(data: ByteArray, len: Int) = lib.Write(handle, data, len)
  override fun canWrite(data: ByteArray, len: Int) = lib.CanWrite(handle, data, len)
  override fun error() = lib.Error(handle)
  override fun len() = lib.Len(handle)
  override fun bytes(out: ByteArray) = lib.Bytes(handle, out)
  override fun rawCompressedSize(data: ByteArray, len: Int) = lib.RawCompressedSize(handle, data, len)
  override fun close() = lib.Free(handle)
}

class GoBackedBlobCompressor private constructor(
  private val nativeInstance: NativeCompressorInstance,
  override val version: BlobCompressorVersion,
) : BlobCompressor, AutoCloseable {

  companion object {
    @JvmStatic
    fun getInstance(compressorVersion: BlobCompressorVersion, dataLimit: Int): GoBackedBlobCompressor {
      require(dataLimit > 0) { "dataLimit=$dataLimit must be greater than 0" }

      val dictPath = GoNativeBlobCompressorFactory.dictionaryPath.toString()
      val nativeInstance = if (compressorVersion == BlobCompressorVersion.V4) {
        val lib = GoNativeBlobCompressorFactory.getInstance(compressorVersion)
        val handle = lib.Init(dataLimit, dictPath)
        if (handle == -1) throw InstantiationException("Failed to initialize compressor")
        HandleNativeCompressorInstance(lib, handle)
      } else {
        val lib = GoNativeBlobCompressorFactory.getLegacyInstance(compressorVersion)
        if (!lib.Init(dataLimit, dictPath)) throw InstantiationException(lib.Error())
        LegacyNativeCompressorInstance(lib)
      }

      return GoBackedBlobCompressor(nativeInstance, compressorVersion)
    }
  }

  private val log = LogManager.getLogger(GoBackedBlobCompressor::class.java)

  override fun close() = nativeInstance.close()

  override fun canAppendBlock(blockRLPEncoded: ByteArray): Boolean {
    return nativeInstance.canWrite(blockRLPEncoded, blockRLPEncoded.size)
  }

  override fun appendBlock(blockRLPEncoded: ByteArray): BlobCompressor.AppendResult {
    val compressionSizeBefore = nativeInstance.len()
    val appended = nativeInstance.write(blockRLPEncoded, blockRLPEncoded.size)
    val compressedSizeAfter = nativeInstance.len()
    log.trace(
      "block compressed: blockRlpSize={} compressionDataBefore={} compressionDataAfter={} compressionRatio={}",
      blockRLPEncoded.size,
      compressionSizeBefore,
      compressedSizeAfter,
      1.0 - ((compressedSizeAfter - compressionSizeBefore).toDouble() / blockRLPEncoded.size),
    )
    val error = nativeInstance.error()
    if (error != null) {
      log.error("Failure while writing the following RLP encoded block: {}", blockRLPEncoded.encodeHex())
      throw BlobCompressionException(error)
    }
    return BlobCompressor.AppendResult(appended, compressionSizeBefore, compressedSizeAfter)
  }

  override fun startNewBatch() {
    nativeInstance.startNewBatch()
  }

  override fun getCompressedData(): ByteArray {
    val compressedData = ByteArray(nativeInstance.len())
    nativeInstance.bytes(compressedData)
    return compressedData
  }

  override fun reset() {
    nativeInstance.reset()
  }

  override fun compressedSize(data: ByteArray): Int {
    return nativeInstance.rawCompressedSize(data, data.size)
  }
}

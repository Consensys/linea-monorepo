package linea.blob

import linea.kotlin.encodeHex
import org.apache.logging.log4j.LogManager

class TxCompressionException(message: String) : RuntimeException(message)

/**
 * Interface for transaction-level compression.
 *
 * This compressor operates at the transaction level rather than the block level,
 * maintaining compression context across transactions for better compression ratios.
 * It is designed for sequencer block building where transactions are added one by one
 * until the compressed size threshold is reached.
 */
interface TxCompressor {

  /**
   * Checks if a transaction can be appended without actually appending it.
   *
   * @param from sender address (20 bytes)
   * @param rlpEncodedTxForSigning RLP-encoded transaction for signing (without signature)
   * @return true if the transaction could be appended, false if it would exceed the limit
   * @throws TxCompressionException if the transaction is invalid
   */
  fun canAppendTransaction(from: ByteArray, rlpEncodedTxForSigning: ByteArray): Boolean

  /**
   * Appends a transaction to the compressed data.
   *
   * @param from sender address (20 bytes)
   * @param rlpEncodedTxForSigning RLP-encoded transaction for signing (without signature)
   * @return AppendResult containing whether the transaction was appended and size information
   * @throws TxCompressionException if the transaction is invalid
   */
  fun appendTransaction(from: ByteArray, rlpEncodedTxForSigning: ByteArray): AppendResult

  /**
   * Returns the current compressed size in bytes (raw LZSS, before field-element packing).
   */
  fun getCompressedSize(): Int

  /**
   * Returns the field-element-packed size of the current compressed data.
   * This is what TxCompressor checks against the blob limit internally, and is directly
   * comparable to the blob size produced by BlobMaker.
   */
  fun getPackedSize(): Int

  /**
   * Returns the number of uncompressed bytes written.
   */
  fun getUncompressedSize(): Int

  /**
   * Returns the compressed data.
   */
  fun getCompressedData(): ByteArray

  /**
   * Resets the compressor to its initial state.
   * Must be called between each block being built.
   */
  fun reset()

  /**
   * Returns the compressed data and resets the compressor.
   */
  fun getCompressedDataAndReset(): ByteArray {
    val compressedData = getCompressedData()
    reset()
    return compressedData
  }

  /**
   * Compresses the (raw) input statelessly and returns the length of the compressed data.
   * The returned length accounts for the "padding" used to fit the data in field elements.
   * Input size must be less than 256kB.
   *
   * This function is stateless and does not affect the compressor's internal state.
   * It is useful for estimating the compressed size of a transaction for profitability calculations.
   *
   * @param data bytes to compress (should be from + rlpEncodedTxForSigning)
   * @return compressed size in bytes, or -1 if an error occurred
   */
  fun compressedSize(data: ByteArray): Int

  data class AppendResult(
    /** Whether the transaction was appended (false if it would exceed the limit) */
    val txAppended: Boolean,
    /** Compressed size before attempting to append */
    val compressedSizeBefore: Int,
    /** Compressed size after (same as before if not appended) */
    val compressedSizeAfter: Int,
  )
}

/**
 * Go-backed implementation of TxCompressor using JNA bindings.
 */
class GoBackedTxCompressor private constructor(
  internal val goNativeTxCompressor: GoNativeTxCompressor,
) : TxCompressor {

  companion object {
    /**
     * Gets a TxCompressor instance with recompression enabled (better compression, slower).
     */
    @JvmStatic
    fun getInstance(compressorVersion: TxCompressorVersion, dataLimit: Int): GoBackedTxCompressor {
      return getInstance(compressorVersion, dataLimit, enableRecompress = true)
    }

    /**
     * Gets a TxCompressor instance with configurable recompression.
     *
     * @param compressorVersion the compressor version to use
     * @param dataLimit maximum compressed size in bytes
     * @param enableRecompress whether to attempt recompression when incremental compression
     *                         exceeds the limit. Set to false for faster operation at the cost
     *                         of slightly worse compression ratios.
     */
    @JvmStatic
    fun getInstance(
      compressorVersion: TxCompressorVersion,
      dataLimit: Int,
      enableRecompress: Boolean,
    ): GoBackedTxCompressor {
      require(dataLimit > 0) { "dataLimit=$dataLimit must be greater than 0" }

      val goNativeTxCompressor = GoNativeTxCompressorFactory.getInstance(compressorVersion)
      val initialized = goNativeTxCompressor.TxInit(
        dataLimit = dataLimit,
        dictPath = GoNativeTxCompressorFactory.dictionaryPath.toString(),
        enableRecompress = enableRecompress,
      )
      if (!initialized) {
        throw InstantiationException(goNativeTxCompressor.TxError())
      }
      return GoBackedTxCompressor(goNativeTxCompressor)
    }
  }

  private val log = LogManager.getLogger(GoBackedTxCompressor::class.java)

  private fun encodeTxData(from: ByteArray, rlpEncodedTxForSigning: ByteArray): ByteArray {
    require(from.size == 20) { "from address must be 20 bytes, got ${from.size}" }
    val txData = ByteArray(from.size + rlpEncodedTxForSigning.size)
    System.arraycopy(from, 0, txData, 0, from.size)
    System.arraycopy(rlpEncodedTxForSigning, 0, txData, from.size, rlpEncodedTxForSigning.size)
    return txData
  }

  private fun checkAndThrowOnError(from: ByteArray, rlpEncodedTxForSigning: ByteArray, operation: String) {
    val error = goNativeTxCompressor.TxError()
    if (error != null) {
      log.error("Failure while {} transaction: from={} rlp={}", operation, from.encodeHex(), rlpEncodedTxForSigning.encodeHex())
      throw TxCompressionException(error)
    }
  }

  override fun canAppendTransaction(from: ByteArray, rlpEncodedTxForSigning: ByteArray): Boolean {
    val txData = encodeTxData(from, rlpEncodedTxForSigning)
    val canWrite = goNativeTxCompressor.TxCanWriteRaw(txData, txData.size)
    checkAndThrowOnError(from, rlpEncodedTxForSigning, "checking")
    return canWrite
  }

  override fun appendTransaction(from: ByteArray, rlpEncodedTxForSigning: ByteArray): TxCompressor.AppendResult {
    val txData = encodeTxData(from, rlpEncodedTxForSigning)
    val compressionSizeBefore = goNativeTxCompressor.TxLen()
    val appended = goNativeTxCompressor.TxWriteRaw(txData, txData.size)
    val compressedSizeAfter = goNativeTxCompressor.TxLen()

    log.trace(
      "transaction compressed: txDataSize={} compressionDataBefore={} compressionDataAfter={} compressionRatio={}",
      txData.size,
      compressionSizeBefore,
      compressedSizeAfter,
      1.0 - ((compressedSizeAfter - compressionSizeBefore).toDouble() / txData.size),
    )

    checkAndThrowOnError(from, rlpEncodedTxForSigning, "writing")
    return TxCompressor.AppendResult(appended, compressionSizeBefore, compressedSizeAfter)
  }

  override fun getCompressedSize(): Int {
    return goNativeTxCompressor.TxLen()
  }

  override fun getPackedSize(): Int {
    return goNativeTxCompressor.TxPackedLen()
  }

  override fun getUncompressedSize(): Int {
    return goNativeTxCompressor.TxWritten()
  }

  override fun getCompressedData(): ByteArray {
    val compressedData = ByteArray(goNativeTxCompressor.TxLen())
    goNativeTxCompressor.TxBytes(compressedData)
    return compressedData
  }

  override fun reset() {
    goNativeTxCompressor.TxReset()
  }

  override fun compressedSize(data: ByteArray): Int {
    return goNativeTxCompressor.TxRawCompressedSize(data, data.size)
  }
}

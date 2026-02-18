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
   * Checks if an RLP-encoded transaction can be appended without actually appending it.
   *
   * @param rlpEncodedTx RLP-encoded transaction bytes
   * @return true if the transaction could be appended, false if it would exceed the limit
   * @throws TxCompressionException if the transaction is invalid
   */
  fun canAppendTransaction(rlpEncodedTx: ByteArray): Boolean

  /**
   * Appends an RLP-encoded transaction to the compressed data.
   *
   * @param rlpEncodedTx RLP-encoded transaction bytes
   * @return AppendResult containing whether the transaction was appended and size information
   * @throws TxCompressionException if the transaction is invalid
   */
  fun appendTransaction(rlpEncodedTx: ByteArray): AppendResult

  /**
   * Returns the current compressed size in bytes.
   */
  fun getCompressedSize(): Int

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
   * @param data bytes to compress
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
    @JvmStatic
    fun getInstance(compressorVersion: TxCompressorVersion, dataLimit: Int): GoBackedTxCompressor {
      require(dataLimit > 0) { "dataLimit=$dataLimit must be greater than 0" }

      val goNativeTxCompressor = GoNativeTxCompressorFactory.getInstance(compressorVersion)
      val initialized = goNativeTxCompressor.TxInit(
        dataLimit = dataLimit,
        dictPath = GoNativeTxCompressorFactory.dictionaryPath.toString(),
      )
      if (!initialized) {
        throw InstantiationException(goNativeTxCompressor.TxError())
      }
      return GoBackedTxCompressor(goNativeTxCompressor)
    }
  }

  private val log = LogManager.getLogger(GoBackedTxCompressor::class.java)

  override fun canAppendTransaction(rlpEncodedTx: ByteArray): Boolean {
    val canWrite = goNativeTxCompressor.TxCanWrite(rlpEncodedTx, rlpEncodedTx.size)
    val error = goNativeTxCompressor.TxError()
    if (error != null) {
      log.error("Failure while checking transaction: {}", rlpEncodedTx.encodeHex())
      throw TxCompressionException(error)
    }
    return canWrite
  }

  override fun appendTransaction(rlpEncodedTx: ByteArray): TxCompressor.AppendResult {
    val compressionSizeBefore = goNativeTxCompressor.TxLen()
    val appended = goNativeTxCompressor.TxWrite(rlpEncodedTx, rlpEncodedTx.size)
    val compressedSizeAfter = goNativeTxCompressor.TxLen()

    log.trace(
      "transaction compressed: txRlpSize={} compressionDataBefore={} compressionDataAfter={} compressionRatio={}",
      rlpEncodedTx.size,
      compressionSizeBefore,
      compressedSizeAfter,
      1.0 - ((compressedSizeAfter - compressionSizeBefore).toDouble() / rlpEncodedTx.size),
    )

    val error = goNativeTxCompressor.TxError()
    if (error != null) {
      log.error("Failure while writing the following RLP encoded transaction: {}", rlpEncodedTx.encodeHex())
      throw TxCompressionException(error)
    }

    return TxCompressor.AppendResult(appended, compressionSizeBefore, compressedSizeAfter)
  }

  override fun getCompressedSize(): Int {
    return goNativeTxCompressor.TxLen()
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

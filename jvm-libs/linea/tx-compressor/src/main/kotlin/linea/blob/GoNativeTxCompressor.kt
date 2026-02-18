package linea.blob

import com.sun.jna.Library
import com.sun.jna.Native
import linea.jvm.ResourcesUtil.copyResourceToTmpDir

/**
 * JNA interface for the Go native transaction compressor library.
 *
 * This compressor operates at the transaction level rather than the block level,
 * maintaining compression context across transactions for better compression ratios.
 * It is designed for sequencer block building where transactions are added one by one
 * until the compressed size threshold is reached.
 */
interface GoNativeTxCompressor {

  /**
   * TxInit initializes the transaction compressor.
   *
   * @param dataLimit Maximum size of compressed data in bytes. The caller should
   *                  account for blob overhead (~100 bytes) when setting this limit.
   * @param dictPath Path to the compression dictionary
   * @return true if the compressor was successfully initialized, false otherwise
   */
  fun TxInit(dataLimit: Int, dictPath: String): Boolean

  /**
   * TxReset resets the compressor to its initial state.
   * Must be called between each block being built.
   */
  fun TxReset()

  /**
   * TxWrite appends an RLP-encoded transaction to the compressed data.
   *
   * @param data bytes of the RLP encoded transaction
   * @param data_len number of bytes
   * @return true if the transaction was appended, false if it would exceed the limit
   *         or if an error occurred (check TxError() for details)
   */
  fun TxWrite(data: ByteArray, data_len: Int): Boolean

  /**
   * TxCanWrite checks if an RLP-encoded transaction can be appended without actually appending it.
   *
   * @param data bytes of the RLP encoded transaction
   * @param data_len number of bytes
   * @return true if the transaction could be appended, false otherwise
   */
  fun TxCanWrite(data: ByteArray, data_len: Int): Boolean

  /**
   * TxLen returns the current length of the compressed data.
   *
   * @return number of bytes of compressed data
   */
  fun TxLen(): Int

  /**
   * TxWritten returns the number of uncompressed bytes written to the compressor.
   *
   * @return number of uncompressed bytes written
   */
  fun TxWritten(): Int

  /**
   * TxBytes fills out with the compressed data.
   * The caller must allocate out and ensure that len(out) == TxLen()
   *
   * @param out The ByteArray to be filled with compressed data
   */
  fun TxBytes(out: ByteArray)

  /**
   * TxError returns the last error message.
   * Should be checked if TxWrite or TxCanWrite returns false.
   *
   * @return error message string, or null if no error
   */
  fun TxError(): String?
}

interface GoNativeTxCompressorJnaLib : GoNativeTxCompressor, Library

enum class TxCompressorVersion(val version: String) {
  V1("v1.0.0"),
}

class GoNativeTxCompressorFactory {
  companion object {
    private const val DICTIONARY_NAME = "compressor-dictionaries/v2025-04-21.bin"
    val dictionaryPath =
      copyResourceToTmpDir(DICTIONARY_NAME, GoNativeTxCompressorFactory::class.java.classLoader)

    private fun getLibFileName(version: String) = "tx_compressor_jna_$version"

    @JvmStatic
    private val loadedVersions = mutableMapOf<TxCompressorVersion, GoNativeTxCompressor>()

    @JvmStatic
    fun getInstance(version: TxCompressorVersion): GoNativeTxCompressor {
      synchronized(loadedVersions) {
        return loadedVersions[version]
          ?: loadLib(version)
            .also { loadedVersions[version] = it }
      }
    }

    private fun loadLib(version: TxCompressorVersion): GoNativeTxCompressor {
      val extractedLibFile = Native.extractFromResourcePath(
        getLibFileName(version.version),
        GoNativeTxCompressorFactory::class.java.classLoader,
      )

      return Native.load(
        /* name = */
        extractedLibFile.toString(),
        /* interfaceClass = */
        GoNativeTxCompressorJnaLib::class.java,
      )
    }
  }
}

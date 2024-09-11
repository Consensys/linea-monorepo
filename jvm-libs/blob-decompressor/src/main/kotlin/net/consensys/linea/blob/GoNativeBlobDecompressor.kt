package net.consensys.linea.blob

import com.sun.jna.Library
import com.sun.jna.Native
import net.consensys.jvm.ResourcesUtil.copyResourceToTmpDir

interface GoNativeBlobDecompressor {

  /**
   * Init initializes the Decompressor.
   */
  fun Init()

  /**
   * LoadDictionary attempts to cache the dictionary from the given path. Returns
   * true if the dictionary is successfully loaded, false if not, in which case Error() will
   * return a description of the error.
   *
   * @param data  dictPath path to the dictionary file
   * @return true if loading was successful else false
   */
  fun LoadDictionary(dictPath: StringArray, nbDicts: Int) Boolean

  /**
   * Decompress a blob b and writes the resulting blocks in out, serialized in the format of
   * prover/backend/ethereum.
   * Returns the number of bytes in out, or -1 in case of failure
   * If -1 is returned, the Error() method will return a string describing the error.
   *
   * @param blob to be decompressed
   * @param blob_len length of the blob
   * @param out buffer to write the decompressed data
    * @param out_max_len maximum length of the out buffer
   * @return number of bytes in out, or -1 in case of failure
   */
  fun Decompress(blob: ByteArray, blob_len: Int, out: ByteArray, out_max_len: Int): Int

  /**
   * Error returns the last error message. Should be checked if Write returns false.
   */
  fun Error(): String?
}

interface GoNativeBlobDecompressorJnaLib : GoNativeBlobDecompressor, Library

enum class BlobDecompressorVersion(val version: String) {
  V1_0_1("v1.0.1")
}

class GoNativeBlobDecompressorFactory {
  companion object {
    private const val DICTIONARY_NAME = "Decompressor_dict.bin"
    val dictionaryPath = copyResourceToTmpDir(DICTIONARY_NAME)

    private fun getLibFileName(version: String) = "blob_decompressor_jna_$version"

    fun getInstance(
      version: BlobDecompressorVersion
    ): GoNativeBlobDecompressor {
      return Native.load(
        Native.extractFromResourcePath(getLibFileName(version.version)).toString(),
        GoNativeBlobDecompressorJnaLib::class.java
      )
    }
  }
}

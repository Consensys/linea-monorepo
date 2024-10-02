package net.consensys.linea.blob

import com.sun.jna.Library
import com.sun.jna.Native
import net.consensys.jvm.ResourcesUtil.copyResourceToTmpDir
import java.nio.file.Path

class DecompressionException(message: String) : RuntimeException(message)

interface BlobDecompressor {
  fun decompress(blob: ByteArray, out: ByteArray): Int
}

internal class Adapter(
  private val delegate: GoNativeBlobDecompressorJnaBinding,
  dictionaries: List<Path>
) : BlobDecompressor {
  init {
    delegate.Init()

    val paths = 
      dictionaries.map{path -> path.toString()}.  // TODO more concise idiom?
      joinToString(":")

    if (delegate.LoadDictionaries(paths) != dictionaries.size) {
      throw DecompressionException("Failed to load dictionaries '${paths}', error='${delegate.Error()}'")
    }
  }

  override fun decompress(blob: ByteArray, out: ByteArray): Int {
    val decompressedSize = delegate.Decompress(blob, blob.size, out, out.size)
    if (decompressedSize < 0) {
      throw DecompressionException("Decompression failed, error='${delegate.Error()}'")
    }
    return decompressedSize
  }
}

internal interface GoNativeBlobDecompressorJnaBinding {

  /**
   * Init initializes the Decompressor. Must be run before anything else.
   */
  fun Init()

  /**
   * LoadDictionaries attempts to cache dictionaries from given paths, separated by colons,
   * e.g. "../compressor_dict.bin:./other_dict"
   * Returns the number of dictionaries successfully loaded, and -1 in case of failure, in which case Error() will
   * return a description of the error.
   *
   * @param dictPaths a colon-separated list of paths to dictionaries, to be loaded into the decompressor
   * @return the number of dictionaries loaded if successful, -1 if not.
   */
  fun LoadDictionaries(dictPaths: String): Int

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

internal interface GoNativeBlobDecompressorJnaLib : GoNativeBlobDecompressorJnaBinding, Library

enum class BlobDecompressorVersion(val version: String) {
  V1_1_0("v1.1.0")
}

class GoNativeBlobDecompressorFactory {
  companion object {
    private const val DICTIONARY_NAME = "compressor_dict.bin"
    val dictionaryPath = copyResourceToTmpDir(DICTIONARY_NAME)

    private fun getLibFileName(version: String) = "blob_decompressor_jna_$version"

    fun getInstance(
      version: BlobDecompressorVersion
    ): BlobDecompressor {
      return Native.load(
        Native.extractFromResourcePath(getLibFileName(version.version)).toString(),
        GoNativeBlobDecompressorJnaLib::class.java
      ).let {
        Adapter(it, listOf(dictionaryPath))
      }
    }
  }
}

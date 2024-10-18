package build.linea.staterecover.core

interface BlobDecompressorToDomain {
  fun decompress(blobs: List<ByteArray>): List<BlockL1RecoveredData>
  fun decompress(blob: ByteArray): List<BlockL1RecoveredData>
}

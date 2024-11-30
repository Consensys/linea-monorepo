package linea.staterecover.clients.blobscan

import linea.staterecover.BlobFetcher
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class BlobInfo(
  val startBlockNumber: ULong,
  val endBlockNumber: ULong,
  val blob: ByteArray
)

class FakeBlobFetcher(
  val blobRecords: List<BlobInfo>
) : BlobFetcher {
  var cursor: Int = 0

  fun moveCursorToBlob(
    startBlockNumber: ULong
  ) {
    val index = blobRecords.indexOfFirst { it.startBlockNumber == startBlockNumber }
    if (index == -1) {
      throw IllegalArgumentException("Blob not found for startBlockNumber: $startBlockNumber")
    }

    cursor = index
  }

  override fun fetchBlobsByHash(blobVersionedHashes: List<ByteArray>): SafeFuture<List<ByteArray>> {
    val blobs = blobRecords
      .subList(cursor, cursor + blobVersionedHashes.size)
      .map { it.blob }

    cursor += blobVersionedHashes.size
    return SafeFuture.completedFuture(blobs)
  }
}

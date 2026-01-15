package net.consensys.zkevm.ethereum.coordination.blob

import linea.domain.BlockInterval
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class ParentBlobData(
  val endBlockNumber: ULong,
  val blobHash: ByteArray,
  val blobShnarf: ByteArray,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ParentBlobData

    if (endBlockNumber != other.endBlockNumber) return false
    if (!blobHash.contentEquals(other.blobHash)) return false
    if (!blobShnarf.contentEquals(other.blobShnarf)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = endBlockNumber.hashCode()
    result = 31 * result + blobHash.contentHashCode()
    result = 31 * result + blobShnarf.contentHashCode()
    return result
  }
}

interface ParentBlobDataProvider {
  fun getParentBlobData(currentBlobRange: BlockInterval): SafeFuture<ParentBlobData>
}

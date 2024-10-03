package net.consensys.linea.staterecover.core

import net.consensys.linea.staterecover.domain.BlockL1RecoveredData
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class ImportResult(
  val blockNumber: ULong,
  val zkStateRootHash: ByteArray
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ImportResult

    if (blockNumber != other.blockNumber) return false
    if (!zkStateRootHash.contentEquals(other.zkStateRootHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = blockNumber.hashCode()
    result = 31 * result + zkStateRootHash.contentHashCode()
    return result
  }
}

interface BlockImporterAndStateVerifier {
  fun importBlocks(blocks: List<BlockL1RecoveredData>): SafeFuture<ImportResult>
}

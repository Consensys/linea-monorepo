package net.consensys.zkevm.ethereum.coordination.blockcreation

import kotlinx.datetime.Instant
import net.consensys.zkevm.toULong
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class BlockHeaderSummary(
  val number: ULong,
  val hash: ByteArray,
  val timestamp: Instant
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BlockHeaderSummary

    if (number != other.number) return false
    if (!hash.contentEquals(other.hash)) return false
    if (timestamp != other.timestamp) return false

    return true
  }

  override fun hashCode(): Int {
    var result = number.hashCode()
    result = 31 * result + hash.contentHashCode()
    result = 31 * result + timestamp.hashCode()
    return result
  }

  override fun toString(): String {
    return "BlockHeaderSummary(number=$number, hash=${hash.contentToString()}, timestamp=$timestamp)"
  }
}

interface SafeBlockProvider {
  fun getLatestSafeBlock(): SafeFuture<ExecutionPayloadV1>
  fun getLatestSafeBlockHeader(): SafeFuture<BlockHeaderSummary> {
    return getLatestSafeBlock().thenApply {
      BlockHeaderSummary(
        it.blockNumber.toULong(),
        it.blockHash.toArray(),
        Instant.fromEpochSeconds(it.timestamp.longValue())
      )
    }
  }
}

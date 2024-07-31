package net.consensys.zkevm.ethereum.coordination.blockcreation

import kotlinx.datetime.Instant
import net.consensys.zkevm.toULong
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class BlockHeaderSummary(
  val number: ULong,
  val hash: Bytes32,
  val timestamp: Instant
)

interface SafeBlockProvider {
  fun getLatestSafeBlock(): SafeFuture<ExecutionPayloadV1>
  fun getLatestSafeBlockHeader(): SafeFuture<BlockHeaderSummary> {
    return getLatestSafeBlock().thenApply {
      BlockHeaderSummary(
        it.blockNumber.toULong(),
        it.blockHash,
        Instant.fromEpochSeconds(it.timestamp.longValue())
      )
    }
  }
}

package net.consensys.zkevm.ethereum.coordination.blockcreation

import kotlinx.datetime.Instant
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class BlockCreated(
  // val tracesCounters: TracesCounters,
  val executionPayload: ExecutionPayloadV1
// val forkChoiceStateV1: ForkChoiceStateV1
)

interface BlockCreationCoordinator {
  fun createBlock(timestamp: Instant, slotNumber: ULong): SafeFuture<BlockCreated>
  fun finalizeBlock(block: BlockHeightAndHash): SafeFuture<Unit>
}

fun interface BlockCreationListener {
  fun acceptBlock(blockEvent: BlockCreated): SafeFuture<Unit>
}

interface BlockEmitter {
  fun onBlockCreated(listenerName: String, listener: BlockCreationListener)
}

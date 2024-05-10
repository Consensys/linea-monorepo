package net.consensys.zkevm.ethereum.coordination.blockcreation

import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class BlockCreated(
  val executionPayload: ExecutionPayloadV1
)

fun interface BlockCreationListener {
  fun acceptBlock(blockEvent: BlockCreated): SafeFuture<Unit>
}

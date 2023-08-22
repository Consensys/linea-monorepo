package net.consensys.zkevm.coordinator.clients

import org.apache.tuweni.units.bigints.UInt64
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture

class PostgresSqlBlocksStore() : BlocksStore<ExecutionPayloadV1> {
  override fun saveBlock(block: ExecutionPayloadV1): SafeFuture<*> {
    TODO("Not yet implemented")
  }

  override fun findBlockByHeight(blockNumber: UInt64): SafeFuture<ExecutionPayloadV1?> {
    TODO("Not yet implemented")
  }

  override fun findHighestBlock(blockNumber: UInt64): SafeFuture<ExecutionPayloadV1?> {
    TODO("Not yet implemented")
  }
}

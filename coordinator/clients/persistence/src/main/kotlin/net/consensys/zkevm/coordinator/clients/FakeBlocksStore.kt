package net.consensys.zkevm.coordinator.clients

import org.apache.tuweni.units.bigints.UInt64
import tech.pegasys.teku.infrastructure.async.SafeFuture

class FakeBlocksStore<T>() : BlocksStore<T> {
  override fun saveBlock(block: T): SafeFuture<*> {
    return SafeFuture.completedFuture(null)
  }

  override fun findBlockByHeight(blockNumber: UInt64): SafeFuture<T?> {
    return SafeFuture.completedFuture(null)
  }

  override fun findHighestBlock(blockNumber: UInt64): SafeFuture<T?> {
    return SafeFuture.completedFuture(null)
  }
}

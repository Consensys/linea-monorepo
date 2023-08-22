package net.consensys.zkevm.coordinator.clients

import org.apache.tuweni.units.bigints.UInt64
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface BlocksStore<T> {
  fun saveBlock(block: T): SafeFuture<*>
  fun findBlockByHeight(blockNumber: UInt64): SafeFuture<T?>
  fun findHighestBlock(blockNumber: UInt64): SafeFuture<T?>
}

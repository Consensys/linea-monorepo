package net.consensys.zkevm.ethereum.finalization

import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

interface FinalizationMonitor {
  data class FinalizationUpdate(
    val zkStateRootHash: Bytes32,
    val blockNumber: UInt64,
    val blockHash: Bytes32
  )

  fun getLastFinalizationUpdate(): FinalizationUpdate
  fun addFinalizationHandler(handlerName: String, handler: (FinalizationUpdate) -> SafeFuture<*>)
  fun removeFinalizationHandler(handlerName: String)
}

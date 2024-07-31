package net.consensys.zkevm.coordinator.clients.smartcontract

import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface ContractVersionProvider<T> {
  fun getVersion(): SafeFuture<T>
}

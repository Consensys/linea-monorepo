package net.consensys.zkevm.ethereum.coordination

import tech.pegasys.teku.infrastructure.async.SafeFuture

class SimpleCompositeSafeFutureHandler<T>(
  private val handlers: List<(T) -> SafeFuture<*>>
) : (T) -> SafeFuture<*> {
  override fun invoke(arg: T): SafeFuture<Unit> {
    val handlingFutures = handlers.map { it.invoke(arg) }
    return SafeFuture.allOf(*handlingFutures.toTypedArray()).thenApply { }
  }
}

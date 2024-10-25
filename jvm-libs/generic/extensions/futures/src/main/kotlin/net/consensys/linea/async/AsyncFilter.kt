package net.consensys.linea.async

import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface AsyncFilter<T> : (List<T>) -> SafeFuture<List<T>> {
  fun then(next: AsyncFilter<T>): AsyncFilter<T> = AsyncFilter { items ->
    this(items).thenCompose(next)
  }

  companion object {
    fun <T> NoOp(): AsyncFilter<T> = AsyncFilter { SafeFuture.completedFuture(it) }
  }
}

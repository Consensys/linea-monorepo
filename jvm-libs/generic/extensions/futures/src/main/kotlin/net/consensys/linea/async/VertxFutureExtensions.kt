package net.consensys.linea.async

import io.vertx.core.Future
import io.vertx.core.Promise
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun <T> Future<T>.get(): T = this.toCompletionStage().toCompletableFuture().get()

fun <T> Future<T>.toSafeFuture(): SafeFuture<T> = SafeFuture.of(this.toCompletionStage())

fun <T> SafeFuture<T>.toVertxFuture(): Future<T> {
  val result = Promise.promise<T>()
  this.thenAccept(result::complete)
  this.handleException(result::fail)
  return result.future()
}

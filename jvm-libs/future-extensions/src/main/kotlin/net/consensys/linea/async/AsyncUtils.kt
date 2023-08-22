package net.consensys.linea.async

import io.vertx.core.Vertx
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds

class RetriedExecutionException(override val message: String) : RuntimeException(message)

fun <T> retry(retries: Int, action: () -> SafeFuture<T>): SafeFuture<T> {
  return action().exceptionallyCompose { throwable ->
    when {
      retries > 0 -> retry(retries - 1, action)
      else -> SafeFuture.failedFuture(throwable)
    }
  }
}

private fun <T> alwaysTruePredicate(@Suppress("UNUSED_PARAMETER") value: T) = true

fun <T> retryWithInterval(
  retries: Int,
  interval: Duration,
  vertx: Vertx,
  action: () -> SafeFuture<T>
): SafeFuture<T> = retryWithInterval(retries, retries, 0.milliseconds, interval, vertx, ::alwaysTruePredicate, action)

fun <T> retryWithInterval(
  retries: Int,
  initialDelay: Duration,
  interval: Duration,
  vertx: Vertx,
  action: () -> SafeFuture<T>
): SafeFuture<T> = retryWithInterval(retries, retries, initialDelay, interval, vertx, ::alwaysTruePredicate, action)

fun <T> retryWithInterval(
  retries: Int,
  interval: Duration,
  vertx: Vertx,
  stopRetriesPredicate: (T) -> Boolean = ::alwaysTruePredicate,
  action: () -> SafeFuture<T>
): SafeFuture<T> = retryWithInterval(retries, retries, 0.milliseconds, interval, vertx, stopRetriesPredicate, action)

fun <T> retryWithInterval(
  retries: Int,
  initialDelay: Duration,
  interval: Duration,
  vertx: Vertx,
  stopRetriesPredicate: (T) -> Boolean = ::alwaysTruePredicate,
  action: () -> SafeFuture<T>
): SafeFuture<T> = retryWithInterval(retries, retries, initialDelay, interval, vertx, stopRetriesPredicate, action)

private fun <T> retryWithInterval(
  maxRetries: Int,
  remainingRetries: Int,
  initialDelay: Duration,
  interval: Duration,
  vertx: Vertx,
  stopRetriesPredicate: (T) -> Boolean = ::alwaysTruePredicate,
  action: () -> SafeFuture<T>
): SafeFuture<T> {
  require(maxRetries >= 0) { "maxRetries must be >= 0. value=$maxRetries" }
  require(initialDelay >= 0.milliseconds) { "initialDelay must be >= 0ms. value=$initialDelay" }
  require(interval > 0.milliseconds) { "interval must be > 0ms. value=$interval" }

  if (initialDelay > 0.milliseconds) {
    return SafeFuture<T>().also { delayedFutureResult ->
      vertx.setTimer(initialDelay.inWholeMilliseconds) {
        retryWithInterval(maxRetries, remainingRetries, 0.milliseconds, interval, vertx, stopRetriesPredicate, action)
          .propagateTo(delayedFutureResult)
      }
    }
  }

  return action().handleComposed { result, throwable ->
    val stopRetrying =
      if (throwable != null) {
        false
      } else {
        stopRetriesPredicate(result)
      }
    if (stopRetrying) {
      SafeFuture.completedFuture(result)
    } else if (remainingRetries > 0) {
      val delayedFutureResult = SafeFuture<T>()
      vertx.setTimer(interval.inWholeMilliseconds) {
        retryWithInterval(
          maxRetries = maxRetries,
          remainingRetries = remainingRetries - 1,
          initialDelay = 0.milliseconds,
          interval = interval,
          vertx = vertx,
          stopRetriesPredicate = stopRetriesPredicate,
          action = action
        )
          .propagateTo(delayedFutureResult)
      }
      delayedFutureResult
    } else {
      if (throwable != null) {
        SafeFuture.failedFuture(throwable)
      } else {
        SafeFuture.failedFuture(
          RetriedExecutionException("Stop condition wasn't met after $maxRetries retries.")
        )
      }
    }
  }
}

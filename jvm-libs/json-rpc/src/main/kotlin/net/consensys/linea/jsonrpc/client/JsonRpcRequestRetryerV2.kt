package net.consensys.linea.jsonrpc.client

import com.fasterxml.jackson.databind.ObjectMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.map
import com.github.michaelbull.result.mapError
import com.github.michaelbull.result.onFailure
import com.github.michaelbull.result.onSuccess
import io.vertx.core.Vertx
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.async.RetriedExecutionException
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.atomic.AtomicReference
import java.util.function.Predicate

class JsonRpcRequestRetryerV2(
  private val vertx: Vertx,
  private val delegate: JsonRpcClient,
  private val requestRetry: RequestRetryConfig,
  private val requestObjectMapper: ObjectMapper = objectMapper,
  private val shallRetryRequestsClientBasePredicate: Predicate<Result<Any?, Throwable>>,
  private val log: Logger = LogManager.getLogger(JsonRpcRequestRetryer::class.java),
  private val failuresLogLevel: Level = Level.WARN
) {
  fun <T> makeRequest(
    request: JsonRpcRequest,
    shallRetryRequestPredicate: Predicate<Result<T, Throwable>>,
    resultMapper: (Any?) -> T
  ): SafeFuture<T> {
    return makeRequestWithRetryer(request, resultMapper, shallRetryRequestPredicate)
  }

  private fun <T> makeRequestWithRetryer(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> T,
    shallRetryRequestPredicate: Predicate<Result<T, Throwable>>
  ): SafeFuture<T> {
    val lastResult = AtomicReference<T>()
    val lastException = AtomicReference<Throwable>()
    val retriesCount = AtomicInteger(0)
    val requestPredicate = Predicate<Result<T, Throwable>> { result ->
      shallRetryRequestsClientBasePredicate.test(result) || shallRetryRequestPredicate.test(result)
    }

    return AsyncRetryer.retry(
      vertx = vertx,
      backoffDelay = requestRetry.backoffDelay,
      maxRetries = requestRetry.maxRetries?.toInt(),
      timeout = requestRetry.timeout,
      stopRetriesPredicate = { result: Result<T, Throwable> ->
        result
          .onSuccess { lastResult.set(it) }
          .onFailure { lastException.set(it) }
        !requestPredicate.test(result)
      }
    ) {
      if (requestRetry.failuresWarningThreshold > 0u &&
        retriesCount.get() > 0 &&
        (retriesCount.get() % requestRetry.failuresWarningThreshold.toInt()) == 0
      ) {
        log.log(
          failuresLogLevel,
          "Request '{}' already retried {} times. lastError={}",
          requestObjectMapper.writeValueAsString(request),
          retriesCount.get(),
          lastException.get()
        )
      }
      retriesCount.incrementAndGet()
      val future: SafeFuture<Result<T, Throwable>> = delegate.makeRequest(request, resultMapper)
        .toSafeFuture()
        .thenApply { unfoldResultValueOrException<T>(it) }
        .exceptionally { th ->
          if (th is Error || th.cause is Error) {
            // Very serious JVM error, we should stop retrying anyway
            throw th
          } else {
            Err(th.cause ?: th)
          }
        }
      future
    }
      .handleComposed { result, throwable ->
        when {
          result is Ok -> SafeFuture.completedFuture(result.value)
          result is Err -> SafeFuture.failedFuture<T>(result.error)
          throwable != null && throwable is RetriedExecutionException -> SafeFuture.failedFuture(lastException.get())
          else -> SafeFuture.failedFuture(throwable)
        }
      }
  }

  companion object {
    fun <T> unfoldResultValueOrException(
      response: Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>
    ): Result<T, Throwable> {
      @Suppress("UNCHECKED_CAST")
      return response
        .map { it.result as T }
        .mapError { it.error.asException() }
    }
  }
}

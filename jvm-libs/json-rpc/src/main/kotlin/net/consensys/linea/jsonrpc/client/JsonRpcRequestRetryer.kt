package net.consensys.linea.jsonrpc.client

import com.fasterxml.jackson.databind.ObjectMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Result
import io.vertx.core.Future
import io.vertx.core.Vertx
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.async.RetriedExecutionException
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.async.toVertxFuture
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds

data class RequestRetryConfig(
  val maxAttempts: UInt? = null,
  val timeout: Duration? = null,
  val backoffDelay: Duration,
  val failuresWarningThreshold: UInt = 0u
) {
  init {
    require(maxAttempts != null || timeout != null) { "maxAttempts or timeout must be specified" }
    maxAttempts?.also {
      require(maxAttempts > 0u) { "maxAttempts must be greater than zero. value=$maxAttempts" }
      require(maxAttempts > failuresWarningThreshold) {
        "maxAttempts must be greater than failuresWarningThreshold." +
          " maxAttempts=$maxAttempts, failuresWarningThreshold=$failuresWarningThreshold"
      }
    }
    timeout?.also {
      require(timeout > 0.milliseconds) { "timeout must be >= 1ms. value=$timeout" }
    }
  }
}

class JsonRpcRequestRetryer(
  private val vertx: Vertx,
  private val delegate: JsonRpcClient,
  private val config: Config,
  private val requestObjectMapper: ObjectMapper = VertxHttpJsonRpcClient.objectMapper,
  private val log: Logger = LogManager.getLogger(JsonRpcRequestRetryer::class.java),
  private val failuresLogLevel: Level = Level.WARN
) : JsonRpcClientWithRetries {

  data class Config(
    val methodsToRetry: Set<String>,
    val requestRetry: RequestRetryConfig
  )

  val retryer: AsyncRetryer<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> = AsyncRetryer.retryer(
    vertx = this.vertx,
    backoffDelay = config.requestRetry.backoffDelay,
    maxRetries = config.requestRetry.maxAttempts?.toInt(),
    timeout = config.requestRetry.timeout
  )

  override fun makeRequest(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> Any?
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    return makeRequest(request, resultMapper, ::isResultOk)
  }

  override fun makeRequest(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> Any?,
    stopRetriesPredicate: (result: Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>) -> Boolean
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    if (request.method in config.methodsToRetry) {
      return makeRequestWithRetryer(request, resultMapper, stopRetriesPredicate)
    } else {
      return delegate.makeRequest(request, resultMapper)
    }
  }

  private fun makeRequestWithRetryer(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> Any?,
    stopRetriesPredicate: (result: Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>) -> Boolean
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    val lastResult = AtomicReference<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>()
    val lastException = AtomicReference<Throwable>()
    val retriesCount = AtomicInteger(0)
    return retryer.retry(
      exceptionConsumer = lastException::set,
      stopRetriesPredicate = { result: Result<JsonRpcSuccessResponse, JsonRpcErrorResponse> ->
        lastResult.set(result)
        stopRetriesPredicate.invoke(result)
      }
    ) {
      if (config.requestRetry.failuresWarningThreshold > 0u &&
        retriesCount.get() > 0 &&
        (retriesCount.get() % config.requestRetry.failuresWarningThreshold.toInt()) == 0
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
      delegate.makeRequest(request, resultMapper).toSafeFuture()
    }
      .exceptionallyCompose { th ->
        log.trace("Got failure: {}", th.message, th)

        if (th is RetriedExecutionException && lastResult.get() is Err) {
          SafeFuture.completedFuture(lastResult.get()!!)
        } else {
          SafeFuture.failedFuture(th)
        }
      }
      .toVertxFuture()
  }
}

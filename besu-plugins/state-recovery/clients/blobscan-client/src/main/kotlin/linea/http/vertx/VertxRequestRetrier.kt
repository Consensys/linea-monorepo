package linea.http.vertx

import io.vertx.core.Vertx
import io.vertx.core.buffer.Buffer
import io.vertx.ext.web.client.HttpRequest
import io.vertx.ext.web.client.HttpResponse
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import tech.pegasys.teku.infrastructure.async.SafeFuture

class VertxRequestRetrier(
  private val vertx: Vertx,
  private val requestRetryConfig: RequestRetryConfig,
  private val requestLogger: VertxRequestLogger,
  private val retryableErrorCodes: Set<Int> = DEFAULT_RETRY_HTTP_CODES,
  private val asyncRetryer: AsyncRetryer<HttpResponse<Buffer>> = AsyncRetryer.retryer(
    backoffDelay = requestRetryConfig.backoffDelay,
    maxRetries = requestRetryConfig.maxRetries?.toInt(),
    timeout = requestRetryConfig.timeout,
    vertx = vertx
  )
) : VertxHttpRequestSender {
  override fun makeRequest(request: HttpRequest<Buffer>): SafeFuture<HttpResponse<Buffer>> {
    return asyncRetryer
      .retry(
        stopRetriesPredicate = { response: HttpResponse<Buffer> ->
          response.statusCode() !in retryableErrorCodes
        }
      ) {
        requestLogger.logRequest(request)
        request.send().toSafeFuture()
          .thenPeek { response -> requestLogger.logResponse(request, response) }
          .whenException { error -> requestLogger.logResponse(request = request, failureCause = error) }
      }
  }

  companion object {
    val DEFAULT_RETRY_HTTP_CODES = setOf(429, 500, 503, 504)
  }
}

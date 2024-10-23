package build.linea.staterecover.clients.blobscan

import io.vertx.core.Vertx
import io.vertx.core.buffer.Buffer
import io.vertx.ext.web.client.HttpRequest
import io.vertx.ext.web.client.HttpResponse
import io.vertx.ext.web.client.WebClient
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import tech.pegasys.teku.infrastructure.async.SafeFuture

// TODO: move to a common module
data class RestResponse<T>(
  val statusCode: Int,
  val body: T?
)

interface RestClient<Response> {
  fun get(path: String): SafeFuture<RestResponse<Response>>
  // add remaining verbs as we need them
}

class VertxRestClient<Response>(
  private val vertx: Vertx,
  private val webClient: WebClient,
  private val responseParser: (Buffer) -> Response,
  private val retryableErrorCodes: Set<Int> = DEFAULT_RETRY_HTTP_CODES,
  private val requestRetryConfig: RequestRetryConfig,
  private val asyncRetryer: AsyncRetryer<HttpResponse<Buffer>> = AsyncRetryer.retryer(
    backoffDelay = requestRetryConfig.backoffDelay,
    maxRetries = requestRetryConfig.maxRetries?.toInt(),
    timeout = requestRetryConfig.timeout,
    vertx = vertx
  ),
  private val requestHeaders: Map<String, String> = mapOf("Accept" to "application/json")
) : RestClient<Response> {
  private fun makeRequestWithRetry(
    request: HttpRequest<Buffer>
  ): SafeFuture<HttpResponse<Buffer>> {
    return asyncRetryer
      .retry(
        stopRetriesPredicate = { response: HttpResponse<Buffer> ->
          response.statusCode() !in retryableErrorCodes
        }
      ) {
        request.send().toSafeFuture()
      }
  }

  override fun get(path: String): SafeFuture<RestResponse<Response>> {
    return makeRequestWithRetry(
      webClient
        .get(path)
        .apply { requestHeaders.forEach(::putHeader) }
    )
      .thenApply { response ->
        val parsedResponse = response.body()?.let(responseParser)
        RestResponse(response.statusCode(), parsedResponse)
      }
  }

  companion object {
    val DEFAULT_RETRY_HTTP_CODES = setOf(429, 500, 503, 504)
  }
}

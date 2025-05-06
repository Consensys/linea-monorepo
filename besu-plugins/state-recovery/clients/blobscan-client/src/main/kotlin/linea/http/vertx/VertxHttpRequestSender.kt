package linea.http.vertx

import io.vertx.core.buffer.Buffer
import io.vertx.ext.web.client.HttpRequest
import io.vertx.ext.web.client.HttpResponse
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface VertxHttpRequestSender {
  fun makeRequest(request: HttpRequest<Buffer>): SafeFuture<HttpResponse<Buffer>>
}

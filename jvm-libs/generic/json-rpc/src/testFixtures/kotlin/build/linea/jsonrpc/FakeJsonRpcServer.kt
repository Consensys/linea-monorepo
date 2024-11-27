package build.linea.jsonrpc

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.DeploymentOptions
import io.vertx.core.Future
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import io.vertx.ext.auth.User
import net.consensys.linea.async.get
import net.consensys.linea.jsonrpc.HttpRequestHandler
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcErrorResponseException
import net.consensys.linea.jsonrpc.JsonRpcMessageHandler
import net.consensys.linea.jsonrpc.JsonRpcMessageProcessor
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcRequestHandler
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.jsonrpc.httpserver.HttpJsonRpcServer
import java.util.concurrent.ConcurrentHashMap

class DynamicRoutingHandler(
  val recordRequestsResponses: Boolean
) : JsonRpcRequestHandler {
  val handlers: MutableMap<String, (JsonRpcRequest) -> Any?> = ConcurrentHashMap()
  var requests: MutableList<Pair<JsonRpcRequest, Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>>> =
    mutableListOf()

  override fun invoke(
    user: User?,
    jsonRpcRequest: JsonRpcRequest,
    requestJson: JsonObject
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    // need this otherwise kotlin compiler/IDE struggle to infer the type
    val result: Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> = (
      handlers[jsonRpcRequest.method]
        ?.let { handler ->
          try {
            Future.succeededFuture(
              Ok(
                JsonRpcSuccessResponse(
                  jsonrpc = "2.0",
                  id = jsonRpcRequest.id,
                  result = handler(jsonRpcRequest)
                )
              )
            )
          } catch (e: JsonRpcErrorResponseException) {
            Future.succeededFuture(Err(JsonRpcErrorResponse(jsonRpcRequest.id, e.asJsonRpcError())))
          } catch (e: Exception) {
            Future.succeededFuture(Err(JsonRpcErrorResponse.internalError(jsonRpcRequest.id, data = e.message)))
          }
        }
        ?: Future.succeededFuture(Err(JsonRpcErrorResponse.methodNotFound(jsonRpcRequest.id, jsonRpcRequest.method)))
      )

    return result
      .also {
        if (recordRequestsResponses) {
          requests.add(jsonRpcRequest to it)
        }
      }
  }
}

class FakeJsonRpcServer(
  port: Int = 0,
  apiPath: String = "/",
  recordRequestsResponses: Boolean = false,
  val vertx: Vertx = Vertx.vertx()
) {
  val meterRegistry = SimpleMeterRegistry()
  val jsonRpcRequestHandler = DynamicRoutingHandler(recordRequestsResponses)
  val messageHandler: JsonRpcMessageHandler = JsonRpcMessageProcessor(jsonRpcRequestHandler, meterRegistry)
  val httpServer: HttpJsonRpcServer = HttpJsonRpcServer(
    port = port.toUInt(),
    path = apiPath,
    HttpRequestHandler(messageHandler)
  )
  val bindedPort: Int
    get() = httpServer.bindedPort
  private var verticleId: String? = null

  init {
    vertx
      .deployVerticle(httpServer, DeploymentOptions().setInstances(1))
      .onSuccess { verticleId: String ->
        this.verticleId = verticleId
      }
      .get()
  }

  fun stop(): Future<Void> {
    return vertx.undeploy(verticleId)
  }

  /**
   * Handler shall return response result or throw [JsonRpcErrorResponseException] if error
   */
  fun handle(
    method: String,
    methodHandler: (jsonRpcRequest: JsonRpcRequest) -> Any?
  ) {
    jsonRpcRequestHandler.handlers[method] = methodHandler
  }

  fun recordedRequests(): List<Pair<JsonRpcRequest, Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>>> {
    return jsonRpcRequestHandler.requests.toList()
  }

  fun cleanRecordedRequests() {
    jsonRpcRequestHandler.requests.clear()
  }
}

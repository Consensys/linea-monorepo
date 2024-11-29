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
import net.consensys.linea.jsonrpc.JsonRpcMessageProcessor
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.jsonrpc.httpserver.HttpJsonRpcServer
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.util.concurrent.ConcurrentHashMap

open class FakeJsonRpcServer(
  port: Int = 0,
  apiPath: String = "/",
  val recordRequestsResponses: Boolean = false,
  serverName: String = "FakeJsonRpcServer",
  loggerName: String = serverName,
  val vertx: Vertx = Vertx.vertx()
) {
  val log: Logger = LogManager.getLogger(loggerName)
  val httpServer: HttpJsonRpcServer = HttpJsonRpcServer(
    port = port.toUInt(),
    path = apiPath,
    requestHandler = HttpRequestHandler(
      JsonRpcMessageProcessor(
        requestsHandler = this::handleRequest,
        meterRegistry = SimpleMeterRegistry(),
        log = log
      )
    ),
    serverName = serverName
  )
  val bindedPort: Int
    get() = httpServer.bindedPort
  private var verticleId: String? = null

  private val handlers: MutableMap<String, (JsonRpcRequest) -> Any?> = ConcurrentHashMap()
  private var requests: MutableList<
    Pair<JsonRpcRequest, Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>>
    > = mutableListOf()

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

  @Suppress("UNUSED_PARAMETER")
  private fun handleRequest(
    user: User?,
    jsonRpcRequest: JsonRpcRequest,
    requestJson: JsonObject
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    // need this otherwise kotlin compiler/IDE struggle to infer the type
    val result: Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> = (
      handlers[jsonRpcRequest.method]
        ?.let { handler ->
          try {
            val result = handler(jsonRpcRequest)
            Future.succeededFuture(
              Ok(
                JsonRpcSuccessResponse(
                  request = jsonRpcRequest,
                  result = result
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

  /**
   * Handler shall return response result or throw [JsonRpcErrorResponseException] if error
   */
  fun handle(
    method: String,
    methodHandler: (jsonRpcRequest: JsonRpcRequest) -> Any?
  ) {
    handlers[method] = methodHandler
  }

  fun recordedRequests(): List<Pair<JsonRpcRequest, Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>>> {
    return requests.toList()
  }

  fun cleanRecordedRequests() {
    requests.clear()
  }

  fun callCountByMethod(method: String): Int {
    return requests.count { it.first.method == method }
  }
}

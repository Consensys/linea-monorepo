package net.consensys.zkevm.coordinator.api

import io.vertx.core.DeploymentOptions
import io.vertx.core.Future
import io.vertx.core.Vertx
import linea.LongRunningService
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.jsonrpc.HttpRequestHandler
import net.consensys.linea.jsonrpc.JsonRpcMessageHandler
import net.consensys.linea.jsonrpc.JsonRpcMessageProcessor
import net.consensys.linea.jsonrpc.JsonRpcRequestRouter
import net.consensys.linea.jsonrpc.httpserver.HttpJsonRpcServer
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.vertx.ObservabilityServer
import net.consensys.zkevm.coordinator.api.requesthandlers.ConflationCreateProverRequestHandler
import net.consensys.zkevm.coordinator.api.requesthandlers.ConflationGetJobStatusRequestHandler
import net.consensys.zkevm.coordinator.app.conflationbacktesting.ConflationBacktestingService
import java.util.concurrent.CompletableFuture

class Api(
  private val configs: Config,
  private val vertx: Vertx,
  private val conflationBacktestingService: ConflationBacktestingService,
  private val metricsFacade: MetricsFacade,
) : LongRunningService {
  data class Config(
    val observabilityPort: UInt,
    val jsonRpcPort: UInt,
    val jsonRpcPath: String,
  )
  private var observabilityServerId: String? = null

  private var jsonRpcServerId: String? = null
  private var serverPort: Int = -1
  val bindedPort: Int
    get() = if (serverPort > 0) {
      serverPort
    } else {
      throw IllegalStateException("Http server not started")
    }

  override fun start(): CompletableFuture<Unit> {
    val requestHandlers = mapOf(
      ConflationCreateProverRequestHandler.METHOD_NAME to
        ConflationCreateProverRequestHandler(conflationBacktestingService = conflationBacktestingService),
      ConflationGetJobStatusRequestHandler.METHOD_NAME to
        ConflationGetJobStatusRequestHandler(conflationBacktestingService = conflationBacktestingService),
    )
    val messageHandler: JsonRpcMessageHandler =
      JsonRpcMessageProcessor(JsonRpcRequestRouter(requestHandlers), metricsFacade)

    val observabilityServer =
      ObservabilityServer(ObservabilityServer.Config("coordinator", configs.observabilityPort.toInt()))

    var httpServer: HttpJsonRpcServer? = null
    return vertx
      .deployVerticle(
        {
          HttpJsonRpcServer(configs.jsonRpcPort.toUInt(), configs.jsonRpcPath, HttpRequestHandler(messageHandler))
            .also {
              httpServer = it
            }
        },
        DeploymentOptions().setInstances(Runtime.getRuntime().availableProcessors()),
      )
      .compose { verticleId: String ->
        jsonRpcServerId = verticleId
        serverPort = httpServer!!.boundPort
        vertx.deployVerticle(observabilityServer).onSuccess { monitorVerticleId ->
          this.observabilityServerId = monitorVerticleId
        }
      }.toSafeFuture().thenApply { }
  }

  override fun stop(): CompletableFuture<Unit> {
    return Future.all(
      this.jsonRpcServerId?.let { vertx.undeploy(it) } ?: Future.succeededFuture(Unit),
      this.observabilityServerId?.let { vertx.undeploy(it) } ?: Future.succeededFuture(Unit),
    ).toSafeFuture().thenApply {}
  }
}

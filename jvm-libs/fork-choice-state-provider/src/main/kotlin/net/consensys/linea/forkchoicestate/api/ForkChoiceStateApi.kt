package net.consensys.linea.forkchoicestate.api

import io.micrometer.core.instrument.MeterRegistry
import io.vertx.core.DeploymentOptions
import io.vertx.core.Future
import io.vertx.core.Vertx
import net.consensys.linea.forkchoicestate.ForkChoiceStateController
import net.consensys.linea.jsonrpc.HttpRequestHandler
import net.consensys.linea.jsonrpc.JsonRpcMessageHandler
import net.consensys.linea.jsonrpc.JsonRpcMessageProcessor
import net.consensys.linea.jsonrpc.JsonRpcRequestRouter
import net.consensys.linea.jsonrpc.httpserver.HttpJsonRpcServer

data class ForkChoiceStateApiConfig(
  val port: UInt,
  val path: String = "/",
  val numberOfVerticles: UInt
)

class ForkChoiceStateApi(
  val configs: ForkChoiceStateApiConfig,
  val vertx: Vertx,
  val meterRegistry: MeterRegistry,
  val controller: ForkChoiceStateController
) {
  private var jsonRpcServerId: String? = null
  fun start(): Future<*> {
    val requestHandlers =
      mapOf(
        ForkChoiceApiMethod.LINEA_GET_FORK_CHOICE_STATE.method to
          GetForkChoiceStateHandler(controller)
      )

    val messageHandler: JsonRpcMessageHandler =
      JsonRpcMessageProcessor(JsonRpcRequestRouter(requestHandlers), meterRegistry)

    val numberOfVerticles: Int =
      if (configs.numberOfVerticles.toInt() > 0) {
        configs.numberOfVerticles.toInt()
      } else {
        Runtime.getRuntime().availableProcessors()
      }

    return vertx
      .deployVerticle(
        { HttpJsonRpcServer(configs.port, configs.path, HttpRequestHandler(messageHandler)) },
        DeploymentOptions().setInstances(numberOfVerticles)
      )
      .onSuccess { verticleId: String -> jsonRpcServerId = verticleId }
  }

  fun stop(): Future<*> {
    return this.jsonRpcServerId?.let { vertx.undeploy(it) } ?: Future.succeededFuture(null)
  }
}

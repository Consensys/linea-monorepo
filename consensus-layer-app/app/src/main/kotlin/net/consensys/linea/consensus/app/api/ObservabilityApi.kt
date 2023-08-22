package net.consensys.linea.consensus.app.api

import io.vertx.core.Future
import io.vertx.core.Vertx
import net.consensys.linea.vertx.ObservabilityServer

class ObservabilityApi(private val config: ObservabilityServer.Config, private val vertx: Vertx) {
  private var monitorServerId: String? = null

  fun start(): Future<*> {
    return vertx.deployVerticle(ObservabilityServer(config)).onSuccess { monitorVerticleId ->
      this.monitorServerId = monitorVerticleId
    }
  }

  fun stop(): Future<*> {
    return this.monitorServerId?.let { vertx.undeploy(it) } ?: Future.succeededFuture(null)
  }
}

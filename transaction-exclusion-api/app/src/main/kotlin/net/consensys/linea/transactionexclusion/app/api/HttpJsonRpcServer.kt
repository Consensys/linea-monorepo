package net.consensys.linea.transactionexclusion.app.api

import io.vertx.core.AbstractVerticle
import io.vertx.core.AsyncResult
import io.vertx.core.Handler
import io.vertx.core.Promise
import io.vertx.core.http.HttpServer
import io.vertx.core.http.HttpServerOptions
import io.vertx.ext.web.Router
import io.vertx.ext.web.RoutingContext
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class HttpJsonRpcServer(
  private val port: UInt,
  private val path: String,
  private val requestHandler: Handler<RoutingContext>
) : AbstractVerticle() {
  private val log: Logger = LogManager.getLogger(this.javaClass)
  private lateinit var httpServer: HttpServer
  val bindedPort: Int
    get() = if (this::httpServer.isInitialized) {
      httpServer.actualPort()
    } else {
      throw IllegalStateException("Http server not started")
    }

  override fun start(startPromise: Promise<Void>) {
    val options = HttpServerOptions().setPort(port.toInt()).setReusePort(true)
    log.debug("Creating Http server on port {}", port)
    httpServer = vertx.createHttpServer(options)
    httpServer.requestHandler(buildRouter())
    httpServer.listen { res: AsyncResult<HttpServer> ->
      if (res.succeeded()) {
        log.info("Http server started and listening on port {}", res.result().actualPort())
        startPromise.complete()
      } else {
        log.error("Creating Http server: {}", res.cause())
        startPromise.fail(res.cause())
      }
    }
  }

  private fun buildRouter(): Router {
    val router = Router.router(vertx)
    router.route(path).produces("application/json").handler(requestHandler)
    return router
  }

  override fun stop(endFuture: Promise<Void>) {
    httpServer.close(endFuture)
  }
}

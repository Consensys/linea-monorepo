package net.consensys.linea.vertx

import io.vertx.core.Vertx
import io.vertx.core.VertxOptions
import io.vertx.core.json.JsonObject

object VertxFactory {
  fun createVertx(
    configsOverrides: JsonObject = JsonObject()
  ): Vertx {
    val defaultConfigs = JsonObject(
      """
      {
        "preferNativeTransport": true,
        "logStacktraceThreshold": 500,
        "blockedThreadCheckIntervalUnit": "MINUTES",
        "maxEventLoopExecuteTime": 5000,
        "maxEventLoopExecuteTimeUnit": "MILLISECONDS",
        "warnEventLoopBlocked": 5000,
        "maxWorkerExecuteTime": 130,
        "maxWorkerExecuteTimeUnit": "SECONDS",
        "metricsOptions": {
          "enabled": true,
          "jvmMetricsEnabled": true,
          "prometheusOptions": {
            "enabled": true,
            "publishQuantiles": true
          }
        }
      }
      """.trimIndent()
    )

    return Vertx.vertx(VertxOptions(defaultConfigs.mergeIn(configsOverrides)))
  }
}

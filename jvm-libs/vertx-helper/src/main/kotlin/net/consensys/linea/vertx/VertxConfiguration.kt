package net.consensys.linea.vertx

import io.vertx.config.ConfigRetriever
import io.vertx.config.ConfigRetrieverOptions
import io.vertx.config.ConfigStoreOptions
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject

fun loadVertxConfig(vertxPropertiesConfigFile: String?, vertx: Vertx = Vertx.vertx()): JsonObject {
  val sysPropsStore = ConfigStoreOptions().setType("sys")
  val options = ConfigRetrieverOptions()

  if (vertxPropertiesConfigFile != null) {
    val fileStore =
      ConfigStoreOptions()
        .setType("file")
        .setFormat("properties")
        .setConfig(
          JsonObject()
            .put("path", vertxPropertiesConfigFile)
            .put("raw-data", false)
            .put("hierarchical", true)
        )
    options.addStore(fileStore)
  }

  options.addStore(sysPropsStore)

  return ConfigRetriever.create(vertx, options)
    .config
    .toCompletionStage()
    .toCompletableFuture()
    .get()
}

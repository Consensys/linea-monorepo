package net.consensys.linea.vertx

import io.vertx.core.http.HttpClientOptions
import net.consensys.getPortWithSchemaDefaults
import java.net.URI

fun <T : HttpClientOptions> T.setDefaultsFrom(uri: URI): T {
  isSsl = uri.scheme.lowercase() == "https"
  defaultHost = uri.host
  defaultPort = uri.getPortWithSchemaDefaults()

  return this
}

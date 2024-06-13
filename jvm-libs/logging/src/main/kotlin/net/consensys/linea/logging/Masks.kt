package net.consensys.linea.logging

import java.net.URL

typealias LogFieldMask = (String) -> String

fun noopMask(field: String): String = field

fun maskEndpointPath(endpoint: String): String {
  val path = URL(endpoint).path
  return endpoint.replace(path, "/" + "*".repeat(path.length - 1))
}

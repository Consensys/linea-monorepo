@file:Suppress("UNCHECKED_CAST")

plugins {
  id("net.consensys.zkevm.kotlin-library-conventions")
}

dependencies {
  val versions: Map<String, String> = project.ext["versions"] as Map<String, String>

  api(project(":coordinator:core"))
  api("tech.pegasys.teku.internal:timer:${versions["teku"]}")
  api("tech.pegasys.teku.internal:serviceutils:${versions["teku"]}")
  api("org.jetbrains.kotlinx:kotlinx-datetime:${versions["kotlinxDatetime"]}")
}

tasks.test {
  useJUnitPlatform()
}

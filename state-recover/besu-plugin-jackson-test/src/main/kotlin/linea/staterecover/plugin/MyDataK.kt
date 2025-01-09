package linea.staterecover.plugin

import kotlin.random.Random

data class MyDataK(
  val description: String = "KotlinDataClass",
  val someNumber: Int = 20,
  val someBytes: ByteArray = Random.nextBytes(20)
)

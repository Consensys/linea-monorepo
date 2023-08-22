package net.consensys.linea.traces

import com.github.michaelbull.result.expect
import io.vertx.core.json.JsonObject
import java.io.File
import java.io.FileWriter
import java.io.PrintWriter
import java.nio.charset.Charset

fun testCount(args: List<String>) {
  val counter = RawJsonTracesCounter("0.1")

  for (arg in args) {
    println("reading $arg")
    val r =
      counter.concreteCountTraces(JsonObject(File(arg).bufferedReader().use { it.readText() }))
    println(r)
  }
}

fun testConflate(args: List<String>) {
  val conflator = RawJsonTracesConflator("0.1")

  val r =
    conflator
      .conflateTraces(
        args.map {
          println("reading $it")
          JsonObject(File(it).bufferedReader().use { it.readText() })
        }
      )
      .expect { "Conflation failed" }
      .result
  try {
    PrintWriter(FileWriter("conflated.json", Charset.defaultCharset())).use {
      it.write(r.toString())
    }
  } catch (e: Exception) {
    e.printStackTrace()
  }
}

fun main(args: Array<String>) {
  when (args[0]) {
    "count" -> testCount(args.slice(1..args.size - 1))
    "conflate" -> testConflate(args.slice(1..args.size - 1))
    else -> println("unknown command $args[0]; expected `count` or `conflate`")
  }
}

package linea.test

import io.vertx.core.Vertx
import net.consensys.linea.async.get
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.core.config.Configurator
import java.util.concurrent.TimeUnit

fun configureLoggers(loggerConfigs: List<Pair<String, Level>>) {
  loggerConfigs.forEach { (loggerName, level) ->
    Configurator.setLevel(loggerName, level)
  }
}

fun main() {
  val lineaSepoliaUrl = "https://linea-sepolia.infura.io/v3/${System.getenv("INFURA_PROJECT_ID")}"
//  val lineaMainnetUrl = "https://linea-mainnet.infura.io/v3/${System.getenv("INFURA_PROJECT_ID")}"
  val vertx = Vertx.vertx()
  vertx.exceptionHandler { error ->
    println("Unhandled exception: message=${error.message}")
    LogManager.getLogger("vertx").error("Unhandled exception: message={}", error.message, error)
  }
  val fetcherAndValidate =
    FetchAndValidationRunner(
      rpcUrl = lineaSepoliaUrl,
      vertx = vertx,
      log = LogManager.getLogger("test.validator")
    )
  configureLoggers(
    listOf(
      "test.client.web3j" to Level.INFO,
      "test.validator" to Level.INFO
    )
  )

//  val startBlockNumber = 924973UL
//  val startBlockNumber = 924976UL
//  val startBlockNumber = 929527UL
  val startBlockNumber = 100_034UL // encoding/decoding does not match
//  val startBlockNumber = 100_000UL
  runCatching {
    fetcherAndValidate.fetchAndValidateBlocks(
      startBlockNumber = startBlockNumber,
//      endBlockNumber = startBlockNumber + 5000u,
      endBlockNumber = startBlockNumber + 0u,
      chuckSize = 500U,
      rlpEncodingDecodingOnly = false
    ).get(2, TimeUnit.MINUTES)
  }.onFailure { error ->
    fetcherAndValidate.log.error("Error fetching and validating blocks", error)
  }
  fetcherAndValidate.awaitValidationFinishes().get()
  println("waited validation finishes")
//  fetcherAndValidate.validator.stop()
  vertx.close().get()
  println("closed vertx")
}

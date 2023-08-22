package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import com.github.michaelbull.result.onFailure
import com.github.michaelbull.result.onSuccess
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.jsonrpc.BaseJsonRpcRequest
import net.consensys.linea.jsonrpc.client.JsonRpcClient
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.net.URL
import java.util.concurrent.atomic.AtomicInteger

class GasPriceUpdaterImpl(
  private val httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory,
  config: Config
) : GasPriceUpdater {
  class Config(
    val endpoints: List<URL>
  )

  private val log: Logger = LogManager.getLogger(this::class.java)
  private val rpcClients: List<JsonRpcClient> = config.endpoints.map {
    httpJsonRpcClientFactory.create(
      it,
      log = log,
      requestDefaultLogLevel = Level.TRACE,
      requestFailureLogLevel = Level.DEBUG
    )
  }
  private var id = AtomicInteger(0)

  override fun updateMinerGasPrice(gasPrice: BigInteger): SafeFuture<List<Unit>> {
    val jsonRequest =
      BaseJsonRpcRequest(
        "2.0",
        id.incrementAndGet(),
        "miner_setGasPrice",
        listOf(
          "0x${gasPrice.toString(16)}"
        )
      )

    return SafeFuture.collectAll(
      rpcClients.map { jsonRpcClient ->
        jsonRpcClient
          .makeRequest(jsonRequest)
          .toSafeFuture()
          .whenException { th ->
            log.error("Error from rpc request of miner_setGasPrice: errorMessage={}", th.message, th)
          }
          .thenApply { result ->
            result
              .onSuccess {
                if (it.result == true) {
                  log.trace("Result of miner_setGasPrice: {}", it.result)
                } else {
                  log.warn("Result of miner_setGasPrice: {}", it.result)
                }
              }.onFailure {
                log.error("Error from miner_setGasPrice: errorMessage={}", it.error)
              }
            Unit
          }
      }.stream()
    )
  }
}

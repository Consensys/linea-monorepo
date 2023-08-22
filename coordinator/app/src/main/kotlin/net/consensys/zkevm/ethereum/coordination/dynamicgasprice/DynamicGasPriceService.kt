package net.consensys.zkevm.ethereum.coordination.dynamicgasprice

import io.vertx.core.TimeoutStream
import io.vertx.core.Vertx
import net.consensys.linea.toGWei
import net.consensys.linea.toIntervalString
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import kotlin.time.Duration

class DynamicGasPriceService(
  private val config: Config,
  private val vertx: Vertx,
  private val feesFetcher: FeesFetcher,
  private val feesCalculator: FeesCalculator,
  private val gasPriceUpdater: GasPriceUpdater
) {
  class Config(
    val pollingInterval: Duration,
    val gasPriceCap: BigInteger
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  @Volatile
  private lateinit var monitorStream: TimeoutStream

  internal fun tick(): SafeFuture<Unit> {
    return feesFetcher
      .getL1EthGasPriceData()
      .thenCompose { feeHistory ->
        val blockRange = feeHistory.blocksRange()
        val gasPrice = feesCalculator.calculateFees(feeHistory)
        val gasPriceToUpdate = if (gasPrice > config.gasPriceCap) {
          log.warn(
            "L2 Gas price update: gasPrice={}GWei, l1Blocks={}. gasPrice is higher than gasPriceCap={}GWei." +
              " Will default to cap value",
            gasPrice.toGWei(),
            blockRange.toIntervalString(),
            config.gasPriceCap.toGWei()
          )
          config.gasPriceCap
        } else {
          log.info(
            "L2 Gas price update: gasPrice={}GWei, l1Blocks={}.",
            gasPrice.toGWei(),
            blockRange.toIntervalString()
          )
          gasPrice
        }
        gasPriceUpdater.updateMinerGasPrice(gasPriceToUpdate)
      }
      .thenApply { log.debug("Fetch, calculate, update new miner gas price are all done.") }
  }

  fun start(): SafeFuture<Unit> {
    monitorStream =
      vertx.periodicStream(config.pollingInterval.inWholeMilliseconds).handler {
        try {
          monitorStream.pause()
          tick().whenComplete { _, _ -> monitorStream.resume() }
        } catch (th: Throwable) {
          log.error(th)
          monitorStream.resume()
        }
      }
    return SafeFuture.completedFuture(Unit)
  }

  fun stop(): SafeFuture<Unit> {
    return if (this::monitorStream.isInitialized) {
      SafeFuture.completedFuture(monitorStream.cancel())
    } else {
      log.warn("Dynamic Gas Price Service hasn't been started to stop it, but Ok")
      SafeFuture.completedFuture(Unit)
    }
  }
}

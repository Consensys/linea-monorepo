package net.consensys.zkevm.coordinator.app

import io.vertx.core.Vertx
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.DynamicGasPriceService
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeeHistoryFetcherImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeesCalculator
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeesFetcher
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.GasPriceUpdater
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.GasPriceUpdaterImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.GasUsageRatioWeightedAverageFeesCalculator
import org.apache.logging.log4j.LogManager
import org.web3j.protocol.Web3j
import java.util.concurrent.CompletableFuture
import kotlin.time.toKotlinDuration

class GasPriceUpdaterApp(
  vertx: Vertx,
  httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory,
  l1Web3jClient: Web3j,
  configs: Config
) : LongRunningService {
  private val log = LogManager.getLogger(this::class.java)

  data class Config(
    val dynamicGasPriceService: DynamicGasPriceServiceConfig
  )

  private val dynamicGasPriceService: DynamicGasPriceService = run {
    val gasServiceFeesFetcher: FeesFetcher = FeeHistoryFetcherImpl(
      l1Web3jClient,
      FeeHistoryFetcherImpl.Config(
        configs.dynamicGasPriceService.feeHistoryBlockCount.toUInt(),
        configs.dynamicGasPriceService.feeHistoryRewardPercentile
      )
    )

    val l2MinMinerTipCalculator: FeesCalculator = GasUsageRatioWeightedAverageFeesCalculator(
      GasUsageRatioWeightedAverageFeesCalculator.Config(
        configs.dynamicGasPriceService.baseFeeCoefficient,
        configs.dynamicGasPriceService.priorityFeeCoefficient
      )
    )

    val l2SetGasPriceUpdater: GasPriceUpdater = GasPriceUpdaterImpl(
      httpJsonRpcClientFactory,
      GasPriceUpdaterImpl.Config(
        configs.dynamicGasPriceService.minerGasPriceUpdateRecipients
      )
    )

    DynamicGasPriceService(
      DynamicGasPriceService.Config(
        configs.dynamicGasPriceService.pollingInterval.toKotlinDuration(),
        configs.dynamicGasPriceService.gasPriceCap
      ),
      vertx,
      gasServiceFeesFetcher,
      l2MinMinerTipCalculator,
      l2SetGasPriceUpdater
    )
  }

  override fun start(): CompletableFuture<Unit> {
    return dynamicGasPriceService.start().thenPeek {
      log.info("GasPriceUpdater started")
    }
  }

  override fun stop(): CompletableFuture<Unit> {
    return dynamicGasPriceService.stop()
      .thenPeek {
        log.info("GasPriceUpdater stopped")
      }
  }
}

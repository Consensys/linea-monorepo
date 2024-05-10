package net.consensys.zkevm.coordinator.app

import io.vertx.core.Vertx
import net.consensys.linea.contract.BoundableFeeCalculator
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.web3j.Web3jBlobExtended
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.DynamicGasPriceService
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeeHistoryFetcherImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.GasPriceUpdaterImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.GasUsageRatioWeightedAverageFeesCalculator
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import net.consensys.zkevm.ethereum.gaspricing.FeesFetcher
import net.consensys.zkevm.ethereum.gaspricing.GasPriceUpdater
import org.apache.logging.log4j.LogManager
import org.web3j.protocol.Web3j
import java.util.concurrent.CompletableFuture
import kotlin.time.toKotlinDuration

class GasPriceUpdaterApp(
  vertx: Vertx,
  httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory,
  l1Web3jClient: Web3j,
  l1Web3jService: Web3jBlobExtended,
  configs: Config
) : LongRunningService {
  private val log = LogManager.getLogger(this::class.java)

  data class Config(
    val dynamicGasPriceService: DynamicGasPriceServiceConfig
  )

  private val dynamicGasPriceService: DynamicGasPriceService = run {
    val gasServiceFeesFetcher: FeesFetcher = FeeHistoryFetcherImpl(
      l1Web3jClient,
      l1Web3jService,
      FeeHistoryFetcherImpl.Config(
        configs.dynamicGasPriceService.feeHistoryBlockCount.toUInt(),
        configs.dynamicGasPriceService.feeHistoryRewardPercentile
      )
    )

    val gasUsageRatioWeightedAverageFeesCalculator = GasUsageRatioWeightedAverageFeesCalculator(
      GasUsageRatioWeightedAverageFeesCalculator.Config(
        configs.dynamicGasPriceService.baseFeeCoefficient,
        configs.dynamicGasPriceService.priorityFeeCoefficient,
        configs.dynamicGasPriceService.baseFeeBlobCoefficient,
        configs.dynamicGasPriceService.blobSubmissionExpectedExecutionGas,
        configs.dynamicGasPriceService.expectedBlobGas
      )
    )

    val l2MinMinerTipCalculator: FeesCalculator = BoundableFeeCalculator(
      BoundableFeeCalculator.Config(
        configs.dynamicGasPriceService.gasPriceUpperBound,
        configs.dynamicGasPriceService.gasPriceLowerBound,
        configs.dynamicGasPriceService.gasPriceFixedCost
      ),
      gasUsageRatioWeightedAverageFeesCalculator
    )

    val l2SetGasPriceUpdater: GasPriceUpdater = GasPriceUpdaterImpl(
      httpJsonRpcClientFactory = httpJsonRpcClientFactory,
      config = GasPriceUpdaterImpl.Config(
        gethEndpoints = configs.dynamicGasPriceService.gethGasPriceUpdateRecipients,
        besuEndPoints = configs.dynamicGasPriceService.besuGasPriceUpdateRecipients,
        retryConfig = configs.dynamicGasPriceService.requestRetryConfig
      )
    )

    DynamicGasPriceService(
      DynamicGasPriceService.Config(configs.dynamicGasPriceService.priceUpdateInterval.toKotlinDuration()),
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

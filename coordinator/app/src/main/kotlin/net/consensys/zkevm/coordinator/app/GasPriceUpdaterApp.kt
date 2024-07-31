package net.consensys.zkevm.coordinator.app

import io.vertx.core.Vertx
import net.consensys.linea.ethereum.gaspricing.BoundableFeeCalculator
import net.consensys.linea.ethereum.gaspricing.FeesCalculator
import net.consensys.linea.ethereum.gaspricing.FeesFetcher
import net.consensys.linea.ethereum.gaspricing.GasPriceUpdater
import net.consensys.linea.ethereum.gaspricing.staticcap.ExtraDataV1PricerService
import net.consensys.linea.ethereum.gaspricing.staticcap.ExtraDataV1UpdaterImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.FeeHistoryFetcherImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.GasPriceUpdaterImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.GasUsageRatioWeightedAverageFeesCalculator
import net.consensys.linea.ethereum.gaspricing.staticcap.MinMineableFeesPricerService
import net.consensys.linea.ethereum.gaspricing.staticcap.MinerExtraDataV1CalculatorImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.VariableFeesCalculator
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.web3j.Web3jBlobExtended
import net.consensys.toKWeiUInt
import net.consensys.zkevm.LongRunningService
import org.apache.logging.log4j.LogManager
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.CompletableFuture
import kotlin.time.toKotlinDuration

class GasPriceUpdaterApp(
  vertx: Vertx,
  httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory,
  l1Web3jClient: Web3j,
  l1Web3jService: Web3jBlobExtended,
  config: DynamicGasPriceServiceConfig
) : LongRunningService {
  private val log = LogManager.getLogger(this::class.java)

  private val gasPricingFeesFetcher: FeesFetcher = FeeHistoryFetcherImpl(
    l1Web3jClient,
    l1Web3jService,
    FeeHistoryFetcherImpl.Config(
      config.feeHistoryBlockCount.toUInt(),
      config.feeHistoryRewardPercentile
    )
  )

  private val minMineableFeesCalculator: FeesCalculator = run {
    val gasUsageRatioWeightedAverageFeesCalculator = GasUsageRatioWeightedAverageFeesCalculator(
      GasUsageRatioWeightedAverageFeesCalculator.Config(
        baseFeeCoefficient = config.baseFeeCoefficient,
        priorityFeeCoefficient = config.priorityFeeCoefficient,
        baseFeeBlobCoefficient = config.baseFeeBlobCoefficient,
        blobSubmissionExpectedExecutionGas = config.blobSubmissionExpectedExecutionGas,
        expectedBlobGas = config.expectedBlobGas
      )
    )
    BoundableFeeCalculator(
      BoundableFeeCalculator.Config(
        config.gasPriceUpperBound.toDouble(),
        config.gasPriceLowerBound.toDouble(),
        config.gasPriceFixedCost.toDouble()
      ),
      gasUsageRatioWeightedAverageFeesCalculator
    )
  }

  private val minMineableFeesPricerService: MinMineableFeesPricerService? =
    if (config.minMineableFeesEnabled) {
      val l2SetGasPriceUpdater: GasPriceUpdater = GasPriceUpdaterImpl(
        httpJsonRpcClientFactory = httpJsonRpcClientFactory,
        config = GasPriceUpdaterImpl.Config(
          gethEndpoints = config.gethGasPriceUpdateRecipients,
          besuEndPoints = config.besuGasPriceUpdateRecipients,
          retryConfig = config.requestRetryConfig
        )
      )

      MinMineableFeesPricerService(
        pollingInterval = config.priceUpdateInterval.toKotlinDuration(),
        vertx = vertx,
        feesFetcher = gasPricingFeesFetcher,
        feesCalculator = minMineableFeesCalculator,
        gasPriceUpdater = l2SetGasPriceUpdater
      )
    } else {
      null
    }

  private val extraDataPricerService: ExtraDataV1PricerService? = if (config.extraDataEnabled) {
    val variableCostCalculator = VariableFeesCalculator(
      VariableFeesCalculator.Config(
        blobSubmissionExpectedExecutionGas = config.blobSubmissionExpectedExecutionGas,
        bytesPerDataSubmission = config.bytesPerDataSubmission,
        expectedBlobGas = config.expectedBlobGas,
        margin = config.margin
      )
    )
    val boundedVariableCostCalculator = BoundableFeeCalculator(
      config = BoundableFeeCalculator.Config(
        feeUpperBound = config.variableCostUpperBound.toDouble(),
        feeLowerBound = config.variableCostLowerBound.toDouble(),
        feeMargin = 0.0
      ),
      feesCalculator = variableCostCalculator
    )
    ExtraDataV1PricerService(
      pollingInterval = config.priceUpdateInterval.toKotlinDuration(),
      vertx = vertx,
      feesFetcher = gasPricingFeesFetcher,
      minerExtraDataCalculatorImpl = MinerExtraDataV1CalculatorImpl(
        config = MinerExtraDataV1CalculatorImpl.Config(
          fixedCostInKWei = config.gasPriceFixedCost.toKWeiUInt(),
          ethGasPriceMultiplier = config.legacyFeesMultiplier
        ),
        variableFeesCalculator = boundedVariableCostCalculator,
        legacyFeesCalculator = minMineableFeesCalculator
      ),
      extraDataUpdater = ExtraDataV1UpdaterImpl(
        httpJsonRpcClientFactory = httpJsonRpcClientFactory,
        config = ExtraDataV1UpdaterImpl.Config(
          config.extraDataUpdateRecipient,
          config.requestRetryConfig
        )
      )
    )
  } else {
    null
  }

  override fun start(): CompletableFuture<Unit> {
    return (minMineableFeesPricerService?.start() ?: SafeFuture.completedFuture(Unit))
      .thenCompose { extraDataPricerService?.start() ?: SafeFuture.completedFuture(Unit) }
      .thenPeek {
        log.info("GasPriceUpdater started")
      }
  }

  override fun stop(): CompletableFuture<Unit> {
    return SafeFuture.allOf(
      minMineableFeesPricerService?.stop() ?: SafeFuture.completedFuture(Unit),
      extraDataPricerService?.stop() ?: SafeFuture.completedFuture(Unit)
    )
      .thenApply { }
      .thenPeek {
        log.info("GasPriceUpdater stopped")
      }
  }
}

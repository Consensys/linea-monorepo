package linea.coordinator.config.v2.toml

import java.net.URL
import kotlin.time.Duration

data class L2NetworkGasPricingConfigToml(
  val disabled: Boolean,
  val priceUpdateInterval: Duration,
  val feeHistoryBlockCount: UInt,
  val feeHistoryRewardPercentile: UInt,
  val minMineableFeesEnabled: Boolean,
  val extraDataEnabled: Boolean,
  val gasPriceUpperBound: ULong,
  val gasPriceLowerBound: ULong,
  val gasPriceFixedCost: ULong,
  val requestRetries: RequestRetriesToml,
  val extraData: ExtraData,
  val minMineable: MinMineable
) {
  data class ExtraData(
    val l1BlobGas: Double,
    val blobSubmissionExpectedExecutionGas: Double,
    val variableCostUpperBound: Long,
    val variableCostLowerBound: Long,
    val margin: Double,
    val extraDataUpdateEndpoint: URL
  )

  data class MinMineable(
    val baseFeeCoefficient: Double,
    val priorityFeeCoefficient: Double,
    val baseFeeBlobCoefficient: Double,
    val legacyFeesMultiplier: Double,
    val gethGasPriceUpdateEndpoints: List<URL>,
    val besuGasPriceUpdateEndpoints: List<URL>
  )
}

package net.consensys.zkevm.coordinator.app.config

import com.sksamuel.hoplite.ConfigAlias
import net.consensys.linea.ethereum.gaspricing.BoundableFeeCalculator
import net.consensys.linea.ethereum.gaspricing.staticcap.ExtraDataV1UpdaterImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.FeeHistoryFetcherImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.GasPriceUpdaterImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.GasUsageRatioWeightedAverageFeesCalculator
import net.consensys.linea.ethereum.gaspricing.staticcap.MinerExtraDataV1CalculatorImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.VariableFeesCalculator
import net.consensys.toKWeiUInt
import net.consensys.zkevm.coordinator.app.L2NetworkGasPricingService
import java.net.URL
import java.time.Duration
import kotlin.time.toKotlinDuration

data class NaiveGasPricingTomlDto(
  val baseFeeCoefficient: Double,
  val priorityFeeCoefficient: Double,
  val baseFeeBlobCoefficient: Double,

  val gasPriceUpperBound: ULong,
  val gasPriceLowerBound: ULong
) {
  init {
    require(gasPriceUpperBound >= gasPriceLowerBound) {
      "gasPriceUpperBound must be greater than or equal to gasPriceLowerBound"
    }
  }
}

data class VariableCostPricingTomlDto(
  val gasPriceFixedCost: ULong,
  val legacyFeesMultiplier: Double,
  val margin: Double,
  val variableCostUpperBound: ULong,
  val variableCostLowerBound: ULong
) {
  init {
    require(variableCostUpperBound >= variableCostLowerBound) {
      "variableCostUpperBound must be greater than or equal to variableCostLowerBound"
    }
  }
}

data class JsonRpcPricingPropagationTomlDto(
  override var disabled: Boolean = false,
  val gethGasPriceUpdateRecipients: List<URL>,
  val besuGasPriceUpdateRecipients: List<URL>
) : FeatureToggleable {
  init {
    require(disabled || (gethGasPriceUpdateRecipients.isNotEmpty() || besuGasPriceUpdateRecipients.isNotEmpty())) {
      "There is no point of enabling JSON RPC pricing propagation if there are no " +
        "gethGasPriceUpdateRecipients or besuGasPriceUpdateRecipients defined"
    }
  }
}

data class ExtraDataPricingPropagationTomlDto(
  override var disabled: Boolean = false,
  val extraDataUpdateRecipient: URL
) : FeatureToggleable

data class L2NetworkGasPricingTomlDto(
  override var disabled: Boolean = false,
  override val requestRetry: RequestRetryConfigTomlFriendly,

  val priceUpdateInterval: Duration,
  val feeHistoryBlockCount: Int,
  val feeHistoryRewardPercentile: Double,

  val blobSubmissionExpectedExecutionGas: Int,
  @ConfigAlias("bytesPerDataSubmission") val _bytesPerDataSubmission: Int?,
  val l1BlobGas: Int,

  val naiveGasPricing: NaiveGasPricingTomlDto,
  val variableCostPricing: VariableCostPricingTomlDto,
  val jsonRpcPricingPropagation: JsonRpcPricingPropagationTomlDto,
  val extraDataPricingPropagation: ExtraDataPricingPropagationTomlDto
) : FeatureToggleable, RequestRetryConfigurable {
  init {
    require(feeHistoryBlockCount > 0) { "feeHistoryBlockCount must be greater than 0" }

    require(disabled || (jsonRpcPricingPropagation.enabled || extraDataPricingPropagation.enabled)) {
      "There is no point of enabling L2 network gas pricing if " +
        "both jsonRpcPricingPropagation and extraDataPricingPropagation are disabled"
    }
  }

  private val bytesPerDataSubmission = _bytesPerDataSubmission ?: l1BlobGas

  fun reified(): L2NetworkGasPricingService.Config {
    return L2NetworkGasPricingService.Config(
      feeHistoryFetcherConfig = FeeHistoryFetcherImpl.Config(
        feeHistoryBlockCount = feeHistoryBlockCount.toUInt(),
        feeHistoryRewardPercentile = feeHistoryRewardPercentile
      ),
      jsonRpcPricingPropagationEnabled = jsonRpcPricingPropagation.enabled,
      naiveGasPricingCalculatorConfig = GasUsageRatioWeightedAverageFeesCalculator.Config(
        baseFeeCoefficient = naiveGasPricing.baseFeeCoefficient,
        priorityFeeCoefficient = naiveGasPricing.priorityFeeCoefficient,
        baseFeeBlobCoefficient = naiveGasPricing.baseFeeBlobCoefficient,
        blobSubmissionExpectedExecutionGas = blobSubmissionExpectedExecutionGas,
        expectedBlobGas = l1BlobGas
      ),
      naiveGasPricingCalculatorBounds = BoundableFeeCalculator.Config(
        naiveGasPricing.gasPriceUpperBound.toDouble(),
        naiveGasPricing.gasPriceLowerBound.toDouble(),
        0.0
      ),
      jsonRpcGasPriceUpdaterConfig = GasPriceUpdaterImpl.Config(
        gethEndpoints = jsonRpcPricingPropagation.gethGasPriceUpdateRecipients,
        besuEndPoints = jsonRpcPricingPropagation.besuGasPriceUpdateRecipients,
        retryConfig = requestRetryConfig
      ),
      jsonRpcPriceUpdateInterval = priceUpdateInterval.toKotlinDuration(),
      extraDataPricingPropagationEnabled = extraDataPricingPropagation.enabled,
      extraDataUpdateInterval = priceUpdateInterval.toKotlinDuration(),
      variableFeesCalculatorConfig = VariableFeesCalculator.Config(
        blobSubmissionExpectedExecutionGas = blobSubmissionExpectedExecutionGas,
        bytesPerDataSubmission = l1BlobGas,
        expectedBlobGas = bytesPerDataSubmission,
        margin = variableCostPricing.margin
      ),
      variableFeesCalculatorBounds = BoundableFeeCalculator.Config(
        feeUpperBound = variableCostPricing.variableCostUpperBound.toDouble(),
        feeLowerBound = variableCostPricing.variableCostLowerBound.toDouble(),
        feeMargin = 0.0
      ),
      extraDataCalculatorConfig = MinerExtraDataV1CalculatorImpl.Config(
        fixedCostInKWei = variableCostPricing.gasPriceFixedCost.toUInt(),
        ethGasPriceMultiplier = variableCostPricing.legacyFeesMultiplier
      ),
      extraDataUpdaterConfig = ExtraDataV1UpdaterImpl.Config(
        extraDataPricingPropagation.extraDataUpdateRecipient,
        requestRetryConfig
      )
    )
  }
}

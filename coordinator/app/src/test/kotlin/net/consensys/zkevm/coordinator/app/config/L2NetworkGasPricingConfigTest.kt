package net.consensys.zkevm.coordinator.app.config

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.toml.TomlPropertySource
import net.consensys.linea.ethereum.gaspricing.BoundableFeeCalculator
import net.consensys.linea.ethereum.gaspricing.staticcap.ExtraDataV1UpdaterImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.FeeHistoryFetcherImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.GasPriceUpdaterImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.GasUsageRatioWeightedAverageFeesCalculator
import net.consensys.linea.ethereum.gaspricing.staticcap.MinerExtraDataV1CalculatorImpl
import net.consensys.linea.ethereum.gaspricing.staticcap.VariableFeesCalculator
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import net.consensys.zkevm.coordinator.app.L2NetworkGasPricingService
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import java.net.URI
import java.time.Duration
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

class L2NetworkGasPricingConfigTest {
  data class Config(
    val l2NetworkGasPricing: L2NetworkGasPricingTomlDto
  )

  private fun parseConfig(toml: String): L2NetworkGasPricingTomlDto {
    return ConfigLoaderBuilder
      .default()
      .addSource(TomlPropertySource(toml))
      .build()
      .loadConfigOrThrow<Config>().l2NetworkGasPricing
  }

  private val l2NetworkGasPricingServiceConfigToml = """
    [l2-network-gas-pricing]
    disabled = false
    price-update-interval = "PT12S"

    fee-history-block-count = 50
    fee-history-reward-percentile = 15

    blob-submission-expected-execution-gas = 213000.0 # Lower to 120k as we improve efficiency
    # Defaults to expected-blob-gas
    #bytes-per-data-submission=131072.0 # 2^17
    l1-blob-gas = 131072 # 2^17

    [l2-network-gas-pricing.request-retry]
    max-retries = 3
    timeout = "PT6S"
    backoff-delay = "PT1S"
    failures-warning-threshold = 2

    [l2-network-gas-pricing.variable-cost-pricing]
    gas-price-fixed-cost = 3000000
    legacy-fees-multiplier = 1.2
    margin = 4.0
    variable-cost-upper-bound = 10000000001 # ~10 GWEI
    variable-cost-lower-bound = 90000001  # ~0.09 GWEI

    [l2-network-gas-pricing.extra-data-pricing-propagation]
    extra-data-update-recipient = "http://sequencer:8545/"

    [l2-network-gas-pricing.naive-gas-pricing]
    base-fee-coefficient = 0.1
    priority-fee-coefficient = 1.0
    base-fee-blob-coefficient = 0.1

    gas-price-upper-bound = 10000000000 # 10 GWEI
    gas-price-lower-bound = 90000000 # 0.09 GWEI

    [l2-network-gas-pricing.json-rpc-pricing-propagation]
    geth-gas-price-update-recipients = [
      "http://traces-node:8545/",
      "http://l2-node:8545/"
    ]
    besu-gas-price-update-recipients = [
      "http://sequencer:8545/"
    ]
  """.trimIndent()

  @Test
  fun `dto is parseable`() {
    val config = parseConfig(l2NetworkGasPricingServiceConfigToml)
    assertThat(config).isEqualTo(
      L2NetworkGasPricingTomlDto(
        requestRetry = RequestRetryConfigTomlFriendly(
          maxRetries = 3,
          timeout = 6.seconds.toJavaDuration(),
          backoffDelay = 1.seconds.toJavaDuration(),
          failuresWarningThreshold = 2
        ),

        priceUpdateInterval = Duration.parse("PT12S"),
        feeHistoryBlockCount = 50,
        feeHistoryRewardPercentile = 15.0,

        blobSubmissionExpectedExecutionGas = 213_000,
        _bytesPerDataSubmission = null,
        l1BlobGas = 131072,

        naiveGasPricing = NaiveGasPricingTomlDto(
          baseFeeCoefficient = 0.1,
          priorityFeeCoefficient = 1.0,
          baseFeeBlobCoefficient = 0.1,

          gasPriceUpperBound = 10_000_000_000u,
          gasPriceLowerBound = 90_000_000u
        ),
        variableCostPricing = VariableCostPricingTomlDto(
          gasPriceFixedCost = 3000000u,
          legacyFeesMultiplier = 1.2,
          margin = 4.0,
          variableCostUpperBound = 10_000_000_001u,
          variableCostLowerBound = 90_000_001u
        ),
        jsonRpcPricingPropagation = JsonRpcPricingPropagationTomlDto(
          gethGasPriceUpdateRecipients = listOf(
            URI("http://traces-node:8545/").toURL(),
            URI("http://l2-node:8545/").toURL()
          ),
          besuGasPriceUpdateRecipients = listOf(
            URI("http://sequencer:8545/").toURL()
          )
        ),
        extraDataPricingPropagation = ExtraDataPricingPropagationTomlDto(
          extraDataUpdateRecipient = URI("http://sequencer:8545/").toURL()
        )
      )
    )
  }

  @Test
  fun `reification is correct`() {
    val config = parseConfig(l2NetworkGasPricingServiceConfigToml).reified()
    val l2NetworkGasPricingRequestretryConfig = RequestRetryConfig(
      maxRetries = 3u,
      timeout = 6.seconds,
      backoffDelay = 1.seconds,
      failuresWarningThreshold = 2u
    )

    assertThat(config).isEqualTo(
      L2NetworkGasPricingService.Config(
        feeHistoryFetcherConfig = FeeHistoryFetcherImpl.Config(
          feeHistoryBlockCount = 50U,
          feeHistoryRewardPercentile = 15.0
        ),
        jsonRpcPricingPropagationEnabled = true,
        naiveGasPricingCalculatorConfig = GasUsageRatioWeightedAverageFeesCalculator.Config(
          baseFeeCoefficient = 0.1,
          priorityFeeCoefficient = 1.0,
          baseFeeBlobCoefficient = 0.1,
          blobSubmissionExpectedExecutionGas = 213_000,
          expectedBlobGas = 131072
        ),
        naiveGasPricingCalculatorBounds = BoundableFeeCalculator.Config(
          10_000_000_000.0,
          90_000_000.0,
          0.0
        ),
        jsonRpcGasPriceUpdaterConfig = GasPriceUpdaterImpl.Config(
          gethEndpoints = listOf(
            URI("http://traces-node:8545/").toURL(),
            URI("http://l2-node:8545/").toURL()
          ),
          besuEndPoints = listOf(
            URI("http://sequencer:8545/").toURL()
          ),
          retryConfig = l2NetworkGasPricingRequestretryConfig,
        ),
        jsonRpcPriceUpdateInterval = 12.seconds,
        extraDataPricingPropagationEnabled = true,
        extraDataUpdateInterval = 12.seconds,
        variableFeesCalculatorConfig = VariableFeesCalculator.Config(
          blobSubmissionExpectedExecutionGas = 213_000,
          bytesPerDataSubmission = 131072,
          expectedBlobGas = 131072,
          margin = 4.0
        ),
        variableFeesCalculatorBounds = BoundableFeeCalculator.Config(
          feeUpperBound = 10_000_000_001.0,
          feeLowerBound = 90_000_001.0,
          feeMargin = 0.0
        ),
        extraDataCalculatorConfig = MinerExtraDataV1CalculatorImpl.Config(
          fixedCostInKWei = 3000000u,
          ethGasPriceMultiplier = 1.2
        ),
        extraDataUpdaterConfig = ExtraDataV1UpdaterImpl.Config(
          sequencerEndpoint = URI(/* str = */ "http://sequencer:8545/").toURL(),
          retryConfig = l2NetworkGasPricingRequestretryConfig
        )
      )
    )
  }
}

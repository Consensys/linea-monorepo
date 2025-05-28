package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.L2NetworkGasPricingConfigToml
import linea.coordinator.config.v2.toml.RequestRetriesToml
import linea.coordinator.config.v2.toml.parseConfig
import linea.kotlin.toURL
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import kotlin.time.Duration.Companion.seconds

class L2NetWorkingGasPricingConfigParsingTest {
  companion object {
    val toml = """
    [l2-network-gas-pricing] # old [dynamic-gas-price-service]
    disabled = false
    price-update-interval = "PT12S"
    fee-history-block-count = 50
    fee-history-reward-percentile = 15
    min-mineable-fees-enabled = true
    extra-data-enabled = true
    ## Usend un both extraDataPricerService and minMineableFeesCalculator,
    # extraDataPricerService just uses minMineableFeesCalculator as delegate to get legacy fees
    gas-price-upper-bound = 10000000000 # 10 GWEI
    gas-price-lower-bound = 90000000 # 0.09 GWEI
    gas-price-fixed-cost = 3000000

    # Defaults to expected-blob-gas
    #bytes-per-data-submission=131072.0 # 2^17
    [l2-network-gas-pricing.request-retries]
    max-retries = 3
    timeout = "PT6S"
    backoff-delay = "PT1S"
    failures-warning-threshold = 2

    [l2-network-gas-pricing.extra-data] # TODO: find proper name for "new", maybe "2D"
    l1-blob-gas = 131072.0 # 2^17 # expected-l1-blob-gas previous name: expected-blob-gas
    blob-submission-expected-execution-gas = 213000.0 # Lower to 120k as we improve efficiency
    variable-cost-upper-bound = 10000000001 # ~10 GWEI
    variable-cost-lower-bound = 90000001  # ~0.09 GWEI
    margin = 4.0
    extra-data-update-endpoint = "http://sequencer:8545/"

    [l2-network-gas-pricing.min-mineable] # current legacy implementation
    base-fee-coefficient = 0.1
    priority-fee-coefficient = 1.0
    base-fee-blob-coefficient = 0.1
    legacy-fees-multiplier = 1.2
    geth-gas-price-update-endpoints = [
      "http://traces-node:8545/",
      "http://l2-node:8545/"
    ]
    besu-gas-price-update-endpoints = [
      "http://sequencer:8545/"
    ]
    """.trimIndent()

    val expectedConfig = L2NetworkGasPricingConfigToml(
      disabled = false,
      priceUpdateInterval = 12.seconds,
      feeHistoryBlockCount = 50u,
      feeHistoryRewardPercentile = 15u,
      minMineableFeesEnabled = true,
      extraDataEnabled = true,
      gasPriceUpperBound = 10_000_000_000UL,
      gasPriceLowerBound = 90_000_000UL,
      gasPriceFixedCost = 3_000_000UL,
      requestRetries = RequestRetriesToml(
        maxRetries = 3u,
        timeout = 6.seconds,
        backoffDelay = 1.seconds,
        failuresWarningThreshold = 2u
      ),
      extraData = L2NetworkGasPricingConfigToml.ExtraData(
        l1BlobGas = 131072.0, // 2^17
        blobSubmissionExpectedExecutionGas = 213000.0, // Lower to 120k as we improve efficiency
        variableCostUpperBound = 10_000_000_001L, // ~10 GWEI
        variableCostLowerBound = 90_000_001L, // ~0.09 GWEI
        margin = 4.0,
        extraDataUpdateEndpoint = "http://sequencer:8545/".toURL()
      ),
      minMineable = L2NetworkGasPricingConfigToml.MinMineable(
        baseFeeCoefficient = 0.1,
        priorityFeeCoefficient = 1.0,
        baseFeeBlobCoefficient = 0.1,
        legacyFeesMultiplier = 1.2,
        gethGasPriceUpdateEndpoints = listOf(
          "http://traces-node:8545/".toURL(),
          "http://l2-node:8545/".toURL()
        ),
        besuGasPriceUpdateEndpoints = listOf("http://sequencer:8545/".toURL())
      )
    )
  }

  data class WrapperConfig(
    val l2NetworkGasPricing: L2NetworkGasPricingConfigToml
  )

  @Test
  fun `should parse full state manager config`() {
    assertThat(parseConfig<WrapperConfig>(toml).l2NetworkGasPricing)
      .isEqualTo(expectedConfig)
  }
}

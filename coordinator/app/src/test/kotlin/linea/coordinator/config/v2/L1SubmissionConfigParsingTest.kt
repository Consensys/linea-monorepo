package linea.coordinator.config.v2

import com.sksamuel.hoplite.Masked
import linea.coordinator.config.v2.toml.L1SubmissionToml
import linea.coordinator.config.v2.toml.SignerConfigToml
import linea.coordinator.config.v2.toml.parseConfig
import linea.kotlin.decodeHex
import linea.kotlin.toURL
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import kotlin.time.Duration.Companion.days
import kotlin.time.Duration.Companion.hours
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds

class L1SubmissionConfigParsingTest {
  companion object {
    val toml = """
    [l1-submission]
    disabled = true
    [l1-submission.dynamic-gas-price-cap]
    disabled = true
    [l1-submission.dynamic-gas-price-cap.gas-price-cap-calculation]
    adjustment-constant = 25
    blob-adjustment-constant = 25
    finalization-target-max-delay = "PT32H"
    base-fee-per-gas-percentile-window = "P7D"
    base-fee-per-gas-percentile-window-leeway = "PT10M"
    base-fee-per-gas-percentile = 10
    gas-price-caps-check-coefficient = 0.9
    [l1-submission.dynamic-gas-price-cap.fee-history-fetcher]
    fetch-interval = "PT1S"
    max-block-count = 1000
    reward-percentiles = [10, 20, 30, 40, 50, 60, 70, 80, 90, 100]
    [l1-submission.dynamic-gas-price-cap.fee-history-storage]
    storage-period = "P10D"

    [l1-submission.fallback-gas-price]
    fee-history-block-count = 10
    fee-history-reward-percentile = 15

    [l1-submission.blob]
    disabled = false
    endpoint = "http://l1-el-node:8545"
    submission-delay = "PT1S"
    submission-tick-interval = "PT10S"
    max-submission-transactions-per-tick = 10
    target-blobs-per-transaction=9
    db-max-blobs-to-return = 100
    [l1-submission.blob.gas]
    gas-limit = 10_000_000
    max-fee-per-gas-cap = 100_000_000_000
    max-fee-per-blob-gas-cap = 100_000_000
    # Note: prefixed with "fallback-", used when dynamic gas price is disabled or DB is not populated yet
    [l1-submission.blob.gas.fallback]
    priority-fee-per-gas-upper-bound = 20_000_000_000 # 20 GWEI
    priority-fee-per-gas-lower-bound = 2_000_000_000 # 2 GWEI

    [l1-submission.blob.signer]
    # Web3j/Web3signer
    type = "Web3j"

    # The account with this private key is in genesis file
    [l1-submission.blob.signer.web3j]
    private-key = "0x0000000000000000000000000000000000000000000000000000000000000001"

    [l1-submission.blob.signer.web3signer]
    endpoint = "http://web3signer:9000"
    max-pool-size = 10
    keep-alive = true
    public-key = "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"

    [l1-submission.aggregation]
    disabled = false
    endpoint = "http://l1-el-node:8545"
    submission-delay = "PT2S"
    submission-tick-interval = "PT12S"
    max-submissions-per-tick = 10
    [l1-submission.aggregation.gas]
    gas-limit = 10_000_001
    max-fee-per-gas-cap = 100_000_000_001

    [l1-submission.aggregation.gas.fallback]
    # Note: prefixed with "fallback-", used when dynamic gas price is disabled or DB is not populated yet
    priority-fee-per-gas-upper-bound = 20_000_000_001 # 20 GWEI
    priority-fee-per-gas-lower-bound = 2_000_000_001 # 2 GWEI

    [l1-submission.aggregation.signer]
    # Web3j/Web3signer
    type = "Web3signer"

    [l1-submission.aggregation.signer.web3j]
    private-key = "0x0000000000000000000000000000000000000000000000000000000000000002"

    [l1-submission.aggregation.signer.web3signer]
    endpoint = "http://web3signer:9000"
    max-pool-size = 10
    keep-alive = true
    public-key = "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002"
    """.trimIndent()

    val config =
      L1SubmissionToml(
        disabled = true,
        dynamicGasPriceCap = L1SubmissionToml.DynamicGasPriceCapToml(
          disabled = true,
          gasPriceCapCalculation = L1SubmissionToml.DynamicGasPriceCapToml.GasPriceCapCalculationToml(
            adjustmentConstant = 25,
            blobAdjustmentConstant = 25,
            finalizationTargetMaxDelay = 32.hours,
            baseFeePerGasPercentileWindow = 7.days,
            baseFeePerGasPercentileWindowLeeway = 10.minutes,
            baseFeePerGasPercentile = 10u,
            gasPriceCapsCheckCoefficient = 0.9
          ),
          feeHistoryFetcher = L1SubmissionToml.DynamicGasPriceCapToml.FeeHistoryFetcherConfig(
            fetchInterval = 1.seconds,
            maxBlockCount = 1000u,
            rewardPercentiles = listOf(10, 20, 30, 40, 50, 60, 70, 80, 90, 100).map { it.toUInt() }
          ),
          feeHistoryStorage = L1SubmissionToml.DynamicGasPriceCapToml.FeeHistoryStorageConfig(
            storagePeriod = 10.days
          )
        ),
        fallbackGasPrice = L1SubmissionToml.FallbackGasPriceToml(
          feeHistoryBlockCount = 10u,
          feeHistoryRewardPercentile = 15u
        ),
        blob = L1SubmissionToml.BlobSubmissionConfigToml(
          disabled = false,
          endpoint = "http://l1-el-node:8545".toURL(),
          submissionDelay = 1.seconds,
          submissionTickInterval = 10.seconds,
          maxSubmissionTransactionsPerTick = 10u,
          targetBlobsPerTransaction = 9u,
          dbMaxBlobsToReturn = 100u,
          gas = L1SubmissionToml.GasConfigToml(
            gasLimit = 10_000_000u,
            maxFeePerGasCap = 100_000_000_000u,
            maxFeePerBlobGasCap = 100_000_000u,
            fallback = L1SubmissionToml.GasConfigToml.FallbackGasConfig(
              priorityFeePerGasUpperBound = 20_000_000_000u,
              priorityFeePerGasLowerBound = 2_000_000_000u
            )
          ),
          signer = SignerConfigToml(
            type = SignerConfigToml.SignerType.WEB3J,
            web3j = SignerConfigToml.Web3jConfig(
              privateKey = Masked("0x0000000000000000000000000000000000000000000000000000000000000001")
            ),
            web3signer = SignerConfigToml.Web3SignerConfig(
              endpoint = "http://web3signer:9000".toURL(),
              publicKey = (
                "0000000000000000000000000000000000000000000000000000000000000000" +
                  "0000000000000000000000000000000000000000000000000000000000000001"
                ).decodeHex(),
              maxPoolSize = 10,
              keepAlive = true
            )
          )
        ),
        aggregation = L1SubmissionToml.AggregationSubmissionToml(
          disabled = false,
          endpoint = "http://l1-el-node:8545".toURL(),
          submissionDelay = 2.seconds,
          submissionTickInterval = 12.seconds,
          maxSubmissionsPerTick = 10u,
          gas = L1SubmissionToml.GasConfigToml(
            gasLimit = 10_000_001u,
            maxFeePerGasCap = 100_000_000_001u,
            maxFeePerBlobGasCap = null,
            fallback = L1SubmissionToml.GasConfigToml.FallbackGasConfig(
              priorityFeePerGasUpperBound = 20_000_000_001u,
              priorityFeePerGasLowerBound = 2_000_000_001u
            )
          ),
          signer = SignerConfigToml(
            type = SignerConfigToml.SignerType.WEB3SIGNER,
            web3j = SignerConfigToml.Web3jConfig(
              privateKey = Masked("0x0000000000000000000000000000000000000000000000000000000000000002")
            ),
            web3signer = SignerConfigToml.Web3SignerConfig(
              endpoint = "http://web3signer:9000".toURL(),
              publicKey = (
                "0000000000000000000000000000000000000000000000000000000000000000" +
                  "0000000000000000000000000000000000000000000000000000000000000002"
                ).decodeHex(),
              maxPoolSize = 10,
              keepAlive = true
            )
          )
        )
      )
  }

  data class WrapperConfig(
    val l1Submission: L1SubmissionToml
  )

  @Test
  fun `should parse full state manager config`() {
    assertThat(parseConfig<WrapperConfig>(toml).l1Submission)
      .isEqualTo(config)
  }
}

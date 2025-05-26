package linea.coordinator.config.v2.toml

import linea.domain.BlockParameter
import net.consensys.zkevm.coordinator.app.config.MessageAnchoringConfig
import java.net.URL
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

data class MessageAnchoringConfigToml(
  var disabled: Boolean = false,
  val l1Endpoint: URL? = null, // shall default to L1 endpoint
  val l1HighestBlockTag: BlockParameter = BlockParameter.Tag.FINALIZED,
  val l1RequestRetries: RequestRetriesToml = RequestRetriesToml.endlessRetry(
    backoffDelay = 1.seconds,
    failuresWarningThreshold = 3u
  ),
  val l1EventScrapping: L1EventScrapping = L1EventScrapping(),
  val l2Endpoint: URL? = null,
  val l2HighestBlockTag: BlockParameter = BlockParameter.Tag.LATEST,
  val l2RequestRetries: RequestRetriesToml = RequestRetriesToml.endlessRetry(
    backoffDelay = 1.seconds,
    failuresWarningThreshold = 3u
  ),
  val anchoringTickInterval: Duration = 2.seconds,
  val messageQueueCapacity: Int = 10_000,
  val maxMessagesToAnchorPerL2Transaction: Int = 100,
  val signer: SignerConfigToml,
  val gas: GasConfig = GasConfig()
) {
  init {
    require(messageQueueCapacity > 0) {
      "messageQueueCapacity must be greater than 0"
    }
    require(maxMessagesToAnchorPerL2Transaction >= 1) {
      "maxMessagesToAnchorPerL2Transaction=$maxMessagesToAnchorPerL2Transaction be equal or greater than 1"
    }

    require(anchoringTickInterval >= 1.milliseconds) {
      "anchoringTickInterval must be equal or greater than 1ms"
    }
  }

  data class L1EventScrapping(
    val pollingInterval: Duration = 2.seconds,
    val pollingTimeout: Duration = 5.seconds,
    val ethLogsSearchSuccessBackoffDelay: Duration = 1.milliseconds,
    val ethLogsSearchBlockChunkSize: Int = 1000
  ) {
    init {

      require(pollingInterval >= 1.milliseconds) {
        "pollingInterval=$pollingInterval must be equal or greater than 1ms"
      }
      require(pollingTimeout >= 1.milliseconds) {
        "pollingTimeout=$pollingTimeout must be equal or greater than 1ms"
      }
      require(ethLogsSearchSuccessBackoffDelay >= 1.milliseconds) {
        "ethLogsSearchSuccessBackoffDelay=$ethLogsSearchSuccessBackoffDelay must be equal or greater than 1ms"
      }
      require(ethLogsSearchBlockChunkSize >= 1) {
        "ethLogsSearchBlockChunkSize=$ethLogsSearchBlockChunkSize must be equal or greater than 1"
      }
    }
  }

  data class GasConfig(
    val maxFeePerGasCap: ULong = 100_000uL,
    val gasLimit: ULong = 1_000_000_000uL,
    val feeHistoryBlockCount: UInt = 4u,
    val feeHistoryRewardPercentile: UInt = 4u
  )

  fun reified(
    l1DefaultEndpoint: URL,
    l2DefaultEndpoint: URL
  ): MessageAnchoringConfig {
    return MessageAnchoringConfig(
      disabled = disabled,
      l1Endpoint = l1Endpoint ?: l1DefaultEndpoint,
      l2Endpoint = l2Endpoint ?: l2DefaultEndpoint,
      l1HighestBlockTag = l1HighestBlockTag,
      l2HighestBlockTag = l2HighestBlockTag,
      l1RequestRetryConfig = l1RequestRetries.asDomain,
      l2RequestRetryConfig = l2RequestRetries.asDomain,
      l1EventPollingInterval = l1EventScrapping.pollingInterval,
      l1EventPollingTimeout = l1EventScrapping.pollingTimeout,
      l1SuccessBackoffDelay = l1EventScrapping.ethLogsSearchSuccessBackoffDelay,
      l1EventSearchBlockChunk = l1EventScrapping.ethLogsSearchBlockChunkSize.toUInt(),
      anchoringTickInterval = anchoringTickInterval,
      messageQueueCapacity = messageQueueCapacity.toUInt(),
      maxMessagesToAnchorPerL2Transaction = maxMessagesToAnchorPerL2Transaction.toUInt()
    )
  }
}

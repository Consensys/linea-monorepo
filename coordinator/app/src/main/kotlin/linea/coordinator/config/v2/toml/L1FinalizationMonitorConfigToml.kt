package linea.coordinator.config.v2.toml

import linea.coordinator.config.v2.L1FinalizationMonitorConfig
import linea.domain.BlockParameter
import java.net.URL
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

data class L1FinalizationMonitorConfigToml(
  val l1Endpoint: URL?,
  val l2Endpoint: URL?,
  val l1PollingInterval: Duration = 6.seconds,
  val l1QueryBlockTag: BlockParameter.Tag = BlockParameter.Tag.FINALIZED,
) {
  fun reified(
    l1DefaultEndpoint: URL?,
    l2DefaultEndpoint: URL?,
  ): L1FinalizationMonitorConfig {
    return L1FinalizationMonitorConfig(
      l1Endpoint = this.l1Endpoint ?: l1DefaultEndpoint ?: throw AssertionError("l1Endpoint missing"),
      l2Endpoint = this.l2Endpoint ?: l2DefaultEndpoint ?: throw AssertionError("l2Endpoint missing"),
      l1PollingInterval = this.l1PollingInterval,
      l1QueryBlockTag = this.l1QueryBlockTag,
    )
  }
}

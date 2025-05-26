package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.L1FinalizationMonitor
import linea.coordinator.config.v2.toml.RequestRetriesToml
import linea.coordinator.config.v2.toml.parseConfig
import linea.domain.BlockParameter
import linea.kotlin.toURL
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import kotlin.time.Duration.Companion.seconds

class L1FinalizationMonitorParsingTest {
  companion object {
    val toml = """
    [l1-finalization-monitor]
    l1-endpoint = "http://l1-el-node:8545"
    l2-endpoint = "http://sequencer:8545"
    l1-polling-interval = "PT1S"
    l1-query-block-tag="FINALIZED"

    [l1-finalization-monitor.request-retries]
    max-retries = 3
    backoff-delay = "PT1S"
    failures-warning-threshold = 2
    """.trimIndent()

    val config = L1FinalizationMonitor(
      l1Endpoint = "http://l1-el-node:8545".toURL(),
      l2Endpoint = "http://sequencer:8545".toURL(),
      l1PollingInterval = 1.seconds,
      l1QueryBlockTag = BlockParameter.Tag.FINALIZED,
      requestRetries = RequestRetriesToml(
        maxRetries = 3u,
        backoffDelay = 1.seconds,
        failuresWarningThreshold = 2u
      )
    )
  }

  data class WrapperConfig(
    val l1FinalizationMonitor: L1FinalizationMonitor
  )

  @Test
  fun `should parse full state manager config`() {
    assertThat(parseConfig<WrapperConfig>(toml).l1FinalizationMonitor)
      .isEqualTo(config)
  }
}

package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.RequestRetriesToml
import linea.coordinator.config.v2.toml.StateManagerToml
import linea.coordinator.config.v2.toml.parseConfig
import linea.kotlin.toURL
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import kotlin.time.Duration.Companion.seconds

class StateManagerParsingTest {
  companion object {
    val toml = """
      [state-manager]
      version = "2.2.0"
      endpoints = ["http://shomei:8888/"]
      request-limit-per-endpoint = 3
      [state-manager.request-retries]
      max-retries = 5
      backoff-delay = "PT2S"
      failures-warning-threshold = 2
    """.trimIndent()

    val config = StateManagerToml(
      version = "2.2.0",
      endpoints = listOf("http://shomei:8888/".toURL()),
      requestLimitPerEndpoint = 3u,
      requestRetries = RequestRetriesToml(
        maxRetries = 5u,
        backoffDelay = 2.seconds,
        failuresWarningThreshold = 2u
      )
    )
  }

  data class WrapperConfig(
    val stateManager: StateManagerToml
  )

  @Test
  fun `should parse full state manager config`() {
    assertThat(parseConfig<WrapperConfig>(toml).stateManager)
      .isEqualTo(config)
  }
}

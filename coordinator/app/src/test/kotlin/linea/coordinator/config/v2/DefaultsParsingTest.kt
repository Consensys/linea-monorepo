package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.DefaultsToml
import linea.coordinator.config.v2.toml.parseConfig
import linea.kotlin.toURL
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class DefaultsParsingTest {
  companion object {
    val toml = """
    [defaults]
    l1-endpoint = "http://l1-el-node:8545"
    l2-endpoint = "http://sequencer:8545"
    """.trimIndent()

    val config = DefaultsToml(
      l1Endpoint = "http://l1-el-node:8545".toURL(),
      l2Endpoint = "http://sequencer:8545".toURL()
    )

    val tomlMinimal = """
    """.trimIndent()

    val configMinimal = DefaultsToml(
      l1Endpoint = null,
      l2Endpoint = null
    )
  }
  internal data class WrapperConfig(val defaults: DefaultsToml = DefaultsToml())

  @Test
  fun `should parse defaults full configs`() {
    assertThat(parseConfig<WrapperConfig>(toml).defaults).isEqualTo(config)
  }

  @Test
  fun `should parse defaults minimal configs`() {
    assertThat(parseConfig<WrapperConfig>(tomlMinimal).defaults).isEqualTo(configMinimal)
  }
}

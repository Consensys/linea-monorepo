package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.ApiConfigToml
import linea.coordinator.config.v2.toml.parseConfig
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class ApiConfigParsingTest {
  companion object {
    val toml = """
      [api]
      observability = 9545
    """.trimIndent()

    val config = ApiConfigToml(
      observabilityPort = 9545u
    )
  }

  data class WrapperConfig(
    val api: ApiConfigToml
  )

  @Test
  fun `should parse full state manager config`() {
    assertThat(parseConfig<WrapperConfig>(toml).api).isEqualTo(config)
  }
}

package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.DataBaseToml
import linea.coordinator.config.v2.toml.parseConfig
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class DataBaseConfigParsingTest {
  companion object {
    val toml = """
      [database]
      hostname = "localhost"
      port = "5432"
      username = "someuser"
      password = "somepassword"
      schema = "linea_coordinator"
      read_pool_size = 10
      read_pipelining_limit = 11
      transactional_pool_size = 12
    """.trimIndent()

    val config = DataBaseToml(
      hostname = "localhost",
      username = "someuser",
      password = "somepassword",
      schema = "linea_coordinator",
      readPoolSize = 10,
      readPipeliningLimit = 11,
      transactionalPoolSize = 12,
      port = 5432
    )
  }

  data class WrapperConfig(
    val database: DataBaseToml
  )

  @Test
  fun `should parse full state manager config`() {
    assertThat(parseConfig<WrapperConfig>(toml).database).isEqualTo(config)
  }
}

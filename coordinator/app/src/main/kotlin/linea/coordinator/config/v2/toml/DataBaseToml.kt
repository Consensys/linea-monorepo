package linea.coordinator.config.v2.toml

data class DataBaseToml(
  val hostname: String,
  val port: Int,
  val username: String,
  val password: String,
  val schema: String,
  val readPoolSize: Int = 10,
  val readPipeliningLimit: Int = 10,
  val transactionalPoolSize: Int = 10
)

package linea.coordinator.config.v2.toml

import linea.coordinator.config.v2.ApiConfig

data class ApiConfigToml(
  val observabilityPort: UInt = 9545u,
  val jsonRpcPort: UInt = 0u,
  val jsonRpcPath: String = "/",
) {
  fun reified(): ApiConfig {
    return ApiConfig(observabilityPort = observabilityPort, jsonRpcPort = jsonRpcPort, jsonRpcPath = jsonRpcPath)
  }
}

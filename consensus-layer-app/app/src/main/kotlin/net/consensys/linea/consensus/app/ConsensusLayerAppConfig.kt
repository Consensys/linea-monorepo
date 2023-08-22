package net.consensys.linea.consensus.app

import com.sksamuel.hoplite.ConfigAlias
import net.consensys.linea.forkchoicestate.api.ForkChoiceStateApiConfig
import java.net.URL
import java.nio.file.Path
import java.time.Duration

data class ExecutionClientConfig(
  val url: URL,
  @ConfigAlias("jwt-secret-file") val jwtSecretFile: Path
)

data class ForkChoiceSource(val url: URL, val pollingInterval: Duration)

data class ApiConfig(
  val observabilityPort: UInt,
  val forkChoiceProvider: ForkChoiceStateApiConfig?
)

data class ConsensusLayerAppConfig(
  val forkChoiceSource: ForkChoiceSource,
  val api: ApiConfig,
  val executionClient: ExecutionClientConfig
)

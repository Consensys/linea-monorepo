package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.CoordinatorConfigFileToml
import linea.coordinator.config.v2.toml.parseConfig
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class CoordinatorConfigTest {
  val toml = """
    ${DefaultsParsingTest.toml}
    ${ProtocolParsingTest.toml}
    ${ConflationParsingTest.toml}
    ${ProverParsingTest.toml}
    ${TracesParsingTest.toml}
    ${StateManagerParsingTest.toml}
    ${Type2StateProofProviderParsingTest.toml}
    ${L1FinalizationMonitorParsingTest.toml}
    ${L1SubmissionConfigParsingTest.toml}
    ${MessageAnchoringConfigParsingTest.toml}
    ${L2NetWorkingGasPricingConfigParsingTest.toml}
    ${DataBaseConfigParsingTest.toml}
    ${ApiConfigParsingTest.toml}
  """.trimIndent()
  val config = CoordinatorConfigFileToml(
    defaults = DefaultsParsingTest.config,
    protocol = ProtocolParsingTest.config,
    conflation = ConflationParsingTest.config,
    prover = ProverParsingTest.config,
    traces = TracesParsingTest.config,
    stateManager = StateManagerParsingTest.config,
    type2StateProofProvider = Type2StateProofProviderParsingTest.config,
    l1FinalizationMonitor = L1FinalizationMonitorParsingTest.config,
    l1Submission = L1SubmissionConfigParsingTest.config,
    messageAnchoring = MessageAnchoringConfigParsingTest.expectedConfig,
    l2NetworkGasPricing = L2NetWorkingGasPricingConfigParsingTest.expectedConfig,
    database = DataBaseConfigParsingTest.config,
    api = ApiConfigParsingTest.config
  )

  @Test
  fun `should parse full configs`() {
    assertThat(parseConfig<CoordinatorConfigFileToml>(toml)).isEqualTo(config)
  }
}

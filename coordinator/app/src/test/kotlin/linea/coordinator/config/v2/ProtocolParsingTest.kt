package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.ProtocolToml
import linea.coordinator.config.v2.toml.parseConfig
import linea.domain.BlockParameter.Companion.toBlockParameter
import linea.kotlin.decodeHex
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class ProtocolParsingTest {
  companion object {
    val toml = """
    [protocol.genesis]
    genesis-state-root-hash = "0x072ead6777750dc20232d1cee8dc9a395c2d350df4bbaa5096c6f59b214dcecd"
    genesis-shnarf = "0x47452a1b9ebadfe02bdd02f580fa1eba17680d57eec968a591644d05d78ee84f"
    [protocol.l1]
    contract-address = "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9"
    contract-deployment-block-number = 2_000_000
    [protocol.l2]
    contract-address = "0xe537D669CA013d86EBeF1D64e40fC74CADC91987"
    contract-deployment-block-number = 1
    """.trimIndent()

    val config = ProtocolToml(
      genesis = ProtocolToml.Genesis(
        genesisStateRootHash = "0x072ead6777750dc20232d1cee8dc9a395c2d350df4bbaa5096c6f59b214dcecd".decodeHex(),
        genesisShnarf = "0x47452a1b9ebadfe02bdd02f580fa1eba17680d57eec968a591644d05d78ee84f".decodeHex()
      ),
      l1 = ProtocolToml.LayerConfig(
        contractAddress = "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9",
        contractDeploymentBlockNumber = 2_000_000UL.toBlockParameter()
      ),
      l2 = ProtocolToml.LayerConfig(
        contractAddress = "0xe537D669CA013d86EBeF1D64e40fC74CADC91987",
        contractDeploymentBlockNumber = 1UL.toBlockParameter()
      )
    )
  }

  @Test
  fun `should parse protocol configs`() {
    data class WrapperConfig(
      val protocol: ProtocolToml
    )
    assertThat(parseConfig<WrapperConfig>(toml).protocol)
      .isEqualTo(config)
  }
}

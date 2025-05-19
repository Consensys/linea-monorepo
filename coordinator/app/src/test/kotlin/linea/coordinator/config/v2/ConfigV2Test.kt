package linea.coordinator.config.v2

import linea.coordinator.config.v2.toml.Protocol
import linea.coordinator.config.v2.toml.parseConfig
import linea.kotlin.decodeHex
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import kotlin.time.Duration.Companion.seconds


class ConfigV2Test {
    @Test
    fun `should parse Kotlin typings`() {
        val toml = """
      l1-contract-address = "0x1234567890abcdef1234567890abcdef1234567890abcdef"
      l2-contract-address = "0x1234567890abcdef1234567890abcdef1234567890abcfff"
      duration = "PT2S"
      [genesis]
      genesis-state-root-hash = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdaa"
      genesis-shnarf = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdbb"
    """.trimIndent()
        assertThat(parseConfig<Protocol>(toml)).isEqualTo(
            Protocol(
                genesis = Protocol.Genesis(
                    genesisStateRootHash = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdaa".decodeHex(),
                    genesisShnarf = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdbb".decodeHex()
                ),
                l1ContractAddress = "0x1234567890abcdef1234567890abcdef1234567890abcdef",
                l2ContractAddress = "0x1234567890abcdef1234567890abcdef1234567890abcfff",
                duration = 2.seconds
            )
        )
    }
}

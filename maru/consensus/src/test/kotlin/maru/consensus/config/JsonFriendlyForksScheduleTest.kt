/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.consensus.config

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.ExperimentalHoplite
import com.sksamuel.hoplite.json.JsonPropertySource
import maru.consensus.ElFork
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.delegated.ElDelegatedConsensus
import maru.consensus.qbft.QbftConsensusConfig
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test

@OptIn(ExperimentalHoplite::class)
class JsonFriendlyForksScheduleTest {
  private val genesisConfig =
    """
    {
      "config": {
        "2": {
          "type": "delegated",
          "blockTimeSeconds": 4
        },
        "4": {
          "type": "qbft",
          "blockTimeSeconds": 6,
          "feeRecipient": "0x0000000000000000000000000000000000000000",
          "elFork": "Prague"
        }
      }
    }
    """.trimIndent()

  @Test
  fun genesisFileIsParseable() {
    val config =
      parseJsonConfig<JsonFriendlyForksSchedule>(
        genesisConfig,
      )
    val expectedDelegatedConsensusMap =
      mapOf(
        "type" to "delegated",
        "blockTimeSeconds" to "4",
      )
    val expectedQbftMap =
      mapOf(
        "type" to "qbft",
        "blockTimeSeconds" to "6",
        "feeRecipient" to "0x0000000000000000000000000000000000000000",
        "elFork" to "Prague",
      )
    assertThat(config).isEqualTo(
      JsonFriendlyForksSchedule(
        mapOf(
          "2" to expectedDelegatedConsensusMap,
          "4" to expectedQbftMap,
        ),
      ),
    )
  }

  @Test
  fun genesisFileIsConvertableToDomain() {
    val config =
      parseJsonConfig<JsonFriendlyForksSchedule>(
        genesisConfig,
      ).domainFriendly()
    assertThat(config).isEqualTo(
      ForksSchedule(
        setOf(
          ForkSpec(
            timestampSeconds = 2,
            blockTimeSeconds = 2,
            ElDelegatedConsensus.ElDelegatedConfig,
          ),
          ForkSpec(
            timestampSeconds = 4,
            blockTimeSeconds = 6,
            configuration =
              QbftConsensusConfig(
                feeRecipient = Bytes.fromHexString("0x0000000000000000000000000000000000000000").toArray(),
                elFork = ElFork.Prague,
              ),
          ),
        ),
      ),
    )
  }

  @Test
  fun parserFailsIfSomeConfigurationIsMissing() {
    val invalidConfiguration =
      """
      {
        "config": {
          "4": {
            "type": "qbft",
            "feeRecipient": "0x0000000000000000000000000000000000000000",
            "elFork": "Prague"
          }
        }
      }
      """.trimIndent()
    assertThatThrownBy {
      parseJsonConfig<JsonFriendlyForksSchedule>(
        invalidConfiguration,
      ).domainFriendly()
    }.isInstanceOf(Exception::class.java)
  }

  private inline fun <reified T : Any> parseJsonConfig(json: String): T =
    ConfigLoaderBuilder
      .default()
      .withExplicitSealedTypes()
      .addSource(JsonPropertySource(json))
      .build()
      .loadConfigOrThrow<T>()
}

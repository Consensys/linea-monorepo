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
import maru.consensus.dummy.DummyConsensusConfig
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

@OptIn(ExperimentalHoplite::class)
class JsonFriendlyForksScheduleTest {
  private val genesisConfig =
    """
    {
      "config": {
        "0": {
          "type": "dummy",
          "blockTimeSeconds": 1,
          "feeRecipient": "0x0000000000000000000000000000000000000000",
          "elFork": "Prague"
        },
        "2": {
          "type": "delegated",
          "blockTimeSeconds": 4
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
    val expectedDummyConsensusMap =
      mapOf(
        "type" to "dummy",
        "blockTimeSeconds" to "1",
        "feeRecipient" to "0x0000000000000000000000000000000000000000",
        "elFork" to "Prague",
      )
    val expectedDelegatedConsensusMap =
      mapOf(
        "type" to "delegated",
        "blockTimeSeconds" to "4",
      )
    assertThat(config).isEqualTo(
      JsonFriendlyForksSchedule(
        mapOf(
          "0" to expectedDummyConsensusMap,
          "2" to expectedDelegatedConsensusMap,
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
            0,
            blockTimeSeconds = 2,
            DummyConsensusConfig(
              feeRecipient = Bytes.fromHexString("0x0000000000000000000000000000000000000000").toArray(),
              elFork = ElFork.Prague,
            ),
          ),
          ForkSpec(
            2,
            blockTimeSeconds = 2,
            ElDelegatedConsensus.ElDelegatedConfig,
          ),
        ),
      ),
    )
  }

  private inline fun <reified T : Any> parseJsonConfig(json: String): T =
    ConfigLoaderBuilder
      .default()
      .withExplicitSealedTypes()
      .addSource(JsonPropertySource(json))
      .build()
      .loadConfigOrThrow<T>()
}

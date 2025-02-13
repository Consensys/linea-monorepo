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
package maru.app.config

import com.sksamuel.hoplite.ExperimentalHoplite
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.dummy.DummyConsensusConfig
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

@OptIn(ExperimentalHoplite::class)
class JsonFriendlinessTest {
  @Test
  fun genesisFileIsParseable() {
    val config =
      Utils.parseJsonConfig<JsonFriendlyForksSchedule>(
        """
        {
          "config": {
            "0": {
              "type": "dummy",
              "blockTimeMillis": 1000,
              "feeRecipient": "0x0000000000000000000000000000000000000000"
            }
          }
        }
        """.trimIndent(),
      )
    val expectedObject =
      mapOf(
        "type" to "dummy",
        "blockTimeMillis" to "1000",
        "feeRecipient" to "0x0000000000000000000000000000000000000000",
      )
    assertThat(config).isEqualTo(
      JsonFriendlyForksSchedule(mapOf("0" to expectedObject)),
    )
  }

  @Test
  fun genesisFileIsConvertableToDomain() {
    val config =
      Utils
        .parseJsonConfig<JsonFriendlyForksSchedule>(
          """
          {
            "config": {
              "0": {
                "type": "dummy",
                "blockTimeMillis": 1000,
                "feeRecipient": "0x0000000000000000000000000000000000000000"
              }
            }
          }
          """.trimIndent(),
        ).domainFriendly()
    assertThat(config).isEqualTo(
      ForksSchedule(
        setOf(
          ForkSpec(
            0u,
            DummyConsensusConfig(
              blockTimeMillis = 1000u,
              feeRecipient = Bytes.fromHexString("0x0000000000000000000000000000000000000000").toArray(),
            ),
          ),
        ),
      ),
    )
  }
}

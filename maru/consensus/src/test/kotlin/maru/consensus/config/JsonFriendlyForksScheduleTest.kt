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

import maru.consensus.ElFork
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.delegated.ElDelegatedConsensus
import maru.consensus.qbft.QbftConsensusConfig
import maru.core.Validator
import maru.extensions.fromHexToByteArray
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test

class JsonFriendlyForksScheduleTest {
  private val genesisConfig =
    """
    {
      "chainId": 1337,
      "config": {
        "2": {
          "type": "delegated",
          "blockTimeSeconds": 4
        },
        "4": {
          "type": "qbft",
          "validatorSet": ["0x1b9abeec3215d8ade8a33607f2cf0f4f60e5f0d0"],
          "blockTimeSeconds": 6,
          "feeRecipient": "0x0000000000000000000000000000000000000000",
          "elFork": "Prague"
        }
      }
    }
    """.trimIndent()

  @Test
  fun genesisFileIsConvertableToDomain() {
    val config =
      Utils
        .parseBeaconChainConfig(
          genesisConfig,
        ).domainFriendly()
    assertThat(config).isEqualTo(
      ForksSchedule(
        1337u,
        setOf(
          ForkSpec(
            timestampSeconds = 2,
            blockTimeSeconds = 4,
            ElDelegatedConsensus.ElDelegatedConfig,
          ),
          ForkSpec(
            timestampSeconds = 4,
            blockTimeSeconds = 6,
            configuration =
              QbftConsensusConfig(
                feeRecipient = Bytes.fromHexString("0x0000000000000000000000000000000000000000").toArray(),
                validatorSet = setOf(Validator("0x1b9abeec3215d8ade8a33607f2cf0f4f60e5f0d0".fromHexToByteArray())),
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
      Utils.parseBeaconChainConfig(
        invalidConfiguration,
      )
    }.isInstanceOf(Exception::class.java)
  }
}

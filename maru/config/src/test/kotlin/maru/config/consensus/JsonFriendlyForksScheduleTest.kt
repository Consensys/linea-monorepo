/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.config.consensus

import maru.config.Utils
import maru.config.consensus.delegated.ElDelegatedConfig
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.core.Validator
import maru.extensions.fromHexToByteArray
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
            ElDelegatedConfig,
          ),
          ForkSpec(
            timestampSeconds = 4,
            blockTimeSeconds = 6,
            configuration =
              QbftConsensusConfig(
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

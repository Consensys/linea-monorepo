/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.serialization

import maru.config.consensus.ElFork
import maru.config.consensus.qbft.DifficultyAwareQbftConfig
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.core.Validator
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class DifficultyAwareQbftConfigSerializerTest {
  private val serializer = DifficultyAwareQbftConfigSerializer

  @Test
  fun `serialization is deterministic for same input`() {
    val validatorSet = DataGenerators.randomValidators()

    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = validatorSet,
        elFork = ElFork.Prague,
      )

    val difficultyAwareQbftConfig =
      DifficultyAwareQbftConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 1000UL,
      )

    val serialized1 = serializer.serialize(difficultyAwareQbftConfig)
    val serialized2 = serializer.serialize(difficultyAwareQbftConfig)

    assertThat(serialized1).isEqualTo(serialized2)
  }

  @Test
  fun `serialization changes when terminal total difficulty changes`() {
    val validator = Validator(ByteArray(20) { 0x01 })
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator),
        elFork = ElFork.Cancun,
      )

    val config1 =
      DifficultyAwareQbftConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 1000UL,
      )

    val config2 =
      DifficultyAwareQbftConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 2000UL,
      )

    val serialized1 = serializer.serialize(config1)
    val serialized2 = serializer.serialize(config2)

    assertThat(serialized1).isNotEqualTo(serialized2)
  }

  @Test
  fun `serialization changes when TTD changes to maximum value`() {
    val validator = Validator(ByteArray(20) { 0x01 })
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator),
        elFork = ElFork.Cancun,
      )

    val normalTtdConfig =
      DifficultyAwareQbftConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 1000UL,
      )

    val maxTtdConfig =
      DifficultyAwareQbftConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = ULong.MAX_VALUE,
      )

    val serializedNormal = serializer.serialize(normalTtdConfig)
    val serializedMax = serializer.serialize(maxTtdConfig)

    assertThat(serializedNormal).isNotEqualTo(serializedMax)
  }

  @Test
  fun `serialization changes when ElFork changes`() {
    val validator = Validator(ByteArray(20) { 0x01 })

    val configParis =
      DifficultyAwareQbftConfig(
        postTtdConfig =
          QbftConsensusConfig(
            validatorSet = setOf(validator),
            elFork = ElFork.Paris,
          ),
        terminalTotalDifficulty = 500UL,
      )

    val configCancun =
      DifficultyAwareQbftConfig(
        postTtdConfig =
          QbftConsensusConfig(
            validatorSet = setOf(validator),
            elFork = ElFork.Cancun,
          ),
        terminalTotalDifficulty = 500UL,
      )

    val configPrague =
      DifficultyAwareQbftConfig(
        postTtdConfig =
          QbftConsensusConfig(
            validatorSet = setOf(validator),
            elFork = ElFork.Prague,
          ),
        terminalTotalDifficulty = 500UL,
      )

    val serializedParis = serializer.serialize(configParis)
    val serializedCancun = serializer.serialize(configCancun)
    val serializedPrague = serializer.serialize(configPrague)

    // Each fork should produce different serialization
    assertThat(serializedParis).isNotEqualTo(serializedCancun)
    assertThat(serializedCancun).isNotEqualTo(serializedPrague)
    assertThat(serializedParis).isNotEqualTo(serializedPrague)
  }
}

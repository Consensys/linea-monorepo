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
import maru.config.consensus.delegated.ElDelegatedConfig
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.core.Validator
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test

class ElDelegatedConfigSerializerTest {
  private val serializer = ElDelegatedConfigSerializer

  @Test
  fun `serialization is deterministic for same input`() {
    val validatorSet = DataGenerators.randomValidators()

    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = validatorSet,
        elFork = ElFork.Prague,
      )

    val elDelegatedConfig =
      ElDelegatedConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 1000UL,
      )

    val serialized1 = serializer.serialize(elDelegatedConfig)
    val serialized2 = serializer.serialize(elDelegatedConfig)

    assertThat(serialized1).isEqualTo(serialized2)
  }

  @Test
  fun `serialization changes when terminal total difficulty changes`() {
    val validator = Validator(ByteArray(20) { 0x01 })
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator),
        elFork = ElFork.Shanghai,
      )

    val config1 =
      ElDelegatedConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 1000UL,
      )

    val config2 =
      ElDelegatedConfig(
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
        elFork = ElFork.Shanghai,
      )

    val normalTtdConfig =
      ElDelegatedConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 1000UL,
      )

    val maxTtdConfig =
      ElDelegatedConfig(
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
      ElDelegatedConfig(
        postTtdConfig =
          QbftConsensusConfig(
            validatorSet = setOf(validator),
            elFork = ElFork.Paris,
          ),
        terminalTotalDifficulty = 500UL,
      )

    val configShanghai =
      ElDelegatedConfig(
        postTtdConfig =
          QbftConsensusConfig(
            validatorSet = setOf(validator),
            elFork = ElFork.Shanghai,
          ),
        terminalTotalDifficulty = 500UL,
      )

    val configPrague =
      ElDelegatedConfig(
        postTtdConfig =
          QbftConsensusConfig(
            validatorSet = setOf(validator),
            elFork = ElFork.Prague,
          ),
        terminalTotalDifficulty = 500UL,
      )

    val serializedParis = serializer.serialize(configParis)
    val serializedShanghai = serializer.serialize(configShanghai)
    val serializedPrague = serializer.serialize(configPrague)

    // Each fork should produce different serialization
    assertThat(serializedParis).isNotEqualTo(serializedShanghai)
    assertThat(serializedShanghai).isNotEqualTo(serializedPrague)
    assertThat(serializedParis).isNotEqualTo(serializedPrague)
  }

  @Test
  fun `throws exception when postTtdConfig is not QbftConsensusConfig`() {
    // Create a mock ConsensusConfig that is not QbftConsensusConfig
    val mockConfig = object : maru.consensus.ConsensusConfig {}

    val elDelegatedConfig =
      ElDelegatedConfig(
        postTtdConfig = mockConfig,
        terminalTotalDifficulty = 1000UL,
      )

    assertThatThrownBy {
      serializer.serialize(elDelegatedConfig)
    }.isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("only QbftConsensusConfig serialization is implemented for postTtdConfig!")
  }
}

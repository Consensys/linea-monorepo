/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.serialization

import java.nio.ByteBuffer
import maru.config.consensus.ElFork
import maru.config.consensus.delegated.ElDelegatedConfig
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForkSpec
import maru.core.Validator
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test

class ForkSpecSerializerTest {
  private val serializer = ForkSpecSerializer

  @Test
  fun `can serialize ForkSpec with QbftConsensusConfig`() {
    val validator = Validator(ByteArray(20) { 0x01 })
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator),
        elFork = ElFork.Prague,
      )

    val forkSpec =
      ForkSpec(
        timestampSeconds = 1000000UL,
        blockTimeSeconds = 12U,
        configuration = qbftConfig,
      )

    val serializedData = serializer.serialize(forkSpec)

    // Verify structure: 4 bytes blockTime + 8 bytes timestamp + QbftConfig serialization
    val expectedConfigBytes = QbftConsensusConfigSerializer.serialize(qbftConfig)
    val expectedSize = 4 + 8 + expectedConfigBytes.size

    assertThat(serializedData).hasSize(expectedSize)

    // Verify the components
    val buffer = ByteBuffer.wrap(serializedData)
    assertThat(buffer.getInt()).isEqualTo(12) // blockTimeSeconds
    assertThat(buffer.getLong()).isEqualTo(1000000L) // timestampSeconds

    val remainingBytes = ByteArray(expectedConfigBytes.size)
    buffer.get(remainingBytes)
    assertThat(remainingBytes).isEqualTo(expectedConfigBytes)
  }

  @Test
  fun `can serialize ForkSpec with ElDelegatedConfig`() {
    val validator1 = Validator(ByteArray(20) { 0x01 })
    val validator2 = Validator(ByteArray(20) { 0x02 })
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator1, validator2),
        elFork = ElFork.Shanghai,
      )

    val elDelegatedConfig =
      ElDelegatedConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 5000UL,
      )

    val forkSpec =
      ForkSpec(
        timestampSeconds = 2000000UL,
        blockTimeSeconds = 15U,
        configuration = elDelegatedConfig,
      )

    val serializedData = serializer.serialize(forkSpec)

    // Verify structure: 4 bytes blockTime + 8 bytes timestamp + ElDelegatedConfig serialization
    val expectedConfigBytes = ElDelegatedConfigSerializer.serialize(elDelegatedConfig)
    val expectedSize = 4 + 8 + expectedConfigBytes.size

    assertThat(serializedData).hasSize(expectedSize)

    // Verify the components
    val buffer = ByteBuffer.wrap(serializedData)
    assertThat(buffer.getInt()).isEqualTo(15) // blockTimeSeconds
    assertThat(buffer.getLong()).isEqualTo(2000000L) // timestampSeconds

    val remainingBytes = ByteArray(expectedConfigBytes.size)
    buffer.get(remainingBytes)
    assertThat(remainingBytes).isEqualTo(expectedConfigBytes)
  }

  @Test
  fun `serialization produces different results for different consensus configs`() {
    val validator = Validator(ByteArray(20) { 0x01 })

    // Create identical ForkSpecs with different consensus configs
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator),
        elFork = ElFork.Prague,
      )

    val elDelegatedConfig =
      ElDelegatedConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 1000UL,
      )

    val forkSpecWithQbft =
      ForkSpec(
        timestampSeconds = 1000UL,
        blockTimeSeconds = 12U,
        configuration = qbftConfig,
      )

    val forkSpecWithElDelegated =
      ForkSpec(
        timestampSeconds = 1000UL,
        blockTimeSeconds = 12U,
        configuration = elDelegatedConfig,
      )

    val serializedQbft = serializer.serialize(forkSpecWithQbft)
    val serializedElDelegated = serializer.serialize(forkSpecWithElDelegated)

    // They should produce different serializations
    assertThat(serializedQbft).isNotEqualTo(serializedElDelegated)

    // ElDelegated should be larger due to the additional TTD field
    assertThat(serializedElDelegated.size).isGreaterThan(serializedQbft.size)
  }

  @Test
  fun `can serialize ForkSpec with various block time and timestamp combinations`() {
    val validator = Validator(ByteArray(20) { 0x01 })
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator),
        elFork = ElFork.Paris,
      )

    val testCases =
      listOf(
        Pair(1U, 0UL),
        Pair(12U, 1000000UL),
        Pair(UInt.MAX_VALUE, ULong.MAX_VALUE),
      )

    testCases.forEach { (blockTime, timestamp) ->
      val forkSpec =
        ForkSpec(
          timestampSeconds = timestamp,
          blockTimeSeconds = blockTime,
          configuration = qbftConfig,
        )

      val serializedData = serializer.serialize(forkSpec)

      val buffer = ByteBuffer.wrap(serializedData)
      assertThat(buffer.getInt()).isEqualTo(blockTime.toInt())
      assertThat(buffer.getLong()).isEqualTo(timestamp.toLong())
    }
  }

  @Test
  fun `can serialize ForkSpec with ElDelegatedConfig containing different TTD values`() {
    val validator = Validator(ByteArray(20) { 0x01 })
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator),
        elFork = ElFork.Shanghai,
      )

    val ttdValues = listOf(0UL, 1UL, 1000UL, ULong.MAX_VALUE)

    ttdValues.forEach { ttd ->
      val elDelegatedConfig =
        ElDelegatedConfig(
          postTtdConfig = qbftConfig,
          terminalTotalDifficulty = ttd,
        )

      val forkSpec =
        ForkSpec(
          timestampSeconds = 1000UL,
          blockTimeSeconds = 12U,
          configuration = elDelegatedConfig,
        )

      val serializedData = serializer.serialize(forkSpec)

      // Skip ForkSpec fields and read the TTD from ElDelegatedConfig serialization
      val buffer = ByteBuffer.wrap(serializedData)
      buffer.getInt() // skip blockTime
      buffer.getLong() // skip timestamp
      val deserializedTtd = buffer.getLong() // TTD from ElDelegatedConfig

      assertThat(deserializedTtd).isEqualTo(ttd.toLong())
    }
  }

  @Test
  fun `serialization is deterministic for ElDelegatedConfig`() {
    val validator1 = Validator(ByteArray(20) { 0x01 })
    val validator2 = Validator(ByteArray(20) { 0x02 })
    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = setOf(validator1, validator2),
        elFork = ElFork.Prague,
      )

    val elDelegatedConfig =
      ElDelegatedConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 12345UL,
      )

    val forkSpec =
      ForkSpec(
        timestampSeconds = 5000UL,
        blockTimeSeconds = 8U,
        configuration = elDelegatedConfig,
      )

    val serialized1 = serializer.serialize(forkSpec)
    val serialized2 = serializer.serialize(forkSpec)

    assertThat(serialized1).isEqualTo(serialized2)
  }

  @Test
  fun `can serialize ForkSpec with ElDelegatedConfig with different ElFork values`() {
    val validator = Validator(ByteArray(20) { 0x01 })

    ElFork.values().forEach { fork ->
      val qbftConfig =
        QbftConsensusConfig(
          validatorSet = setOf(validator),
          elFork = fork,
        )

      val elDelegatedConfig =
        ElDelegatedConfig(
          postTtdConfig = qbftConfig,
          terminalTotalDifficulty = 1000UL,
        )

      val forkSpec =
        ForkSpec(
          timestampSeconds = 1000UL,
          blockTimeSeconds = 12U,
          configuration = elDelegatedConfig,
        )

      val serializedData = serializer.serialize(forkSpec)

      // Verify that different forks produce different serializations
      // The fork ordinal should be at the end of the serialization
      val buffer = ByteBuffer.wrap(serializedData)
      buffer.position(serializedData.size - 4) // Last 4 bytes should be fork ordinal
      val forkOrdinal = buffer.getInt()

      assertThat(forkOrdinal).isEqualTo(fork.ordinal)
    }
  }

  @Test
  fun `throws exception when configuration type is not supported`() {
    val unsupportedConfig = object : maru.consensus.ConsensusConfig {}

    val forkSpec =
      ForkSpec(
        timestampSeconds = 1000UL,
        blockTimeSeconds = 12U,
        configuration = unsupportedConfig,
      )

    assertThatThrownBy {
      serializer.serialize(forkSpec)
    }.isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("is not supported!")
  }

  @Test
  fun `can serialize large ForkSpec with ElDelegatedConfig and many validators`() {
    val validators =
      (1..100)
        .map { i ->
          Validator(ByteArray(20) { (i % 256).toByte() })
        }.toSet()

    val qbftConfig =
      QbftConsensusConfig(
        validatorSet = validators,
        elFork = ElFork.Shanghai,
      )

    val elDelegatedConfig =
      ElDelegatedConfig(
        postTtdConfig = qbftConfig,
        terminalTotalDifficulty = 999999UL,
      )

    val forkSpec =
      ForkSpec(
        timestampSeconds = 9999999UL,
        blockTimeSeconds = 30U,
        configuration = elDelegatedConfig,
      )

    val serializedData = serializer.serialize(forkSpec)

    // Verify expected size: 4 + 8 + 8 + (100 * 20) + 4 = 2024 bytes
    val expectedSize = 4 + 8 + 8 + (100 * 20) + 4
    assertThat(serializedData).hasSize(expectedSize)

    // Verify basic structure
    val buffer = ByteBuffer.wrap(serializedData)
    assertThat(buffer.getInt()).isEqualTo(30) // blockTimeSeconds
    assertThat(buffer.getLong()).isEqualTo(9999999L) // timestampSeconds
    assertThat(buffer.getLong()).isEqualTo(999999L) // TTD
  }

  private val config =
    ElDelegatedConfig(
      postTtdConfig =
        QbftConsensusConfig(
          setOf(DataGenerators.randomValidator(), DataGenerators.randomValidator()),
          elFork = ElFork.Paris,
        ),
      terminalTotalDifficulty = 123UL,
    )

  @Test
  fun `serialization for Qbft is deterministic for same input`() {
    val v1 = DataGenerators.randomValidator()
    val v2 = DataGenerators.randomValidator()
    val config1 =
      QbftConsensusConfig(
        validatorSet = setOf(v1, v2),
        elFork = ElFork.Prague,
      )
    val config2 =
      QbftConsensusConfig(
        validatorSet = setOf(v2, v1),
        elFork = ElFork.Prague,
      )
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5u,
        timestampSeconds = 123456789UL,
        configuration = config1,
      )
    val forkSpec2 =
      ForkSpec(
        blockTimeSeconds = 5u,
        timestampSeconds = 123456789UL,
        configuration = config2,
      )
    val bytes1 = ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isEqualTo(bytes2)
  }

  @Test
  fun `serialization for ELDelegated is deterministic for same input`() {
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5u,
        timestampSeconds = 123456789UL,
        configuration = config,
      )
    val forkSpec2 =
      ForkSpec(
        blockTimeSeconds = 5u,
        timestampSeconds = 123456789UL,
        configuration = config,
      )
    val bytes1 = ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isEqualTo(bytes2)
  }

  @Test
  fun `serialization changes for ELDelegated when blockTimeSeconds changes`() {
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5u,
        timestampSeconds = 123456789UL,
        configuration = config,
      )
    val forkSpec2 =
      forkSpec1.copy(blockTimeSeconds = 10U)
    val bytes1 = ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isNotEqualTo(bytes2)
  }

  @Test
  fun `serialization changes for QBFT when blockTimeSeconds changes`() {
    val v = DataGenerators.randomValidator()
    val config =
      QbftConsensusConfig(
        validatorSet = setOf(v),
        elFork = ElFork.Prague,
      )
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5U,
        timestampSeconds = 123456789UL,
        configuration = config,
      )
    val forkSpec2 =
      forkSpec1.copy(blockTimeSeconds = 10U)
    val bytes1 = ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isNotEqualTo(bytes2)
  }

  @Test
  fun `serialization for QBFT changes when timestampSeconds changes`() {
    val v = DataGenerators.randomValidator()
    val config =
      QbftConsensusConfig(
        validatorSet = setOf(v),
        elFork = ElFork.Prague,
      )
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5U,
        timestampSeconds = 123456789UL,
        configuration = config,
      )
    val forkSpec2 = forkSpec1.copy(timestampSeconds = 3123UL)
    val bytes1 = ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isNotEqualTo(bytes2)
  }

  @Test
  fun `serialization for ELDelegated changes when timestampSeconds changes`() {
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5U,
        timestampSeconds = 123456789UL,
        configuration = config,
      )
    val forkSpec2 = forkSpec1.copy(timestampSeconds = 3123UL)
    val bytes1 = ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isNotEqualTo(bytes2)
  }

  @Test
  fun `serialization changes when consensus config changes`() {
    val v1 = DataGenerators.randomValidator()
    val v2 = DataGenerators.randomValidator()
    val config1 =
      QbftConsensusConfig(
        validatorSet = setOf(v1),
        elFork = ElFork.Prague,
      )
    val config2 =
      QbftConsensusConfig(
        validatorSet = setOf(v1, v2),
        elFork = ElFork.Prague,
      )
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5U,
        timestampSeconds = 123456789UL,
        configuration = config1,
      )
    val forkSpec2 =
      ForkSpec(
        blockTimeSeconds = 5U,
        timestampSeconds = 123456789UL,
        configuration = config2,
      )
    val bytes1 = ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isNotEqualTo(bytes2)
  }
}

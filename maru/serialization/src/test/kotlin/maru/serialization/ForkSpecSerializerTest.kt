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
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForkSpec
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class ForkSpecSerializerTest {
  @Test
  fun `serialization is deterministic for same input`() {
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
        blockTimeSeconds = 5,
        timestampSeconds = 123456789L,
        configuration = config1,
      )
    val forkSpec2 =
      ForkSpec(
        blockTimeSeconds = 5,
        timestampSeconds = 123456789L,
        configuration = config2,
      )
    val bytes1 = ForkIdSerializers.ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkIdSerializers.ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isEqualTo(bytes2)
  }

  @Test
  fun `serialization changes when blockTimeSeconds changes`() {
    val v = DataGenerators.randomValidator()
    val config =
      QbftConsensusConfig(
        validatorSet = setOf(v),
        elFork = ElFork.Prague,
      )
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5,
        timestampSeconds = 123456789L,
        configuration = config,
      )
    val forkSpec2 =
      forkSpec1.copy(blockTimeSeconds = 10)
    val bytes1 = ForkIdSerializers.ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkIdSerializers.ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isNotEqualTo(bytes2)
  }

  @Test
  fun `serialization changes when timestampSeconds changes`() {
    val v = DataGenerators.randomValidator()
    val config =
      QbftConsensusConfig(
        validatorSet = setOf(v),
        elFork = ElFork.Prague,
      )
    val forkSpec1 =
      ForkSpec(
        blockTimeSeconds = 5,
        timestampSeconds = 123456789L,
        configuration = config,
      )
    val forkSpec2 = forkSpec1.copy(timestampSeconds = 3123L)
    val bytes1 = ForkIdSerializers.ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkIdSerializers.ForkSpecSerializer.serialize(forkSpec2)
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
        blockTimeSeconds = 5,
        timestampSeconds = 123456789L,
        configuration = config1,
      )
    val forkSpec2 =
      ForkSpec(
        blockTimeSeconds = 5,
        timestampSeconds = 123456789L,
        configuration = config2,
      )
    val bytes1 = ForkIdSerializers.ForkSpecSerializer.serialize(forkSpec1)
    val bytes2 = ForkIdSerializers.ForkSpecSerializer.serialize(forkSpec2)
    assertThat(bytes1).isNotEqualTo(bytes2)
  }
}

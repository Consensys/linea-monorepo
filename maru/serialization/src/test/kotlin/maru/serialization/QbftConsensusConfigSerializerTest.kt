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
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class QbftConsensusConfigSerializerTest {
  @Test
  fun `serialization is deterministic regardless of validator set order`() {
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
    val bytes1 = QbftConsensusConfigSerializer.serialize(config1)
    val bytes2 = QbftConsensusConfigSerializer.serialize(config2)
    assertThat(bytes1).isEqualTo(bytes2)
  }

  @Test
  fun `serialization changes when validator set changes`() {
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
    val bytes1 = QbftConsensusConfigSerializer.serialize(config1)
    val bytes2 = QbftConsensusConfigSerializer.serialize(config2)
    assertThat(bytes1).isNotEqualTo(bytes2)
  }
}

/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import maru.core.Validator

data class QbftConsensusConfig(
  val validatorSet: Set<Validator>,
  val elFork: ElFork,
) : ConsensusConfig {
  override fun toString(): String = "QbftConsensusConfig(validatorSet=$validatorSet, elFork=$elFork)"
}

data class DifficultyAwareQbftConfig(
  val postTtdConfig: QbftConsensusConfig,
  val terminalTotalDifficulty: ULong,
) : ConsensusConfig {
  override fun toString(): String =
    "DifficultyAwareQbftConfig(postTtdConfig=$postTtdConfig, terminalTotalDifficulty=$terminalTotalDifficulty)"
}

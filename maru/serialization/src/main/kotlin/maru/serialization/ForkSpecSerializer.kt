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
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForkSpec
import maru.extensions.encodeHex

object ForkSpecSerializer {
  object QbftConsensusConfigSerializer : Serializer<QbftConsensusConfig> {
    override fun serialize(value: QbftConsensusConfig): ByteArray {
      // Sort validators deterministically by address hex
      val validatorsSorted = value.validatorSet.sortedBy { it.address.encodeHex(prefix = false) }
      // Allocate buffer: 20 bytes per validator + 4 for elFork.ordinal
      val buffer = ByteBuffer.allocate(validatorsSorted.size * 20 + 4)
      for (validator in validatorsSorted) {
        buffer.put(validator.address)
      }
      buffer.putInt(value.elFork.ordinal)
      return buffer.array()
    }
  }

  object ForkSpecSerializer : Serializer<ForkSpec> {
    override fun serialize(value: ForkSpec): ByteArray =
      when (value.configuration) {
        is QbftConsensusConfig -> {
          val serializedConsensusConfig =
            QbftConsensusConfigSerializer.serialize(value.configuration as QbftConsensusConfig)
          ByteBuffer
            .allocate(4 + 8 + serializedConsensusConfig.size)
            .putInt(value.blockTimeSeconds)
            .putLong(value.timestampSeconds)
            .put(serializedConsensusConfig)
            .array()
        }

        else -> throw IllegalArgumentException("${value.configuration.javaClass.simpleName} is not supported!")
      }
  }
}

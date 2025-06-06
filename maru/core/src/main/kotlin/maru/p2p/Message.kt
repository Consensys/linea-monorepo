/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import maru.core.SealedBeaconBlock

enum class Version : Comparable<Version> {
  V1,
}

enum class MessageType {
  QBFT, // Won't be supported until Milestone 6
  BEACON_BLOCK,
}

data class Message<T : Any>(
  val type: MessageType,
  val version: Version = Version.V1,
  val payload: T,
) {
  init {
    when (type) {
      MessageType.QBFT -> Unit // require(payload is BftMessage≤*>) Not adding this to avoid dependency on QBFT
      MessageType.BEACON_BLOCK -> require(payload is SealedBeaconBlock)
    }
  }
}

interface TopicIdGenerator {
  fun topicId(
    messageType: MessageType,
    version: Version,
  ): String
}

class LineaTopicIdGenerator(
  private val chainId: UInt,
) : TopicIdGenerator {
  override fun topicId(
    messageType: MessageType,
    version: Version,
  ): String = "/linea/$chainId/${messageType.toString().lowercase()}/$version"
}

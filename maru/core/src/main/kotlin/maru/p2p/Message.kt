/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

enum class Version : Comparable<Version> {
  V1,
}

enum class GossipMessageType : MessageType {
  QBFT,
  BEACON_BLOCK,
}

enum class RpcMessageType : MessageType {
  STATUS(),
}

sealed interface MessageType

data class Message<TPayload, TMessageType : MessageType>(
  val type: TMessageType,
  val version: Version = Version.V1,
  val payload: TPayload,
)

interface MessageIdGenerator {
  fun id(
    messageName: String,
    version: Version,
  ): String
}

class LineaMessageIdGenerator(
  private val chainId: UInt,
) : MessageIdGenerator {
  override fun id(
    messageName: String,
    version: Version,
  ): String = "/linea/$chainId/${messageName.lowercase()}/$version"
}

class LineaRpcProtocolIdGenerator(
  private val chainId: UInt,
) : MessageIdGenerator {
  override fun id(
    messageName: String,
    version: Version,
  ): String = "/linea/req/$chainId/${messageName.lowercase()}/$version"
}

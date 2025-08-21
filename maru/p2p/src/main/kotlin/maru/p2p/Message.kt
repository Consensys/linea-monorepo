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
  STATUS,
  BEACON_BLOCKS_BY_RANGE,
}

enum class Encoding {
  RLP,
  RLP_SNAPPY,
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
    encoding: Encoding = Encoding.RLP,
  ): String
}

class LineaMessageIdGenerator(
  private val chainId: UInt,
) : MessageIdGenerator {
  override fun id(
    messageName: String,
    version: Version,
    encoding: Encoding,
  ): String = "/linea/$chainId/${messageName.lowercase()}/$version/${encoding.name.lowercase()}"
}

class LineaRpcProtocolIdGenerator(
  private val chainId: UInt,
) : MessageIdGenerator {
  override fun id(
    messageName: String,
    version: Version,
    encoding: Encoding,
  ): String = "/linea/req/$chainId/${messageName.lowercase()}/$version/${encoding.name.lowercase()}"
}

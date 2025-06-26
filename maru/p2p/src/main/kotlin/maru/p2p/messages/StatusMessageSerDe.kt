/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.messages

import maru.p2p.Message
import maru.p2p.RpcMessageType
import maru.p2p.Version
import maru.serialization.SerDe

class StatusMessageSerDe(
  private val statusSerDe: SerDe<Status>,
) : SerDe<Message<Status, RpcMessageType>> {
  override fun serialize(value: Message<Status, RpcMessageType>): ByteArray = statusSerDe.serialize(value.payload)

  override fun deserialize(bytes: ByteArray): Message<Status, RpcMessageType> =
    Message(RpcMessageType.STATUS, Version.V1, statusSerDe.deserialize(bytes))
}

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
import maru.serialization.rlp.RLPSerDe
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class StatusMessageSerDe(
  private val statusSerDe: RLPSerDe<Status>,
) : RLPSerDe<Message<Status, RpcMessageType>> {
  override fun writeTo(
    value: Message<Status, RpcMessageType>,
    rlpOutput: RLPOutput,
  ) {
    statusSerDe.writeTo(value.payload, rlpOutput)
  }

  override fun readFrom(rlpInput: RLPInput): Message<Status, RpcMessageType> {
    val status = statusSerDe.readFrom(rlpInput)

    return Message(
      RpcMessageType.STATUS,
      Version.V1,
      status,
    )
  }
}

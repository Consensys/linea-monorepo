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
import maru.serialization.rlp.SealedBeaconBlockSerDe
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.rlp.RLP
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class BeaconBlocksByRangeResponseMessageSerDe(
  private val sealedBeaconBlockSerDe: SealedBeaconBlockSerDe,
) : SerDe<Message<BeaconBlocksByRangeResponse, RpcMessageType>> {
  override fun serialize(value: Message<BeaconBlocksByRangeResponse, RpcMessageType>): ByteArray =
    RLP
      .encode { rlpOutput ->
        writeTo(value.payload, rlpOutput)
      }.toArray()

  override fun deserialize(bytes: ByteArray): Message<BeaconBlocksByRangeResponse, RpcMessageType> =
    Message(RpcMessageType.BEACON_BLOCKS_BY_RANGE, Version.V1, readFrom(RLP.input(Bytes.wrap(bytes))))

  private fun writeTo(
    value: BeaconBlocksByRangeResponse,
    rlpOutput: RLPOutput,
  ) {
    rlpOutput.startList()
    rlpOutput.writeList(value.blocks) { block, output ->
      sealedBeaconBlockSerDe.writeTo(block, output)
    }
    rlpOutput.endList()
  }

  private fun readFrom(rlpInput: RLPInput): BeaconBlocksByRangeResponse {
    rlpInput.enterList()
    val blocks = rlpInput.readList { sealedBeaconBlockSerDe.readFrom(it) }
    rlpInput.leaveList()

    return BeaconBlocksByRangeResponse(blocks = blocks)
  }
}

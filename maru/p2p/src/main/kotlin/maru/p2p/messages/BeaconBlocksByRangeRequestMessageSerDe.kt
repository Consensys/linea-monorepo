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
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.rlp.RLP
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class BeaconBlocksByRangeRequestMessageSerDe : SerDe<Message<BeaconBlocksByRangeRequest, RpcMessageType>> {
  override fun serialize(value: Message<BeaconBlocksByRangeRequest, RpcMessageType>): ByteArray =
    RLP
      .encode { rlpOutput ->
        writeTo(value.payload, rlpOutput)
      }.toArray()

  override fun deserialize(bytes: ByteArray): Message<BeaconBlocksByRangeRequest, RpcMessageType> =
    Message(RpcMessageType.BEACON_BLOCKS_BY_RANGE, Version.V1, readFrom(RLP.input(Bytes.wrap(bytes))))

  private fun writeTo(
    value: BeaconBlocksByRangeRequest,
    rlpOutput: RLPOutput,
  ) {
    rlpOutput.startList()
    rlpOutput.writeLong(value.startBlockNumber.toLong())
    rlpOutput.writeLong(value.count.toLong())
    rlpOutput.endList()
  }

  private fun readFrom(rlpInput: RLPInput): BeaconBlocksByRangeRequest {
    rlpInput.enterList()
    val startBlockNumber = rlpInput.readLong().toULong()
    val count = rlpInput.readLong().toULong()
    rlpInput.leaveList()

    return BeaconBlocksByRangeRequest(
      startBlockNumber = startBlockNumber,
      count = count,
    )
  }
}

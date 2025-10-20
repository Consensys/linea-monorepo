/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft.adapters

import maru.core.BeaconBlock
import maru.core.BeaconBlockHeader
import maru.core.SealedBeaconBlock
import maru.p2p.GossipMessageType
import maru.p2p.Message
import maru.p2p.MessageData
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockHeader
import org.hyperledger.besu.ethereum.p2p.rlpx.wire.MessageData as BesuMessageData

/**
 * Convert a QBFT block to a BeaconBlock
 *
 * @param this the QBFT block to convert
 */
fun QbftBlock.toBeaconBlock(): BeaconBlock =
  when (this) {
    is QbftBlockAdapter -> this.beaconBlock
    is QbftSealedBlockAdapter -> this.sealedBeaconBlock.beaconBlock
    else -> throw IllegalArgumentException("Unsupported block type")
  }

/**
 * Convert a QBFT block to a SealedBeaconBlock
 *
 * @param this the QBFT block to convert
 */
fun QbftBlock.toSealedBeaconBlock(): SealedBeaconBlock {
  require(this is QbftSealedBlockAdapter) {
    "Unsupported block type"
  }
  return this.sealedBeaconBlock
}

/**
 * Convert a QBFT block header to a BeaconBlockHeader
 *
 * @param this the QBFT block header to convert
 */
fun QbftBlockHeader.toBeaconBlockHeader(): BeaconBlockHeader {
  require(this is QbftBlockHeaderAdapter) {
    "Unsupported block header type"
  }
  return this.beaconBlockHeader
}

fun BesuMessageData.toDomain(): Message<BesuMessageData, GossipMessageType> =
  MessageData(
    type = GossipMessageType.QBFT,
    payload = this,
  )

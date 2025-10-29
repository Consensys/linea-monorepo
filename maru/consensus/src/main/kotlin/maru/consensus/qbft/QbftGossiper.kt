/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft

import maru.consensus.qbft.adapters.P2PValidatorMulticaster
import org.hyperledger.besu.consensus.qbft.core.types.QbftGossiper
import org.hyperledger.besu.consensus.qbft.core.types.QbftMessage

/**
 * Gossiper that only rebroadcasts messages from the future message buffer since these may have been discarded.
 * Since we are using libp2p with topics there is no rebroadcasting the current messages.
 */
class QbftGossiper(
  val p2PValidatorMulticaster: P2PValidatorMulticaster,
) : QbftGossiper {
  override fun send(
    message: QbftMessage,
    replayed: Boolean,
  ) {
    // Only send if the message is being replayed from the future message buffer
    // LibP2P takes care of gossiping current messages.
    if (replayed) {
      p2PValidatorMulticaster.send(message.data)
    }
  }
}

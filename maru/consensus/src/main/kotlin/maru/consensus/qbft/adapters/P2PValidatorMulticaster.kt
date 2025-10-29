/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft.adapters

import io.libp2p.pubsub.NoPeersForOutboundMessageException
import java.util.concurrent.ExecutionException
import maru.p2p.P2PNetwork
import org.hyperledger.besu.consensus.common.bft.network.ValidatorMulticaster
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.ethereum.p2p.rlpx.wire.MessageData as BesuMessageData

/**
 * Adapter that implements the Hyperledger Besu ValidatorMulticaster interface and delegates to a P2PNetwork.
 *
 * This adapter is used by the QBFT consensus protocol to send messages to validators.
 */
class P2PValidatorMulticaster(
  private val p2pNetwork: P2PNetwork,
) : ValidatorMulticaster {
  /**
   * Send a message to all connected validators.
   *
   * @param message The message to send.
   */
  override fun send(message: BesuMessageData) {
    try {
      p2pNetwork.broadcastMessage(message.toDomain()).get()
    } catch (ee: ExecutionException) {
      if (ee.cause?.javaClass == NoPeersForOutboundMessageException::class.java) {
        // No peers to send to, just ignore
      } else {
        throw ee
      }
    }
  }

  /**
   * Send a message to all connected validators except those in the denyList.
   *
   * @param message The message to send.
   * @param denyList This becomes irrelevant because it's a broadcasting under the hood, but needs to be there for the
   * completeness of the interface
   */
  override fun send(
    message: BesuMessageData,
    denyList: Collection<Address>,
  ) {
    send(message)
  }
}

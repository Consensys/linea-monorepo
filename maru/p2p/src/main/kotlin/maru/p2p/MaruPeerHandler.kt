/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import tech.pegasys.teku.networking.p2p.network.PeerHandler
import tech.pegasys.teku.networking.p2p.peer.Peer

class MaruPeerHandler : PeerHandler {
  override fun onConnect(peer: Peer) {
    // TODO("Not yet implemented1")
  }

  override fun onDisconnect(peer: Peer) {
    // TODO("Not yet implemented2")
  }
}

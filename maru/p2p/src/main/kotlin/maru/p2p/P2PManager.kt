/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.p2p

import io.libp2p.core.crypto.unmarshalPrivateKey
import java.util.Optional
import java.util.concurrent.TimeUnit
import maru.config.P2P
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.p2p.libp2p.MultiaddrPeerAddress
import tech.pegasys.teku.networking.p2p.libp2p.PeerAlreadyConnectedException
import tech.pegasys.teku.networking.p2p.network.P2PNetwork
import tech.pegasys.teku.networking.p2p.network.PeerAddress
import tech.pegasys.teku.networking.p2p.peer.DisconnectReason
import tech.pegasys.teku.networking.p2p.peer.NodeId
import tech.pegasys.teku.networking.p2p.peer.Peer

class P2PManager(
  private val privateKeyBytes: ByteArray,
  private val p2pConfig: P2P,
) {
  companion object {
    private const val DEFAULT_RECONNECT_DELAY_MILLI_SECONDS = 5000L // TODO: Do we want to make this configurable?
  }

  val p2pNetwork: P2PNetwork<Peer> = buildP2PNetwork()

  private val log: Logger = LogManager.getLogger(this::class.java)
  private val delayedExecutor = SafeFuture.delayedExecutor(DEFAULT_RECONNECT_DELAY_MILLI_SECONDS, TimeUnit.MILLISECONDS)
  private val staticPeerMap = mutableMapOf<NodeId, MultiaddrPeerAddress>()

  fun start() {
    p2pNetwork
      .start()
      ?.thenApply {
        p2pConfig.staticPeers.forEach { peer ->
          p2pNetwork.createPeerAddress(peer)?.let { address -> addStaticPeer(address as MultiaddrPeerAddress) }
        }
      }
  }

  fun stop() {
    p2pNetwork.stop().get(5L, TimeUnit.SECONDS)
  }

  private fun buildP2PNetwork(): P2PNetwork<Peer> {
    val privateKey = unmarshalPrivateKey(privateKeyBytes)

    return P2PNetworkFactory.build(
      privateKey = privateKey,
      ipAddress = p2pConfig.ipAddress,
      port = p2pConfig.port,
    )
  }

  fun addStaticPeer(peerAddress: MultiaddrPeerAddress) {
    if (peerAddress.id == p2pNetwork.nodeId) { // Don't connect to self
      return
    }
    synchronized(this) {
      if (staticPeerMap.containsKey(peerAddress.id)) {
        return
      }
      staticPeerMap[peerAddress.id] = peerAddress
    }
    maintainPersistentConnection(peerAddress)
  }

  fun removeStaticPeer(peerAddress: PeerAddress) {
    synchronized(this) {
      staticPeerMap.remove(peerAddress.id)
      p2pNetwork.getPeer(peerAddress.id).ifPresent { peer -> peer.disconnectImmediately(Optional.empty(), true) }
    }
  }

  private fun maintainPersistentConnection(peerAddress: MultiaddrPeerAddress): SafeFuture<Unit> =
    p2pNetwork
      .connect(peerAddress)
      .whenComplete { peer: Peer?, t: Throwable? ->
        if (t != null) {
          if (t is PeerAlreadyConnectedException) {
            log.info("Already connected to peer $peerAddress. Error: ${t.message}")
            reconnectWhenDisconnected(peer!!, peerAddress)
          } else {
            log.trace(
              "Failed to connect to peer {}, retrying after {} ms. Error: {}",
              peerAddress,
              DEFAULT_RECONNECT_DELAY_MILLI_SECONDS,
              t.message,
            )
            SafeFuture
              .runAsync({ maintainPersistentConnection(peerAddress) }, delayedExecutor)
          }
        } else {
          log.info("Created persistent connection to {}", peerAddress)
          reconnectWhenDisconnected(peer!!, peerAddress)
        }
      }.thenApply {}

  private fun reconnectWhenDisconnected(
    peer: Peer,
    peerAddress: MultiaddrPeerAddress,
  ) {
    peer.subscribeDisconnect { _: Optional<DisconnectReason>, _: Boolean ->
      if (staticPeerMap.containsKey(peerAddress.id)) {
        SafeFuture.runAsync({ maintainPersistentConnection(peerAddress) }, delayedExecutor)
      }
    }
  }
}

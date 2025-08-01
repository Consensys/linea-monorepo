/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import io.libp2p.core.PeerId
import io.libp2p.crypto.keys.unmarshalSecp256k1PublicKey
import java.time.Duration
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.Executors
import java.util.concurrent.ScheduledExecutorService
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean
import maru.config.P2P
import maru.p2p.discovery.MaruDiscoveryPeer
import maru.p2p.discovery.MaruDiscoveryService
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.p2p.discovery.DiscoveryPeer
import tech.pegasys.teku.networking.p2p.libp2p.LibP2PNodeId
import tech.pegasys.teku.networking.p2p.network.P2PNetwork
import tech.pegasys.teku.networking.p2p.network.PeerAddress
import tech.pegasys.teku.networking.p2p.network.PeerHandler
import tech.pegasys.teku.networking.p2p.peer.DisconnectReason
import tech.pegasys.teku.networking.p2p.peer.NodeId
import tech.pegasys.teku.networking.p2p.peer.Peer

class MaruPeerManager(
  private val scheduler: ScheduledExecutorService = Executors.newSingleThreadScheduledExecutor(),
  private val maruPeerFactory: MaruPeerFactory,
  private val p2pConfig: P2P,
) : PeerHandler,
  PeerLookup {
  init {
    scheduler.scheduleAtFixedRate({
      logConnectedPeers()
    }, 20000, 20000, TimeUnit.MILLISECONDS)
  }

  private val log: Logger = LogManager.getLogger(this.javaClass)
  private val maxPeers = p2pConfig.maxPeers
  private val currentlySearching = AtomicBoolean(false)

  private val connectionInProgress = mutableListOf<Bytes>()
  private var discoveryService: MaruDiscoveryService? = null

  private lateinit var p2pNetwork: P2PNetwork<Peer>

  @Volatile
  private var stopCalled = false

  fun start(
    discoveryService: MaruDiscoveryService?,
    p2pNetwork: P2PNetwork<Peer>,
  ) {
    this.discoveryService = discoveryService
    this.p2pNetwork = p2pNetwork
    searchForPeersUntilMaxReached()
  }

  fun stop(): SafeFuture<Unit> {
    stopCalled = true
    return SafeFuture.completedFuture(Unit)
  }

  private fun searchForPeersUntilMaxReached() {
    if (!stopCalled && currentlySearching.compareAndSet(false, true)) {
      discoveryService?.let { discoveryService ->
        discoveryService
          .searchForPeers()
          .orTimeout(Duration.ofSeconds(30L))
          .whenComplete { availablePeers, throwable ->
            log.trace("Finished searching for peers. Found {} peers.", availablePeers.size)
            if (!stopCalled) {
              if (throwable != null) {
                log.debug("Failed to discover peers: {}", throwable.message, throwable)
                discoveryService.getKnownPeers()
              } else {
                availablePeers
              }.forEach { peer ->
                tryToConnectIfNotFull(peer)
              }
            }
          }.whenComplete { _, _ ->
            currentlySearching.set(false)
            log.trace("peerCount={}. maxPeers={}", p2pNetwork.peerCount, maxPeers)
            if (!stopCalled && p2pNetwork.peerCount < maxPeers) {
              scheduler.schedule({ searchForPeersUntilMaxReached() }, 1, TimeUnit.SECONDS)
            }
          }
      }
    }
  }

  private val connectedPeers: ConcurrentHashMap<NodeId, MaruPeer> = ConcurrentHashMap()

  private fun logConnectedPeers() {
    val peerIds = connectedPeers.keys.joinToString(", ") { it.toString() }
    log.info("Currently connected peers: [$peerIds]")
    log.info("Discovered nodes: [${discoveryService?.getKnownPeers()}]")
  }

  override fun onConnect(peer: Peer) {
    // TODO: here we could check if we want to be connected to that peer
    if (p2pNetwork.peerCount > maxPeers) {
      // TODO: We could disconnect another peer here, based on some criteria
      peer.disconnectCleanly(DisconnectReason.TOO_MANY_PEERS)
    }
    val maruPeer = maruPeerFactory.createMaruPeer(peer)
    connectedPeers.put(peer.id, maruPeer)
    if (maruPeer.connectionInitiatedLocally()) {
      maruPeer.sendStatus()
    } else {
      maruPeer.scheduleDisconnectIfStatusNotReceived(p2pConfig.statusUpdate.timeout)
    }
  }

  override fun onDisconnect(peer: Peer) {
    connectedPeers.remove(peer.id)
    log.trace("Peer={} disconnected", peer.id)
    if (!stopCalled && p2pNetwork.peerCount < maxPeers) {
      searchForPeersUntilMaxReached()
    }
  }

  override fun getPeer(nodeId: NodeId): MaruPeer? = connectedPeers[nodeId]

  override fun getPeers(): List<MaruPeer> = connectedPeers.values.toList()

  private fun tryToConnectIfNotFull(peer: MaruDiscoveryPeer) {
    try {
      val peerId = getNodeId(peer)
      if (stopCalled || p2pNetwork.isConnected(PeerAddress(peerId))) {
        return
      }
      synchronized(connectionInProgress) {
        if (p2pNetwork.peerCount >= maxPeers || connectionInProgress.contains(peer.nodeIdBytes)) {
          return
        }
        connectionInProgress.add(peer.nodeIdBytes)
      }

      p2pNetwork
        .connect(p2pNetwork.createPeerAddress(peer))
        .orTimeout(30, TimeUnit.SECONDS)
        .whenComplete { _, throwable ->
          try {
            if (throwable != null) {
              log.error("Failed to connect to peer={}", peer.nodeIdBytes, throwable)
            }
          } finally {
            synchronized(connectionInProgress) {
              connectionInProgress.remove(peer.nodeIdBytes)
            }
          }
        }
    } catch (e: Exception) {
      log.debug("Failed to initiate connection to peer={}. errorMessage={}", peer.nodeIdBytes, e.message, e)
      synchronized(connectionInProgress) {
        connectionInProgress.remove(peer.nodeIdBytes)
      }
    }
  }

  fun getNodeId(peer: DiscoveryPeer): LibP2PNodeId {
    val pubKey = unmarshalSecp256k1PublicKey(peer.publicKey.toArrayUnsafe())
    return LibP2PNodeId(PeerId.fromPubKey(pubKey))
  }
}

/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import java.util.concurrent.ScheduledExecutorService
import maru.config.P2PConfig
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.never
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.p2p.peer.NodeId
import tech.pegasys.teku.networking.p2p.peer.Peer

class MaruPeerManagerTest {
  @Test
  fun `does not schedule timeout when connection is initiated locally`() {
    val mockScheduler = mock<ScheduledExecutorService>()
    val nodeId = mock<NodeId>()
    val peer = mock<Peer>()
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = mock<MaruPeer>()
    val p2pConfig = mock<P2PConfig>()

    whenever(peer.id).thenReturn(nodeId)
    whenever(peer.connectionInitiatedLocally()).thenReturn(true)
    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)
    whenever(maruPeer.connectionInitiatedLocally()).thenReturn(true)
    whenever(p2pConfig.maxPeers).thenReturn(10)

    val manager = MaruPeerManager(maruPeerFactory = maruPeerFactory, p2pConfig = p2pConfig)
    manager.scheduler = mockScheduler
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    verify(mockScheduler, never()).schedule(any<Runnable>(), any(), any())
    verify(maruPeer).sendStatus()
  }

  @Test
  fun `sends status message immediately for locally initiated connections`() {
    val mockScheduler = mock<ScheduledExecutorService>()
    val nodeId = mock<NodeId>()
    val peer = mock<Peer>()
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = mock<MaruPeer>()
    val mockFutureStatus = mock<SafeFuture<Unit>>()
    val p2pConfig = mock<P2PConfig>()

    whenever(peer.id).thenReturn(nodeId)
    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)
    whenever(maruPeer.connectionInitiatedLocally()).thenReturn(true)
    whenever(maruPeer.sendStatus()).thenReturn(mockFutureStatus)
    whenever(p2pConfig.maxPeers).thenReturn(10)

    val manager = MaruPeerManager(maruPeerFactory = maruPeerFactory, p2pConfig = p2pConfig)
    manager.scheduler = mockScheduler
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    verify(maruPeer).sendStatus()
    verify(mockScheduler, never()).schedule(any<Runnable>(), any(), any())
  }

  @Test
  fun `does not send status message for remotely initiated connections`() {
    val nodeId = mock<NodeId>()
    val peer = mock<Peer>()
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = mock<MaruPeer>()
    val p2pConfig = mock<P2PConfig>()

    whenever(peer.id).thenReturn(nodeId)
    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)
    whenever(maruPeer.connectionInitiatedLocally()).thenReturn(false)
    whenever(maruPeer.getStatus()).thenReturn(null)
    whenever(p2pConfig.maxPeers).thenReturn(10)
    whenever(p2pConfig.statusUpdate).thenReturn(P2PConfig.StatusUpdateConfig())

    val manager = MaruPeerManager(maruPeerFactory = maruPeerFactory, p2pConfig = p2pConfig)
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    verify(maruPeer, never()).sendStatus()
  }

  @Test
  fun `creates maru peer through factory when peer connects`() {
    val nodeId = mock<NodeId>()
    val peer = mock<Peer>()
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = mock<MaruPeer>()
    val p2pConfig = mock<P2PConfig>()

    whenever(peer.id).thenReturn(nodeId)
    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)
    whenever(maruPeer.connectionInitiatedLocally()).thenReturn(true)
    whenever(p2pConfig.maxPeers).thenReturn(10)

    val manager = MaruPeerManager(maruPeerFactory = maruPeerFactory, p2pConfig = p2pConfig)
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    verify(maruPeerFactory).createMaruPeer(peer)
  }

  @Test
  fun `stores connected peer in manager for retrieval`() {
    val nodeId = mock<NodeId>()
    val peer = mock<Peer>()
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = mock<MaruPeer>()
    val p2pConfig = mock<P2PConfig>()

    whenever(peer.id).thenReturn(nodeId)
    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)
    whenever(maruPeer.connectionInitiatedLocally()).thenReturn(false)
    whenever(maruPeer.getStatus()).thenReturn(null)
    whenever(p2pConfig.maxPeers).thenReturn(10)
    whenever(p2pConfig.statusUpdate).thenReturn(P2PConfig.StatusUpdateConfig())

    val manager = MaruPeerManager(maruPeerFactory = maruPeerFactory, p2pConfig = p2pConfig)
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    assertThat(manager.getPeer(nodeId)).isEqualTo(maruPeer)
  }
}

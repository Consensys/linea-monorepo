/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import maru.config.P2PConfig
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem
import org.junit.jupiter.api.Test
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.never
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider
import tech.pegasys.teku.networking.p2p.network.PeerAddress
import tech.pegasys.teku.networking.p2p.peer.NodeId
import tech.pegasys.teku.networking.p2p.peer.Peer
import tech.pegasys.teku.networking.p2p.network.P2PNetwork as TekuP2PNetwork

class MaruPeerManagerTest {
  private object Fixtures {
    fun peerAddress(id: NodeId = mock()): PeerAddress = mock<PeerAddress>().also { whenever(it.id).thenReturn(id) }

    fun tekuPeer(
      id: NodeId = mock(),
      address: PeerAddress = peerAddress(id),
      initiatedLocally: Boolean = true,
    ): Peer =
      mock<Peer>().also {
        whenever(it.id).thenReturn(id)
        whenever(it.address).thenReturn(address)
        whenever(it.connectionInitiatedLocally()).thenReturn(initiatedLocally)
        whenever(it.connectionInitiatedRemotely()).thenReturn(!initiatedLocally)
      }

    fun maruPeer(
      initiatedLocally: Boolean = true,
      isConnected: Boolean = true,
      address: PeerAddress = mock(),
      sendStatusFuture: SafeFuture<Unit>? = null,
    ): MaruPeer =
      mock<MaruPeer>().also {
        whenever(it.connectionInitiatedLocally()).thenReturn(initiatedLocally)
        whenever(it.isConnected).thenReturn(isConnected)
        whenever(it.address).thenReturn(address)
        sendStatusFuture?.let { fut -> whenever(it.sendStatus()).thenReturn(fut) }
      }
  }

  companion object {
    val reputationManager =
      MaruReputationManager(
        metricsSystem = NoOpMetricsSystem(),
        timeProvider = SystemTimeProvider(),
        isStaticPeer = { _: NodeId -> false },
        reputationConfig = P2PConfig.Reputation(),
      )
  }

  @Test
  fun `does not schedule timeout when connection is initiated locally`() {
    // Arrange
    val peer = Fixtures.tekuPeer(id = mock(), initiatedLocally = true)
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = Fixtures.maruPeer(initiatedLocally = true)
    val p2pConfig = P2PConfig()

    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)

    val manager =
      MaruPeerManager(
        maruPeerFactory = maruPeerFactory,
        p2pConfig = p2pConfig,
        reputationManager = reputationManager,
        isStaticPeer = { false },
      )

    // Act
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    // Assert
    verify(maruPeer).sendStatus()
    verify(maruPeer, never()).scheduleDisconnectIfStatusNotReceived(any())
  }

  @Test
  fun `sends status message immediately for locally initiated connections`() {
    // Arrange
    val nodeId: NodeId = mock()
    val peer = Fixtures.tekuPeer(id = nodeId, initiatedLocally = true)
    val maruPeerFactory = mock<MaruPeerFactory>()
    val mockFutureStatus = mock<SafeFuture<Unit>>()
    val maruPeer = Fixtures.maruPeer(initiatedLocally = true, sendStatusFuture = mockFutureStatus)
    val p2pConfig = P2PConfig()

    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)

    val manager =
      MaruPeerManager(
        maruPeerFactory = maruPeerFactory,
        p2pConfig = p2pConfig,
        reputationManager = reputationManager,
        isStaticPeer = { false },
      )

    // Act
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    // Assert
    verify(maruPeer).sendStatus()
  }

  @Test
  fun `does not send status message for remotely initiated connections`() {
    // Arrange
    val nodeId: NodeId = mock()
    val address = Fixtures.peerAddress(nodeId)
    val peer = Fixtures.tekuPeer(id = nodeId, address = address, initiatedLocally = false)
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = Fixtures.maruPeer(initiatedLocally = false, address = address)
    val p2pConfig = P2PConfig()

    whenever(maruPeer.getStatus()).thenReturn(null)
    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)

    val manager =
      MaruPeerManager(
        maruPeerFactory = maruPeerFactory,
        p2pConfig = p2pConfig,
        reputationManager = reputationManager,
        isStaticPeer = { false },
      )

    // Act
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    // Assert
    verify(maruPeer, never()).sendStatus()
    verify(maruPeer).scheduleDisconnectIfStatusNotReceived(any())
  }

  @Test
  fun `creates maru peer through factory when peer connects`() {
    // Arrange
    val peer = Fixtures.tekuPeer(id = mock(), initiatedLocally = true)
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = Fixtures.maruPeer(initiatedLocally = true)
    val p2pConfig = P2PConfig()

    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)

    val manager =
      MaruPeerManager(
        maruPeerFactory = maruPeerFactory,
        p2pConfig = p2pConfig,
        reputationManager = reputationManager,
        isStaticPeer = { false },
      )

    // Act
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    // Assert
    verify(maruPeerFactory).createMaruPeer(peer)
  }

  @Test
  fun `stores connected peer in manager for retrieval`() {
    // Arrange
    val nodeId: NodeId = mock()
    val address = Fixtures.peerAddress(nodeId)
    val peer = Fixtures.tekuPeer(id = nodeId, address = address, initiatedLocally = false)
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = Fixtures.maruPeer(initiatedLocally = false, isConnected = true, address = address)
    val p2pConfig = P2PConfig()

    whenever(maruPeer.getStatus()).thenReturn(null)
    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)

    val manager =
      MaruPeerManager(
        maruPeerFactory = maruPeerFactory,
        p2pConfig = p2pConfig,
        reputationManager = reputationManager,
        isStaticPeer = { false },
      )

    // Act
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(peer)

    // Assert
    assertThat(manager.getPeer(nodeId)).isEqualTo(maruPeer)
  }

  @Test
  fun `does not connect or add peer if reputation manager disallows connection`() {
    // Arrange
    val nodeId: NodeId = mock()
    val address = Fixtures.peerAddress(nodeId)
    val peer = Fixtures.tekuPeer(id = nodeId, address = address, initiatedLocally = true)
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = Fixtures.maruPeer(initiatedLocally = true)
    val p2pConfig = P2PConfig()
    val p2pNetwork = mock<TekuP2PNetwork<Peer>>()

    val reputationManager = mock<MaruReputationManager>()
    whenever(reputationManager.isConnectionInitiationAllowed(address)).thenReturn(false)

    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)

    val manager =
      MaruPeerManager(
        maruPeerFactory = maruPeerFactory,
        p2pConfig = p2pConfig,
        reputationManager = reputationManager,
        isStaticPeer = { false },
      )

    // Act
    manager.start(discoveryService = null, p2pNetwork = p2pNetwork)
    manager.onConnect(peer)

    // Assert
    verify(maruPeerFactory, never()).createMaruPeer(peer)
    assertThat(manager.getPeer(nodeId)).isNull()
  }

  @Test
  fun `connects and adds peer if reputation manager allows connection`() {
    // Arrange
    val nodeId: NodeId = mock()
    val address = Fixtures.peerAddress(nodeId)
    val peer = Fixtures.tekuPeer(id = nodeId, address = address, initiatedLocally = true)
    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer = Fixtures.maruPeer(initiatedLocally = true, isConnected = true)
    val p2pConfig = P2PConfig()
    val reputationManager = mock<MaruReputationManager>()
    val p2pNetwork = mock<TekuP2PNetwork<Peer>>()

    whenever(reputationManager.isConnectionInitiationAllowed(address)).thenReturn(true)
    whenever(maruPeerFactory.createMaruPeer(peer)).thenReturn(maruPeer)

    val manager =
      MaruPeerManager(
        maruPeerFactory = maruPeerFactory,
        p2pConfig = p2pConfig,
        reputationManager = reputationManager,
        isStaticPeer = { false },
      )

    // Act
    manager.start(discoveryService = null, p2pNetwork = p2pNetwork)
    manager.onConnect(peer)

    // Assert
    assertThat(manager.getPeer(nodeId)).isEqualTo(maruPeer)
    verify(maruPeerFactory).createMaruPeer(peer)
  }

  @Test
  fun `getPeers and peerCount only include actually connected peers`() {
    // Arrange
    val nodeId1: NodeId = mock()
    val nodeId2: NodeId = mock()

    val libp2pPeer1 = Fixtures.tekuPeer(id = nodeId1, initiatedLocally = true)
    val libp2pPeer2 = Fixtures.tekuPeer(id = nodeId2, initiatedLocally = true)

    val maruPeerFactory = mock<MaruPeerFactory>()
    val maruPeer1 = Fixtures.maruPeer(initiatedLocally = true, isConnected = true)
    val maruPeer2 = Fixtures.maruPeer(initiatedLocally = true, isConnected = true)
    val p2pConfig = P2PConfig()

    whenever(maruPeerFactory.createMaruPeer(libp2pPeer1)).thenReturn(maruPeer1)
    whenever(maruPeerFactory.createMaruPeer(libp2pPeer2)).thenReturn(maruPeer2)
    whenever(maruPeer1.sendStatus()).thenReturn(SafeFuture.completedFuture(Unit))
    whenever(maruPeer2.sendStatus()).thenReturn(SafeFuture.completedFuture(Unit))

    val manager =
      MaruPeerManager(
        maruPeerFactory = maruPeerFactory,
        p2pConfig = p2pConfig,
        reputationManager = reputationManager,
        isStaticPeer = { true }, // bypass reputation checks for the test
      )

    // Act
    manager.start(discoveryService = null, p2pNetwork = mock())
    manager.onConnect(libp2pPeer1)
    manager.onConnect(libp2pPeer2)

    // Assert
    assertThat(manager.peerCount).isEqualTo(2)
    assertThat(manager.getPeers()).containsExactlyInAnyOrder(maruPeer1, maruPeer2)

    // Arrange 2
    whenever(maruPeer2.isConnected).thenReturn(false)

    // Assert 2
    assertThat(manager.peerCount).isEqualTo(1)
    assertThat(manager.getPeers()).containsExactly(maruPeer1)
    assertThat(manager.getPeer(nodeId2)).isNull()
    assertThat(manager.getPeer(nodeId1)).isEqualTo(maruPeer1)
  }
}

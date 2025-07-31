/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.beaconchain.pipeline

import java.util.concurrent.ExecutionException
import java.util.concurrent.TimeoutException
import kotlin.time.Duration.Companion.seconds
import maru.core.SealedBeaconBlock
import maru.core.ext.DataGenerators.randomSealedBeaconBlock
import maru.p2p.MaruPeer
import maru.p2p.PeerLookup
import maru.p2p.messages.BeaconBlocksByRangeResponse
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.any
import org.mockito.kotlin.atLeastOnce
import org.mockito.kotlin.mock
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.async.SafeFuture.completedFuture
import tech.pegasys.teku.networking.p2p.peer.NodeId
import tech.pegasys.teku.networking.p2p.reputation.ReputationAdjustment

class DownloadBlocksTest {
  @Test
  fun `downloads blocks successfully from peer`() {
    val peer = mock<MaruPeer>()
    val peerLookup = mock<PeerLookup>()
    val blocks = listOf(randomSealedBeaconBlock(10u), randomSealedBeaconBlock(11uL))
    whenever(peer.sendBeaconBlocksByRange(10uL, 2uL)).thenReturn(completedFuture(BeaconBlocksByRangeResponse(blocks)))
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val task = DownloadBlocksStep(peerLookup, maxRetries = 5u, blockRangeRequestTimeout = 5.seconds)
    val range = SyncTargetRange(10uL, 11uL)
    val result = task.apply(range).get()

    assertThat(result.size).isEqualTo(2)
    assertThat(result.stream().map({ it.sealedBeaconBlock }).toList()).containsAll(blocks)
  }

  @Test
  fun `downloads blocks in multiple requests when peer returns partial results`() {
    val peer = mock<MaruPeer>()
    val peerLookup = mock<PeerLookup>()
    val block1 = randomSealedBeaconBlock(10uL)
    val block2 = randomSealedBeaconBlock(11uL)
    val block3 = randomSealedBeaconBlock(12uL)

    // First call returns only block1, second call returns block2 and block3
    whenever(peer.sendBeaconBlocksByRange(10uL, 3uL))
      .thenReturn(completedFuture(BeaconBlocksByRangeResponse(listOf(block1))))
    whenever(peer.sendBeaconBlocksByRange(11uL, 2uL))
      .thenReturn(completedFuture(BeaconBlocksByRangeResponse(listOf(block2, block3))))
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val task =
      DownloadBlocksStep(
        peerLookup,
        maxRetries = 5u,
        blockRangeRequestTimeout = 5.seconds,
      )
    val range = SyncTargetRange(10uL, 12uL)
    val result = task.apply(range).get()

    assertThat(result.stream().map({ it.sealedBeaconBlock }).toList()).containsExactly(block1, block2, block3)
  }

  @Test
  fun `applies small penalty when peer returns empty blocks`() {
    val peer = mock<MaruPeer>()
    val peerLookup = mock<PeerLookup>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))
    whenever(peer.sendBeaconBlocksByRange(1uL, 2uL)).thenReturn(
      completedFuture(
        BeaconBlocksByRangeResponse(emptyList()),
      ),
    )

    val task = DownloadBlocksStep(peerLookup, maxRetries = 5u, blockRangeRequestTimeout = 5.seconds)
    val range = SyncTargetRange(1uL, 2uL)

    assertThrows<Exception> { task.apply(range).get() }
    verify(peer, atLeastOnce()).adjustReputation(ReputationAdjustment.SMALL_PENALTY)
  }

  @Test
  fun `applies large penalty on timeout exception`() {
    val peer = mock<MaruPeer>()
    val peerLookup = mock<PeerLookup>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))
    val future = SafeFuture<BeaconBlocksByRangeResponse>()
    future.completeExceptionally(TimeoutException("timeout"))
    whenever(peer.sendBeaconBlocksByRange(1uL, 1uL)).thenReturn(future)

    val task = DownloadBlocksStep(peerLookup, maxRetries = 5u, blockRangeRequestTimeout = 5.seconds)
    val range = SyncTargetRange(1uL, 1uL)

    assertThrows<Exception> { task.apply(range).get() }
    verify(peer, atLeastOnce()).adjustReputation(ReputationAdjustment.LARGE_PENALTY)
  }

  @Test
  fun `throws after exceeding max retries`() {
    val nodeId = mock<NodeId>()
    val peer = mock<MaruPeer>()
    val peerLookup = mock<PeerLookup>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))
    whenever(peer.id).thenReturn(nodeId)
    val future = SafeFuture<BeaconBlocksByRangeResponse>()
    future.completeExceptionally(Exception("fail"))
    whenever(peer.sendBeaconBlocksByRange(any(), any())).thenReturn(future)

    val task = DownloadBlocksStep(peerLookup, maxRetries = 5u, blockRangeRequestTimeout = 5.seconds)
    val range = SyncTargetRange(1uL, 1uL)

    val ex = assertThrows<Exception> { task.apply(range).get() }
    assertThat(ex.message).contains("Maximum retries reached.")
  }

  @Test
  fun `throws when no peers available`() {
    val peerLookup = mock<PeerLookup>()
    whenever(peerLookup.getPeers()).thenReturn(emptyList())

    val task = DownloadBlocksStep(peerLookup, maxRetries = 5u, blockRangeRequestTimeout = 5.seconds)
    val range = SyncTargetRange(1uL, 1uL)

    assertThrows<ExecutionException> { task.apply(range).get() }
  }

  @Test
  fun `downloads blocks from a random peer`() {
    val peer = mock<MaruPeer>()
    val peerLookup = mock<PeerLookup>()
    val blocks = listOf(randomSealedBeaconBlock(10u), randomSealedBeaconBlock(11u))
    val response = mock<BeaconBlocksByRangeResponse>()
    whenever(response.blocks).thenReturn(blocks)
    whenever(peer.sendBeaconBlocksByRange(10u, 2u)).thenReturn(completedFuture(response))
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val step = DownloadBlocksStep(peerLookup, maxRetries = 5u, blockRangeRequestTimeout = 5.seconds)
    val range = SyncTargetRange(10u, 11u)
    val result = step.apply(range).get()
    assertThat(result).isEqualTo(createSealedBlocksWithPeers(blocks, peer))
  }

  @Test
  fun `throws if no peers are available`() {
    val peerLookup = mock<PeerLookup>()
    whenever(peerLookup.getPeers()).thenReturn(emptyList())
    val step = DownloadBlocksStep(peerLookup, maxRetries = 5u, blockRangeRequestTimeout = 5.seconds)
    val range = SyncTargetRange(0u, 0u)
    try {
      step.apply(range).get()
      assert(false) { "Expected exception" }
    } catch (e: Exception) {
      assertThat(e.cause).isInstanceOf(NoSuchElementException::class.java)
    }
  }

  private fun createSealedBlocksWithPeers(
    blocks: List<SealedBeaconBlock>,
    peer: MaruPeer,
  ): List<SealedBlockWithPeer> =
    blocks
      .stream()
      .map({
        SealedBlockWithPeer(it, peer)
      })
      .toList()
}

/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.pipeline

import maru.core.ext.DataGenerators.randomSealedBeaconBlock
import maru.p2p.MaruPeer
import maru.p2p.PeerLookup
import maru.p2p.messages.BeaconBlocksByRangeResponse
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture.completedFuture

class DownloadBlocksStepTest {
  @Test
  fun `downloads blocks from a random peer`() {
    val peer = mock<MaruPeer>()
    val peerLookup = mock<PeerLookup>()
    val blocks = listOf(randomSealedBeaconBlock(10u), randomSealedBeaconBlock(11u))
    val response = mock<BeaconBlocksByRangeResponse>()
    whenever(response.blocks).thenReturn(blocks)
    whenever(peer.sendBeaconBlocksByRange(10u, 2u)).thenReturn(completedFuture(response))
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val step = DownloadBlocksStep(peerLookup)
    val range = SyncTargetRange(10u, 11u)
    val result = step.apply(range).get()
    assertThat(result).isEqualTo(blocks)
  }

  @Test
  fun `returns empty list if peer returns empty response`() {
    val peer = mock<MaruPeer>()
    val peerLookup = mock<PeerLookup>()
    val response = mock<BeaconBlocksByRangeResponse>()
    whenever(response.blocks).thenReturn(emptyList())
    whenever(peer.sendBeaconBlocksByRange(0u, 1u)).thenReturn(completedFuture(response))
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val step = DownloadBlocksStep(peerLookup)
    val range = SyncTargetRange(0u, 0u)
    val result = step.apply(range).get()
    assertThat(result).isEmpty()
  }

  @Test
  fun `throws if no peers are available`() {
    val peerLookup = mock<PeerLookup>()
    whenever(peerLookup.getPeers()).thenReturn(emptyList())
    val step = DownloadBlocksStep(peerLookup)
    val range = SyncTargetRange(0u, 0u)
    try {
      step.apply(range).get()
      assert(false) { "Expected exception" }
    } catch (e: Exception) {
      assertThat(e).isInstanceOf(NoSuchElementException::class.java)
    }
  }
}

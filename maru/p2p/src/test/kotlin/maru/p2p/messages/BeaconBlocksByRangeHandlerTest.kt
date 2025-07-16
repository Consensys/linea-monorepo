/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.messages

import maru.core.ext.DataGenerators
import maru.database.BeaconChain
import maru.p2p.MaruPeer
import maru.p2p.Message
import maru.p2p.RpcMessageType
import maru.p2p.Version
import maru.p2p.messages.BeaconBlocksByRangeHandler.Companion.MAX_BLOCKS_PER_REQUEST
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.argumentCaptor
import org.mockito.kotlin.mock
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.networking.eth2.rpc.core.ResponseCallback

class BeaconBlocksByRangeHandlerTest {
  private lateinit var beaconChain: BeaconChain
  private lateinit var handler: BeaconBlocksByRangeHandler
  private lateinit var peer: MaruPeer
  private lateinit var callback: ResponseCallback<Message<BeaconBlocksByRangeResponse, RpcMessageType>>

  @BeforeEach
  fun setup() {
    beaconChain = mock()
    handler = BeaconBlocksByRangeHandler(beaconChain)
    peer = mock()
    callback = mock()
  }

  @Test
  fun `handles request with no blocks available`() {
    val request = BeaconBlocksByRangeRequest(startBlockNumber = 100UL, count = 10UL)
    val message =
      Message(
        type = RpcMessageType.BEACON_BLOCKS_BY_RANGE,
        version = Version.V1,
        payload = request,
      )

    whenever(beaconChain.getSealedBeaconBlocks(100UL, 10UL)).thenReturn(emptyList())

    handler.handleIncomingMessage(peer, message, callback)

    val responseCaptor = argumentCaptor<Message<BeaconBlocksByRangeResponse, RpcMessageType>>()
    verify(callback).respondAndCompleteSuccessfully(responseCaptor.capture())

    val response = responseCaptor.firstValue
    assertThat(response.type).isEqualTo(RpcMessageType.BEACON_BLOCKS_BY_RANGE)
    assertThat(response.payload.blocks).isEmpty()
  }

  @Test
  fun `handles request with blocks available`() {
    val request = BeaconBlocksByRangeRequest(startBlockNumber = 100UL, count = 3UL)
    val message =
      Message(
        type = RpcMessageType.BEACON_BLOCKS_BY_RANGE,
        version = Version.V1,
        payload = request,
      )

    val blocks =
      listOf(
        DataGenerators.randomSealedBeaconBlock(number = 100UL),
        DataGenerators.randomSealedBeaconBlock(number = 101UL),
        DataGenerators.randomSealedBeaconBlock(number = 102UL),
      )

    whenever(beaconChain.getSealedBeaconBlocks(100UL, 3UL)).thenReturn(blocks)

    handler.handleIncomingMessage(peer, message, callback)

    val responseCaptor = argumentCaptor<Message<BeaconBlocksByRangeResponse, RpcMessageType>>()
    verify(callback).respondAndCompleteSuccessfully(responseCaptor.capture())

    val response = responseCaptor.firstValue
    assertThat(response.type).isEqualTo(RpcMessageType.BEACON_BLOCKS_BY_RANGE)
    assertThat(response.payload.blocks).hasSize(3)
    assertThat(response.payload.blocks).isEqualTo(blocks)
  }

  @Test
  fun `handles large count request`() {
    val request = BeaconBlocksByRangeRequest(startBlockNumber = 0UL, count = 1000UL)
    val message =
      Message(
        type = RpcMessageType.BEACON_BLOCKS_BY_RANGE,
        version = Version.V1,
        payload = request,
      )

    // Handler should limit to MAX_BLOCKS_PER_REQUEST
    val limitedBlocks =
      (0UL until MAX_BLOCKS_PER_REQUEST).map { i ->
        DataGenerators.randomSealedBeaconBlock(number = i)
      }

    // handler should limit to MAX_BLOCKS_PER_REQUEST
    whenever(beaconChain.getSealedBeaconBlocks(0UL, MAX_BLOCKS_PER_REQUEST)).thenReturn(limitedBlocks)

    handler.handleIncomingMessage(peer, message, callback)

    val responseCaptor = argumentCaptor<Message<BeaconBlocksByRangeResponse, RpcMessageType>>()
    verify(callback).respondAndCompleteSuccessfully(responseCaptor.capture())

    val response = responseCaptor.firstValue
    assertThat(response.payload.blocks).hasSize(256)

    // Verify that the handler limited the request to 64 blocks
    verify(beaconChain).getSealedBeaconBlocks(0UL, MAX_BLOCKS_PER_REQUEST)
  }
}

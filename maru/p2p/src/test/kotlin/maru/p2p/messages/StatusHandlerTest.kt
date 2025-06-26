/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.messages

import kotlin.random.Random
import kotlin.random.nextULong
import maru.consensus.ForkIdHashProvider
import maru.core.ext.DataGenerators
import maru.database.BeaconChain
import maru.p2p.Message
import maru.p2p.RpcMessageType
import maru.p2p.Version
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.networking.eth2.rpc.core.ResponseCallback
import tech.pegasys.teku.networking.p2p.peer.Peer

class StatusHandlerTest {
  @Test
  fun `responds with current status`() {
    val beaconChain = mock<BeaconChain>()
    val forkIdHashProvider = mock<ForkIdHashProvider>()
    val latestBeaconState = DataGenerators.randomBeaconState(0U)
    val forkIdHash = Random.nextBytes(32)
    whenever(beaconChain.getLatestBeaconState()).thenReturn(latestBeaconState)
    whenever(forkIdHashProvider.currentForkIdHash()).thenReturn(forkIdHash)

    val peer = mock<Peer>()
    val message =
      Message(
        RpcMessageType.STATUS,
        Version.V1,
        Status(Random.nextBytes(32), Random.nextBytes(32), Random.nextULong()),
      )
    val callback = mock<ResponseCallback<Message<Status, RpcMessageType>>>()
    val statusHandler = StatusHandler(beaconChain, forkIdHashProvider)
    statusHandler.handleIncomingMessage(peer, message, callback)

    val expectedMessage =
      Message(
        RpcMessageType.STATUS,
        Version.V1,
        Status(
          forkIdHash,
          latestBeaconState.latestBeaconBlockHeader.hash,
          latestBeaconState.latestBeaconBlockHeader.number,
        ),
      )
    verify(callback).respondAndCompleteSuccessfully(expectedMessage)
  }
}

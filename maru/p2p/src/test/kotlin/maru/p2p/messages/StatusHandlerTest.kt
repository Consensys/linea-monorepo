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
import maru.core.ext.DataGenerators
import maru.p2p.MaruPeer
import maru.p2p.Message
import maru.p2p.RpcMessageType
import maru.p2p.Version
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.networking.eth2.rpc.core.ResponseCallback

class StatusHandlerTest {
  @Test
  fun `responds with current status`() {
    val statusMessageFactory = mock<StatusMessageFactory>()
    val localBeaconState = DataGenerators.randomBeaconState(0U)
    val localStatusMessage =
      Message(
        RpcMessageType.STATUS,
        Version.V1,
        Status(
          Random.nextBytes(32),
          localBeaconState.latestBeaconBlockHeader.hash,
          localBeaconState.latestBeaconBlockHeader.number,
        ),
      )
    whenever(statusMessageFactory.createStatusMessage()).thenReturn(localStatusMessage)

    val peer = mock<MaruPeer>()
    val callback = mock<ResponseCallback<Message<Status, RpcMessageType>>>()
    val statusHandler = StatusHandler(statusMessageFactory)
    val remoteBeaconState = DataGenerators.randomBeaconState(0U)
    val remoteStatusMessage =
      Message(
        RpcMessageType.STATUS,
        Version.V1,
        Status(
          Random.nextBytes(32),
          remoteBeaconState.latestBeaconBlockHeader.hash,
          remoteBeaconState.latestBeaconBlockHeader.number,
        ),
      )
    statusHandler.handleIncomingMessage(peer, remoteStatusMessage, callback)
    verify(callback).respondAndCompleteSuccessfully(localStatusMessage)
  }

  @Test
  fun `updates peer status`() {
    val statusMessageFactory = mock<StatusMessageFactory>()

    val peer = mock<MaruPeer>()
    val callback = mock<ResponseCallback<Message<Status, RpcMessageType>>>()
    val statusHandler = StatusHandler(statusMessageFactory)
    val remoteBeaconState = DataGenerators.randomBeaconState(0U)
    val remoteStatusMessage =
      Message(
        RpcMessageType.STATUS,
        Version.V1,
        Status(
          Random.nextBytes(32),
          remoteBeaconState.latestBeaconBlockHeader.hash,
          remoteBeaconState.latestBeaconBlockHeader.number,
        ),
      )
    statusHandler.handleIncomingMessage(peer, remoteStatusMessage, callback)
    verify(peer).updateStatus(remoteStatusMessage.payload)
  }
}

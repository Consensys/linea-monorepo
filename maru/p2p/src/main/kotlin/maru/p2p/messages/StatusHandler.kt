/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.messages

import maru.p2p.MaruPeer
import maru.p2p.Message
import maru.p2p.RpcMessageHandler
import maru.p2p.RpcMessageType
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.networking.eth2.rpc.core.ResponseCallback
import tech.pegasys.teku.networking.eth2.rpc.core.RpcException
import tech.pegasys.teku.networking.eth2.rpc.core.RpcResponseStatus

class StatusHandler(
  private val statusMessageFactory: StatusMessageFactory,
) : RpcMessageHandler<Message<Status, RpcMessageType>, Message<Status, RpcMessageType>> {
  private val log = LogManager.getLogger(this.javaClass)

  override fun handleIncomingMessage(
    peer: MaruPeer,
    message: Message<Status, RpcMessageType>,
    callback: ResponseCallback<Message<Status, RpcMessageType>>,
  ) {
    try {
      peer.updateStatus(message.payload)
      val localStatusMessage = statusMessageFactory.createStatusMessage()
      callback.respondAndCompleteSuccessfully(localStatusMessage)
    } catch (e: RpcException) {
      log.error("handling request failed with RpcException", e)
      callback.completeWithErrorResponse(e)
    } catch (th: Throwable) {
      log.error("handling request failed with unexpected error", th)
      callback.completeWithUnexpectedError(
        RpcException(
          RpcResponseStatus.SERVER_ERROR_CODE,
          "Handling request failed with unexpected error: " + th.message,
        ),
      )
    }
  }
}

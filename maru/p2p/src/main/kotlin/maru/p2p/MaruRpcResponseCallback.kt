/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import java.nio.channels.ClosedChannelException
import maru.serialization.Serializer
import org.apache.logging.log4j.LogManager
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.async.RootCauseExceptionHandler
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.eth2.rpc.core.ResponseCallback
import tech.pegasys.teku.networking.eth2.rpc.core.RpcException
import tech.pegasys.teku.networking.p2p.rpc.RpcStream

class MaruRpcResponseCallback<TResponse : Message<*, *>>(
  private val rpcStream: RpcStream,
  private val messageSerializer: Serializer<TResponse>,
) : ResponseCallback<TResponse> {
  private val log = LogManager.getLogger(this.javaClass)

  override fun respond(data: TResponse): SafeFuture<Void> =
    rpcStream.writeBytes(Bytes.wrap(messageSerializer.serialize(data)))

  override fun respondAndCompleteSuccessfully(data: TResponse) {
    respond(data)
      .thenRun { completeSuccessfully() }
      .finish(
        RootCauseExceptionHandler
          .builder()
          .addCatch(
            ClosedChannelException::class.java,
          ) { err -> log.trace("Failed to write because channel was closed", err) }
          .defaultCatch { err -> log.error("Failed to write req/resp response", err) },
      )
  }

  override fun completeSuccessfully() {
    rpcStream.closeWriteStream().ifExceptionGetsHereRaiseABug()
  }

  override fun completeWithErrorResponse(error: RpcException) {
  }

  override fun completeWithUnexpectedError(error: Throwable) {
  }
}

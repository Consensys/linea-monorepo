/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import java.util.Optional
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.p2p.rpc.RpcResponseHandler

class MaruRpcResponseHandler : RpcResponseHandler<Bytes> {
  val future = SafeFuture<Bytes>()

  override fun onResponse(response: Bytes): SafeFuture<*> {
    future.complete(response)
    return SafeFuture.completedFuture(response)
  }

  override fun onCompleted(error: Optional<out Throwable>) {
    if (error.isEmpty) {
      return // if needed do something when the response is completed successfully
    } else {
      throw error.get()
    }
  }

  fun response(): SafeFuture<Bytes> = future
}

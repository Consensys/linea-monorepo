/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import maru.p2p.messages.StatusHandler
import maru.p2p.messages.StatusMessageFactory
import maru.p2p.messages.StatusMessageSerDe
import maru.p2p.messages.StatusSerDe

class RpcMethods(
  statusMessageFactory: StatusMessageFactory,
  lineaRpcProtocolIdGenerator: LineaRpcProtocolIdGenerator,
  private val peerLookup: () -> PeerLookup,
) {
  val statusMessageSerDe = StatusMessageSerDe(StatusSerDe())
  val statusRpcMethod by lazy {
    MaruRpcMethod(
      messageType = RpcMessageType.STATUS,
      rpcMessageHandler = StatusHandler(statusMessageFactory),
      requestMessageSerDe = statusMessageSerDe,
      responseMessageSerDe = statusMessageSerDe,
      peerLookup = peerLookup.invoke(),
      protocolIdGenerator = lineaRpcProtocolIdGenerator,
      version = Version.V1,
    )
  }

  fun status() = statusRpcMethod

  fun all(): List<MaruRpcMethod<*, *>> = listOf(statusRpcMethod)
}

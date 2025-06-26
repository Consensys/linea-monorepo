/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import maru.consensus.ForkIdHashProvider
import maru.database.BeaconChain
import maru.p2p.messages.StatusHandler
import maru.p2p.messages.StatusMessageSerDe
import maru.p2p.messages.StatusSerDe

class RpcMethodFactory(
  private val beaconChain: BeaconChain,
  private val forkIdHashProvider: ForkIdHashProvider,
  chainId: UInt,
) {
  private val protocolIdGenerator = LineaRpcProtocolIdGenerator(chainId = chainId)

  fun createRpcMethods(peerLookup: PeerLookup): Map<RpcMessageType, MaruRpcMethod<*, *>> {
    val statusMessageSerDe = StatusMessageSerDe(StatusSerDe())
    val statusRpcMethod =
      MaruRpcMethod(
        messageType = RpcMessageType.STATUS,
        rpcMessageHandler = StatusHandler(beaconChain, forkIdHashProvider),
        requestMessageSerDe = statusMessageSerDe,
        responseMessageSerDe = statusMessageSerDe,
        peerLookup = peerLookup,
        protocolIdGenerator = protocolIdGenerator,
        version = Version.V1,
      )

    return mapOf(RpcMessageType.STATUS to statusRpcMethod)
  }
}

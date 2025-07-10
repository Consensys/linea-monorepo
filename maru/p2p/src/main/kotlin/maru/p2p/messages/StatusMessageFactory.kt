/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.messages

import maru.consensus.ForkIdHashProvider
import maru.database.BeaconChain
import maru.p2p.Message
import maru.p2p.RpcMessageType
import maru.p2p.Version

class StatusMessageFactory(
  private val beaconChain: BeaconChain,
  private val forkIdHashProvider: ForkIdHashProvider,
) {
  fun createStatusMessage(): Message<Status, RpcMessageType> {
    val forkIdHash = forkIdHashProvider.currentForkIdHash()
    val latestBeaconBlockHeader = beaconChain.getLatestBeaconState().latestBeaconBlockHeader
    val statusPayload = Status(forkIdHash = forkIdHash, latestBeaconBlockHeader.hash, latestBeaconBlockHeader.number)
    val statusMessage =
      Message(
        type = RpcMessageType.STATUS,
        version = Version.V1,
        payload = statusPayload,
      )
    return statusMessage
  }
}

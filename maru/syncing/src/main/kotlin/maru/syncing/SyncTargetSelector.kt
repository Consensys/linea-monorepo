/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import maru.p2p.PeersHeadBlockProvider
import maru.services.LongRunningService

/**
 * Responsible to keep track of peer's STATUS and select the head of the chain
 */
interface SyncTargetSelector {
  fun selectBestSyncTarget(peerHeads: List<ULong>): ULong
}

class MostFrequentHeadTargetSelector : SyncTargetSelector {
  override fun selectBestSyncTarget(peerHeads: List<ULong>): ULong {
    TODO("Not implemented yet")
  }
}

fun interface SyncTargetUpdateHandler {
  fun onChainHeadUpdated(beaconBlockNumber: ULong)
}

/**
 * polls periodically peers chain head
 * and notify SyncTargetSelector when chain head of any peer changes
 *
 * it only notifies syncTargetUpdateHandler when there is actual update, not on every tick/calculation
 *
 * MaruPeerManager -> PeerChainTracker -> SyncController
 */
class PeerChainTracker(
  val peersHeadsProvider: PeersHeadBlockProvider,
  val syncTargetUpdateHandler: SyncTargetUpdateHandler,
  val targetChainHeadCalculator: SyncTargetSelector,
) : LongRunningService {
  override fun start() {
    TODO("Not yet implemented")
  }

  override fun stop() {
    TODO("Not yet implemented")
  }
}

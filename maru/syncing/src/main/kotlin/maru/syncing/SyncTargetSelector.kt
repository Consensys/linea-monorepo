/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

/**
 * Responsible to keep track of peer's STATUS and select the head of the chain
 */
fun interface SyncTargetSelector {
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

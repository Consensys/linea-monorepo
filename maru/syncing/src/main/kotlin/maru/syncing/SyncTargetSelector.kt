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
    require(peerHeads.isNotEmpty()) { "Peer heads list cannot be empty" }

    val frequencyMap = peerHeads.groupingBy { it }.eachCount()
    val maxFrequency = frequencyMap.values.max()

    // Among all heads with max frequency, return the highest value
    return frequencyMap
      .filterValues { it == maxFrequency }
      .keys
      .max()
  }
}

fun interface BeaconSyncTargetUpdateHandler {
  fun onBeaconChainSyncTargetUpdated(syncTargetBlockNumber: ULong)
}

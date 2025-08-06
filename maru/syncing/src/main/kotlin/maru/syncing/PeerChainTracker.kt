/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import java.util.Timer
import java.util.UUID
import kotlin.concurrent.timerTask
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds
import maru.p2p.PeersHeadBlockProvider
import maru.services.LongRunningService
import org.apache.logging.log4j.LogManager

/**
 * polls periodically peers chain head
 * and notify SyncTargetSelector when chain head of any peer changes
 *
 * it only notifies syncTargetUpdateHandler when there is actual update, not on every tick/calculation
 *
 * MaruPeerManager -> PeerChainTracker -> SyncController
 */
class PeerChainTracker(
  private val peersHeadsProvider: PeersHeadBlockProvider,
  private val beaconSyncTargetUpdateHandler: BeaconSyncTargetUpdateHandler,
  private val targetChainHeadCalculator: SyncTargetSelector,
  private val config: Config,
  private val timerFactory: (String, Boolean) -> Timer = { name, isDaemon ->
    Timer(
      "$name-${UUID.randomUUID()}",
      isDaemon,
    )
  },
) : LongRunningService {
  private val log = LogManager.getLogger(this.javaClass)

  data class Config(
    val pollingUpdateInterval: Duration,
    val granularity: UInt, // Resolution of the peer heights
  ) {
    init {
      require(granularity > 0U) { "Granularity should not be 0!" }
    }
  }

  private var peers = mutableMapOf<String, ULong>()
  private var lastNotifiedTarget: ULong? = null // 0 is an Ok magic number, since it represents Genesis
  private var isRunning = false

  // Marked as volatile to ensure visibility across threads
  private var poller: Timer? = null

  /**
   * Rounds the block height according to the configured granularity
   */
  private fun roundHeight(height: ULong): ULong {
    if (config.granularity <= 1u) return height

    return (height / config.granularity.toULong()) * config.granularity.toULong()
  }

  /**
   * Updates the peer view and triggers sync target updates if needed
   */
  private fun updatePeerView() {
    log.trace("Updating peer view")
    val newPeerHeads = peersHeadsProvider.getPeersHeads()

    val roundedNewPeerHeads = newPeerHeads.mapValues { roundHeight(it.value) }
    peers = roundedNewPeerHeads.toMutableMap()
    log.debug("Rounded peers peersSize={}", peers.size)
    // Update the state and recalculate the sync target
    val newSyncTarget =
      if (peers.isNotEmpty()) {
        targetChainHeadCalculator.selectBestSyncTarget(peers.values.toList())
      } else {
        0UL // If there are no peers, let's just say we're synced, because we don't know better
      }
    log.trace("Selected best syncTarget={} lastNotifiedTarget={}", newSyncTarget, lastNotifiedTarget)
    if (newSyncTarget != lastNotifiedTarget) { // Only send an update if there's an actual target change
      beaconSyncTargetUpdateHandler.onBeaconChainSyncTargetUpdated(newSyncTarget)
      log.debug("Notified about the new syncTarget={}", newSyncTarget)
      lastNotifiedTarget = newSyncTarget
    }
  }

  override fun start() {
    synchronized(this) {
      if (isRunning) {
        return // Already running, don't start again
      }

      // Always create a new timer when starting
      poller = timerFactory("peer-chain-tracker", true)

      poller!!.scheduleAtFixedRate(
        /* task = */ timerTask { updatePeerView() },
        /* delay = */ 0.seconds.inWholeMilliseconds,
        /* period = */ config.pollingUpdateInterval.inWholeMilliseconds,
      )

      log.info("PeerChainTracker is started")
      isRunning = true
    }
  }

  override fun stop() {
    synchronized(this) {
      if (!isRunning) {
        return // Already stopped, don't stop again
      }

      poller?.cancel()
      poller = null
      isRunning = false
      lastNotifiedTarget = null
      log.info("PeerChainTracker is stopped")
    }
  }
}

/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.beaconchain.pipeline

import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit
import java.util.concurrent.TimeoutException
import java.util.concurrent.atomic.AtomicBoolean
import java.util.function.Function
import kotlin.time.Duration
import maru.p2p.MaruPeer
import maru.p2p.PeerLookup
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.util.log.LogUtil
import tech.pegasys.teku.networking.p2p.peer.DisconnectReason
import tech.pegasys.teku.networking.p2p.reputation.ReputationAdjustment

class DownloadBlocksStep(
  private val peerLookup: PeerLookup,
  private val maxRetries: UInt,
  private val blockRangeRequestTimeout: Duration,
  private val syncTargetProvider: () -> ULong,
) : Function<SyncTargetRange, CompletableFuture<List<SealedBlockWithPeer>>> {
  private val log: Logger = LogManager.getLogger(this.javaClass)
  private val shouldLog = AtomicBoolean(true)

  override fun apply(targetRange: SyncTargetRange): CompletableFuture<List<SealedBlockWithPeer>> =
    CompletableFuture.supplyAsync {
      var startBlockNumber = targetRange.startBlock
      val count = targetRange.endBlock - targetRange.startBlock + 1uL
      var remaining = count
      val downloadedBlocks = mutableListOf<SealedBlockWithPeer>()
      var retries = 0u
      var peer: MaruPeer?

      LogUtil.throttledLog(
        log::info,
        "Downloading blocks: start clBlockNumber=$startBlockNumber count=$count",
        shouldLog,
        30,
      )

      do {
        val currentSyncTarget = syncTargetProvider()
        peer =
          peerLookup
            .getPeers()
            .filter {
              it.getStatus() != null &&
                it.getStatus()!!.latestBlockNumber >=
                currentSyncTarget
            }.random()
        try {
          peer
            .sendBeaconBlocksByRange(startBlockNumber, remaining)
            .orTimeout(blockRangeRequestTimeout.inWholeMilliseconds, TimeUnit.MILLISECONDS)
            .thenApply { response ->
              if (response.blocks.isEmpty()) {
                peer.adjustReputation(ReputationAdjustment.SMALL_PENALTY)
                log.debug("No blocks received from peer: {}", peer.id)
                retries++
              } else {
                val numBlocks = response.blocks.size.toULong()
                if (numBlocks > remaining) {
                  peer.disconnectCleanly(DisconnectReason.REMOTE_FAULT)
                  log.debug(
                    "Received more blocks than requested from peer: {}. Expected: {}, Received: {}",
                    peer.id,
                    remaining,
                    numBlocks,
                  )
                  retries++
                } else {
                  response.blocks.forEach { downloadedBlocks.add(SealedBlockWithPeer(it, peer)) }
                  startBlockNumber += numBlocks
                  remaining -= numBlocks
                  retries = 0u // Reset retries on successful download
                  peer.adjustReputation(ReputationAdjustment.SMALL_REWARD)
                }
              }
            }.join()
        } catch (e: Exception) {
          if (e.cause is TimeoutException) {
            log.debug("Timed out while downloading blocks from peer: {}", peer.id)
            peer.adjustReputation(ReputationAdjustment.LARGE_PENALTY)
          } else {
            log.debug("Failed to download blocks from peer: {}", peer.id, e)
          }
          retries++
        }
        if (retries >= maxRetries) {
          log.debug("Maximum retries reached.")
          throw Exception("Maximum retries reached.")
        }
      } while (downloadedBlocks.size.toULong() < count)
      downloadedBlocks
    }
}

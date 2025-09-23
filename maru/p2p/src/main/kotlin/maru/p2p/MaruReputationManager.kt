/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import com.google.common.base.MoreObjects
import java.util.EnumSet
import java.util.Optional
import java.util.concurrent.atomic.AtomicInteger
import kotlin.concurrent.Volatile
import kotlin.math.min
import maru.config.P2PConfig
import maru.metrics.BesuMetricsCategoryAdapter
import maru.metrics.MaruMetricsCategory
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.plugin.services.MetricsSystem
import tech.pegasys.teku.infrastructure.collections.cache.Cache
import tech.pegasys.teku.infrastructure.collections.cache.LRUCache
import tech.pegasys.teku.infrastructure.time.TimeProvider
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.networking.p2p.network.PeerAddress
import tech.pegasys.teku.networking.p2p.peer.DisconnectReason
import tech.pegasys.teku.networking.p2p.peer.NodeId
import tech.pegasys.teku.networking.p2p.reputation.ReputationAdjustment
import tech.pegasys.teku.networking.p2p.reputation.ReputationManager

class MaruReputationManager(
  metricsSystem: MetricsSystem,
  private val timeProvider: TimeProvider,
  private val isStaticPeer: (NodeId) -> Boolean,
  reputationConfig: P2PConfig.Reputation,
) : ReputationManager {
  companion object {
    private const val DEFAULT_SCORE = 0
    private val BAN_REASONS: EnumSet<DisconnectReason?> =
      EnumSet.of(
        DisconnectReason.IRRELEVANT_NETWORK,
        DisconnectReason.UNABLE_TO_VERIFY_NETWORK,
        DisconnectReason.REMOTE_FAULT,
      )
  }

  private val log: Logger = LogManager.getLogger(this.javaClass)

  private val cooldownPeriod: Long = reputationConfig.cooldownPeriod.inWholeMilliseconds
  private val banPeriod: Long = reputationConfig.banPeriod.inWholeMilliseconds
  private val disconnectScoreThreshold: Int = reputationConfig.disconnectScoreThreshold
  private val maxReputationScore: Int = reputationConfig.maxReputation
  private val largeChange: Int = reputationConfig.largeChange
  private val smallChange: Int = reputationConfig.smallChange
  private val peerReputations: Cache<NodeId, Reputation> = LRUCache.create(reputationConfig.capacity)

  init {
    metricsSystem.createIntegerGauge(
      BesuMetricsCategoryAdapter.from(MaruMetricsCategory.STORAGE),
      "peer_reputation_cache_size",
      "Total number of peer reputations tracked",
      peerReputations::size,
    )
  }

  override fun reportInitiatedConnectionFailed(peerAddress: PeerAddress) {
    getOrCreateReputation(peerAddress)
      .reportInitiatedConnectionFailed(timeProvider.timeInMillis)
  }

  override fun isConnectionInitiationAllowed(peerAddress: PeerAddress): Boolean =
    peerReputations
      .getCached(peerAddress.id)
      .map { it.shouldInitiateConnection(timeProvider.timeInMillis) }
      .orElse(true)

  override fun reportInitiatedConnectionSuccessful(peerAddress: PeerAddress) {
    getOrCreateReputation(peerAddress).reportInitiatedConnectionSuccessful()
  }

  override fun reportDisconnection(
    peerAddress: PeerAddress,
    reason: Optional<DisconnectReason>,
    locallyInitiated: Boolean,
  ) {
    log.trace(
      "Reporting disconnection: peer={}, reason={}, locallyInitiated={}",
      peerAddress,
      reason.orElse(null),
      locallyInitiated,
    )
    getOrCreateReputation(peerAddress)
      .reportDisconnection(timeProvider.timeInMillis, reason, locallyInitiated)
  }

  override fun adjustReputation(
    peerAddress: PeerAddress,
    effect: ReputationAdjustment,
  ): Boolean {
    if (isStaticPeer(peerAddress.id)) {
      return false
    }
    return getOrCreateReputation(peerAddress)
      .adjustReputation(effect, timeProvider.timeInMillis)
      .also { if (it) log.debug("Disconnecting peer={} after adjustment", peerAddress) }
  }

  private fun getOrCreateReputation(peerAddress: PeerAddress): Reputation =
    peerReputations.get(peerAddress.id) {
      Reputation()
    }

  private fun toScoreDelta(adjustment: ReputationAdjustment): Int =
    when (adjustment) {
      ReputationAdjustment.LARGE_PENALTY -> -largeChange
      ReputationAdjustment.SMALL_PENALTY -> -smallChange
      ReputationAdjustment.SMALL_REWARD -> smallChange
      ReputationAdjustment.LARGE_REWARD -> largeChange
    }

  inner class Reputation {
    @Volatile
    private var suitableAfter: Optional<UInt64> = Optional.empty()
    private val score = AtomicInteger(DEFAULT_SCORE)

    fun reportInitiatedConnectionFailed(failureTime: UInt64) {
      suitableAfter = Optional.of(failureTime.plus(cooldownPeriod))
    }

    fun shouldInitiateConnection(currentTime: UInt64): Boolean = isSuitableAt(currentTime)

    private fun isSuitableAt(someTime: UInt64): Boolean = suitableAfter.map { it < someTime }.orElse(true)

    fun reportInitiatedConnectionSuccessful() {
      suitableAfter = Optional.empty()
    }

    fun reportDisconnection(
      disconnectTime: UInt64,
      reason: Optional<DisconnectReason>,
      locallyInitiated: Boolean,
    ) {
      if (isSuitableAt(disconnectTime)) {
        if (isLocallyConsideredUnsuitable(reason, locallyInitiated) ||
          reason.map { it.isPermanent }.orElse(false)
        ) {
          suitableAfter = Optional.of(disconnectTime.plus(banPeriod))
          score.set(DEFAULT_SCORE)
        } else {
          suitableAfter = Optional.of(disconnectTime.plus(cooldownPeriod))
        }
      }
    }

    fun isLocallyConsideredUnsuitable(
      reason: Optional<DisconnectReason>,
      locallyInitiated: Boolean,
    ): Boolean = locallyInitiated && reason.map { BAN_REASONS.contains(it) }.orElse(false)

    fun adjustReputation(
      effect: ReputationAdjustment,
      currentTime: UInt64,
    ): Boolean {
      if (!isSuitableAt(currentTime)) {
        return score.get() <= disconnectScoreThreshold
      }
      val newScore =
        score.updateAndGet { current ->
          min(maxReputationScore, current + toScoreDelta(effect))
        }
      val shouldDisconnect = newScore <= disconnectScoreThreshold
      if (shouldDisconnect) {
        suitableAfter = Optional.of(currentTime.plus(banPeriod))
        score.set(DEFAULT_SCORE)
      }
      return shouldDisconnect
    }

    override fun toString(): String =
      MoreObjects
        .toStringHelper(this)
        .add("suitableAfter", suitableAfter)
        .add("score", score)
        .toString()
  }
}

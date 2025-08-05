/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.beaconchain

import java.util.concurrent.CancellationException
import java.util.concurrent.ExecutorService
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicReference
import maru.consensus.ValidatorProvider
import maru.database.BeaconChain
import maru.metrics.MaruMetricsCategory
import maru.p2p.PeerLookup
import maru.services.LongRunningService
import maru.subscription.InOrderFanoutSubscriptionManager
import maru.subscription.SubscriptionManager
import maru.syncing.CLSyncService
import maru.syncing.beaconchain.pipeline.BeaconChainDownloadPipelineFactory
import maru.syncing.beaconchain.pipeline.BeaconChainPipeline
import net.consensys.linea.metrics.MetricsFacade
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.plugin.services.MetricsSystem

class CLSyncServiceImpl(
  private val beaconChain: BeaconChain,
  private val validatorProvider: ValidatorProvider,
  private val allowEmptyBlocks: Boolean,
  private var executorService: ExecutorService,
  pipelineConfig: BeaconChainDownloadPipelineFactory.Config,
  peerLookup: PeerLookup,
  besuMetrics: MetricsSystem,
  metricsFacade: MetricsFacade,
) : CLSyncService,
  LongRunningService {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val beaconChainPipeline = AtomicReference<BeaconChainPipeline?>(null)
  private val syncTarget = AtomicReference(0UL)
  private val started = AtomicBoolean(false)
  private val syncCompleteHanders: SubscriptionManager<ULong> = InOrderFanoutSubscriptionManager()
  private val blockImporter =
    SyncSealedBlockImporterFactory()
      .create(
        beaconChain = beaconChain,
        validatorProvider = validatorProvider,
        allowEmptyBlocks = allowEmptyBlocks,
      )
  private var pipelineFactory =
    BeaconChainDownloadPipelineFactory(blockImporter, besuMetrics, peerLookup, pipelineConfig) {
      syncTarget.get()
    }
  private val pipelineRestartCounter =
    metricsFacade.createCounter(
      category = MaruMetricsCategory.SYNCHRONIZER,
      name = "beaconchain.restart.counter",
      description = "Count of chain pipeline restarts",
    )

  override fun setSyncTarget(syncTarget: ULong) {
    check(started.get()) { "Sync service must be started before setting sync target" }

    val oldTarget = this.syncTarget.getAndSet(syncTarget)
    if (oldTarget != syncTarget) {
      log.info("Syncing target updated from {} to {}", oldTarget, syncTarget)
    }

    // If the pipeline is already running, we don't need to start a new one
    if (beaconChainPipeline.get() == null) {
      startSync()
    }
  }

  @Synchronized
  private fun startSync() {
    val startBlock = beaconChain.getLatestBeaconState().latestBeaconBlockHeader.number + 1UL
    val pipeline = pipelineFactory.createPipeline(startBlock)

    if (beaconChainPipeline.compareAndSet(null, pipeline)) {
      pipeline.pipeline.start(executorService).handle { _, ex ->
        if (ex != null && ex !is CancellationException) {
          log.error("Sync pipeline failed, restarting", ex)
          pipelineRestartCounter.increment()
          if (beaconChainPipeline.compareAndSet(pipeline, null)) {
            startSync()
          }
        } else {
          val completedSyncTarget = pipeline.target()
          beaconChainPipeline.compareAndSet(pipeline, null)
          log.info("Sync completed syncTarget={}", completedSyncTarget)
          syncCompleteHanders.notifySubscribers(completedSyncTarget)
        }
      }
    }
  }

  override fun onSyncComplete(handler: (ULong) -> Unit) {
    syncCompleteHanders.addSyncSubscriber(handler)
  }

  override fun start() {
    if (!started.compareAndSet(false, true)) {
      log.warn("Sync service is already started")
    }
  }

  override fun stop() {
    if (started.compareAndSet(true, false)) {
      beaconChainPipeline.get()?.pipeline?.abort()
    }
  }
}

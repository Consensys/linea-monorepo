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
import net.consensys.linea.metrics.MetricsFacade
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.plugin.services.MetricsSystem
import org.hyperledger.besu.services.pipeline.Pipeline

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
  private var pipeline: Pipeline<*>? = null
  private val syncTarget: AtomicReference<ULong> = AtomicReference(0UL)
  private val syncHandlerSubscriptionIds = mutableListOf<String>()
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
    log.info("Syncing started syncTarget={}", syncTarget)
    this.syncTarget.set(syncTarget)

    // If the pipeline is already running, we don't need to start a new one
    if (pipeline == null) {
      startSync()
    }
  }

  private fun startSync() {
    val startBlock = beaconChain.getLatestBeaconState().latestBeaconBlockHeader.number + 1UL
    val pipeline = pipelineFactory.createPipeline(startBlock)
    this.pipeline = pipeline

    pipeline.start(executorService).handle { _, ex ->
      if (ex != null && ex !is CancellationException) {
        log.error("Sync pipeline failed, restarting", ex)
        pipelineRestartCounter.increment()
        this.pipeline = null
        startSync()
      } else {
        log.info("Sync pipeline completed successfully")
        this.pipeline = null
        val completedSyncTarget = syncTarget.get() // Capture current target at completion
        syncCompleteHanders.notifySubscribers(completedSyncTarget)
      }
    }
  }

  override fun onSyncComplete(handler: (ULong) -> Unit) {
    val subscriptionId = handler.toString()
    syncHandlerSubscriptionIds.add(subscriptionId)
    syncCompleteHanders.addSyncSubscriber(subscriptionId, handler)
  }

  override fun start() {
  }

  override fun stop() {
    pipeline?.abort()
  }
}

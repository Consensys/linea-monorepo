/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import java.util.concurrent.Executors
import java.util.concurrent.locks.ReentrantReadWriteLock
import kotlin.concurrent.read
import kotlin.concurrent.write
import kotlin.time.Duration.Companion.milliseconds
import maru.consensus.ValidatorProvider
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import maru.p2p.PeerLookup
import maru.p2p.PeersHeadBlockProvider
import maru.services.LongRunningService
import maru.subscription.InOrderFanoutSubscriptionManager
import maru.syncing.beaconchain.CLSyncServiceImpl
import maru.syncing.beaconchain.pipeline.BeaconChainDownloadPipelineFactory
import net.consensys.linea.metrics.MetricsFacade
import org.hyperledger.besu.plugin.services.MetricsSystem

internal data class SyncState(
  val clStatus: CLSyncStatus,
  val elStatus: ELSyncStatus,
)

class BeaconSyncControllerImpl(
  private val beaconChain: BeaconChain,
  private val clSyncService: CLSyncService,
  clState: CLSyncStatus = CLSyncStatus.SYNCING,
  elState: ELSyncStatus = ELSyncStatus.SYNCING,
) : SyncStatusProvider,
  BeaconSyncTargetUpdateHandler {
  private val lock = ReentrantReadWriteLock()
  private var currentState = SyncState(clState, elState)
  private var currentSyncTarget: ULong? = null

  private val clSyncHandlers = InOrderFanoutSubscriptionManager<CLSyncStatus>()
  private val elSyncHandlers = InOrderFanoutSubscriptionManager<ELSyncStatus>()
  private val beaconSyncCompleteHandlers = InOrderFanoutSubscriptionManager<Unit>()
  private val fullSyncCompleteHandlers = InOrderFanoutSubscriptionManager<Unit>()

  init {
    clSyncService.onSyncComplete { syncTarget ->
      updateClSyncStatus(CLSyncStatus.SYNCED)
    }
  }

  override fun getCLSyncStatus(): CLSyncStatus = lock.read { currentState.clStatus }

  override fun getElSyncStatus(): ELSyncStatus = lock.read { currentState.elStatus }

  override fun isNodeFullInSync(): Boolean = lock.read { isELSynced() && isBeaconChainSynced() }

  override fun isBeaconChainSynced(): Boolean = lock.read { currentState.clStatus == CLSyncStatus.SYNCED }

  override fun isELSynced(): Boolean = lock.read { currentState.elStatus == ELSyncStatus.SYNCED }

  override fun onClSyncStatusUpdate(handler: (CLSyncStatus) -> Unit) {
    clSyncHandlers.addSyncSubscriber(handler.toString(), handler)
  }

  override fun onElSyncStatusUpdate(handler: (ELSyncStatus) -> Unit) {
    elSyncHandlers.addSyncSubscriber(handler.toString(), handler)
  }

  override fun onBeaconSyncComplete(handler: () -> Unit) {
    beaconSyncCompleteHandlers.addSyncSubscriber(handler.toString()) { handler() }
  }

  override fun onFullSyncComplete(handler: () -> Unit) {
    fullSyncCompleteHandlers.addSyncSubscriber(handler.toString()) { handler() }
  }

  fun updateClSyncStatus(newStatus: CLSyncStatus) {
    val callbacks: List<() -> Unit> =
      lock.write {
        val previousState = currentState

        if (previousState.clStatus == newStatus) return@write emptyList()

        // When CL starts syncing, EL must also be syncing (EL follows CL rule)
        val newElStatus =
          when {
            newStatus == CLSyncStatus.SYNCING -> ELSyncStatus.SYNCING
            else -> previousState.elStatus
          }

        currentState = SyncState(newStatus, newElStatus)

        buildList {
          add { clSyncHandlers.notifySubscribers(newStatus) }

          // If EL status changed due to CL change, notify EL handlers
          if (newElStatus != previousState.elStatus) {
            add { elSyncHandlers.notifySubscribers(newElStatus) }
          }

          // If CL moved from SYNCING to SYNCED, beacon sync is complete
          if (previousState.clStatus == CLSyncStatus.SYNCING && newStatus == CLSyncStatus.SYNCED) {
            add { beaconSyncCompleteHandlers.notifySubscribers(Unit) }

            // Check if this makes the node fully synced
            if (isNodeFullInSync()) {
              add { fullSyncCompleteHandlers.notifySubscribers(Unit) }
            }
          }
        }
      }

    // Fire callbacks outside of lock
    callbacks.forEach { it() }
  }

  fun updateElSyncStatus(newStatus: ELSyncStatus) {
    val callbacks: List<() -> Unit> =
      lock.write {
        val previousState = currentState

        // EL can't be synced if CL is still syncing (EL follows CL rule)
        if (previousState.clStatus == CLSyncStatus.SYNCING ||
          previousState.elStatus == newStatus
        ) {
          return@write emptyList()
        }

        currentState = SyncState(previousState.clStatus, newStatus)

        buildList {
          add { elSyncHandlers.notifySubscribers(newStatus) }

          // Check if this transition from EL SYNCING->SYNCED when CL is already SYNCED makes us fully synced
          if (previousState.elStatus == ELSyncStatus.SYNCING &&
            newStatus == ELSyncStatus.SYNCED &&
            currentState.clStatus == CLSyncStatus.SYNCED
          ) {
            add { fullSyncCompleteHandlers.notifySubscribers(Unit) }
          }
        }
      }

    // Fire callbacks outside of lock
    callbacks.forEach { it() }
  }

  override fun onBeaconChainSyncTargetUpdated(syncTargetBlockNumber: ULong) {
    val previousTarget =
      lock.write {
        val prev = currentSyncTarget
        currentSyncTarget = syncTargetBlockNumber
        prev
      }

    // Early return if same target (prevents redundant operations)
    if (previousTarget == syncTargetBlockNumber) {
      return
    }

    val currentHead = beaconChain.getLatestBeaconState().latestBeaconBlockHeader.number

    if (syncTargetBlockNumber > currentHead) {
      updateClSyncStatus(CLSyncStatus.SYNCING)
      clSyncService.setSyncTarget(syncTargetBlockNumber)
    } else {
      // We're caught up or ahead, but check if we were previously syncing
      val currentClStatus = lock.read { currentState.clStatus }
      if (currentClStatus == CLSyncStatus.SYNCING) {
        // Transition from SYNCING to SYNCED
        clSyncService.setSyncTarget(syncTargetBlockNumber)
      }
      // If already SYNCED, do nothing
    }
  }

  // Helper method for testing
  internal fun captureStateSnapshot(): SyncState = lock.read { currentState }

  companion object {
    fun create(
      beaconChain: BeaconChain,
      elManager: ExecutionLayerManager,
      peersHeadsProvider: PeersHeadBlockProvider,
      targetChainHeadCalculator: SyncTargetSelector = MostFrequentHeadTargetSelector(),
      peerChainTrackerConfig: PeerChainTracker.Config,
      validatorProvider: ValidatorProvider,
      peerLookup: PeerLookup,
      besuMetrics: MetricsSystem,
      metricsFacade: MetricsFacade,
      allowEmptyBlocks: Boolean = true,
    ): SyncControllerManager {
      val clSyncPipeline = CLSyncPipelineImpl()

      val controller =
        BeaconSyncControllerImpl(
          beaconChain = beaconChain,
          clSyncService = clSyncPipeline,
        )

      val elSyncService =
        ELSyncService(
          beaconChain = beaconChain,
          executionLayerManager = elManager,
          onStatusChange = controller::updateElSyncStatus,
          config =
            ELSyncService.Config(
              pollingInterval = 5000.milliseconds,
            ),
        )
      val clSyncService =
        CLSyncServiceImpl(
          beaconChain = beaconChain,
          validatorProvider = validatorProvider,
          allowEmptyBlocks = allowEmptyBlocks,
          executorService = Executors.newCachedThreadPool(),
          pipelineConfig = BeaconChainDownloadPipelineFactory.Config(),
          peerLookup = peerLookup,
          besuMetrics = besuMetrics,
          metricsFacade = metricsFacade,
        )

      val peerChainTracker =
        PeerChainTracker(
          peersHeadsProvider = peersHeadsProvider,
          beaconSyncTargetUpdateHandler = controller,
          targetChainHeadCalculator = targetChainHeadCalculator,
          config = peerChainTrackerConfig,
        )

      return SyncControllerManager(
        syncStatusController = controller,
        elSyncService = elSyncService,
        clSyncService = clSyncService,
        peerChainTracker = peerChainTracker,
      )
    }
  }
}

class SyncControllerManager(
  val syncStatusController: BeaconSyncControllerImpl,
  val elSyncService: LongRunningService,
  val clSyncService: LongRunningService,
  val peerChainTracker: PeerChainTracker,
) : SyncStatusProvider by syncStatusController,
  LongRunningService {
  override fun start() {
    // TODO: remove when clSyncService is implemented
    syncStatusController.updateClSyncStatus(CLSyncStatus.SYNCED)
    // TODO: remove when elSyncService is implemented
    syncStatusController.updateElSyncStatus(ELSyncStatus.SYNCED)
    clSyncService.start()
    elSyncService.start()
    peerChainTracker.start()
  }

  override fun stop() {
    clSyncService.stop()
    elSyncService.stop()
    peerChainTracker.stop()
    // Setting to default status so that SYNCING -> SYNCED will actually trigger the callbacks
    syncStatusController.updateClSyncStatus(CLSyncStatus.SYNCING)
    syncStatusController.updateElSyncStatus(ELSyncStatus.SYNCING)
  }
}

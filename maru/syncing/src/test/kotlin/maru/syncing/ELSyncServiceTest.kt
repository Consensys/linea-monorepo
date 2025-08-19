/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import kotlin.concurrent.atomics.AtomicInt
import kotlin.concurrent.atomics.ExperimentalAtomicApi
import kotlin.concurrent.atomics.incrementAndFetch
import kotlin.time.Duration.Companion.seconds
import maru.consensus.state.FinalizationProvider
import maru.consensus.state.FinalizationState
import maru.consensus.state.InstantFinalizationProvider
import maru.core.ext.DataGenerators
import maru.database.InMemoryBeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import maru.executionlayer.manager.ExecutionPayloadStatus
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import maru.executionlayer.manager.PayloadStatus
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.Mockito.verify
import org.mockito.kotlin.any
import org.mockito.kotlin.eq
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import testutils.maru.TestablePeriodicTimer

@OptIn(ExperimentalAtomicApi::class)
class ELSyncServiceTest {
  @Test
  fun `should set sync status to Synced for genesis block`() {
    var elSyncStatus: ELSyncStatus? = null
    val onStatusChangeCount = AtomicInt(0)
    val onStatusChange: (ELSyncStatus) -> Unit = {
      elSyncStatus = it
      onStatusChangeCount.incrementAndFetch()
    }
    val config = ELSyncService.Config(pollingInterval = 1.seconds)
    val timer = TestablePeriodicTimer()
    val beaconChain =
      DataGenerators.genesisState(0uL, emptySet()).let {
        InMemoryBeaconChain(initialBeaconState = it.first, initialBeaconBlock = it.second)
      }
    val executionLayerManager = mock<ExecutionLayerManager>()
    val finalizationProvider: FinalizationProvider = InstantFinalizationProvider

    val elSyncService =
      ELSyncService(
        beaconChain = beaconChain,
        executionLayerManager = executionLayerManager,
        onStatusChange = onStatusChange,
        config = config,
        finalizationProvider = finalizationProvider,
        timerFactory = { _, _ -> timer },
      )

    elSyncService.start()
    assertThat(elSyncStatus).isNull()
    timer.runNextTask()
    assertThat(elSyncStatus).isEqualTo(ELSyncStatus.SYNCED)

    timer.runNextTask()
    timer.runNextTask()
    assertThat(elSyncStatus).isEqualTo(ELSyncStatus.SYNCED)
    elSyncService.stop()
  }

  @Test
  fun `should change el sync status when el is syncing and synced`() {
    var elSyncStatus: ELSyncStatus? = null
    val onStatusChangeCount = AtomicInt(0)
    val onStatusChange: (ELSyncStatus) -> Unit = {
      elSyncStatus = it
      onStatusChangeCount.incrementAndFetch()
    }
    val config = ELSyncService.Config(pollingInterval = 1.seconds)
    val timer = TestablePeriodicTimer()
    val beaconChain =
      DataGenerators.genesisState(0uL, emptySet()).let {
        InMemoryBeaconChain(initialBeaconState = it.first, initialBeaconBlock = it.second)
      }
    val executionLayerManager = mock<ExecutionLayerManager>()
    val finalizationProvider: FinalizationProvider = InstantFinalizationProvider
    val elSyncService =
      ELSyncService(
        beaconChain = beaconChain,
        executionLayerManager = executionLayerManager,
        onStatusChange = onStatusChange,
        config = config,
        finalizationProvider = finalizationProvider,
        timerFactory = { _, _ -> timer },
      )

    elSyncService.start()
    assertThat(elSyncStatus).isNull()
    timer.runNextTask()
    assertThat(elSyncStatus).isEqualTo(ELSyncStatus.SYNCED)

    beaconChain
      .newUpdater()
      .putBeaconState(DataGenerators.randomBeaconState(3uL))
      .putSealedBeaconBlock(DataGenerators.randomSealedBeaconBlock(3UL))
      .commit()

    whenever(executionLayerManager.newPayload(any()))
      .thenReturn(
        SafeFuture.completedFuture(
          PayloadStatus(ExecutionPayloadStatus.VALID, null, null),
        ),
      )

    whenever(executionLayerManager.setHead(any(), any(), any()))
      .thenReturn(
        SafeFuture.completedFuture(
          ForkChoiceUpdatedResult(PayloadStatus(ExecutionPayloadStatus.SYNCING, null, null), null),
        ),
      )
    timer.runNextTask()
    assertThat(elSyncStatus).isEqualTo(ELSyncStatus.SYNCING)

    whenever(executionLayerManager.newPayload(any()))
      .thenReturn(
        SafeFuture.completedFuture(
          PayloadStatus(ExecutionPayloadStatus.VALID, null, null),
        ),
      )

    whenever(executionLayerManager.setHead(any(), any(), any()))
      .thenReturn(
        SafeFuture.completedFuture(
          ForkChoiceUpdatedResult(PayloadStatus(ExecutionPayloadStatus.VALID, null, null), null),
        ),
      )
    timer.runNextTask()
    assertThat(elSyncStatus).isEqualTo(ELSyncStatus.SYNCED)
    assertThat(onStatusChangeCount.load()).isEqualTo(3)

    timer.runNextTask()
    timer.runNextTask()
    assertThat(elSyncStatus).isEqualTo(ELSyncStatus.SYNCED)
    elSyncService.stop()
  }

  @Test
  fun `should respect finalization provider when calling setHead`() {
    val config = ELSyncService.Config(pollingInterval = 1.seconds)
    val timer = TestablePeriodicTimer()

    val beaconChain =
      DataGenerators.genesisState(0uL, emptySet()).let {
        InMemoryBeaconChain(initialBeaconState = it.first, initialBeaconBlock = it.second)
      }
    val executionLayerManager = mock<ExecutionLayerManager>()

    // Create custom finalization provider that returns specific values
    val customSafeBlockHash = ByteArray(32) { 0xAA.toByte() }
    val customFinalizedBlockHash = ByteArray(32) { 0xBB.toByte() }
    val finalizationProvider: FinalizationProvider = {
      FinalizationState(
        safeBlockHash = customSafeBlockHash,
        finalizedBlockHash = customFinalizedBlockHash,
      )
    }

    val elSyncService =
      ELSyncService(
        beaconChain = beaconChain,
        executionLayerManager = executionLayerManager,
        onStatusChange = { },
        config = config,
        finalizationProvider = finalizationProvider,
        timerFactory = { _, _ -> timer },
      )

    // Add a block to the beacon chain to trigger EL sync
    val sealedBlock = DataGenerators.randomSealedBeaconBlock(3UL)
    beaconChain
      .newUpdater()
      .putBeaconState(DataGenerators.randomBeaconState(3uL))
      .putSealedBeaconBlock(sealedBlock)
      .commit()

    // Mock newPayload to return a valid PayloadStatus
    whenever(executionLayerManager.newPayload(any()))
      .thenReturn(
        SafeFuture.completedFuture(
          PayloadStatus(ExecutionPayloadStatus.VALID, null, null),
        ),
      )

    whenever(executionLayerManager.setHead(any(), any(), any()))
      .thenReturn(
        SafeFuture.completedFuture(
          ForkChoiceUpdatedResult(PayloadStatus(ExecutionPayloadStatus.VALID, null, null), null),
        ),
      )

    elSyncService.start()
    timer.runNextTask()

    // Verify that newPayload was called with the execution payload
    verify(executionLayerManager).newPayload(eq(sealedBlock.beaconBlock.beaconBlockBody.executionPayload))

    // Verify that setHead was called with the finalization provider's values
    verify(executionLayerManager).setHead(
      eq(sealedBlock.beaconBlock.beaconBlockBody.executionPayload.blockHash),
      eq(customSafeBlockHash),
      eq(customFinalizedBlockHash),
    )

    elSyncService.stop()
  }
}

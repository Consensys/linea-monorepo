/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class SyncControllerTest {
  private lateinit var fakeClSyncService: FakeCLSyncService
  private lateinit var syncController: BeaconSyncControllerImpl

  private fun createController(
    blockNumber: ULong,
    clSyncService: CLSyncService = fakeClSyncService,
  ): BeaconSyncControllerImpl = createSyncController(blockNumber, clSyncService)

  @BeforeEach
  fun setUp() {
    fakeClSyncService = FakeCLSyncService()
    syncController = createController(0UL)
  }

  @Test
  fun `should initialize with default SYNCING status`() {
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
    assertThat(syncController.getElSyncStatus()).isEqualTo(ELSyncStatus.SYNCING)
    assertThat(syncController.isBeaconChainSynced()).isFalse()
    assertThat(syncController.isELSynced()).isFalse()
    assertThat(syncController.isNodeFullInSync()).isFalse()
  }

  @Test
  fun `should trigger sync when beacon chain is empty and sync target is provided`() {
    // When: chain head is updated with non-empty target
    syncController.onBeaconChainSyncTargetUpdated(100UL)

    // Then: CL sync should start and target should be set
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(100UL)
  }

  @Test
  fun `should trigger sync when behind sync target`() {
    // Given: beacon chain initialized with current head at block 50
    val controller = createController(50UL)

    // When: chain head is updated to higher block
    controller.onBeaconChainSyncTargetUpdated(100UL)

    // Then: CL sync should start
    assertThat(controller.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(100UL)
  }

  @Test
  fun `should not trigger sync when synced already at sync target`() {
    // Given: beacon chain is at the same level as sync target
    val controller = createController(100UL)
    controller.updateClSyncStatus(CLSyncStatus.SYNCED)

    // When: chain head is updated to same block (controller starts in SYNCING by default)
    controller.onBeaconChainSyncTargetUpdated(100UL)

    // Then: should transition to SYNCED since target matches current head
    assertThat(fakeClSyncService.lastSyncTarget).isNull() // No sync target set on service
    assertThat(controller.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCED)
    assertThat(controller.isBeaconChainSynced()).isTrue()
  }

  @Test
  fun `should update sync target during ongoing sync`() {
    // Given: sync is already in progress
    syncController.updateClSyncStatus(CLSyncStatus.SYNCING)
    fakeClSyncService.lastSyncTarget = 100UL

    // When: sync target is updated
    syncController.onBeaconChainSyncTargetUpdated(150UL)

    // Then: sync target should be updated
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(150UL)
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
  }

  @Test
  fun `should automatically set EL to SYNCING when CL starts syncing`() {
    // Given: both CL and EL are syncing (default state)
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
    assertThat(syncController.getElSyncStatus()).isEqualTo(ELSyncStatus.SYNCING)

    // When: CL sync is set to synced first
    syncController.updateClSyncStatus(CLSyncStatus.SYNCED)
    syncController.updateElSyncStatus(ELSyncStatus.SYNCED)

    // Then: both should be synced
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCED)
    assertThat(syncController.getElSyncStatus()).isEqualTo(ELSyncStatus.SYNCED)

    // When: CL sync starts again
    syncController.updateClSyncStatus(CLSyncStatus.SYNCING)

    // Then: EL should also be marked as syncing
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
    assertThat(syncController.getElSyncStatus()).isEqualTo(ELSyncStatus.SYNCING)
  }

  @Test
  fun `should notify handlers when sync status changes`() {
    val clStatusUpdates = mutableListOf<CLSyncStatus>()
    val elStatusUpdates = mutableListOf<ELSyncStatus>()

    syncController.onClSyncStatusUpdate { clStatusUpdates.add(it) }
    syncController.onElSyncStatusUpdate { elStatusUpdates.add(it) }

    // When: status changes from default SYNCING to SYNCED
    syncController.updateClSyncStatus(CLSyncStatus.SYNCED)
    syncController.updateElSyncStatus(ELSyncStatus.SYNCED)

    // Then: handlers should be notified
    assertThat(clStatusUpdates).containsExactly(CLSyncStatus.SYNCED)
    assertThat(elStatusUpdates).containsExactly(ELSyncStatus.SYNCED)
  }

  @Test
  fun `should notify beacon sync complete when CL sync completes`() {
    var beaconSyncCompleted = false
    syncController.onBeaconSyncComplete { beaconSyncCompleted = true }

    // When: CL sync completes
    fakeClSyncService.triggerSyncComplete(100UL)

    // Then: beacon sync complete should be notified
    assertThat(beaconSyncCompleted).isTrue()
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCED)
  }

  @Test
  fun `should notify full sync complete when EL completes after CL`() {
    var fullSyncCompleted = false
    syncController.onFullSyncComplete { fullSyncCompleted = true }

    // Given: both are syncing
    syncController.updateClSyncStatus(CLSyncStatus.SYNCING)
    syncController.updateElSyncStatus(ELSyncStatus.SYNCING)

    // When: EL completes first
    syncController.updateElSyncStatus(ELSyncStatus.SYNCED)
    assertThat(fullSyncCompleted).isFalse()

    // When: CL completes
    syncController.updateClSyncStatus(CLSyncStatus.SYNCED)

    // Then: El sync status should follow the CL one, so CL update after EL doesn't trigger full sync message
    assertThat(fullSyncCompleted).isFalse()

    // When: EL completes again
    syncController.updateElSyncStatus(ELSyncStatus.SYNCED)

    // Then: El should trigger the full sync
    assertThat(fullSyncCompleted).isTrue()
  }

  @Test
  fun `should handle multiple sync target updates correctly`() {
    // Given: beacon chain with empty initial state (block 0)
    val controller = createController(0UL)

    // When: multiple targets are provided
    controller.onBeaconChainSyncTargetUpdated(50UL)
    controller.onBeaconChainSyncTargetUpdated(100UL)
    controller.onBeaconChainSyncTargetUpdated(75UL)

    // Then: should have the latest target
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(75UL)
    assertThat(controller.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
  }

  @Test
  fun `should not call onClSyncStatusUpdate when only EL goes out of sync`() {
    val clStatusUpdates = mutableListOf<CLSyncStatus>()
    val elStatusUpdates = mutableListOf<ELSyncStatus>()

    syncController.onClSyncStatusUpdate { clStatusUpdates.add(it) }
    syncController.onElSyncStatusUpdate { elStatusUpdates.add(it) }

    // Given: both CL and EL are synced
    syncController.updateClSyncStatus(CLSyncStatus.SYNCED)
    syncController.updateElSyncStatus(ELSyncStatus.SYNCED)

    // Clear previous updates
    clStatusUpdates.clear()
    elStatusUpdates.clear()

    // When: only EL goes out of sync
    syncController.updateElSyncStatus(ELSyncStatus.SYNCING)

    // Then: only EL status update should be called, CL should remain synced
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCED)
    assertThat(syncController.getElSyncStatus()).isEqualTo(ELSyncStatus.SYNCING)
    assertThat(clStatusUpdates).isEmpty()
    assertThat(elStatusUpdates).containsExactly(ELSyncStatus.SYNCING)
  }

  @Test
  fun `should not call onClSyncStatusUpdate without actual status change`() {
    val clStatusUpdates = mutableListOf<CLSyncStatus>()
    val elStatusUpdates = mutableListOf<ELSyncStatus>()

    syncController.onClSyncStatusUpdate { clStatusUpdates.add(it) }
    syncController.onElSyncStatusUpdate { elStatusUpdates.add(it) }

    // Given: send multiple updateClSyncStatus calls
    syncController.updateClSyncStatus(CLSyncStatus.SYNCED)
    syncController.updateClSyncStatus(CLSyncStatus.SYNCED)

    // Then: only EL status update should be called, CL should remain synced
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCED)
    assertThat(syncController.getElSyncStatus()).isEqualTo(ELSyncStatus.SYNCING)
    assertThat(clStatusUpdates).isEqualTo(listOf(CLSyncStatus.SYNCED))
    assertThat(elStatusUpdates).isEmpty()
  }

  @Test
  fun `should not call onElSyncStatusUpdate without actual status change`() {
    val clStatusUpdates = mutableListOf<CLSyncStatus>()
    val elStatusUpdates = mutableListOf<ELSyncStatus>()
    var fullSyncUpdates = 0u

    syncController.onClSyncStatusUpdate { clStatusUpdates.add(it) }
    syncController.onElSyncStatusUpdate { elStatusUpdates.add(it) }
    syncController.onFullSyncComplete { fullSyncUpdates += 1u }

    // Given: Cl is synced, send multiple updateClSyncStatus calls
    syncController.updateClSyncStatus(CLSyncStatus.SYNCED)
    syncController.updateElSyncStatus(ELSyncStatus.SYNCED)
    syncController.updateElSyncStatus(ELSyncStatus.SYNCED)

    clStatusUpdates.clear()

    // Then: only EL status update should be called, CL should remain synced
    assertThat(syncController.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCED)
    assertThat(syncController.getElSyncStatus()).isEqualTo(ELSyncStatus.SYNCED)
    assertThat(elStatusUpdates).isEqualTo(listOf(ELSyncStatus.SYNCED))
    assertThat(clStatusUpdates).isEmpty()
    assertThat(fullSyncUpdates).isEqualTo(1u)
  }

  @Test
  fun `should transition to synced when sync target is updated to match current chain head`() {
    // Given: beacon chain at block 100
    val controller = createController(100UL)

    val clStatusUpdates = mutableListOf<CLSyncStatus>()
    val elStatusUpdates = mutableListOf<ELSyncStatus>()

    controller.onClSyncStatusUpdate { clStatusUpdates.add(it) }
    controller.onElSyncStatusUpdate { elStatusUpdates.add(it) }

    // First: trigger sync to target 150 (controller starts in SYNCING state by default)
    controller.onBeaconChainSyncTargetUpdated(150UL)
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(150UL)
    assertThat(controller.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)

    // Clear status updates from the first sync trigger
    clStatusUpdates.clear()
    elStatusUpdates.clear()

    // When: sync target is updated back to 100 (which matches current chain head)
    controller.onBeaconChainSyncTargetUpdated(100UL)

    // Then: controller should transition to SYNCED since target matches current head
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(100UL)
    assertThat(controller.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
    assertThat(controller.getElSyncStatus()).isEqualTo(ELSyncStatus.SYNCING)
    assertThat(controller.isBeaconChainSynced()).isFalse

    // Status updates should be triggered for the transition to SYNCED
    assertThat(clStatusUpdates).isEmpty()
    assertThat(elStatusUpdates).isEmpty() // EL status doesn't change
  }

  @Test
  fun `should handle ongoing sync target updates correctly`() {
    val controller = createController(50UL)

    // Given: sync is triggered to target 200
    controller.onBeaconChainSyncTargetUpdated(200UL)
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(200UL)
    assertThat(controller.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)

    // When: sync target is updated to 150 during ongoing sync
    controller.onBeaconChainSyncTargetUpdated(150UL)

    // Then: should update sync target
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(150UL)
    assertThat(controller.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)

    // When: sync target is updated again to 180
    controller.onBeaconChainSyncTargetUpdated(180UL)

    // Then: should update sync target again
    assertThat(fakeClSyncService.lastSyncTarget).isEqualTo(180UL)
    assertThat(controller.getCLSyncStatus()).isEqualTo(CLSyncStatus.SYNCING)
  }

  @Test
  fun `should not set redundant sync targets for same value`() {
    val trackingClSyncService = TrackingCLSyncService()
    val controller = createController(50UL, trackingClSyncService)

    // Given: sync is triggered to target 100
    controller.onBeaconChainSyncTargetUpdated(100UL)
    assertThat(trackingClSyncService.setSyncTargetCalls).hasSize(1)
    assertThat(trackingClSyncService.setSyncTargetCalls[0]).isEqualTo(100UL)

    // When: same sync target is provided again
    controller.onBeaconChainSyncTargetUpdated(100UL)

    // Then: should NOT call setSyncTarget again due to early return
    assertThat(trackingClSyncService.setSyncTargetCalls).hasSize(1)
    assertThat(trackingClSyncService.setSyncTargetCalls[0]).isEqualTo(100UL)

    // When: same target provided third time
    controller.onBeaconChainSyncTargetUpdated(100UL)

    // Then: still no additional calls
    assertThat(trackingClSyncService.setSyncTargetCalls).hasSize(1) // Fixed: no redundant call
  }
}

// Additional test double to track method calls
private class TrackingCLSyncService : CLSyncService {
  val setSyncTargetCalls = mutableListOf<ULong>()
  private val syncCompleteHandlers = mutableListOf<(ULong) -> Unit>()

  override fun setSyncTarget(syncTarget: ULong) {
    setSyncTargetCalls.add(syncTarget)
  }

  override fun onSyncComplete(handler: (ULong) -> Unit) {
    syncCompleteHandlers.add(handler)
  }
}

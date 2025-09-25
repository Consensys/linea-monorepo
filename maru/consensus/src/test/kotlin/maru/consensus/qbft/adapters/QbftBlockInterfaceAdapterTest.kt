/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft.adapters

import com.github.michaelbull.result.Ok
import java.util.SequencedSet
import maru.consensus.ValidatorProvider
import maru.consensus.state.StateTransitionImpl
import maru.consensus.validation.StateRootValidator
import maru.core.BeaconBlock
import maru.core.Validator
import maru.core.ext.DataGenerators
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.kotlin.any
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class QbftBlockInterfaceAdapterTest {
  @Test
  fun `can replace round number in header`() {
    val beaconBlock =
      BeaconBlock(
        beaconBlockHeader = DataGenerators.randomBeaconBlockHeader(1UL).copy(round = 10u),
        beaconBlockBody = DataGenerators.randomBeaconBlockBody(),
      )
    val qbftBlock = QbftBlockAdapter(beaconBlock)
    val stateTransition = createMockStateTransition()
    val adapter = QbftBlockInterfaceAdapter(stateTransition)
    val updatedBlock =
      adapter.replaceRoundInBlock(qbftBlock, 20)
    val updatedBeaconBlockHeader = updatedBlock.header.toBeaconBlockHeader()
    assertEquals(updatedBeaconBlockHeader.round, 20u)
  }

  @Test
  fun `updates state root when replacing round number`() {
    val validators = DataGenerators.randomValidators().toSortedSet()
    val beaconBlock =
      BeaconBlock(
        beaconBlockHeader = DataGenerators.randomBeaconBlockHeader(1UL).copy(round = 10u),
        beaconBlockBody = DataGenerators.randomBeaconBlockBody(),
      )
    val qbftBlock = QbftBlockAdapter(beaconBlock)
    val stateTransition = createMockStateTransition(validators)
    val adapter = QbftBlockInterfaceAdapter(stateTransition)
    val updatedBlock = adapter.replaceRoundInBlock(qbftBlock, 20)

    val updatedBeaconBlock = updatedBlock.toBeaconBlock()
    val validationResult = StateRootValidator(stateTransition).validateBlock(updatedBeaconBlock).get()
    assertEquals(Ok(Unit), validationResult)
    assertEquals(20u, updatedBeaconBlock.beaconBlockHeader.round)
  }

  private fun createMockStateTransition(
    validators: SequencedSet<Validator> = DataGenerators.randomValidators().toSortedSet(),
  ): StateTransitionImpl {
    val validatorProvider = mock<ValidatorProvider>()
    whenever(validatorProvider.getValidatorsForBlock(any()))
      .thenReturn(SafeFuture.completedFuture(validators))
    return StateTransitionImpl(validatorProvider = validatorProvider)
  }
}

/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft

import kotlin.random.Random
import maru.consensus.ValidatorProvider
import maru.consensus.state.FinalizationState
import maru.core.ext.DataGenerators
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.consensus.common.bft.blockcreation.ProposerSelector
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.whenever

class QbftBlockCreatorFactoryTest {
  private val executionLayerManager = Mockito.mock(ExecutionLayerManager::class.java)
  private val proposerSelector = Mockito.mock(ProposerSelector::class.java)
  private val validatorProvider = Mockito.mock(ValidatorProvider::class.java)
  private val beaconChain = Mockito.mock(BeaconChain::class.java)
  private val finalizationState = Mockito.mock(FinalizationState::class.java)
  private val feeRecipient = Random.nextBytes(20)
  private val eagerQbftBlockCreatorConfig = Mockito.mock(EagerQbftBlockCreator.Config::class.java)
  private val prevRandaoProvider = { a: ULong, b: ByteArray -> Bytes32.random().toArray() }

  @Test
  fun `uses eager block creator for first block`() {
    whenever(beaconChain.getLatestBeaconState()).thenReturn(
      DataGenerators.randomBeaconState(0u),
    )

    val qbftBlockCreatorFactory =
      QbftBlockCreatorFactory(
        manager = executionLayerManager,
        proposerSelector = proposerSelector,
        validatorProvider = validatorProvider,
        beaconChain = beaconChain,
        finalizationStateProvider = { (_) -> finalizationState },
        prevRandaoProvider = prevRandaoProvider,
        feeRecipient = feeRecipient,
        eagerQbftBlockCreatorConfig = eagerQbftBlockCreatorConfig,
      )

    val blockCreator = qbftBlockCreatorFactory.create(0)
    assertThat(blockCreator).isInstanceOf(EagerQbftBlockCreator::class.java)
  }

  @Test
  fun `uses eager block creator for round greater than zero`() {
    whenever(beaconChain.getLatestBeaconState()).thenReturn(
      DataGenerators.randomBeaconState(0u),
    )

    val qbftBlockCreatorFactory =
      QbftBlockCreatorFactory(
        manager = executionLayerManager,
        proposerSelector = proposerSelector,
        validatorProvider = validatorProvider,
        beaconChain = beaconChain,
        finalizationStateProvider = { (_) -> finalizationState },
        prevRandaoProvider = prevRandaoProvider,
        feeRecipient = feeRecipient,
        eagerQbftBlockCreatorConfig = eagerQbftBlockCreatorConfig,
      )

    val blockCreator = qbftBlockCreatorFactory.create(1)
    assertThat(blockCreator).isInstanceOf(EagerQbftBlockCreator::class.java)
  }

  @Test
  fun `uses delayed block creator for round 0 after first block`() {
    whenever(beaconChain.getLatestBeaconState()).thenReturn(
      DataGenerators.randomBeaconState(0u),
    )

    val qbftBlockCreatorFactory =
      QbftBlockCreatorFactory(
        manager = executionLayerManager,
        proposerSelector = proposerSelector,
        validatorProvider = validatorProvider,
        beaconChain = beaconChain,
        finalizationStateProvider = { (_) -> finalizationState },
        prevRandaoProvider = prevRandaoProvider,
        feeRecipient = feeRecipient,
        eagerQbftBlockCreatorConfig = eagerQbftBlockCreatorConfig,
      )

    // trigger first use of block creator
    qbftBlockCreatorFactory.create(0)

    val blockCreator = qbftBlockCreatorFactory.create(0)
    assertThat(blockCreator).isInstanceOf(DelayedQbftBlockCreator::class.java)
  }
}

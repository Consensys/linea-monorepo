/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.consensus

import java.time.Clock
import kotlin.random.Random
import maru.consensus.dummy.DummyConsensusState
import maru.consensus.dummy.DummyEngineApiBlockCreator
import maru.consensus.state.FinalizationState
import maru.core.ExecutionPayload
import maru.core.ext.DataGenerators
import maru.core.ext.DataGenerators.randomValidForkChoiceUpdatedResult
import maru.executionlayer.manager.ExecutionLayerManager
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito.atLeastOnce
import org.mockito.Mockito.inOrder
import org.mockito.Mockito.mock
import org.mockito.Mockito.reset
import org.mockito.Mockito.verify
import org.mockito.Mockito.verifyNoMoreInteractions
import org.mockito.kotlin.any
import org.mockito.kotlin.eq
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class EngineApiBlockCreatorTest {
  private lateinit var executionLayerManager: ExecutionLayerManager

  @BeforeEach
  fun setUp() {
    executionLayerManager = mock()
  }

  @AfterEach
  fun tearDown() {
    reset(executionLayerManager)
  }

  private fun createDummyConsensusState(finalizationState: FinalizationState): DummyConsensusState =
    DummyConsensusState(
      Clock.systemUTC(),
      finalizationState,
      Random.nextBytes(32),
    )

  private fun mockFinishBlockBuilding(result: ExecutionPayload) {
    whenever(executionLayerManager.finishBlockBuilding()).thenReturn(
      SafeFuture.completedFuture(result),
    )
  }

  private fun mockSetHeadAndStartBlockBuilding(result: ForkChoiceUpdatedResult) {
    whenever(executionLayerManager.setHeadAndStartBlockBuilding(any(), any(), any(), any(), any())).thenReturn(
      SafeFuture.completedFuture(result),
    )
  }

  @Test
  fun `initialization triggers setHeadAndStartBlockBuilding with latest known state`() {
    val finalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))
    val dummyConsensusState = createDummyConsensusState(finalizationState)
    val feeRecipient = Random.nextBytes(20)
    mockSetHeadAndStartBlockBuilding(randomValidForkChoiceUpdatedResult())
    val nextTimestamp = dummyConsensusState.clock.millis()
    DummyEngineApiBlockCreator(
      manager = executionLayerManager,
      state = dummyConsensusState,
      nextBlockTimestamp = nextTimestamp,
      feeRecipientProvider = { feeRecipient },
    )

    verify(executionLayerManager, atLeastOnce()).setHeadAndStartBlockBuilding(
      eq(dummyConsensusState.latestBlockHash),
      eq(finalizationState.safeBlockHash),
      eq(finalizationState.finalizedBlockHash),
      eq(nextTimestamp),
      eq(feeRecipient),
    )
  }

  @Test
  fun `block creator doesn't call setHeadAndStartBlockBuilding or finishBlockBuilding twice in a row`() {
    val finalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))
    val dummyConsensusState = createDummyConsensusState(finalizationState)
    val expectedBlockBuildingResult = DataGenerators.randomExecutionPayload()
    mockFinishBlockBuilding(expectedBlockBuildingResult)
    val forkChoiceUpdatedResult = randomValidForkChoiceUpdatedResult()
    mockSetHeadAndStartBlockBuilding(forkChoiceUpdatedResult)
    val nextTimestamp = dummyConsensusState.clock.millis()

    val blockCreator =
      DummyEngineApiBlockCreator(
        manager = executionLayerManager,
        state = dummyConsensusState,
        nextBlockTimestamp = nextTimestamp,
        feeRecipientProvider = { Random.nextBytes(20) },
      )
    blockCreator.createBlock(1L)
    blockCreator.createBlock(2L)
    blockCreator.createBlock(3L)

    val inOrder = inOrder(executionLayerManager)
    repeat((1..3).count()) {
      inOrder.verify(executionLayerManager).setHeadAndStartBlockBuilding(
        any(),
        any(),
        any(),
        any(),
        any(),
      )
      inOrder.verify(executionLayerManager).finishBlockBuilding()
    }
    inOrder.verify(executionLayerManager).setHeadAndStartBlockBuilding(
      any(),
      any(),
      any(),
      any(),
      any(),
    )
    verifyNoMoreInteractions(executionLayerManager)
  }

  @Test
  fun `finalization updates are respected`() {
    val finalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))
    val dummyConsensusState = createDummyConsensusState(finalizationState)
    val expectedBlockBuildingResult = DataGenerators.randomExecutionPayload()
    mockFinishBlockBuilding(expectedBlockBuildingResult)
    val forkChoiceUpdatedResult = randomValidForkChoiceUpdatedResult()
    mockSetHeadAndStartBlockBuilding(forkChoiceUpdatedResult)
    val nextBlockTimestamp = dummyConsensusState.clock.millis()
    val feeRecipient = Random.nextBytes(20)

    val blockCreator =
      DummyEngineApiBlockCreator(
        manager = executionLayerManager,
        state = dummyConsensusState,
        nextBlockTimestamp = nextBlockTimestamp,
        feeRecipientProvider = { feeRecipient },
      )
    val nextTimestamp1 = 123L
    blockCreator.createBlock(nextTimestamp1)
    verify(executionLayerManager, atLeastOnce()).setHeadAndStartBlockBuilding(
      eq(expectedBlockBuildingResult.blockHash),
      eq(finalizationState.safeBlockHash),
      eq(finalizationState.finalizedBlockHash),
      eq(nextTimestamp1),
      eq(feeRecipient),
    )
    val newFinalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))
    dummyConsensusState.updateFinalizationState(newFinalizationState)

    val otherTimestamp2 = 124L
    blockCreator.createBlock(otherTimestamp2)

    verify(executionLayerManager, atLeastOnce()).setHeadAndStartBlockBuilding(
      eq(expectedBlockBuildingResult.blockHash),
      eq(newFinalizationState.safeBlockHash),
      eq(newFinalizationState.finalizedBlockHash),
      eq(otherTimestamp2),
      eq(feeRecipient),
    )
  }
}

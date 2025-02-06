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
import maru.consensus.dummy.FinalizationState
import maru.core.ext.DataGenerators
import maru.executionlayer.manager.BlockBuildingResult
import maru.executionlayer.manager.DataGenerators.randomValidForkChoiceUpdatedResult
import maru.executionlayer.manager.ExecutionLayerManager
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
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
      mock(),
      Clock.systemUTC(),
      finalizationState,
      Random.nextBytes(32),
    )

  private fun mockFinishBlockBuilding(result: BlockBuildingResult) {
    whenever(executionLayerManager.finishBlockBuilding()).thenReturn(
      SafeFuture.completedFuture(result),
    )
  }

  private fun mockSetHeadAndStartBlockBuilding(result: ForkChoiceUpdatedResult) {
    whenever(executionLayerManager.setHeadAndStartBlockBuilding(any(), any(), any())).thenReturn(
      SafeFuture.completedFuture(result),
    )
  }

  @Test
  fun `initialization triggers setHeadAndStartBlockBuilding with latest known state`() {
    val finalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))
    val dummyConsensusState = createDummyConsensusState(finalizationState)
    EngineApiBlockCreator(executionLayerManager, dummyConsensusState, MainnetBlockHeaderFunctions())

    verify(executionLayerManager, atLeastOnce()).setHeadAndStartBlockBuilding(
      eq(dummyConsensusState.latestBlockHash),
      eq(finalizationState.safeBlockHash),
      eq(finalizationState.finalizedBlockHash),
    )
  }

  @Test
  fun `block creator doesn't call setHeadAndStartBlockBuilding or finishBlockBuilding twice in a row`() {
    val finalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))
    val dummyConsensusState = createDummyConsensusState(finalizationState)
    val expectedBlockBuildingResult =
      BlockBuildingResult(
        DataGenerators.randomExecutionPayload(),
        Random.nextBytes(32),
      )
    mockFinishBlockBuilding(expectedBlockBuildingResult)
    val forkChoiceUpdatedResult = randomValidForkChoiceUpdatedResult()
    mockSetHeadAndStartBlockBuilding(forkChoiceUpdatedResult)

    val blockCreator = EngineApiBlockCreator(executionLayerManager, dummyConsensusState, MainnetBlockHeaderFunctions())
    blockCreator.createEmptyWithdrawalsBlock(1L, null)
    blockCreator.createEmptyWithdrawalsBlock(2L, null)
    blockCreator.createEmptyWithdrawalsBlock(3L, null)

    val inOrder = inOrder(executionLayerManager)
    repeat((1..3).count()) {
      inOrder.verify(executionLayerManager).setHeadAndStartBlockBuilding(
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
    )
    verifyNoMoreInteractions(executionLayerManager)
  }

  @Test
  fun `finalization updates are respected`() {
    val finalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))
    val dummyConsensusState = createDummyConsensusState(finalizationState)
    val expectedBlockBuildingResult =
      BlockBuildingResult(
        DataGenerators.randomExecutionPayload(),
        Random.nextBytes(32),
      )
    mockFinishBlockBuilding(expectedBlockBuildingResult)
    val forkChoiceUpdatedResult = randomValidForkChoiceUpdatedResult()
    mockSetHeadAndStartBlockBuilding(forkChoiceUpdatedResult)

    val blockCreator = EngineApiBlockCreator(executionLayerManager, dummyConsensusState, MainnetBlockHeaderFunctions())
    blockCreator.createEmptyWithdrawalsBlock(1L, null)
    verify(executionLayerManager, atLeastOnce()).setHeadAndStartBlockBuilding(
      eq(expectedBlockBuildingResult.resultingBlockHash),
      eq(finalizationState.safeBlockHash),
      eq(finalizationState.finalizedBlockHash),
    )
    val newFinalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))
    dummyConsensusState.updateFinalizationState(newFinalizationState)

    blockCreator.createEmptyWithdrawalsBlock(2L, null)

    verify(executionLayerManager, atLeastOnce()).setHeadAndStartBlockBuilding(
      eq(expectedBlockBuildingResult.resultingBlockHash),
      eq(newFinalizationState.safeBlockHash),
      eq(newFinalizationState.finalizedBlockHash),
    )
  }
}

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
package maru.consensus.qbft

import kotlin.random.Random
import kotlin.test.assertEquals
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.blockimport.BlockBuildingBeaconBlockImporter
import maru.consensus.state.FinalizationState
import maru.core.BeaconState
import maru.core.Validator
import maru.core.ext.DataGenerators
import maru.executionlayer.manager.ExecutionLayerManager
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.Mockito.times
import org.mockito.Mockito.verify
import org.mockito.kotlin.eq
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class FollowerBeaconBlockImporterTest {
  private lateinit var executionLayerManager: ExecutionLayerManager
  private var nextBlockTimestamp: Long = 123456789L
  private var shouldBuildNextBlock: Boolean = false
  private var blockBuilderIdentity: Validator = Validator(Random.nextBytes(20))
  private lateinit var beaconBlockImporter: BlockBuildingBeaconBlockImporter
  private lateinit var finalizationState: FinalizationState

  @BeforeEach
  fun setUp() {
    executionLayerManager = mock(ExecutionLayerManager::class.java)
    finalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))

    beaconBlockImporter =
      BlockBuildingBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = { finalizationState },
        nextBlockTimestampProvider = { nextBlockTimestamp },
        shouldBuildNextBlock = { _: BeaconState, _: ConsensusRoundIdentifier ->
          shouldBuildNextBlock
        },
        blockBuilderIdentity = blockBuilderIdentity,
      )
  }

  @Test
  fun `importBlock should call setHeadAndStartBlockBuilding when shouldBuildNextBlock returns true`() {
    shouldBuildNextBlock = true
    val randomBeaconBlock = DataGenerators.randomBeaconBlock(1UL)
    val randomBeaconState = DataGenerators.randomBeaconState(1UL)

    val expectedResponse = SafeFuture.completedFuture(DataGenerators.randomValidForkChoiceUpdatedResult())
    whenever(
      executionLayerManager.setHeadAndStartBlockBuilding(
        headHash = eq(randomBeaconBlock.beaconBlockBody.executionPayload.blockHash),
        safeHash = eq(finalizationState.safeBlockHash),
        finalizedHash = eq(finalizationState.finalizedBlockHash),
        nextBlockTimestamp = eq(nextBlockTimestamp),
        feeRecipient = eq(blockBuilderIdentity.address),
      ),
    ).thenReturn(expectedResponse)

    val result = beaconBlockImporter.importBlock(randomBeaconState, randomBeaconBlock)
    assertEquals(expectedResponse, result)
    verify(executionLayerManager).setHeadAndStartBlockBuilding(
      headHash = eq(randomBeaconBlock.beaconBlockBody.executionPayload.blockHash),
      safeHash = eq(finalizationState.safeBlockHash),
      finalizedHash = eq(finalizationState.finalizedBlockHash),
      nextBlockTimestamp = eq(nextBlockTimestamp),
      feeRecipient = eq(blockBuilderIdentity.address),
    )
  }

  @Test
  fun `importBlock should call setHead when shouldBuildNextBlock returns false`() {
    val randomBeaconBlock = DataGenerators.randomBeaconBlock(1UL)
    val randomBeaconState = DataGenerators.randomBeaconState(1UL)

    val expectedResponse = SafeFuture.completedFuture(DataGenerators.randomValidForkChoiceUpdatedResult())
    whenever(
      executionLayerManager.setHead(
        headHash = eq(randomBeaconBlock.beaconBlockBody.executionPayload.blockHash),
        safeHash = eq(finalizationState.safeBlockHash),
        finalizedHash = eq(finalizationState.finalizedBlockHash),
      ),
    ).thenReturn(expectedResponse)

    val result = beaconBlockImporter.importBlock(randomBeaconState, randomBeaconBlock)
    assertEquals(expectedResponse, result)
    verify(executionLayerManager).setHead(
      headHash = eq(randomBeaconBlock.beaconBlockBody.executionPayload.blockHash),
      safeHash = eq(finalizationState.safeBlockHash),
      finalizedHash = eq(finalizationState.finalizedBlockHash),
    )
  }

  @Test
  fun `importBlock should pass last block timestamp and next block's round identifier`() {
    val randomBeaconBlock = DataGenerators.randomBeaconBlock(1UL)
    val randomBeaconState = DataGenerators.randomBeaconState(1UL)
    val expectedConsensusRoundIdentifier = ConsensusRoundIdentifier(2, 0)
    val shouldBuildNextBlockPredicate: (BeaconState, ConsensusRoundIdentifier) -> Boolean = mock()
    whenever(
      shouldBuildNextBlockPredicate.invoke(eq(randomBeaconState), eq(expectedConsensusRoundIdentifier)),
    ).thenReturn(true)
    val nextBlockTimestampProvider: NextBlockTimestampProvider = mock()
    val expectedParentTimestamp = randomBeaconState.latestBeaconBlockHeader.timestamp.toLong()
    whenever(
      nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(eq(expectedParentTimestamp)),
    ).thenReturn(nextBlockTimestamp)
    beaconBlockImporter =
      BlockBuildingBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = { finalizationState },
        nextBlockTimestampProvider = nextBlockTimestampProvider,
        shouldBuildNextBlock = shouldBuildNextBlockPredicate,
        blockBuilderIdentity = blockBuilderIdentity,
      )

    beaconBlockImporter.importBlock(randomBeaconState, randomBeaconBlock)

    verify(shouldBuildNextBlockPredicate, times(1)).invoke(eq(randomBeaconState), eq(expectedConsensusRoundIdentifier))
    verify(nextBlockTimestampProvider, times(1)).nextTargetBlockUnixTimestamp(eq(expectedParentTimestamp))
  }
}

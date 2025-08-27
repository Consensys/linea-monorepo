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
import kotlin.test.assertEquals
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.blockimport.BlockBuildingBeaconBlockImporter
import maru.consensus.state.FinalizationState
import maru.core.BeaconState
import maru.core.ext.DataGenerators
import maru.executionlayer.manager.ExecutionLayerManager
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.Mockito.times
import org.mockito.Mockito.verify
import org.mockito.kotlin.any
import org.mockito.kotlin.eq
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class FollowerBeaconBlockImporterTest {
  private lateinit var executionLayerManager: ExecutionLayerManager
  private var nextBlockTimestamp: Long = 123456789L
  private var shouldBuildNextBlock: Boolean = false
  private lateinit var beaconBlockImporter: BlockBuildingBeaconBlockImporter
  private lateinit var finalizationState: FinalizationState
  private val prevRandaoProvider = { a: ULong, b: ByteArray -> Bytes32.random().toArray() }
  private val feeRecipient = Random.nextBytes(20)

  @BeforeEach
  fun setUp() {
    executionLayerManager = mock(ExecutionLayerManager::class.java)
    finalizationState = FinalizationState(Random.nextBytes(32), Random.nextBytes(32))

    beaconBlockImporter =
      BlockBuildingBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = { finalizationState },
        nextBlockTimestampProvider = { nextBlockTimestamp },
        prevRandaoProvider = prevRandaoProvider,
        shouldBuildNextBlock = { _: BeaconState, _: ConsensusRoundIdentifier, _: Long ->
          shouldBuildNextBlock
        },
        feeRecipient = feeRecipient,
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
        feeRecipient = eq(feeRecipient),
        prevRandao = any(),
      ),
    ).thenReturn(expectedResponse)

    val result = beaconBlockImporter.importBlock(randomBeaconState, randomBeaconBlock)
    assertEquals(expectedResponse, result)
    verify(executionLayerManager).setHeadAndStartBlockBuilding(
      headHash = eq(randomBeaconBlock.beaconBlockBody.executionPayload.blockHash),
      safeHash = eq(finalizationState.safeBlockHash),
      finalizedHash = eq(finalizationState.finalizedBlockHash),
      nextBlockTimestamp = eq(nextBlockTimestamp),
      feeRecipient = eq(feeRecipient),
      prevRandao = any(),
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
    val shouldBuildNextBlockPredicate: (BeaconState, ConsensusRoundIdentifier, Long) -> Boolean = mock()
    whenever(
      shouldBuildNextBlockPredicate.invoke(
        eq(randomBeaconState),
        eq(expectedConsensusRoundIdentifier),
        eq(nextBlockTimestamp),
      ),
    ).thenReturn(true)
    val nextBlockTimestampProvider: NextBlockTimestampProvider = mock()
    val expectedParentTimestamp = randomBeaconState.beaconBlockHeader.timestamp.toLong()
    whenever(
      nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(eq(expectedParentTimestamp)),
    ).thenReturn(nextBlockTimestamp)
    beaconBlockImporter =
      BlockBuildingBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = { finalizationState },
        nextBlockTimestampProvider = nextBlockTimestampProvider,
        prevRandaoProvider = prevRandaoProvider,
        shouldBuildNextBlock = shouldBuildNextBlockPredicate,
        feeRecipient = feeRecipient,
      )

    beaconBlockImporter.importBlock(randomBeaconState, randomBeaconBlock)

    verify(shouldBuildNextBlockPredicate, times(1)).invoke(
      eq(randomBeaconState),
      eq
        (expectedConsensusRoundIdentifier),
      eq(nextBlockTimestamp),
    )
    verify(nextBlockTimestampProvider, times(1)).nextTargetBlockUnixTimestamp(eq(expectedParentTimestamp))
  }
}

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
package maru.executionlayer.manager

import kotlin.random.Random
import maru.executionlayer.client.ExecutionLayerClient
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito.atLeastOnce
import org.mockito.Mockito.mock
import org.mockito.Mockito.reset
import org.mockito.Mockito.verify
import org.mockito.kotlin.any
import org.mockito.kotlin.argThat
import org.mockito.kotlin.eq
import org.mockito.kotlin.whenever
import tech.pegasys.teku.ethereum.executionclient.schema.BlobsBundleV1
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV3
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.GetPayloadV3Response
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV3
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadStatusV1
import tech.pegasys.teku.ethereum.executionclient.schema.Response
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.bytes.Bytes8
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.spec.TestSpecFactory
import tech.pegasys.teku.spec.executionlayer.ExecutionPayloadStatus
import tech.pegasys.teku.spec.util.DataStructureUtil
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult as TekuForkChoiceUpdatedResult

class JsonRpcExecutionLayerManagerTest {
  private val dataStructureUtil = DataStructureUtil(TestSpecFactory.createMinimalDeneb())
  private lateinit var executionLayerClient: ExecutionLayerClient
  private lateinit var executionLayerManager: ExecutionLayerManager

  private val feeRecipient: ByteArray = Random.nextBytes(20)
  private val initialBlockHeight = 1UL

  @Volatile
  private var latestBlockHash = Bytes32.random().toArray()

  @BeforeEach
  fun setUp() {
    executionLayerClient = mock()
    executionLayerManager = createExecutionLayerManager()
  }

  @AfterEach
  fun tearDown() {
    reset(executionLayerClient)
  }

  private fun createExecutionLayerManager(): ExecutionLayerManager {
    whenever(executionLayerClient.getLatestBlockMetadata()).thenReturn(
      SafeFuture.completedFuture(BlockMetadata(initialBlockHeight, latestBlockHash, 0L)),
    )
    return JsonRpcExecutionLayerManager
      .create(
        executionLayerClient,
        { feeRecipient },
        NoopValidator,
      ).get()
  }

  private fun mockForkChoiceUpdateWithValidStatus(payloadId: Bytes8): PayloadStatusV1 {
    val latestValidHash = Bytes32.random()
    val executionStatus = ExecutionPayloadStatus.VALID
    val payloadStatus = PayloadStatusV1(executionStatus, latestValidHash, null)
    whenever(executionLayerClient.forkChoiceUpdate(any(), any()))
      .thenReturn(
        SafeFuture.completedFuture(Response(TekuForkChoiceUpdatedResult(payloadStatus, payloadId))),
      )
    return payloadStatus
  }

  private fun mockGetPayloadWithRandomData(
    payloadId: Bytes8,
    executionPayload: ExecutionPayloadV3,
  ) {
    val blobsBundle = BlobsBundleV1.fromInternalBlobsBundle(dataStructureUtil.randomBlobsBundle())
    val getPayloadResponse =
      Response(
        GetPayloadV3Response(
          /* executionPayload = */ executionPayload,
          /* blockValue = */ UInt256.ZERO,
          /* blobsBundle = */ blobsBundle,
          /* shouldOverrideBuilder = */ false,
        ),
      )
    whenever(executionLayerClient.getPayload(eq(payloadId)))
      .thenReturn(SafeFuture.completedFuture(getPayloadResponse))
  }

  private fun mockNewPayloadWithStatus(payloadStatus: PayloadStatusV1) {
    whenever(executionLayerClient.newPayload(any())).thenReturn(
      SafeFuture.completedFuture(Response(payloadStatus)),
    )
  }

  @Test
  fun `setHeadAndStartBlockBuilding stores payloadId for finishBlockBuilding`() {
    val newHeadHash = Bytes32.random()
    val newSafeHash = Bytes32.random()
    val newFinalizedHash = Bytes32.random()
    val nextTimestamp = 0L

    val payloadId = Bytes8(Bytes.random(8))
    val payloadStatus = mockForkChoiceUpdateWithValidStatus(payloadId)

    val result =
      executionLayerManager
        .setHeadAndStartBlockBuilding(
          headHash = newHeadHash.toArray(),
          safeHash = newSafeHash.toArray(),
          finalizedHash = newFinalizedHash.toArray(),
          nextBlockTimestamp = nextTimestamp,
        ).get()

    val expectedPayloadStatus =
      PayloadStatus(
        executionPayloadStatus = "VALID",
        latestValidHash =
          payloadStatus
            .asInternalExecutionPayload()
            .latestValidHash
            .get()
            .toArray(),
        validationError = null,
        failureCause = null,
      )
    val expectedResult = ForkChoiceUpdatedResult(expectedPayloadStatus, payloadId.wrappedBytes.toArray())
    assertThat(result).isEqualTo(expectedResult)

    val executionPayload = ExecutionPayloadV3.fromInternalExecutionPayload(dataStructureUtil.randomExecutionPayload())
    mockGetPayloadWithRandomData(payloadId, executionPayload)
    mockNewPayloadWithStatus(payloadStatus)

    executionLayerManager.finishBlockBuilding().get()
    verify(executionLayerClient, atLeastOnce()).getPayload(eq(payloadId))
  }

  @Test
  fun `setHeadAndStartBlockBuilding passes arguments to FCU correctly`() {
    val newHeadHash = Bytes32.random()
    val newSafeHash = Bytes32.random()
    val newFinalizedHash = Bytes32.random()
    val nextTimestamp = Random.nextLong(0, Long.MAX_VALUE)

    val payloadId = Bytes8(Bytes.random(8))
    val payloadStatus = mockForkChoiceUpdateWithValidStatus(payloadId)

    val result =
      executionLayerManager
        .setHeadAndStartBlockBuilding(
          headHash = newHeadHash.toArray(),
          safeHash = newSafeHash.toArray(),
          finalizedHash = newFinalizedHash.toArray(),
          nextBlockTimestamp = nextTimestamp,
        ).get()

    val expectedPayloadStatus =
      PayloadStatus(
        executionPayloadStatus = "VALID",
        latestValidHash =
          payloadStatus
            .asInternalExecutionPayload()
            .latestValidHash
            .get()
            .toArray(),
        validationError = null,
        failureCause = null,
      )
    val expectedResult = ForkChoiceUpdatedResult(expectedPayloadStatus, payloadId.wrappedBytes.toArray())
    assertThat(result).isEqualTo(expectedResult)
    verify(executionLayerClient, atLeastOnce()).forkChoiceUpdate(
      argThat { forkChoiceState ->
        forkChoiceState == ForkChoiceStateV1(newHeadHash, newSafeHash, newFinalizedHash)
      },
      argThat { payloadAttributes ->
        payloadAttributes ==
          PayloadAttributesV3(
            UInt64.fromLongBits(nextTimestamp),
            Bytes32.ZERO,
            Bytes20(Bytes.wrap(feeRecipient)),
            emptyList(),
            Bytes32.ZERO,
          )
      },
    )
  }

  // TODO: Add a test for validator

  @Test
  fun `finishBlockBuilding can't be called before setHeadAndStartBlockBuilding`() {
    val result = executionLayerManager.finishBlockBuilding()
    assertThat(result.isCompletedExceptionally).isTrue()
  }

  @Test
  fun `latestBlockHeight is initialized`() {
    val result = executionLayerManager.latestBlockMetadata()
    assertThat(result).isEqualTo(BlockMetadata(initialBlockHeight, latestBlockHash, 0L))
  }

  @Test
  fun `latestBlockHeight is updated by setHeadAndStartBlockBuilding`() {
    val newHeadHash = Bytes32.random()
    val newSafeHash = Bytes32.random()
    val newFinalizedHash = Bytes32.random()
    val nextTimestamp = 0L

    val payloadId = Bytes8(Bytes.random(8))
    val payloadStatus = mockForkChoiceUpdateWithValidStatus(payloadId)

    executionLayerManager
      .setHeadAndStartBlockBuilding(
        headHash = newHeadHash.toArray(),
        safeHash = newSafeHash.toArray(),
        finalizedHash = newFinalizedHash.toArray(),
        nextBlockTimestamp = nextTimestamp,
      ).get()

    val executionPayload = ExecutionPayloadV3.fromInternalExecutionPayload(dataStructureUtil.randomExecutionPayload())
    mockGetPayloadWithRandomData(payloadId, executionPayload)
    mockNewPayloadWithStatus(payloadStatus)

    executionLayerManager.finishBlockBuilding().get()

    executionLayerManager
      .setHeadAndStartBlockBuilding(
        headHash = Bytes32.random().toArray(),
        safeHash = Bytes32.random().toArray(),
        finalizedHash = Bytes32.random().toArray(),
        nextBlockTimestamp = nextTimestamp,
      ).get()

    val expectedBlockMetadata =
      BlockMetadata(
        blockNumber = executionPayload.blockNumber.longValue().toULong(),
        blockHash = executionPayload.blockHash.toArray(),
        unixTimestamp = executionPayload.timestamp.longValue(),
      )
    assertThat(executionLayerManager.latestBlockMetadata()).isEqualTo(expectedBlockMetadata)
  }
}

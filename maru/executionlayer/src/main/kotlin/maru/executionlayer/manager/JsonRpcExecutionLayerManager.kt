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

import kotlin.jvm.optionals.getOrNull
import maru.core.ExecutionPayload
import maru.executionlayer.client.BlockNumberAndHash
import maru.executionlayer.client.ExecutionLayerClient
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV3
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV3
import tech.pegasys.teku.ethereum.executionclient.schema.Response
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.bytes.Bytes8
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult as TekuForkChoiceUpdatedResult

class JsonRpcExecutionLayerManager private constructor(
  private val executionLayerClient: ExecutionLayerClient,
  private val newBlockTimestampProvider: () -> ULong,
  private val feeRecipientProvider: () -> ByteArray,
  private val currentBlockNumberAndHash: BlockNumberAndHash,
) : ExecutionLayerManager {
  companion object {
    fun create(
      executionLayerClient: ExecutionLayerClient,
      newBlockTimestampProvider: () -> ULong,
      feeRecipientProvider: () -> ByteArray,
    ): SafeFuture<JsonRpcExecutionLayerManager> =
      executionLayerClient.getLatestBlockMetadata().thenApply {
        val currentBlockNumberAndHash = BlockNumberAndHash(it.blockNumber, it.blockHash)
        JsonRpcExecutionLayerManager(
          executionLayerClient = executionLayerClient,
          newBlockTimestampProvider = newBlockTimestampProvider,
          feeRecipientProvider = feeRecipientProvider,
          currentBlockNumberAndHash = currentBlockNumberAndHash,
        )
      }
  }

  private class ElHeightMetadata(
    var nextBlockNumberAndHash: BlockNumberAndHash,
    var currentBlockNumberAndHash: BlockNumberAndHash,
  ) {
    @Synchronized
    fun updateNext(blockNumberAndHash: BlockNumberAndHash) {
      nextBlockNumberAndHash = blockNumberAndHash
    }

    @Synchronized
    fun promoteBlockNumberAndHash() {
      currentBlockNumberAndHash = nextBlockNumberAndHash
    }
  }

  private var payloadId: ByteArray? = null
  private var latestBlockCache: ElHeightMetadata =
    ElHeightMetadata(
      nextBlockNumberAndHash = currentBlockNumberAndHash,
      currentBlockNumberAndHash = currentBlockNumberAndHash,
    )

  override fun setHeadAndStartBlockBuilding(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
  ): SafeFuture<ForkChoiceUpdatedResult> {
    val payloadAttributes =
      PayloadAttributesV3(
        UInt64.fromLongBits(newBlockTimestampProvider().toLong()),
        Bytes32.ZERO,
        Bytes20(Bytes.wrap(feeRecipientProvider())),
        emptyList(),
        Bytes32.ZERO,
      )
    return executionLayerClient
      .forkChoiceUpdate(
        ForkChoiceStateV1(
          Bytes32.wrap(headHash),
          Bytes32.wrap(safeHash),
          Bytes32.wrap(finalizedHash),
        ),
        payloadAttributes,
      ).thenApply(::mapForkChoiceUpdatedResultToDomain)
      .thenPeek {
        latestBlockCache.promoteBlockNumberAndHash()
        payloadId = it.payloadId
      }
  }

  private fun mapForkChoiceUpdatedResultToDomain(
    forkChoiceUpdatedResult: Response<TekuForkChoiceUpdatedResult>,
  ): ForkChoiceUpdatedResult {
    val payload = forkChoiceUpdatedResult.payload.asInternalExecutionPayload()
    val parsedPayloadId =
      payload.payloadId
        .getOrNull()
        ?.wrappedBytes
        ?.toArray()
    val payloadStatusV1 = payload.payloadStatus
    val domainPayloadStatus =
      PayloadStatus(
        payloadStatusV1.status.getOrNull()?.name,
        payloadStatusV1.latestValidHash.getOrNull()?.toArray(),
        payloadStatusV1.validationError.getOrNull(),
        payloadStatusV1.failureCause.getOrNull(),
      )
    return ForkChoiceUpdatedResult(domainPayloadStatus, parsedPayloadId)
  }

  override fun finishBlockBuilding(): SafeFuture<BlockBuildingResult> {
    if (payloadId == null) {
      return SafeFuture.failedFuture(
        IllegalStateException(
          "finishBlockBuilding is called before setHeadAndStartBlockBuilding was completed",
        ),
      )
    }
    return executionLayerClient
      .getPayload(Bytes8(Bytes.wrap(payloadId!!)))
      .thenCompose { payloadResponse ->
        if (payloadResponse.isSuccess) {
          val executionPayload = payloadResponse.payload.executionPayload
          latestBlockCache.updateNext(
            BlockNumberAndHash(
              executionPayload.blockNumber
                .longValue()
                .toULong(),
              executionPayload.blockHash,
            ),
          )
          executionLayerClient.newPayload(executionPayload).thenApply { payloadStatus ->
            if (payloadStatus.isSuccess) {
              payloadId = null // Not necessary, but it helps to reinforce the order of calls
              BlockBuildingResult(
                executionPayloadV3ToDomain(executionPayload),
                payloadStatus.payload
                  .asInternalExecutionPayload()
                  .latestValidHash
                  .get()
                  .toArray(),
              )
            } else {
              throw IllegalStateException("engine_newPayload request failed! Cause: " + payloadStatus.errorMessage)
            }
          }
        } else {
          SafeFuture.failedFuture(
            IllegalStateException("engine_getPayload request failed! Cause: " + payloadResponse.errorMessage),
          )
        }
      }
  }

  private fun executionPayloadV3ToDomain(executionPayloadV3: ExecutionPayloadV3): ExecutionPayload =
    ExecutionPayload(
      parentHash = executionPayloadV3.parentHash.toArray(),
      feeRecipient = executionPayloadV3.feeRecipient.wrappedBytes.toArray(),
      stateRoot = executionPayloadV3.stateRoot.toArray(),
      receiptsRoot = executionPayloadV3.receiptsRoot.toArray(),
      logsBloom = executionPayloadV3.logsBloom.toArray(),
      prevRandao = executionPayloadV3.prevRandao.toArray(),
      blockNumber = executionPayloadV3.blockNumber.longValue().toULong(),
      gasLimit = executionPayloadV3.gasLimit.longValue().toULong(),
      gasUsed = executionPayloadV3.gasUsed.longValue().toULong(),
      timestamp = executionPayloadV3.timestamp.longValue().toULong(),
      extraData = executionPayloadV3.extraData.toArray(),
      // Intentional cropping, UInt256 doesn't fit into ULong
      baseFeePerGas =
        executionPayloadV3.baseFeePerGas.toBigInteger(),
      blockHash = executionPayloadV3.blockHash.toArray(),
      transactions = executionPayloadV3.transactions.map { it.toArray() },
    )

  override fun latestBlockHeight(): ULong = latestBlockCache.currentBlockNumberAndHash.blockNumber

  override fun setHead(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
  ): SafeFuture<ForkChoiceUpdatedResult> {
    TODO("Will implement once the chain following feature is implemented")
  }

  override fun importBlock(executionPayload: ExecutionPayload): ByteArray {
    TODO("Will implement once the chain following feature is implemented")
  }
}
